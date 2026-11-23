package app

import (
	"reflect"
	"testing"
	"time"

	plan "wholesale.local/orbit/generated/CommandFlow"
	d "wholesale.local/orbit/generated/dafny"
	"wholesale.local/orbit/internal/domain"
	"wholesale.local/orbit/internal/store"
)

func TestCartPlanAndExpiryBoundary(t *testing.T) {
	actor, product, other := store.ID("customer-a"), store.ID("P-1001"), store.ID("P-1002")
	cart := domain.CartSnapshot{CustomerID: actor, Items: map[string]int64{}}
	a := commandAttempt{actor: domain.Actor{ID: actor, Role: "customer"}}
	decide := func(kind string, seconds, quantity int64) plan.CommandPlan {
		a.command.Type = kind
		a.command.Payload = Payload{ProductID: product, Quantity: quantity}
		a.now = time.Unix(seconds, 0)
		state := plan.Companion_CartState_.Create_CartState_(
			textValue(actor), d.One, itemMap(cart.Items), d.IntOfInt64(cart.Expires),
			plan.Companion_RequestedDay_.Create_Unspecified_(), d.Zero, plan.Companion_Coupon_.Create_NoCoupon_(),
		)
		return plan.Companion_Default___.DecideCommand(a.input(), plan.Companion_Snapshot_.Create_Cart_(state))
	}
	first := decide("cart.set", 100, 1)
	if !first.Is_UpdateCart() || first.Dtor_expiry().Int64() != 3700 ||
		!reflect.DeepEqual(readItems(first.Dtor_reservations()), map[string]int64{product: 1}) {
		t.Fatalf("initial reservation: %v", first)
	}
	cart.Items = readItems(first.Dtor_nextItems())
	cart.Expires = first.Dtor_expiry().Int64()
	if unchanged := decide("cart.set", 200, 1); !unchanged.Is_Unchanged() {
		t.Fatalf("same quantity extended expiry: %v", unchanged)
	}
	changed := decide("cart.set", 200, 2)
	if !changed.Is_UpdateCart() || changed.Dtor_expiry().Int64() != 3800 {
		t.Fatalf("changed quantity did not extend expiry: %v", changed)
	}
	if checkout := decide("order.checkout", 3699, 0); !checkout.Is_Checkout() {
		t.Fatalf("checkout before expiry: %v", checkout)
	}
	if denied := decide("order.checkout", 3700, 0); !denied.Is_Reject() {
		t.Fatalf("checkout at expiry: %v", denied)
	}
	cart.Items = map[string]int64{product: 2, other: 2}
	const maxTime = int64(1<<63 - 1)
	if denied := decide("cart.set", maxTime, 3); !denied.Is_Reject() {
		t.Fatalf("overflow accepted: %v", denied)
	}
	reset := decide("cart.set", 3700, 3)
	if !reset.Is_UpdateCart() || reset.Dtor_expiry().Int64() != 7300 ||
		!reflect.DeepEqual(readItems(reset.Dtor_reservations()), map[string]int64{product: 1, other: -2}) ||
		!reflect.DeepEqual(readItems(reset.Dtor_nextItems()), map[string]int64{product: 3}) {
		t.Fatalf("expired cart reset: %v", reset)
	}
}

func TestExecuteCommandInjectsOnlyRequestedReads(t *testing.T) {
	customer, product := store.ID("customer-a"), store.ID("P-1001")
	for _, role := range []string{"customer", "operator"} {
		t.Run(role, func(t *testing.T) {
			a := commandAttempt{
				actor: domain.Actor{ID: customer, Role: role},
				command: Command{ID: store.ID("test-checkout"), Type: "order.checkout",
					Payload: Payload{CustomerID: store.ID("customer-b"), Coupon: "WELCOME10"}},
				now: time.Unix(100, 0),
			}
			e := &checkoutEffects{t: t, input: a.input(), customer: customer, product: product}
			status := plan.Companion_Default___.ExecuteCommand(e)
			if !status.Is_Ok() || !e.committed {
				t.Fatalf("status=%v committed=%t", status, e.committed)
			}
			want := []string{"rejected"}
			if role == "customer" {
				want = []string{"cart", "coupon", "dependencies", "applied", "succeeded"}
			}
			if !reflect.DeepEqual(e.calls, want) {
				t.Fatalf("calls = %v, want %v", e.calls, want)
			}
		})
	}
}

type checkoutEffects struct {
	plan.Effects
	t                 *testing.T
	input             plan.CommandInput
	customer, product string
	calls             []string
	committed         bool
}

func (*checkoutEffects) Begin() plan.Status { return plan.Companion_Status_.Create_Ok_() }

func (*checkoutEffects) LockCommand() (plan.Status, d.Sequence) {
	return plan.Companion_Status_.Create_Ok_(), textValue("queued")
}

func (e *checkoutEffects) ReadInput() (plan.Status, plan.CommandInput) {
	return plan.Companion_Status_.Create_Ok_(), e.input
}

func (e *checkoutEffects) Read(request plan.ReadRequest) (plan.Status, plan.Snapshot) {
	e.calls = append(e.calls, "cart")
	if !request.Is_ReadCart() || request.Dtor_customerId().VerbatimString(false) != e.customer {
		e.t.Fatalf("read request = %v", request)
	}
	state := plan.Companion_CartState_.Create_CartState_(
		textValue(e.customer), d.One, itemMap(map[string]int64{e.product: 2}), d.IntOf(1000),
		plan.Companion_RequestedDay_.Create_Unspecified_(), d.Zero, plan.Companion_Coupon_.Create_NoCoupon_(),
	)
	return plan.Companion_Status_.Create_Ok_(), plan.Companion_Snapshot_.Create_Cart_(state)
}

func (e *checkoutEffects) ReadCoupon(code d.Sequence) (plan.Status, plan.Coupon) {
	e.calls = append(e.calls, "coupon")
	if code.VerbatimString(false) != "WELCOME10" {
		e.t.Fatalf("coupon code = %v", code)
	}
	return plan.Companion_Status_.Create_Ok_(), plan.Companion_Coupon_.Create_CouponRate_(d.IntOf(10))
}

func (e *checkoutEffects) ReadDependencies(reads plan.Dependencies) plan.Status {
	e.calls = append(e.calls, "dependencies")
	products, order := reads.Dtor_products(), reads.Dtor_newOrder()
	if products.CardinalityInt() != 1 || !products.Contains(textValue(e.product)) ||
		!order.Is_Id() || order.Dtor_value().VerbatimString(false) != store.ID("test-checkout") {
		e.t.Fatalf("dependencies = %v", reads)
	}
	return plan.Companion_Status_.Create_Ok_()
}

func (e *checkoutEffects) Apply(p plan.CommandPlan) plan.Status {
	e.calls = append(e.calls, "applied")
	if !p.Is_Checkout() || p.Dtor_discount().Int() != 10 {
		e.t.Fatalf("plan = %v", p)
	}
	return plan.Companion_Status_.Create_Ok_()
}

func (*checkoutEffects) SetRejection(d.Sequence) {}

func (e *checkoutEffects) RecordRejection() plan.Status {
	e.calls = append(e.calls, "rejected")
	return plan.Companion_Status_.Create_Ok_()
}

func (e *checkoutEffects) RecordSuccess() plan.Status {
	e.calls = append(e.calls, "succeeded")
	return plan.Companion_Status_.Create_Ok_()
}

func (e *checkoutEffects) Commit() plan.Status {
	e.committed = true
	return plan.Companion_Status_.Create_Ok_()
}
