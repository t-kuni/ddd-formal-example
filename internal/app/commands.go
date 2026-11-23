package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"time"
	"wholesale.local/orbit/internal/domain"
	"wholesale.local/orbit/internal/store"
)

type Service struct{ DB *store.Store }
type Payload struct {
	ProductID       string         `json:"productId,omitempty"`
	Quantity        int64          `json:"quantity,omitempty"`
	Delta           int64          `json:"delta,omitempty"`
	OrderID         string         `json:"orderId,omitempty"`
	CustomerID      string         `json:"customerId,omitempty"`
	ExpectedVersion int64          `json:"expectedVersion,omitempty"`
	Expires         int64          `json:"expires,omitempty"`
	Coupon          string         `json:"coupon,omitempty"`
	Address         domain.Address `json:"address,omitempty"`
	RequestedDate   string         `json:"requestedDate,omitempty"`
	Note            string         `json:"note,omitempty"`
	Success         bool           `json:"success,omitempty"`
	BoxCount        int            `json:"boxCount,omitempty"`
	WeightGrams     int64          `json:"weightGrams,omitempty"`
	Tracking        string         `json:"tracking,omitempty"`
	CarrierID       string         `json:"carrierId,omitempty"`
}
type Command struct {
	ID        string          `json:"id"`
	ActorID   string          `json:"actorId"`
	Type      string          `json:"type"`
	Payload   Payload         `json:"payload"`
	Status    string          `json:"status"`
	Result    json.RawMessage `json:"result"`
	Error     string          `json:"error"`
	Attempts  int             `json:"attempts"`
	CreatedAt time.Time       `json:"createdAt"`
}
type Rejection struct{ Message string }

func (e *Rejection) Error() string { return e.Message }
func reject(s string) error        { return &Rejection{s} }
func checked(e error) error {
	if e != nil {
		return reject(e.Error())
	}
	return nil
}

var ErrKeyReuse = errors.New("同じIdempotency-Keyに異なる内容は送れません")

func (s *Service) Submit(ctx context.Context, c Command) (string, error) {
	if _, e := uuid.Parse(c.ID); e != nil {
		return "", reject("Idempotency-KeyはUUIDにしてください")
	}
	if _, e := uuid.Parse(c.ActorID); e != nil {
		return "", reject("利用者IDが不正です")
	}
	b, _ := json.Marshal(c.Payload)
	h := sha256.Sum256(append([]byte(c.ActorID+"/"+c.Type+"/"), b...))
	finger := hex.EncodeToString(h[:])
	_, e := s.DB.Pool.Exec(ctx, "INSERT INTO commands(id,actor_id,type,payload,fingerprint) VALUES($1,$2,$3,$4,$5) ON CONFLICT DO NOTHING", c.ID, c.ActorID, c.Type, b, finger)
	if e != nil {
		return "", e
	}
	var old, status string
	e = s.DB.Pool.QueryRow(ctx, "SELECT fingerprint,status FROM commands WHERE id=$1", c.ID).Scan(&old, &status)
	if e != nil {
		return "", e
	}
	if old != finger {
		return "", ErrKeyReuse
	}
	return status, nil
}
func loadStock(u *store.Unit, id string) (*domain.Stock, error) {
	var st domain.Stock
	e := u.Load("stock", id, &st)
	if e != nil {
		return nil, e
	}
	if u.Versions["stock"+id] == 0 {
		return nil, reject("商品が存在しません")
	}
	return &st, nil
}
