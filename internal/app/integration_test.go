package app

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"os"
	"os/exec"
	"testing"
	"time"
	"wholesale.local/orbit/internal/domain"
	"wholesale.local/orbit/internal/store"
)

func testService(t *testing.T) *Service {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL required for real PostgreSQL tests")
	}
	s, e := store.Open(context.Background(), url)
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(s.Pool.Close)
	if e = s.Migrate(context.Background()); e != nil {
		t.Fatal(e)
	}
	if e = s.Seed(context.Background()); e != nil {
		t.Fatal(e)
	}
	return &Service{DB: s}
}
func run(t *testing.T, s *Service, actor, typ string, p Payload) Command {
	t.Helper()
	ctx := context.Background()
	c := Command{ID: uuid.NewString(), ActorID: store.ID(actor), Type: typ, Payload: p}
	if _, e := s.Submit(ctx, c); e != nil {
		t.Fatal(e)
	}
	if e := s.Process(ctx, c.ID); e != nil {
		t.Fatal(e)
	}
	if e := s.DB.Pool.QueryRow(ctx, "SELECT status,error FROM commands WHERE id=$1", c.ID).Scan(&c.Status, &c.Error); e != nil {
		t.Fatal(e)
	}
	return c
}
func success(t *testing.T, c Command) {
	t.Helper()
	if c.Status != "succeeded" {
		t.Fatalf("%s: %s %s", c.Type, c.Status, c.Error)
	}
}
func load(t *testing.T, s *Service, kind, id string, dst any) {
	t.Helper()
	var b []byte
	if e := s.DB.Pool.QueryRow(context.Background(), "SELECT body FROM aggregates WHERE kind=$1 AND id=$2", kind, id).Scan(&b); e != nil {
		t.Fatal(e)
	}
	if e := json.Unmarshal(b, dst); e != nil {
		t.Fatal(e)
	}
}
func TestLifecycleAndIdempotency(t *testing.T) {
	s := testService(t)
	ctx := context.Background()
	sku := store.ID("P-1001")
	var initial domain.Stock
	load(t, s, "stock", sku, &initial)
	success(t, run(t, s, "customer-a", "cart.set", Payload{ProductID: sku, Quantity: 2}))
	addr := domain.Address{Company: "テスト商事", Recipient: "検証担当", Postal: "100-0001", Text: "東京都千代田区1", Phone: "03-0000-0000"}
	c := run(t, s, "customer-a", "order.checkout", Payload{Address: addr, Coupon: "WELCOME10"})
	success(t, c)
	id := c.ID
	if _, e := s.Submit(ctx, c); e != nil {
		t.Fatal(e)
	}
	if e := s.Process(ctx, id); e != nil {
		t.Fatal(e)
	}
	bad := c
	bad.Payload.Coupon = "PARTNER5"
	if _, e := s.Submit(ctx, bad); !errors.Is(e, ErrKeyReuse) {
		t.Fatal("idempotency mismatch not rejected", e)
	}
	denied := run(t, s, "customer-b", "order.pay", Payload{OrderID: id, Success: true})
	if denied.Status != "rejected" {
		t.Fatal("ownership not enforced")
	}
	for _, x := range []struct {
		a, k string
		p    Payload
	}{{"customer-a", "order.pay", Payload{OrderID: id, Success: true}}, {"warehouse-a", "order.pick", Payload{OrderID: id}}, {"warehouse-a", "order.block", Payload{OrderID: id, Note: "破損・代替品確認"}}, {"warehouse-a", "order.pick", Payload{OrderID: id}}, {"warehouse-a", "order.pack", Payload{OrderID: id, BoxCount: 2, WeightGrams: 1500}}, {"carrier", "order.dispatch", Payload{OrderID: id, Tracking: "TEST-001"}}, {"carrier", "order.delivery_failed", Payload{OrderID: id, Note: "不在"}}, {"carrier", "order.redeliver", Payload{OrderID: id}}, {"carrier", "order.deliver", Payload{OrderID: id}}} {
		success(t, run(t, s, x.a, x.k, x.p))
	}
	var o domain.Order
	load(t, s, "order", id, &o)
	if o.Status() != domain.Delivered || o.Snapshot().Total != 4752 || o.Snapshot().Points != 43 {
		t.Fatalf("unexpected order %+v", o)
	}
	var loyalty domain.Loyalty
	load(t, s, "loyalty", store.ID("customer-a"), &loyalty)
	before := loyalty.Balance().Value()
	duplicate := run(t, s, "carrier", "order.deliver", Payload{OrderID: id})
	if duplicate.Status != "rejected" {
		t.Fatal("second delivery allowed")
	}
	load(t, s, "loyalty", store.ID("customer-a"), &loyalty)
	if loyalty.Balance().Value() != before {
		t.Fatal("double points")
	}
	var st domain.Stock
	load(t, s, "stock", sku, &st)
	if st.Snapshot().OnHand != initial.Snapshot().OnHand-2 || st.Snapshot().Reserved != initial.Snapshot().Reserved {
		t.Fatal("incorrect shipment stock", st, initial)
	}
	var pdf []byte
	if e := s.DB.Pool.QueryRow(ctx, "SELECT pdf FROM invoices WHERE order_id=$1", id).Scan(&pdf); e != nil {
		t.Fatal(e)
	}
	if len(pdf) < 1000 || string(pdf[:4]) != "%PDF" {
		t.Fatal("invalid PDF")
	}
	if e := s.DB.Project(ctx); e != nil {
		t.Fatal(e)
	}
}
func TestExpiryAndRollback(t *testing.T) {
	s := testService(t)
	ctx := context.Background()
	cid := store.ID("customer-b")
	sku := store.ID("P-2001")
	success(t, run(t, s, "customer-b", "cart.set", Payload{ProductID: sku, Quantity: 3}))
	var c domain.Cart
	load(t, s, "cart", cid, &c)
	old := c.Snapshot().Expires
	// Old scheduled message must not release a renewed cart.
	success(t, run(t, s, "system", "cart.expire", Payload{CustomerID: cid, Expires: old - 1}))
	load(t, s, "cart", cid, &c)
	if c.Snapshot().Items[sku] != 3 {
		t.Fatal("stale expiry released cart")
	}
	var before domain.Stock
	load(t, s, "stock", sku, &before)
	rejected := run(t, s, "customer-b", "cart.set", Payload{ProductID: sku, Quantity: 10000})
	if rejected.Status != "rejected" {
		t.Fatal("shortage not rejected")
	}
	var after domain.Stock
	load(t, s, "stock", sku, &after)
	if before != after {
		t.Fatal("failed reservation changed stock")
	}
	past := time.Now().Unix() - 1
	_, e := s.DB.Pool.Exec(ctx, `UPDATE aggregates SET body=jsonb_set(body,'{expires}',to_jsonb($2::bigint)),version=version+1 WHERE kind='cart' AND id=$1`, cid, past)
	if e != nil {
		t.Fatal(e)
	}
	success(t, run(t, s, "system", "cart.expire", Payload{CustomerID: cid, Expires: past}))
	success(t, run(t, s, "system", "cart.expire", Payload{CustomerID: cid, Expires: past}))
	load(t, s, "stock", sku, &after)
	if after.Snapshot().Reserved != before.Snapshot().Reserved-3 {
		t.Fatal("incorrect release", before, after)
	}
}
func TestOptimisticLock(t *testing.T) {
	s := testService(t)
	ctx := context.Background()
	id := store.ID("P-3001")
	tx1, e := s.DB.Pool.Begin(ctx)
	if e != nil {
		t.Fatal(e)
	}
	defer tx1.Rollback(ctx)
	tx2, e := s.DB.Pool.Begin(ctx)
	if e != nil {
		t.Fatal(e)
	}
	defer tx2.Rollback(ctx)
	u1, u2 := store.NewUnit(ctx, tx1), store.NewUnit(ctx, tx2)
	var a, b domain.Stock
	if e = u1.Load("stock", id, &a); e != nil {
		t.Fatal(e)
	}
	if e = u2.Load("stock", id, &b); e != nil {
		t.Fatal(e)
	}
	if e = a.Adjust(adjustment(1)); e != nil {
		t.Fatal(e)
	}
	if e = u1.Save("stock", id, a); e != nil {
		t.Fatal(e)
	}
	if e = tx1.Commit(ctx); e != nil {
		t.Fatal(e)
	}
	if e = b.Adjust(adjustment(1)); e != nil {
		t.Fatal(e)
	}
	if e = u2.Save("stock", id, b); !errors.Is(e, domain.ErrConflict) {
		t.Fatal("lost update not detected", e)
	}
}

