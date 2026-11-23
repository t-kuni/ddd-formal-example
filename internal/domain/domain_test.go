package domain

import (
	"encoding/json"
	"math/rand"
	"testing"
	"time"

	"github.com/google/uuid"

	core "wholesale.local/orbit/generated/Domain"
	d "wholesale.local/orbit/generated/dafny"
)

func TestInventoryInvariantSequences(t *testing.T) {
	for seed := int64(0); seed < 100; seed++ {
		r := rand.New(rand.NewSource(seed))
		h, _ := NewQuantity(100)
		s, _ := NewStock(h, Quantity{})
		for i := 0; i < 1000; i++ {
			before := s
			q := int64(r.Intn(301) - 150)
			var e error
			switch r.Intn(3) {
			case 0:
				delta, _ := NewReservationDelta(q)
				e = s.Reserve(delta)
			case 1:
				delta, _ := NewStockAdjustment(q)
				e = s.Adjust(delta)
			case 2:
				qty, err := NewOrderQuantity(q)
				e = err
				if e == nil {
					e = s.Ship(qty)
				}
			}
			if !s.Valid() {
				t.Fatalf("seed %d step %d invalid %+v", seed, i, s)
			}
			if e != nil && s != before {
				t.Fatal("rejected mutation changed state")
			}
		}
	}
}
func TestGeneratedStateMachine(t *testing.T) {
	allowed := map[int][]int{0: {1, 8, 9}, 1: {2, 8}, 2: {3, 4, 8}, 3: {2, 8}, 4: {5}, 5: {6, 7}, 6: {5, 10}}
	for from := 0; from <= 10; from++ {
		for to := -1; to <= 11; to++ {
			want := false
			for _, x := range allowed[from] {
				if x == to {
					want = true
				}
			}
			ok, next := core.Companion_Default___.Transition(d.IntOf(from), d.IntOf(to))
			if ok != want {
				t.Fatalf("%d -> %d", from, to)
			}
			if !want && next.Int64() != int64(from) {
				t.Fatal("rejection mutated state")
			}
		}
	}
}
func fixture(t *testing.T) Order {
	t.Helper()
	id, _ := ParseID[OrderTag](uuid.NewString())
	actor, _ := ParseID[ActorTag](uuid.NewString())
	pid, _ := ParseID[ProductTag](uuid.NewString())
	qty, _ := NewOrderQuantity(3)
	price, _ := NewUnitPrice(999)
	line, e := NewLine(pid, "P-1", "試験商品", "個", qty, price)
	if e != nil {
		t.Fatal(e)
	}
	addr, _ := NewShippingAddress(Address{"試験商事", "担当", "1000001", "東京都", "0300000000"})
	rate, _ := NewDiscountRate(10)
	o, e := NewOrder(id, actor, []Line{line}, addr, rate, "WELCOME10", "", "", time.Now())
	if e != nil {
		t.Fatal(e)
	}
	return o
}
func TestOrderPrice(t *testing.T) {
	o := fixture(t).Snapshot()
	if o.Discount != 299 || o.Net != 2698 || o.Tax != 269 || o.Total != 2967 || o.Points != 26 {
		t.Fatal(o)
	}
}
func TestValueBoundariesAndRestore(t *testing.T) {
	if _, e := NewOrderQuantity(0); e == nil {
		t.Fatal("zero order quantity")
	}
	if _, e := NewQuantity(-1); e == nil {
		t.Fatal("negative stock")
	}
	if _, e := NewMoney(-1); e == nil {
		t.Fatal("negative money")
	}
	if _, e := NewDiscountRate(101); e == nil {
		t.Fatal("invalid rate")
	}
	if _, e := ParseID[ProductTag]("not-a-uuid"); e == nil {
		t.Fatal("invalid ID")
	}
	if _, e := ParseOrderStatus(255); e == nil {
		t.Fatal("invalid state")
	}
	var stock Stock
	if e := json.Unmarshal([]byte(`{"onHand":1,"reserved":2}`), &stock); e == nil {
		t.Fatal("corrupt DB stock accepted")
	}
	o := fixture(t)
	snapshot := o.Snapshot()
	snapshot.Total++
	b, _ := json.Marshal(snapshot)
	var restored Order
	if e := json.Unmarshal(b, &restored); e == nil {
		t.Fatal("corrupt money accepted")
	}
	snapshot = o.Snapshot()
	snapshot.Status = Packed.Code()
	b, _ = json.Marshal(snapshot)
	if e := json.Unmarshal(b, &restored); e == nil {
		t.Fatal("packing without boxes accepted")
	}
	snapshot = o.Snapshot()
	snapshot.Lines[0].Quantity = 999
	snapshot.History[0].Note = "mutated"
	if o.Snapshot().Lines[0].Quantity != 3 || o.Snapshot().History[0].Note == "mutated" {
		t.Fatal("snapshot leaked mutable state")
	}
	actor, _ := ParseID[ActorTag](uuid.NewString())
	if e := o.Move(Packed, actor, "", time.Now()); e == nil {
		t.Fatal("illegal transition allowed")
	}
	b, e := json.Marshal(o)
	if e != nil {
		t.Fatal(e)
	}
	if e = json.Unmarshal(b, &restored); e != nil {
		t.Fatal(e)
	}
	if restored.Snapshot().Total != o.Snapshot().Total {
		t.Fatal("roundtrip mismatch")
	}
}

func TestOrderRestoreFulfillmentRequirements(t *testing.T) {
	order := fixture(t)
	actor, _ := ParseID[ActorTag](uuid.NewString())
	now := time.Now()
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	must(order.Move(Accepted, actor, "", now))
	must(order.AssignWorker(actor))
	must(order.Move(Picking, actor, "", now))
	must(order.Pack(BoxCount{2}, WeightGrams{1200}))
	must(order.Move(Packed, actor, "", now))
	must(order.AssignCarrier(actor, "TRACK-1"))
	must(order.Move(Shipping, actor, "", now))
	for _, tc := range []struct {
		name   string
		change func(*OrderSnapshot)
	}{
		{"valid", nil},
		{"missing worker", func(s *OrderSnapshot) { s.WorkerID = "" }},
		{"missing boxes", func(s *OrderSnapshot) { s.Boxes = nil }},
		{"invalid weight", func(s *OrderSnapshot) { s.Boxes[0].Weight = 500001 }},
		{"missing tracking", func(s *OrderSnapshot) { s.Tracking = "" }},
		{"broken history", func(s *OrderSnapshot) { s.History[2].From = PaymentPending.Code() }},
		{"history status mismatch", func(s *OrderSnapshot) { s.Status = Delivered.Code() }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := order.Snapshot()
			if tc.change != nil {
				tc.change(&s)
			}
			b, err := json.Marshal(s)
			if err != nil {
				t.Fatal(err)
			}
			var restored Order
			if err := json.Unmarshal(b, &restored); (err != nil) != (tc.change != nil) {
				t.Fatalf("restore error=%v", err)
			}
		})
	}
}
