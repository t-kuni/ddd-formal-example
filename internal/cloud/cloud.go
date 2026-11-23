package cloud

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/scheduler"
	st "github.com/aws/aws-sdk-go-v2/service/scheduler/types"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/smithy-go"
	"io"
	"net/url"
	"os"
	"wholesale.local/orbit/internal/store"
)

const Bucket = "orbit-products"

type AWS struct {
	IAM       *iam.Client
	S3        *s3.Client
	SQS       *sqs.Client
	Scheduler *scheduler.Client
	Queue     string
	TickQueue string
	Endpoint  string
}

func New(ctx context.Context) (*AWS, error) {
	endpoint := os.Getenv("AWS_ENDPOINT_URL")
	if endpoint == "" {
		return nil, errors.New("AWS_ENDPOINT_URL is mandatory: real AWS is disabled")
	}
	u, e := url.Parse(endpoint)
	if e != nil || u.Scheme != "https" {
		return nil, errors.New("local emulator requires HTTPS endpoint")
	}
	cfg, e := config.LoadDefaultConfig(ctx, config.WithRegion("us-east-1"))
	if e != nil {
		return nil, e
	}
	a := &AWS{Endpoint: endpoint}
	a.IAM = iam.NewFromConfig(cfg, func(o *iam.Options) { o.BaseEndpoint = aws.String(endpoint) })
	a.S3 = s3.NewFromConfig(cfg, func(o *s3.Options) { o.BaseEndpoint = aws.String(endpoint); o.UsePathStyle = true })
	a.SQS = sqs.NewFromConfig(cfg, func(o *sqs.Options) { o.BaseEndpoint = aws.String(endpoint) })
	a.Scheduler = scheduler.NewFromConfig(cfg, func(o *scheduler.Options) { o.BaseEndpoint = aws.String(endpoint) })
	return a, nil
}
func (a *AWS) Init(ctx context.Context, db *store.Store) error {
	_, e := a.S3.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(Bucket)})
	if e != nil {
		if _, e = a.S3.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(Bucket)}); e != nil {
			return e
		}
	}
	for _, name := range []string{"orbit-commands-dlq", "orbit-commands", "orbit-maintenance"} {
		attrs := map[string]string{"VisibilityTimeout": "60", "MessageRetentionPeriod": "1209600"}
		if name == "orbit-commands" {
			attrs["RedrivePolicy"] = `{"deadLetterTargetArn":"arn:aws:sqs:us-east-1:000000000000:orbit-commands-dlq","maxReceiveCount":"8"}`
		}
		q, e := a.SQS.CreateQueue(ctx, &sqs.CreateQueueInput{QueueName: aws.String(name), Attributes: attrs})
		if e != nil {
			return e
		}
		if name == "orbit-commands" {
			a.Queue = a.localQueue(*q.QueueUrl)
		}
		if name == "orbit-maintenance" {
			a.TickQueue = a.localQueue(*q.QueueUrl)
		}
	}
	rows, e := db.Pool.Query(ctx, "SELECT id,image FROM products")
	if e != nil {
		return e
	}
	type pic struct {
		id string
		b  []byte
	}
	var pics []pic
	for rows.Next() {
		var p pic
		if e = rows.Scan(&p.id, &p.b); e != nil {
			rows.Close()
			return e
		}
		pics = append(pics, p)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return e
	}
	for _, p := range pics {
		if _, e = a.S3.PutObject(ctx, &s3.PutObjectInput{Bucket: aws.String(Bucket), Key: aws.String(p.id + ".svg"), Body: bytes.NewReader(p.b), ContentType: aws.String("image/svg+xml")}); e != nil {
			return e
		}
	}
	_, e = a.IAM.CreateRole(ctx, &iam.CreateRoleInput{RoleName: aws.String("orbit-scheduler"), AssumeRolePolicyDocument: aws.String(`{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"Service":"scheduler.amazonaws.com"},"Action":"sts:AssumeRole"}]}`)})
	if e != nil {
		var api smithy.APIError
		if !errors.As(e, &api) || api.ErrorCode() != "EntityAlreadyExists" {
			return e
		}
	}
	_, e = a.IAM.PutRolePolicy(ctx, &iam.PutRolePolicyInput{RoleName: aws.String("orbit-scheduler"), PolicyName: aws.String("send-maintenance"), PolicyDocument: aws.String(`{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":"sqs:SendMessage","Resource":"arn:aws:sqs:us-east-1:000000000000:orbit-maintenance"}]}`)})
	if e != nil {
		return e
	}
	_, e = a.Scheduler.CreateSchedule(ctx, &scheduler.CreateScheduleInput{Name: aws.String("orbit-maintenance"), ScheduleExpression: aws.String("rate(1 minute)"), FlexibleTimeWindow: &st.FlexibleTimeWindow{Mode: st.FlexibleTimeWindowModeOff}, Target: &st.Target{Arn: aws.String("arn:aws:sqs:us-east-1:000000000000:orbit-maintenance"), RoleArn: aws.String("arn:aws:iam::000000000000:role/orbit-scheduler"), Input: aws.String(`{"type":"maintenance"}`)}})
	if e != nil {
		var api smithy.APIError
		if !errors.As(e, &api) || api.ErrorCode() != "ConflictException" {
			return fmt.Errorf("scheduler: %w", e)
		}
	}
	return nil
}
func (a *AWS) localQueue(raw string) string {
	u, e := url.Parse(raw)
	if e != nil {
		return raw
	}
	base, _ := url.Parse(a.Endpoint)
	u.Host = base.Host
	u.Scheme = base.Scheme
	return u.String()
}
func (a *AWS) Queues(ctx context.Context) error {
	q, e := a.SQS.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{QueueName: aws.String("orbit-commands")})
	if e != nil {
		return e
	}
	a.Queue = a.localQueue(*q.QueueUrl)
	q, e = a.SQS.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{QueueName: aws.String("orbit-maintenance")})
	if e != nil {
		return e
	}
	a.TickQueue = a.localQueue(*q.QueueUrl)
	return nil
}
func (a *AWS) Send(ctx context.Context, id string) error {
	_, e := a.SQS.SendMessage(ctx, &sqs.SendMessageInput{QueueUrl: aws.String(a.Queue), MessageBody: aws.String(id)})
	return e
}
func (a *AWS) Image(ctx context.Context, id string) ([]byte, error) {
	o, e := a.S3.GetObject(ctx, &s3.GetObjectInput{Bucket: aws.String(Bucket), Key: aws.String(id + ".svg")})
	if e != nil {
		return nil, e
	}
	defer o.Body.Close()
	return io.ReadAll(o.Body)
}
func (a *AWS) Dispatch(ctx context.Context, db *store.Store) error {
	rows, e := db.Pool.Query(ctx, "SELECT id FROM commands WHERE status='queued' AND (last_dispatched IS NULL OR last_dispatched<now()-interval '30 seconds') ORDER BY created_at LIMIT 100")
	if e != nil {
		return e
	}
	var ids []string
	for rows.Next() {
		var id string
		if e = rows.Scan(&id); e != nil {
			rows.Close()
			return e
		}
		ids = append(ids, id)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return e
	}
	for _, id := range ids {
		if e = a.Send(ctx, id); e != nil {
			return e
		}
		if _, e = db.Pool.Exec(ctx, "UPDATE commands SET last_dispatched=now() WHERE id=$1", id); e != nil {
			return e
		}
	}
	return nil
}
func (a *AWS) Tick(ctx context.Context) error {
	b, _ := json.Marshal(map[string]string{"type": "maintenance"})
	_, e := a.SQS.SendMessage(ctx, &sqs.SendMessageInput{QueueUrl: aws.String(a.TickQueue), MessageBody: aws.String(string(b))})
	return e
}
