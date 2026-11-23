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
