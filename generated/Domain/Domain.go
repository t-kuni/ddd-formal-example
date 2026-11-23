// Package Domain
// Dafny module Domain compiled into Go

package Domain

import (
	os "os"

	m__System "wholesale.local/orbit/generated/System_"
	_dafny "wholesale.local/orbit/generated/dafny"
)

var _ = os.Args
var _ _dafny.Dummy__
var _ m__System.Dummy__

type Dummy__ struct{}

// Definition of class Default__
type Default__ struct {
	dummy byte
}

func New_Default___() *Default__ {
	_this := Default__{}

	return &_this
}

type CompanionStruct_Default___ struct {
}

var Companion_Default___ = CompanionStruct_Default___{}

func (_this *Default__) Equals(other *Default__) bool {
	return _this == other
}

func (_this *Default__) EqualsGeneric(x interface{}) bool {
	other, ok := x.(*Default__)
	return ok && _this.Equals(other)
}

func (*Default__) String() string {
	return "_module.Default__"
}
func (_this *Default__) ParentTraits_() []*_dafny.TraitID {
	return [](*_dafny.TraitID){}
}

var _ _dafny.TraitOffspring = &Default__{}

func (_static *CompanionStruct_Default___) StockValid(h _dafny.Int, r _dafny.Int) bool {
	return (((r).Sign() != -1) && ((r).Cmp(h) <= 0)) && ((h).Cmp(_dafny.IntOfInt64(1000000000)) <= 0)
}
func (_static *CompanionStruct_Default___) Reserve(h _dafny.Int, r _dafny.Int, delta _dafny.Int) (bool, _dafny.Int) {
	var ok bool = false
	_ = ok
	var next _dafny.Int = _dafny.Zero
	_ = next
	ok = (((r).Plus(delta)).Sign() != -1) && (((r).Plus(delta)).Cmp(h) <= 0)
	if ok {
		next = (r).Plus(delta)
	} else {
		next = r
	}
	return ok, next
}
func (_static *CompanionStruct_Default___) Adjust(h _dafny.Int, r _dafny.Int, delta _dafny.Int) (bool, _dafny.Int) {
	var ok bool = false
	_ = ok
	var next _dafny.Int = _dafny.Zero
	_ = next
	ok = ((r).Cmp((h).Plus(delta)) <= 0) && (((h).Plus(delta)).Cmp(_dafny.IntOfInt64(1000000000)) <= 0)
	if ok {
		next = (h).Plus(delta)
	} else {
		next = h
	}
	return ok, next
}
func (_static *CompanionStruct_Default___) Ship(h _dafny.Int, r _dafny.Int, qty _dafny.Int) (bool, _dafny.Int, _dafny.Int) {
	var ok bool = false
	_ = ok
	var nh _dafny.Int = _dafny.Zero
	_ = nh
	var nr _dafny.Int = _dafny.Zero
	_ = nr
	ok = ((qty).Sign() == 1) && ((qty).Cmp(r) <= 0)
	if ok {
		nh = (h).Minus(qty)
	} else {
		nh = h
	}
	if ok {
		nr = (r).Minus(qty)
	} else {
		nr = r
	}
	return ok, nh, nr
}
func (_static *CompanionStruct_Default___) Allowed(s _dafny.Int, t _dafny.Int) bool {
	return ((((((((s).Sign() == 0) && ((((t).Cmp(_dafny.One) == 0) || ((t).Cmp(_dafny.IntOfInt64(8)) == 0)) || ((t).Cmp(_dafny.IntOfInt64(9)) == 0))) || (((s).Cmp(_dafny.One) == 0) && (((t).Cmp(_dafny.IntOfInt64(2)) == 0) || ((t).Cmp(_dafny.IntOfInt64(8)) == 0)))) || (((s).Cmp(_dafny.IntOfInt64(2)) == 0) && ((((t).Cmp(_dafny.IntOfInt64(3)) == 0) || ((t).Cmp(_dafny.IntOfInt64(4)) == 0)) || ((t).Cmp(_dafny.IntOfInt64(8)) == 0)))) || (((s).Cmp(_dafny.IntOfInt64(3)) == 0) && (((t).Cmp(_dafny.IntOfInt64(2)) == 0) || ((t).Cmp(_dafny.IntOfInt64(8)) == 0)))) || (((s).Cmp(_dafny.IntOfInt64(4)) == 0) && ((t).Cmp(_dafny.IntOfInt64(5)) == 0))) || (((s).Cmp(_dafny.IntOfInt64(5)) == 0) && (((t).Cmp(_dafny.IntOfInt64(6)) == 0) || ((t).Cmp(_dafny.IntOfInt64(7)) == 0)))) || (((s).Cmp(_dafny.IntOfInt64(6)) == 0) && (((t).Cmp(_dafny.IntOfInt64(5)) == 0) || ((t).Cmp(_dafny.IntOfInt64(10)) == 0)))
}
func (_static *CompanionStruct_Default___) Transition(s _dafny.Int, t _dafny.Int) (bool, _dafny.Int) {
	var ok bool = false
	_ = ok
	var next _dafny.Int = _dafny.Zero
	_ = next
	ok = Companion_Default___.Allowed(s, t)
	if ok {
		next = t
	} else {
		next = s
	}
	return ok, next
}
func (_static *CompanionStruct_Default___) Touch(now _dafny.Int) _dafny.Int {
	var expires _dafny.Int = _dafny.Zero
	_ = expires
	expires = (now).Plus(_dafny.IntOfInt64(3600))
	return expires
}
func (_static *CompanionStruct_Default___) Expired(now _dafny.Int, expires _dafny.Int) bool {
	var expired bool = false
	_ = expired
	expired = (now).Cmp(expires) >= 0
	return expired
}
func (_static *CompanionStruct_Default___) MayExpire(now _dafny.Int, current _dafny.Int, requested _dafny.Int, hasItems bool) bool {
	var expire bool = false
	_ = expire
	expire = (((now).Cmp(current) >= 0) && ((current).Cmp(requested) == 0)) && (hasItems)
	return expire
}
func (_static *CompanionStruct_Default___) MayCheckout(now _dafny.Int, expires _dafny.Int, hasItems bool) bool {
	var allowed bool = false
	_ = allowed
	allowed = ((now).Cmp(expires) < 0) && (hasItems)
	return allowed
}
func (_static *CompanionStruct_Default___) Price(subtotal _dafny.Int, percent _dafny.Int) (_dafny.Int, _dafny.Int, _dafny.Int, _dafny.Int, _dafny.Int) {
	var discount _dafny.Int = _dafny.Zero
	_ = discount
	var net _dafny.Int = _dafny.Zero
	_ = net
	var tax _dafny.Int = _dafny.Zero
	_ = tax
	var total _dafny.Int = _dafny.Zero
	_ = total
	var points _dafny.Int = _dafny.Zero
	_ = points
	discount = ((subtotal).Times(percent)).DivBy(_dafny.IntOfInt64(100))
	net = (subtotal).Minus(discount)
	tax = (net).DivBy(_dafny.IntOfInt64(10))
	total = (net).Plus(tax)
	points = (net).DivBy(_dafny.IntOfInt64(100))
	return discount, net, tax, total, points
}

// End of class Default__