func TestCancelPaymentFailureAndAtomicMultiSKU(t *testing.T) {
	s := testService(t)
	ctx := context.Background()
	cid := store.ID("customer-b")
	first, second := store.ID("P-1002"), store.ID("P-3002")
	// Ensure this customer's cart is empty before building the scenario.
	var cart domain.Cart
	var e error
	success(t, run(t, s, "customer-b", "cart.set", Payload{ProductID: first, Quantity: 2}))
	success(t, run(t, s, "customer-b", "cart.set", Payload{ProductID: second, Quantity: 3}))
	load(t, s, "cart", cid, &cart)
	addr := domain.Address{Company: "試験", Recipient: "担当", Postal: "1000001", Text: "東京都", Phone: "0300000000"}
	c := run(t, s, "customer-b", "order.checkout", Payload{Address: addr})
	success(t, c)
	success(t, run(t, s, "customer-b", "order.pay", Payload{OrderID: c.ID, Success: false}))
	for _, id := range []string{first, second} {
		var stock domain.Stock
		load(t, s, "stock", id, &stock)
		if stock.Snapshot().Reserved != 0 {
			t.Fatal("payment failure did not release", stock)
		}
	}
	success(t, run(t, s, "customer-b", "cart.set", Payload{ProductID: first, Quantity: 2}))
	c = run(t, s, "customer-b", "order.checkout", Payload{Address: addr})
	success(t, c)
	success(t, run(t, s, "customer-b", "order.pay", Payload{OrderID: c.ID, Success: true}))
	success(t, run(t, s, "warehouse-b", "order.pick", Payload{OrderID: c.ID}))
	denied := run(t, s, "customer-b", "order.cancel", Payload{OrderID: c.ID})
	if denied.Status != "rejected" {
		t.Fatal("customer cancelled warehouse work")
	}
	success(t, run(t, s, "operator", "order.cancel", Payload{OrderID: c.ID, Note: "倉庫欠品のため取消"}))
	var stock domain.Stock
	load(t, s, "stock", first, &stock)
	if stock.Snapshot().Reserved != 0 {
		t.Fatal("cancel did not release")
	}
	// Save two aggregate updates but force failure of the second. The first must roll back.
	tx, e := s.DB.Pool.Begin(ctx)
	if e != nil {
		t.Fatal(e)
	}
	u := store.NewUnit(ctx, tx)
	var a, b domain.Stock
	if e = u.Load("stock", first, &a); e != nil {
		t.Fatal(e)
	}
	if e = u.Load("stock", second, &b); e != nil {
		t.Fatal(e)
	}
	before := a
	if e = a.Reserve(reservation(1)); e != nil {
		t.Fatal(e)
	}
	if e = u.Save("stock", first, a); e != nil {
		t.Fatal(e)
	}
	if e = b.Reserve(reservation(b.Snapshot().OnHand + 1)); e == nil {
		t.Fatal("invalid reservation accepted")
	}
	if e = tx.Rollback(ctx); e != nil {
		t.Fatal(e)
	}
	load(t, s, "stock", first, &a)
	if a != before {
		t.Fatal("partial update committed")
	}
}

