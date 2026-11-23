package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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
func (s *Service) Process(ctx context.Context, id string) error {
	for attempt := 0; attempt < 5; attempt++ {
		e := s.processOnce(ctx, id)
		if errors.Is(e, domain.ErrConflict) {
			continue
		}
		if e != nil {
			_, _ = s.DB.Pool.Exec(ctx, `UPDATE commands SET attempts=attempts+1,error=$2,status=CASE WHEN attempts>=7 THEN 'dead' ELSE status END WHERE id=$1 AND status='queued'`, id, e.Error())
		}
		return e
	}
	return domain.ErrConflict
}
func (s *Service) processOnce(ctx context.Context, id string) error {
	tx, e := s.DB.Pool.Begin(ctx)
	if e != nil {
		return e
	}
	defer tx.Rollback(ctx)
	var c Command
	var b []byte
	e = tx.QueryRow(ctx, "SELECT id,actor_id,type,payload,status FROM commands WHERE id=$1", id).Scan(&c.ID, &c.ActorID, &c.Type, &b, &c.Status)
	if errors.Is(e, pgx.ErrNoRows) {
		return nil
	}
	if e != nil {
		return e
	}
	if c.Status != "queued" {
		return nil
	}
	if e = json.Unmarshal(b, &c.Payload); e != nil {
		return e
	}
	var actor domain.Actor
	var ab []byte
	if c.ActorID == store.ID("system") {
		actor = domain.Actor{ID: c.ActorID, Role: "system", Name: "system"}
	} else {
		if e = tx.QueryRow(ctx, store.ActorQuery+" WHERE a.id=$1", c.ActorID).Scan(&ab); e != nil {
			return e
		}
		if e = json.Unmarshal(ab, &actor); e != nil {
			return e
		}
	}
	var now time.Time
	if e = tx.QueryRow(ctx, "SELECT clock_timestamp()").Scan(&now); e != nil {
		return e
	}
	u := store.NewUnit(ctx, tx)
	result, e := s.apply(u, c, actor, now)
	if e != nil {
		var r *Rejection
		if !errors.As(e, &r) {
			return e
		}
		if e = tx.Rollback(ctx); e != nil {
			return e
		}
		_, e = s.DB.Pool.Exec(ctx, "UPDATE commands SET status='rejected',error=$2,finished_at=now() WHERE id=$1 AND status='queued'", id, r.Message)
		return e
	}
	rb, _ := json.Marshal(result)
	tag, e := tx.Exec(ctx, "UPDATE commands SET status='succeeded',result=$2,error='',finished_at=now() WHERE id=$1 AND status='queued'", id, rb)
	if e != nil {
		return e
	}
	if tag.RowsAffected() != 1 {
		return domain.ErrConflict
	}
	return tx.Commit(ctx)
}
func role(a domain.Actor, roles ...string) error {
	for _, r := range roles {
		if a.Role == r {
			return nil
		}
	}
	return reject("この利用者には操作権限がありません")
}
func validID(s string) bool { _, e := uuid.Parse(s); return e == nil }
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
func changeStock(u *store.Unit, items map[string]int64, mode string) error {
	// A deterministic ordering avoids lock-order inversions in multi-SKU transactions.
	keys := sortedKeys(items)
	for _, id := range keys {
		st, e := loadStock(u, id)
		if e != nil {
			return e
		}
		q := items[id]
		if mode == "ship" {
			qty, err := domain.NewOrderQuantity(q)
			if err != nil {
				return checked(err)
			}
			e = st.Ship(qty)
		} else {
			delta, err := domain.NewReservationDelta(q)
			if err != nil {
				return checked(err)
			}
			e = st.Reserve(delta)
		}
		if e != nil {
			return checked(e)
		}
		if e = u.Save("stock", id, st); e != nil {
			return e
		}
	}
	return nil
}
func (s *Service) apply(u *store.Unit, c Command, a domain.Actor, now time.Time) (any, error) {
	p := c.Payload
	if len(p.Note) > 1000 || len(p.Tracking) > 100 || len(p.RequestedDate) > 10 {
		return nil, reject("入力が長すぎます")
	}
	if c.Type == "cart.set" || c.Type == "cart.expire" || c.Type == "order.checkout" {
		customer := a.ID
		if c.Type == "cart.expire" {
			if e := role(a, "system"); e != nil {
				return nil, e
			}
			customer = p.CustomerID
		} else {
			if e := role(a, "customer"); e != nil {
				return nil, e
			}
		}
		if !validID(customer) {
			return nil, reject("顧客IDが不正です")
		}

		customerID, e := domain.ParseID[domain.ActorTag](customer)
		if e != nil {
			return nil, checked(e)
		}
		cart := domain.NewCart(customerID)
		if e := u.Load("cart", customer, &cart); e != nil {
			return nil, e
		}
		if p.ExpectedVersion > 0 && p.ExpectedVersion != u.Versions["cart"+customer] {
			return nil, reject(domain.ErrConflict.Error())
		}
		if c.Type == "cart.expire" {
			delta, changed := cart.Expire(p.Expires, now.Unix())
			if !changed {
				return map[string]any{"noop": true}, nil
			}
			if e := changeStock(u, delta, "reserve"); e != nil {
				return nil, e
			}
			return map[string]any{"expired": true}, u.Save("cart", customer, cart)
		}
		if c.Type == "cart.set" {
			pid, e := domain.ParseID[domain.ProductTag](p.ProductID)
			if e != nil {
				return nil, checked(e)
			}
			quantity, e := domain.NewCartQuantity(p.Quantity)
			if e != nil {
				return nil, checked(e)
			}
			delta, changed, e := cart.Change(pid, quantity, now.Unix())
			if e != nil {
				return nil, checked(e)
			}
			if !changed {
				return map[string]any{"noop": true}, nil
			}
			if e := changeStock(u, delta, "reserve"); e != nil {
				return nil, e
			}
			return map[string]any{"customerId": customer}, u.Save("cart", customer, cart)
		}
		items, e := cart.Checkout(now.Unix())
		if e != nil {
			return nil, checked(e)
		}
		address, e := domain.NewShippingAddress(p.Address)
		if e != nil {
			return nil, checked(e)
		}
		if p.RequestedDate != "" {
			d, e := time.Parse("2006-01-02", p.RequestedDate)
			if e != nil || d.Format("2006-01-02") < now.In(time.FixedZone("JST", 9*3600)).Format("2006-01-02") {
				return nil, reject("希望納品日が不正です")
			}
		}
		var percent int64
		if p.Coupon != "" {
			if e := u.Tx.QueryRow(u.Ctx, "SELECT percent FROM coupons WHERE code=$1", p.Coupon).Scan(&percent); errors.Is(e, pgx.ErrNoRows) {
				return nil, reject("クーポンが存在しません")
			} else if e != nil {
				return nil, e
			}
		}
		rate, e := domain.NewDiscountRate(percent)
		if e != nil {
			return nil, checked(e)
		}
		lines := []domain.Line{}
		for _, id := range sortedKeys(items) {
			var pb []byte
			var pr domain.Product
			if e := u.Tx.QueryRow(u.Ctx, store.ProductQuery+" WHERE id=$1", id).Scan(&pb); e != nil {
				return nil, e
			}
			if e := json.Unmarshal(pb, &pr); e != nil {
				return nil, e
			}
			pid, e := domain.ParseID[domain.ProductTag](id)
			if e != nil {
				return nil, e
			}
			qty, e := domain.NewOrderQuantity(items[id])
			if e != nil {
				return nil, checked(e)
			}
			price, e := domain.NewUnitPrice(pr.Price)
			if e != nil {
				return nil, checked(e)
			}
			line, e := domain.NewLine(pid, pr.SKU, pr.Name, pr.Unit, qty, price)
			if e != nil {
				return nil, checked(e)
			}
			lines = append(lines, line)
		}
		oid, e := domain.ParseID[domain.OrderTag](c.ID)
		if e != nil {
			return nil, checked(e)
		}
		o, e := domain.NewOrder(oid, customerID, lines, address, rate, p.Coupon, p.RequestedDate, p.Note, now)
		if e != nil {
			return nil, checked(e)
		}
		var existing domain.Order
		if e := u.Load("order", c.ID, &existing); e != nil {
			return nil, e
		}
		if e := u.Save("order", c.ID, o); e != nil {
			return nil, e
		}
		return map[string]any{"orderId": c.ID}, u.Save("cart", customer, cart)
	}

	if c.Type == "stock.adjust" {
		if e := role(a, "operator"); e != nil {
			return nil, e
		}
		if !validID(p.ProductID) || p.Delta < -1000000000 || p.Delta > 1000000000 || p.Note == "" {
			return nil, reject("商品・調整数・理由を入力してください")
		}
		st, e := loadStock(u, p.ProductID)
		if e != nil {
			return nil, e
		}
		if p.ExpectedVersion > 0 && p.ExpectedVersion != u.Versions["stock"+p.ProductID] {
			return nil, reject(domain.ErrConflict.Error())
		}
		delta, e := domain.NewStockAdjustment(p.Delta)
		if e != nil {
			return nil, checked(e)
		}
		if e = checked(st.Adjust(delta)); e != nil {
			return nil, e
		}
		return map[string]any{"productId": p.ProductID}, u.Save("stock", p.ProductID, st)
	}
	if !validID(p.OrderID) {
		return nil, reject("受注IDが不正です")
	}
	var o domain.Order
	if e := u.Load("order", p.OrderID, &o); e != nil {
		return nil, e
	}
	if !o.Exists() {
		return nil, reject("受注が存在しません")
	}
	x := o.Snapshot()
	if p.ExpectedVersion > 0 && p.ExpectedVersion != u.Versions["order"+x.ID] {
		return nil, reject(domain.ErrConflict.Error())
	}
	if a.Role == "customer" && a.ID != x.CustomerID {
		return nil, reject("他の顧客の受注は操作できません")
	}
	actorID, e := domain.ParseID[domain.ActorTag](a.ID)
	if e != nil {
		return nil, checked(e)
	}
	var target domain.OrderStatus
	switch c.Type {
	case "order.pay":
		if e := role(a, "customer"); e != nil {
			return nil, e
		}
		target = domain.Accepted
		if !p.Success {
			target = domain.PaymentFailed
		}
	case "order.cancel":
		if e := role(a, "customer", "operator", "system"); e != nil {
			return nil, e
		}
		if a.Role == "system" && (o.Status() != domain.PaymentPending || now.Sub(x.CreatedAt) < 15*time.Minute) {
			return map[string]any{"noop": true}, nil
		}
		if a.Role == "customer" && o.Status() != domain.PaymentPending && o.Status() != domain.Accepted {
			return nil, reject("作業開始後のキャンセルは運用者に依頼してください")
		}
		target = domain.Cancelled
	case "order.pick":
		if e := role(a, "warehouse"); e != nil {
			return nil, e
		}
		if x.WorkerID != "" && x.WorkerID != a.ID {
			return nil, reject("別の担当者が作業中です")
		}
		if e := checked(o.AssignWorker(actorID)); e != nil {
			return nil, e
		}
		target = domain.Picking
	case "order.block":
		if e := role(a, "warehouse"); e != nil {
			return nil, e
		}
		if x.WorkerID != a.ID {
			return nil, reject("担当外の作業です")
		}
		if p.Note == "" {
			return nil, reject("保留理由が必要です")
		}
		target = domain.PickBlocked
	case "order.pack":
		if e := role(a, "warehouse"); e != nil {
			return nil, e
		}
		if x.WorkerID != a.ID {
			return nil, reject("担当外の作業です")
		}
		if p.BoxCount < 1 || p.BoxCount > 100 || p.WeightGrams < 1 || p.WeightGrams > 500000 {
			return nil, reject("箱数は1〜100、1箱の重量は1〜500000gです")
		}
		boxCount, e := domain.NewBoxCount(p.BoxCount)
		if e != nil {
			return nil, checked(e)
		}
		weight, e := domain.NewWeightGrams(p.WeightGrams)
		if e != nil {
			return nil, checked(e)
		}
		if e := checked(o.Pack(boxCount, weight)); e != nil {
			return nil, e
		}
		target = domain.Packed
	case "order.dispatch":
		if e := role(a, "carrier"); e != nil {
			return nil, e
		}
		if p.Tracking == "" {
			return nil, reject("追跡番号を入力してください")
		}
		if o.Status() != domain.Packed {
			return nil, reject("梱包完了後に出荷してください")
		}
		if e := checked(o.AssignCarrier(actorID, p.Tracking)); e != nil {
			return nil, e
		}
		target = domain.Shipping
	case "order.delivery_failed", "order.redeliver", "order.deliver", "order.return":
		if e := role(a, "carrier"); e != nil {
			return nil, e
		}
		if x.CarrierID != a.ID {
			return nil, reject("担当外の配送です")
		}
		switch c.Type {
		case "order.delivery_failed":
			if p.Note == "" {
				return nil, reject("配送失敗理由が必要です")
			}
			target = domain.DeliveryFailed
		case "order.redeliver":
			if o.Status() != domain.DeliveryFailed {
				return nil, reject("再配送できる状態ではありません")
			}
			target = domain.Shipping
		case "order.deliver":
			target = domain.Delivered
		case "order.return":
			if p.Note == "" {
				return nil, reject("返送理由が必要です")
			}
			target = domain.Returned
		}
	default:
		return nil, reject("未対応のCommandです")
	}
	if e := checked(o.Move(target, actorID, p.Note, now)); e != nil {
		return nil, e
	}
	items := map[string]int64{}
	for _, l := range x.Lines {
		items[l.ProductID] = l.Quantity
	}
	if target == domain.PaymentFailed || target == domain.Cancelled {
		for id, q := range items {
			items[id] = -q
		}
		if e := changeStock(u, items, "reserve"); e != nil {
			return nil, e
		}
	}
	if c.Type == "order.dispatch" {
		if e := changeStock(u, items, "ship"); e != nil {
			return nil, e
		}
	}
	if e := u.Save("order", x.ID, o); e != nil {
		return nil, e
	}
	if target == domain.Delivered {
		if _, e := u.Tx.Exec(u.Ctx, "INSERT INTO order_point_awards(order_id) VALUES($1)", x.ID); e != nil {
			return nil, e
		}
		var l domain.Loyalty
		if e := u.Load("loyalty", x.CustomerID, &l); e != nil {
			return nil, e
		}
		if e := l.Credit(o.EarnedPoints()); e != nil {
			return nil, e
		}
		if e := u.Save("loyalty", x.CustomerID, l); e != nil {
			return nil, e
		}
		pdf, e := InvoicePDF(o)
		if e != nil {
			return nil, e
		}
		if _, e = u.Tx.Exec(u.Ctx, "INSERT INTO invoices(id,order_id,pdf) VALUES($1,$2,$3) ON CONFLICT(order_id) DO NOTHING", store.ID("invoice/"+x.ID), x.ID, pdf); e != nil {
			return nil, e
		}
	}
	return map[string]any{"orderId": x.ID, "status": target.Code()}, nil
}
