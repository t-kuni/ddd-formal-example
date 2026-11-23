package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"wholesale.local/orbit/internal/app"
	"wholesale.local/orbit/internal/cloud"
	"wholesale.local/orbit/internal/store"
	"wholesale.local/orbit/internal/web"
)

type event struct {
	Records []events.SQSMessage             `json:"Records"`
	HTTP    *events.APIGatewayV2HTTPRequest `json:"http,omitempty"`
}
type runtime struct {
	s       *app.Service
	a       *cloud.AWS
	handler http.Handler
}

func connect(ctx context.Context) (*runtime, error) {
	db, e := store.Open(ctx, os.Getenv("DATABASE_URL"))
	if e != nil {
		return nil, e
	}
	a, e := cloud.New(ctx)
	if e != nil {
		return nil, e
	}
	s := &app.Service{DB: db}
	return &runtime{s, a, (&web.API{Service: s, AWS: a}).Handler()}, nil
}
func (r *runtime) handle(ctx context.Context, ev event) (any, error) {
	if ev.HTTP != nil {
		q := ev.HTTP
		var body []byte
		var e error
		if q.IsBase64Encoded {
			body, e = base64.StdEncoding.DecodeString(q.Body)
			if e != nil {
				return nil, e
			}
		} else {
			body = []byte(q.Body)
		}
		path := q.RawPath
		if q.RawQueryString != "" {
			path += "?" + q.RawQueryString
		}
		req := httptest.NewRequest(q.RequestContext.HTTP.Method, path, bytes.NewReader(body)).WithContext(ctx)
		for k, v := range q.Headers {
			req.Header.Set(k, v)
		}
		w := httptest.NewRecorder()
		r.handler.ServeHTTP(w, req)
		h := map[string]string{}
		for k, v := range w.Header() {
			h[k] = strings.Join(v, ",")
		}
		return events.APIGatewayV2HTTPResponse{StatusCode: w.Code, Headers: h, Body: base64.StdEncoding.EncodeToString(w.Body.Bytes()), IsBase64Encoded: true}, nil
	}
	failed := []events.SQSBatchItemFailure{}
	for _, record := range ev.Records {
		var e error
		if strings.Contains(record.EventSourceARN, "orbit-maintenance") {
			e = r.maintenance(ctx)
		} else {
			e = r.s.Process(ctx, record.Body)
		}
		if e != nil {
			log.Printf("record %s: %v", record.MessageId, e)
			failed = append(failed, events.SQSBatchItemFailure{ItemIdentifier: record.MessageId})
		}
	}
	return events.SQSEventResponse{BatchItemFailures: failed}, nil
}
func (r *runtime) maintenance(ctx context.Context) error {
	if e := r.s.Expire(ctx); e != nil {
		return e
	}
	if e := r.a.Dispatch(ctx, r.s.DB); e != nil {
		return e
	}
	return r.s.SendInvoice(ctx)
}
func main() {
	log.SetFlags(log.LstdFlags | log.LUTC)
	mode := "lambda"
	if len(os.Args) > 1 {
		mode = os.Args[1]
	}
	ctx := context.Background()
	if mode == "gateway" {
		log.Fatal(http.ListenAndServe(":8088", http.HandlerFunc(gateway)))
		return
	}
	r, e := connect(ctx)
	if e != nil {
		log.Fatal(e)
	}
	if mode == "init" {
		if e = r.s.DB.Migrate(ctx); e == nil {
			e = r.s.DB.Seed(ctx)
		}
		if e == nil {
			e = r.a.Init(ctx, r.s.DB)
		}
		if e != nil {
			log.Fatal(e)
		}
		log.Print("initialized PostgreSQL, S3, SQS, Scheduler")
		return
	}
	if e = r.a.Queues(ctx); e != nil {
		log.Fatal(e)
	}
	switch mode {
	case "lambda":
		lambda.Start(r.handle)
	case "http":
		log.Fatal(http.ListenAndServe(":8088", r.handler))
	case "invoke":
		log.Fatal(http.ListenAndServe(":8080", http.HandlerFunc(func(w http.ResponseWriter, q *http.Request) {
			var ev event
			if e := json.NewDecoder(q.Body).Decode(&ev); e != nil {
				http.Error(w, e.Error(), 400)
				return
			}
			out, e := r.handle(q.Context(), ev)
			if e != nil {
				http.Error(w, e.Error(), 500)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(out)
		})))
	case "bridge":
		if e = bridge(ctx, r.a); e != nil {
			log.Fatal(e)
		}
	case "tick":
		if e = r.a.Tick(ctx); e != nil {
			log.Fatal(e)
		}
	case "maintenance":
		if e = r.maintenance(ctx); e != nil {
			log.Fatal(e)
		}
	default:
		log.Fatal("unknown mode: " + mode)
	}
}

var invokeClient = &http.Client{Timeout: 55 * time.Second}

func invoke(ctx context.Context, ev event, out any) error {
	// RIE has one execution environment and crashes on overlapping invocations.
	// This file lock is local adapter plumbing, not a business/database lock.
	path := os.Getenv("INVOKE_LOCK_PATH")
	if path == "" {
		path = filepath.Join(os.TempDir(), "orbit-rie-invoke.lock")
	}
	lock, e := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
	if e != nil {
		return e
	}
	defer lock.Close()
	if e = syscall.Flock(int(lock.Fd()), syscall.LOCK_EX); e != nil {
		return e
	}
	defer syscall.Flock(int(lock.Fd()), syscall.LOCK_UN)

	b, _ := json.Marshal(ev)
	endpoint := os.Getenv("LAMBDA_INVOKE_URL")
	if endpoint == "" {
		endpoint = "http://127.0.0.1:8080/2015-03-31/functions/function/invocations"
	}
	q, e := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(b))
	if e != nil {
		return e
	}
	res, e := invokeClient.Do(q)
	if e != nil {
		return e
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf("Lambda status %d", res.StatusCode)
	}
	var raw json.RawMessage
	if e = json.NewDecoder(res.Body).Decode(&raw); e != nil {
		return e
	}
	var failure map[string]any
	_ = json.Unmarshal(raw, &failure)
	if failure["errorMessage"] != nil {
		return fmt.Errorf("Lambda: %v", failure["errorMessage"])
	}
	return json.Unmarshal(raw, out)
}
func gateway(w http.ResponseWriter, r *http.Request) {
	body, e := io.ReadAll(http.MaxBytesReader(w, r.Body, 32768))
	if e != nil {
		http.Error(w, e.Error(), 400)
		return
	}
	headers := map[string]string{}
	for k, v := range r.Header {
		headers[k] = strings.Join(v, ",")
	}
	req := events.APIGatewayV2HTTPRequest{Version: "2.0", RawPath: r.URL.Path, RawQueryString: r.URL.RawQuery, Headers: headers, Body: base64.StdEncoding.EncodeToString(body), IsBase64Encoded: true}
	req.RequestContext.HTTP.Method = r.Method
	var out events.APIGatewayV2HTTPResponse
	if e = invoke(r.Context(), event{HTTP: &req}, &out); e != nil {
		http.Error(w, "Lambda connection: "+e.Error(), 502)
		return
	}
	for k, v := range out.Headers {
		w.Header().Set(k, v)
	}
	data := []byte(out.Body)
	if out.IsBase64Encoded {
		data, e = base64.StdEncoding.DecodeString(out.Body)
		if e != nil {
			http.Error(w, e.Error(), 502)
			return
		}
	}
	w.WriteHeader(out.StatusCode)
	_, _ = w.Write(data)
}
func bridge(ctx context.Context, a *cloud.AWS) error {
	// Only the local AWS event-source mapping adapter. No business logic, no goroutines.
	_ = a.Tick(ctx)
	for {
		for _, q := range []struct{ url, arn string }{{a.Queue, "arn:aws:sqs:us-east-1:000000000000:orbit-commands"}, {a.TickQueue, "arn:aws:sqs:us-east-1:000000000000:orbit-maintenance"}} {
			out, e := a.SQS.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{QueueUrl: aws.String(q.url), MaxNumberOfMessages: 1, WaitTimeSeconds: 1, VisibilityTimeout: 60})
			if e != nil {
				log.Print(e)
				time.Sleep(time.Second)
				continue
			}
			for _, m := range out.Messages {
				ev := event{Records: []events.SQSMessage{{MessageId: aws.ToString(m.MessageId), Body: aws.ToString(m.Body), EventSource: "aws:sqs", EventSourceARN: q.arn, AWSRegion: "us-east-1"}}}
				var reply events.SQSEventResponse
				if e = invoke(ctx, ev, &reply); e != nil {
					log.Print(e)
					continue
				}
				if len(reply.BatchItemFailures) > 0 {
					continue
				}
				if _, e = a.SQS.DeleteMessage(ctx, &sqs.DeleteMessageInput{QueueUrl: aws.String(q.url), ReceiptHandle: m.ReceiptHandle}); e != nil {
					log.Print(e)
				}
			}
		}
		if errors.Is(ctx.Err(), context.Canceled) {
			return ctx.Err()
		}
	}
}