func TestConcurrentDuplicateCommandProcesses(t *testing.T) {
	// Child processes execute the same persisted command without application goroutines.
	if id := os.Getenv("ORBIT_CHILD_COMMAND"); id != "" {
		s := testService(t)
		if e := s.Process(context.Background(), id); e != nil {
			t.Fatal(e)
		}
		return
	}
	s := testService(t)
	ctx := context.Background()
	id := store.ID("P-2002")
	var before domain.Stock
	load(t, s, "stock", id, &before)
	c := Command{ID: uuid.NewString(), ActorID: store.ID("operator"), Type: "stock.adjust", Payload: Payload{ProductID: id, Delta: 1, Note: "concurrent duplicate test"}}
	if _, e := s.Submit(ctx, c); e != nil {
		t.Fatal(e)
	}
	var children []*exec.Cmd
	for i := 0; i < 4; i++ {
		child := exec.Command(os.Args[0], "-test.run=^TestConcurrentDuplicateCommandProcesses$", "-test.v")
		child.Env = append(os.Environ(), "ORBIT_CHILD_COMMAND="+c.ID)
		child.Stdout = os.Stdout
		child.Stderr = os.Stderr
		if e := child.Start(); e != nil {
			t.Fatal(e)
		}
		children = append(children, child)
	}
	for _, child := range children {
		if e := child.Wait(); e != nil {
			t.Fatal(e)
		}
	}
	var after domain.Stock
	load(t, s, "stock", id, &after)
	if after.Snapshot().OnHand != before.Snapshot().OnHand+1 {
		t.Fatalf("duplicate update: before=%+v after=%+v", before, after)
	}
}

func adjustment(v int64) domain.StockAdjustment {
	x, e := domain.NewStockAdjustment(v)
	if e != nil {
		panic(e)
	}
	return x
}
func reservation(v int64) domain.ReservationDelta {
	x, e := domain.NewReservationDelta(v)
	if e != nil {
		panic(e)
	}
	return x
}
