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
	return "Domain.Default__"
}
func (_this *Default__) ParentTraits_() []*_dafny.TraitID {
	return [](*_dafny.TraitID){}
}

var _ _dafny.TraitOffspring = &Default__{}

func (_static *CompanionStruct_Default___) QuantityValid(v _dafny.Int) bool {
	return ((v).Sign() != -1) && ((v).Cmp(Companion_Default___.StockLimit()) <= 0)
}
func (_static *CompanionStruct_Default___) StockDeltaValid(v _dafny.Int) bool {
	return (((_dafny.Zero).Minus(Companion_Default___.StockLimit())).Cmp(v) <= 0) && ((v).Cmp(Companion_Default___.StockLimit()) <= 0)
}
func (_static *CompanionStruct_Default___) OrderQuantityValid(v _dafny.Int) bool {
	return ((_dafny.One).Cmp(v) <= 0) && ((v).Cmp(Companion_Default___.LineQuantityLimit()) <= 0)
}
func (_static *CompanionStruct_Default___) CartQuantityValid(v _dafny.Int) bool {
	return ((v).Sign() != -1) && ((v).Cmp(Companion_Default___.LineQuantityLimit()) <= 0)
}
func (_static *CompanionStruct_Default___) MoneyValid(v _dafny.Int) bool {
	return ((v).Sign() != -1) && ((v).Cmp(_dafny.IntOfInt64(1100000000000)) <= 0)
}
func (_static *CompanionStruct_Default___) UnitPriceValid(v _dafny.Int) bool {
	return ((v).Sign() != -1) && ((v).Cmp(_dafny.IntOfInt64(10000000)) <= 0)
}
func (_static *CompanionStruct_Default___) DiscountRateValid(v _dafny.Int) bool {
	return ((v).Sign() != -1) && ((v).Cmp(_dafny.IntOfInt64(100)) <= 0)
}
func (_static *CompanionStruct_Default___) PointsValid(v _dafny.Int) bool {
	return ((v).Sign() != -1) && ((v).Cmp(_dafny.IntOfInt64(1000000000000000)) <= 0)
}
func (_static *CompanionStruct_Default___) BoxCountValid(v _dafny.Int) bool {
	return ((_dafny.One).Cmp(v) <= 0) && ((v).Cmp(_dafny.IntOfInt64(100)) <= 0)
}
func (_static *CompanionStruct_Default___) WeightValid(v _dafny.Int) bool {
	return ((_dafny.One).Cmp(v) <= 0) && ((v).Cmp(_dafny.IntOfInt64(500000)) <= 0)
}
func (_static *CompanionStruct_Default___) TextFieldValid(nonBlank bool, byteCount _dafny.Int) bool {
	return (nonBlank) && (((_dafny.One).Cmp(byteCount) <= 0) && ((byteCount).Cmp(_dafny.IntOfInt64(750)) <= 0))
}
func (_static *CompanionStruct_Default___) Credit(balance _dafny.Int, award _dafny.Int) (bool, _dafny.Int) {
	var ok bool = false
	_ = ok
	var next _dafny.Int = _dafny.Zero
	_ = next
	ok = Companion_Default___.PointsValid((balance).Plus(award))
	if ok {
		next = (balance).Plus(award)
	} else {
		next = balance
	}
	return ok, next
}
func (_static *CompanionStruct_Default___) StockValid(h _dafny.Int, r _dafny.Int) bool {
	return (((r).Sign() != -1) && ((r).Cmp(h) <= 0)) && ((h).Cmp(Companion_Default___.StockLimit()) <= 0)
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
	ok = ((r).Cmp((h).Plus(delta)) <= 0) && (((h).Plus(delta)).Cmp(Companion_Default___.StockLimit()) <= 0)
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
func (_static *CompanionStruct_Default___) StatusValid(s _dafny.Int) bool {
	return ((s).Sign() != -1) && ((s).Cmp(_dafny.IntOfInt64(10)) <= 0)
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
func (_static *CompanionStruct_Default___) TimestampValid(seconds _dafny.Int) bool {
	return (seconds).Sign() != -1
}
func (_static *CompanionStruct_Default___) CartShapeValid(count _dafny.Int, expires _dafny.Int) bool {
	return (((count).Sign() != -1) && ((count).Cmp(Companion_Default___.LineCountLimit()) <= 0)) && (Companion_Default___.TimestampValid(expires))
}
func (_static *CompanionStruct_Default___) ReleasedReservation(quantity _dafny.Int) _dafny.Int {
	return (_dafny.Zero).Minus(quantity)
}
func (_static *CompanionStruct_Default___) PlanCartChange(count _dafny.Int, previousQuantity _dafny.Int, requested _dafny.Int, now _dafny.Int, expires _dafny.Int) CartChange {
	var plan CartChange = Companion_CartChange_.Default()
	_ = plan
	if ((((count).Cmp(Companion_Default___.LineCountLimit()) == 0) && ((previousQuantity).Sign() == 0)) && ((requested).Sign() == 1)) && ((now).Cmp(expires) < 0) {
		plan = Companion_CartChange_.Create_CartFull_()
	} else if (((count).Sign() == 0) && ((requested).Sign() == 0)) || (((previousQuantity).Cmp(requested) == 0) && ((now).Cmp(expires) < 0)) {
		plan = Companion_CartChange_.Create_CartUnchanged_()
	} else if (now).Cmp((_dafny.IntOfInt64(9223372036854775807)).Minus(_dafny.IntOfInt64(3600))) > 0 {
		plan = Companion_CartChange_.Create_CartTimeOutOfRange_()
	} else {
		var _0_expiry _dafny.Int
		_ = _0_expiry
		var _out0 _dafny.Int
		_ = _out0
		_out0 = Companion_Default___.Touch(now)
		_0_expiry = _out0
		plan = Companion_CartChange_.Create_CartChanged_((now).Cmp(expires) >= 0, (requested).Minus(previousQuantity), _0_expiry)
	}
	return plan
}
func (_static *CompanionStruct_Default___) OrderLineCountValid(count _dafny.Int) bool {
	return ((_dafny.One).Cmp(count) <= 0) && ((count).Cmp(Companion_Default___.LineCountLimit()) <= 0)
}
func (_static *CompanionStruct_Default___) LineValid(hasProduct bool, hasName bool, quantity _dafny.Int, unitPrice _dafny.Int) bool {
	return (((hasProduct) && (hasName)) && (Companion_Default___.OrderQuantityValid(quantity))) && (Companion_Default___.UnitPriceValid(unitPrice))
}
func (_static *CompanionStruct_Default___) RegisterProduct(seen _dafny.Set, product _dafny.Sequence) (bool, _dafny.Set) {
	var ok bool = false
	_ = ok
	var next _dafny.Set = _dafny.EmptySet
	_ = next
	ok = !(seen).Contains(product)
	if ok {
		next = (seen).Union(_dafny.SetOf(product))
	} else {
		next = seen
	}
	return ok, next
}
func (_static *CompanionStruct_Default___) LineAmount(quantity _dafny.Int, unitPrice _dafny.Int) _dafny.Int {
	return (quantity).Times(unitPrice)
}
func (_static *CompanionStruct_Default___) AddLineAmount(subtotal _dafny.Int, quantity _dafny.Int, unitPrice _dafny.Int) (bool, _dafny.Int) {
	var ok bool = false
	_ = ok
	var next _dafny.Int = _dafny.Zero
	_ = next
	var _0_candidate _dafny.Int
	_ = _0_candidate
	_0_candidate = (subtotal).Plus(Companion_Default___.LineAmount(quantity, unitPrice))
	ok = (_0_candidate).Cmp(Companion_Default___.SubtotalLimit()) <= 0
	if ok {
		next = _0_candidate
	} else {
		next = subtotal
	}
	return ok, next
}
func (_static *CompanionStruct_Default___) InitialHistory(from _dafny.Int, to _dafny.Int) bool {
	return ((from).Cmp(_dafny.IntOfInt64(-1)) == 0) && ((to).Sign() == 0)
}
func (_static *CompanionStruct_Default___) HistoryStepsValid(edges _dafny.Sequence) bool {
	return _dafny.Quantifier(_dafny.IntegerRange(_dafny.One, _dafny.IntOfUint32((edges).Cardinality())), true, func(_forall_var_0 _dafny.Int) bool {
		var _0_i _dafny.Int
		_0_i = interface{}(_forall_var_0).(_dafny.Int)
		return !(((_dafny.One).Cmp(_0_i) <= 0) && ((_0_i).Cmp(_dafny.IntOfUint32((edges).Cardinality())) < 0)) || (((((edges).Select((_0_i).Uint32()).(HistoryEdge)).Dtor_from()).Cmp(((edges).Select(((_0_i).Minus(_dafny.One)).Uint32()).(HistoryEdge)).Dtor_to()) == 0) && (Companion_Default___.Allowed(((edges).Select((_0_i).Uint32()).(HistoryEdge)).Dtor_from(), ((edges).Select((_0_i).Uint32()).(HistoryEdge)).Dtor_to())))
	})
}
func (_static *CompanionStruct_Default___) CheckHistory(edges _dafny.Sequence, status _dafny.Int) HistoryCheck {
	var result HistoryCheck = Companion_HistoryCheck_.Default()
	_ = result
	if ((_dafny.IntOfUint32((edges).Cardinality())).Sign() == 0) || (!(Companion_Default___.InitialHistory(((edges).Select(0).(HistoryEdge)).Dtor_from(), ((edges).Select(0).(HistoryEdge)).Dtor_to()))) {
		result = Companion_HistoryCheck_.Create_InvalidStart_()
		return result
	}
	if !(Companion_Default___.HistoryStepsValid(edges)) {
		result = Companion_HistoryCheck_.Create_InvalidStep_()
		return result
	}
	if (((edges).Select(((_dafny.IntOfUint32((edges).Cardinality())).Minus(_dafny.One)).Uint32()).(HistoryEdge)).Dtor_to()).Cmp(status) == 0 {
		result = Companion_HistoryCheck_.Create_ValidHistory_()
	} else {
		result = Companion_HistoryCheck_.Create_StatusMismatch_()
	}
	return result
	return result
}
func (_static *CompanionStruct_Default___) NeedsWorker(status _dafny.Int) bool {
	return (_dafny.SetOf(_dafny.IntOfInt64(2), _dafny.IntOfInt64(3), _dafny.IntOfInt64(4), _dafny.IntOfInt64(5), _dafny.IntOfInt64(6), _dafny.IntOfInt64(7), _dafny.IntOfInt64(10))).Contains(status)
}
func (_static *CompanionStruct_Default___) NeedsBoxes(status _dafny.Int) bool {
	return (_dafny.SetOf(_dafny.IntOfInt64(4), _dafny.IntOfInt64(5), _dafny.IntOfInt64(6), _dafny.IntOfInt64(7), _dafny.IntOfInt64(10))).Contains(status)
}
func (_static *CompanionStruct_Default___) NeedsCarrier(status _dafny.Int) bool {
	return (_dafny.SetOf(_dafny.IntOfInt64(5), _dafny.IntOfInt64(6), _dafny.IntOfInt64(7), _dafny.IntOfInt64(10))).Contains(status)
}
func (_static *CompanionStruct_Default___) WorkerReady(status _dafny.Int, hasAssignee bool) bool {
	return (!(Companion_Default___.NeedsWorker(status))) || (hasAssignee)
}
func (_static *CompanionStruct_Default___) BoxesReady(status _dafny.Int, count _dafny.Int) bool {
	return (!(Companion_Default___.NeedsBoxes(status))) || (Companion_Default___.BoxCountValid(count))
}
func (_static *CompanionStruct_Default___) BoxValid(parsedID bool, weight _dafny.Int) bool {
	return (parsedID) && (Companion_Default___.WeightValid(weight))
}
func (_static *CompanionStruct_Default___) CarrierReady(status _dafny.Int, hasAssignee bool, hasTracking bool) bool {
	return (!(Companion_Default___.NeedsCarrier(status))) || ((hasAssignee) && (hasTracking))
}
func (_static *CompanionStruct_Default___) MayStartWork(status _dafny.Int, validActor bool) bool {
	return (validActor) && (Companion_Default___.Allowed(status, _dafny.IntOfInt64(2)))
}
func (_static *CompanionStruct_Default___) MayAssignWorker(hasAssignee bool, sameActor bool) bool {
	return (!(hasAssignee)) || (sameActor)
}
func (_static *CompanionStruct_Default___) MayPack(status _dafny.Int, count _dafny.Int, weight _dafny.Int) bool {
	return ((Companion_Default___.Allowed(status, _dafny.IntOfInt64(4))) && (Companion_Default___.BoxCountValid(count))) && (Companion_Default___.WeightValid(weight))
}
func (_static *CompanionStruct_Default___) MayAssignCarrier(status _dafny.Int, validActor bool, trackingBytes _dafny.Int) bool {
	return (((status).Cmp(_dafny.IntOfInt64(4)) == 0) && (validActor)) && (((_dafny.One).Cmp(trackingBytes) <= 0) && ((trackingBytes).Cmp(_dafny.IntOfInt64(100)) <= 0))
}
func (_static *CompanionStruct_Default___) StockLimit() _dafny.Int {
	return _dafny.IntOfInt64(1000000000)
}
func (_static *CompanionStruct_Default___) LineQuantityLimit() _dafny.Int {
	return _dafny.IntOfInt64(10000)
}
func (_static *CompanionStruct_Default___) SubtotalLimit() _dafny.Int {
	return _dafny.IntOfInt64(1000000000000)
}
func (_static *CompanionStruct_Default___) LineCountLimit() _dafny.Int {
	return _dafny.IntOfInt64(100)
}

// End of class Default__

// Definition of datatype CartChange
type CartChange struct {
	Data_CartChange_
}

func (_this CartChange) Get_() Data_CartChange_ {
	return _this.Data_CartChange_
}

type Data_CartChange_ interface {
	isCartChange()
}

type CompanionStruct_CartChange_ struct {
}

var Companion_CartChange_ = CompanionStruct_CartChange_{}

type CartChange_CartFull struct {
}

func (CartChange_CartFull) isCartChange() {}

func (CompanionStruct_CartChange_) Create_CartFull_() CartChange {
	return CartChange{CartChange_CartFull{}}
}

func (_this CartChange) Is_CartFull() bool {
	_, ok := _this.Get_().(CartChange_CartFull)
	return ok
}

type CartChange_CartUnchanged struct {
}

func (CartChange_CartUnchanged) isCartChange() {}

func (CompanionStruct_CartChange_) Create_CartUnchanged_() CartChange {
	return CartChange{CartChange_CartUnchanged{}}
}

func (_this CartChange) Is_CartUnchanged() bool {
	_, ok := _this.Get_().(CartChange_CartUnchanged)
	return ok
}

type CartChange_CartTimeOutOfRange struct {
}

func (CartChange_CartTimeOutOfRange) isCartChange() {}

func (CompanionStruct_CartChange_) Create_CartTimeOutOfRange_() CartChange {
	return CartChange{CartChange_CartTimeOutOfRange{}}
}

func (_this CartChange) Is_CartTimeOutOfRange() bool {
	_, ok := _this.Get_().(CartChange_CartTimeOutOfRange)
	return ok
}

type CartChange_CartChanged struct {
	Reset  bool
	Delta  _dafny.Int
	Expiry _dafny.Int
}

func (CartChange_CartChanged) isCartChange() {}

func (CompanionStruct_CartChange_) Create_CartChanged_(Reset bool, Delta _dafny.Int, Expiry _dafny.Int) CartChange {
	return CartChange{CartChange_CartChanged{Reset, Delta, Expiry}}
}

func (_this CartChange) Is_CartChanged() bool {
	_, ok := _this.Get_().(CartChange_CartChanged)
	return ok
}

func (CompanionStruct_CartChange_) Default() CartChange {
	return Companion_CartChange_.Create_CartFull_()
}

func (_this CartChange) Dtor_reset() bool {
	return _this.Get_().(CartChange_CartChanged).Reset
}

func (_this CartChange) Dtor_delta() _dafny.Int {
	return _this.Get_().(CartChange_CartChanged).Delta
}

func (_this CartChange) Dtor_expiry() _dafny.Int {
	return _this.Get_().(CartChange_CartChanged).Expiry
}

func (_this CartChange) String() string {
	switch data := _this.Get_().(type) {
	case nil:
		return "null"
	case CartChange_CartFull:
		{
			return "Domain.CartChange.CartFull"
		}
	case CartChange_CartUnchanged:
		{
			return "Domain.CartChange.CartUnchanged"
		}
	case CartChange_CartTimeOutOfRange:
		{
			return "Domain.CartChange.CartTimeOutOfRange"
		}
	case CartChange_CartChanged:
		{
			return "Domain.CartChange.CartChanged" + "(" + _dafny.String(data.Reset) + ", " + _dafny.String(data.Delta) + ", " + _dafny.String(data.Expiry) + ")"
		}
	default:
		{
			return "<unexpected>"
		}
	}
}

func (_this CartChange) Equals(other CartChange) bool {
	switch data1 := _this.Get_().(type) {
	case CartChange_CartFull:
		{
			_, ok := other.Get_().(CartChange_CartFull)
			return ok
		}
	case CartChange_CartUnchanged:
		{
			_, ok := other.Get_().(CartChange_CartUnchanged)
			return ok
		}
	case CartChange_CartTimeOutOfRange:
		{
			_, ok := other.Get_().(CartChange_CartTimeOutOfRange)
			return ok
		}
	case CartChange_CartChanged:
		{
			data2, ok := other.Get_().(CartChange_CartChanged)
			return ok && data1.Reset == data2.Reset && data1.Delta.Cmp(data2.Delta) == 0 && data1.Expiry.Cmp(data2.Expiry) == 0
		}
	default:
		{
			return false // unexpected
		}
	}
}

func (_this CartChange) EqualsGeneric(other interface{}) bool {
	typed, ok := other.(CartChange)
	return ok && _this.Equals(typed)
}

func Type_CartChange_() _dafny.TypeDescriptor {
	return type_CartChange_{}
}

type type_CartChange_ struct {
}

func (_this type_CartChange_) Default() interface{} {
	return Companion_CartChange_.Default()
}

func (_this type_CartChange_) String() string {
	return "Domain.CartChange"
}
func (_this CartChange) ParentTraits_() []*_dafny.TraitID {
	return [](*_dafny.TraitID){}
}

var _ _dafny.TraitOffspring = CartChange{}

// End of datatype CartChange

// Definition of datatype HistoryEdge
type HistoryEdge struct {
	Data_HistoryEdge_
}

func (_this HistoryEdge) Get_() Data_HistoryEdge_ {
	return _this.Data_HistoryEdge_
}

type Data_HistoryEdge_ interface {
	isHistoryEdge()
}

type CompanionStruct_HistoryEdge_ struct {
}

var Companion_HistoryEdge_ = CompanionStruct_HistoryEdge_{}

type HistoryEdge_HistoryEdge struct {
	From _dafny.Int
	To   _dafny.Int
}

func (HistoryEdge_HistoryEdge) isHistoryEdge() {}

func (CompanionStruct_HistoryEdge_) Create_HistoryEdge_(From _dafny.Int, To _dafny.Int) HistoryEdge {
	return HistoryEdge{HistoryEdge_HistoryEdge{From, To}}
}

func (_this HistoryEdge) Is_HistoryEdge() bool {
	_, ok := _this.Get_().(HistoryEdge_HistoryEdge)
	return ok
}

func (CompanionStruct_HistoryEdge_) Default() HistoryEdge {
	return Companion_HistoryEdge_.Create_HistoryEdge_(_dafny.Zero, _dafny.Zero)
}

func (_this HistoryEdge) Dtor_from() _dafny.Int {
	return _this.Get_().(HistoryEdge_HistoryEdge).From
}

func (_this HistoryEdge) Dtor_to() _dafny.Int {
	return _this.Get_().(HistoryEdge_HistoryEdge).To
}

func (_this HistoryEdge) String() string {
	switch data := _this.Get_().(type) {
	case nil:
		return "null"
	case HistoryEdge_HistoryEdge:
		{
			return "Domain.HistoryEdge.HistoryEdge" + "(" + _dafny.String(data.From) + ", " + _dafny.String(data.To) + ")"
		}
	default:
		{
			return "<unexpected>"
		}
	}
}

func (_this HistoryEdge) Equals(other HistoryEdge) bool {
	switch data1 := _this.Get_().(type) {
	case HistoryEdge_HistoryEdge:
		{
			data2, ok := other.Get_().(HistoryEdge_HistoryEdge)
			return ok && data1.From.Cmp(data2.From) == 0 && data1.To.Cmp(data2.To) == 0
		}
	default:
		{
			return false // unexpected
		}
	}
}

func (_this HistoryEdge) EqualsGeneric(other interface{}) bool {
	typed, ok := other.(HistoryEdge)
	return ok && _this.Equals(typed)
}

func Type_HistoryEdge_() _dafny.TypeDescriptor {
	return type_HistoryEdge_{}
}

type type_HistoryEdge_ struct {
}

func (_this type_HistoryEdge_) Default() interface{} {
	return Companion_HistoryEdge_.Default()
}

func (_this type_HistoryEdge_) String() string {
	return "Domain.HistoryEdge"
}
func (_this HistoryEdge) ParentTraits_() []*_dafny.TraitID {
	return [](*_dafny.TraitID){}
}

var _ _dafny.TraitOffspring = HistoryEdge{}

// End of datatype HistoryEdge

// Definition of datatype HistoryCheck
type HistoryCheck struct {
	Data_HistoryCheck_
}

func (_this HistoryCheck) Get_() Data_HistoryCheck_ {
	return _this.Data_HistoryCheck_
}

type Data_HistoryCheck_ interface {
	isHistoryCheck()
}

type CompanionStruct_HistoryCheck_ struct {
}

var Companion_HistoryCheck_ = CompanionStruct_HistoryCheck_{}

type HistoryCheck_InvalidStart struct {
}

func (HistoryCheck_InvalidStart) isHistoryCheck() {}

func (CompanionStruct_HistoryCheck_) Create_InvalidStart_() HistoryCheck {
	return HistoryCheck{HistoryCheck_InvalidStart{}}
}

func (_this HistoryCheck) Is_InvalidStart() bool {
	_, ok := _this.Get_().(HistoryCheck_InvalidStart)
	return ok
}

type HistoryCheck_InvalidStep struct {
}

func (HistoryCheck_InvalidStep) isHistoryCheck() {}

func (CompanionStruct_HistoryCheck_) Create_InvalidStep_() HistoryCheck {
	return HistoryCheck{HistoryCheck_InvalidStep{}}
}

func (_this HistoryCheck) Is_InvalidStep() bool {
	_, ok := _this.Get_().(HistoryCheck_InvalidStep)
	return ok
}

type HistoryCheck_StatusMismatch struct {
}

func (HistoryCheck_StatusMismatch) isHistoryCheck() {}

func (CompanionStruct_HistoryCheck_) Create_StatusMismatch_() HistoryCheck {
	return HistoryCheck{HistoryCheck_StatusMismatch{}}
}

func (_this HistoryCheck) Is_StatusMismatch() bool {
	_, ok := _this.Get_().(HistoryCheck_StatusMismatch)
	return ok
}

type HistoryCheck_ValidHistory struct {
}

func (HistoryCheck_ValidHistory) isHistoryCheck() {}

func (CompanionStruct_HistoryCheck_) Create_ValidHistory_() HistoryCheck {
	return HistoryCheck{HistoryCheck_ValidHistory{}}
}

func (_this HistoryCheck) Is_ValidHistory() bool {
	_, ok := _this.Get_().(HistoryCheck_ValidHistory)
	return ok
}

func (CompanionStruct_HistoryCheck_) Default() HistoryCheck {
	return Companion_HistoryCheck_.Create_InvalidStart_()
}

func (_ CompanionStruct_HistoryCheck_) AllSingletonConstructors() _dafny.Iterator {
	i := -1
	return func() (interface{}, bool) {
		i++
		switch i {
		case 0:
			return Companion_HistoryCheck_.Create_InvalidStart_(), true
		case 1:
			return Companion_HistoryCheck_.Create_InvalidStep_(), true
		case 2:
			return Companion_HistoryCheck_.Create_StatusMismatch_(), true
		case 3:
			return Companion_HistoryCheck_.Create_ValidHistory_(), true
		default:
			return HistoryCheck{}, false
		}
	}
}

func (_this HistoryCheck) String() string {
	switch _this.Get_().(type) {
	case nil:
		return "null"
	case HistoryCheck_InvalidStart:
		{
			return "Domain.HistoryCheck.InvalidStart"
		}
	case HistoryCheck_InvalidStep:
		{
			return "Domain.HistoryCheck.InvalidStep"
		}
	case HistoryCheck_StatusMismatch:
		{
			return "Domain.HistoryCheck.StatusMismatch"
		}
	case HistoryCheck_ValidHistory:
		{
			return "Domain.HistoryCheck.ValidHistory"
		}
	default:
		{
			return "<unexpected>"
		}
	}
}

func (_this HistoryCheck) Equals(other HistoryCheck) bool {
	switch _this.Get_().(type) {
	case HistoryCheck_InvalidStart:
		{
			_, ok := other.Get_().(HistoryCheck_InvalidStart)
			return ok
		}
	case HistoryCheck_InvalidStep:
		{
			_, ok := other.Get_().(HistoryCheck_InvalidStep)
			return ok
		}
	case HistoryCheck_StatusMismatch:
		{
			_, ok := other.Get_().(HistoryCheck_StatusMismatch)
			return ok
		}
	case HistoryCheck_ValidHistory:
		{
			_, ok := other.Get_().(HistoryCheck_ValidHistory)
			return ok
		}
	default:
		{
			return false // unexpected
		}
	}
}

func (_this HistoryCheck) EqualsGeneric(other interface{}) bool {
	typed, ok := other.(HistoryCheck)
	return ok && _this.Equals(typed)
}

func Type_HistoryCheck_() _dafny.TypeDescriptor {
	return type_HistoryCheck_{}
}

type type_HistoryCheck_ struct {
}

func (_this type_HistoryCheck_) Default() interface{} {
	return Companion_HistoryCheck_.Default()
}

func (_this type_HistoryCheck_) String() string {
	return "Domain.HistoryCheck"
}
func (_this HistoryCheck) ParentTraits_() []*_dafny.TraitID {
	return [](*_dafny.TraitID){}
}

var _ _dafny.TraitOffspring = HistoryCheck{}

// End of datatype HistoryCheck
