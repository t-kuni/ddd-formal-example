// Package CommandModel
// Dafny module CommandModel compiled into Go

package CommandModel

import (
	os "os"

	m_Domain "wholesale.local/orbit/generated/Domain"
	m__System "wholesale.local/orbit/generated/System_"
	_dafny "wholesale.local/orbit/generated/dafny"
)

var _ = os.Args
var _ _dafny.Dummy__
var _ m__System.Dummy__
var _ m_Domain.Dummy__

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
	return "CommandModel.Default__"
}
func (_this *Default__) ParentTraits_() []*_dafny.TraitID {
	return [](*_dafny.TraitID){}
}

var _ _dafny.TraitOffspring = &Default__{}

func (_static *CompanionStruct_Default___) VersionMatches(expected _dafny.Int, actual _dafny.Int) bool {
	return ((expected).Sign() != 1) || ((expected).Cmp(actual) == 0)
}
func (_static *CompanionStruct_Default___) Release(lines _dafny.Map) _dafny.Map {
	return func() _dafny.Map {
		var _coll0 = _dafny.NewMapBuilder()
		_ = _coll0
		for _iter0 := _dafny.Iterate((lines).Keys().Elements()); ; {
			_compr_0, _ok0 := _iter0()
			if !_ok0 {
				break
			}
			var _0_id _dafny.Sequence
			_0_id = interface{}(_compr_0).(_dafny.Sequence)
			if (lines).Contains(_0_id) {
				_coll0.Add(_0_id, (_dafny.Zero).Minus((lines).Get(_0_id).(_dafny.Int)))
			}
		}
		return _coll0.ToMap()
	}()
}
func (_static *CompanionStruct_Default___) WithQuantity(items _dafny.Map, product _dafny.Sequence, quantity _dafny.Int) _dafny.Map {
	if (quantity).Sign() == 0 {
		return (items).Subtract(_dafny.SetOf(product))
	} else {
		return (items).Update(product, quantity)
	}
}
func (_static *CompanionStruct_Default___) InputSizeValid(input CommandInput) bool {
	return ((((input).Dtor_noteBytes()).Cmp(_dafny.IntOfInt64(1000)) <= 0) && (((input).Dtor_trackingBytes()).Cmp(_dafny.IntOfInt64(100)) <= 0)) && (((input).Dtor_requestedDateBytes()).Cmp(_dafny.IntOfInt64(10)) <= 0)
}
func (_static *CompanionStruct_Default___) PaymentPending() _dafny.Int {
	return _dafny.Zero
}
func (_static *CompanionStruct_Default___) Accepted() _dafny.Int {
	return _dafny.One
}
func (_static *CompanionStruct_Default___) Picking() _dafny.Int {
	return _dafny.IntOfInt64(2)
}
func (_static *CompanionStruct_Default___) PickBlocked() _dafny.Int {
	return _dafny.IntOfInt64(3)
}
func (_static *CompanionStruct_Default___) Packed() _dafny.Int {
	return _dafny.IntOfInt64(4)
}
func (_static *CompanionStruct_Default___) Shipping() _dafny.Int {
	return _dafny.IntOfInt64(5)
}
func (_static *CompanionStruct_Default___) DeliveryFailed() _dafny.Int {
	return _dafny.IntOfInt64(6)
}
func (_static *CompanionStruct_Default___) Delivered() _dafny.Int {
	return _dafny.IntOfInt64(7)
}
func (_static *CompanionStruct_Default___) Cancelled() _dafny.Int {
	return _dafny.IntOfInt64(8)
}
func (_static *CompanionStruct_Default___) PaymentFailed() _dafny.Int {
	return _dafny.IntOfInt64(9)
}
func (_static *CompanionStruct_Default___) Returned() _dafny.Int {
	return _dafny.IntOfInt64(10)
}
func (_static *CompanionStruct_Default___) PaymentTimeoutNanos() _dafny.Int {
	return ((_dafny.IntOfInt64(15)).Times(_dafny.IntOfInt64(60))).Times(_dafny.IntOfInt64(1000000000))
}
func (_static *CompanionStruct_Default___) StaleVersion() _dafny.Sequence {
	return _dafny.UnicodeSeqOfUtf8Bytes("更新が競合しました。再読込してください")
}

// End of class Default__

// Definition of datatype CommandInput
type CommandInput struct {
	Data_CommandInput_
}

func (_this CommandInput) Get_() Data_CommandInput_ {
	return _this.Data_CommandInput_
}

type Data_CommandInput_ interface {
	isCommandInput()
}

type CompanionStruct_CommandInput_ struct {
}

var Companion_CommandInput_ = CompanionStruct_CommandInput_{}

type CommandInput_CommandInput struct {
	Kind               _dafny.Sequence
	Role               _dafny.Sequence
	Actor              _dafny.Sequence
	Customer           _dafny.Sequence
	NoteBytes          _dafny.Int
	TrackingBytes      _dafny.Int
	RequestedDateBytes _dafny.Int
	ExpectedVersion    _dafny.Int
	Product            _dafny.Sequence
	ProductValid       bool
	StockDelta         _dafny.Int
	Quantity           _dafny.Int
	Now                _dafny.Int
	ExpireAt           _dafny.Int
	PaymentSucceeded   bool
	Tracking           _dafny.Sequence
	Boxes              _dafny.Int
	Weight             _dafny.Int
	OrderId            _dafny.Sequence
	CouponCode         _dafny.Sequence
	CommandId          _dafny.Sequence
}

func (CommandInput_CommandInput) isCommandInput() {}

func (CompanionStruct_CommandInput_) Create_CommandInput_(Kind _dafny.Sequence, Role _dafny.Sequence, Actor _dafny.Sequence, Customer _dafny.Sequence, NoteBytes _dafny.Int, TrackingBytes _dafny.Int, RequestedDateBytes _dafny.Int, ExpectedVersion _dafny.Int, Product _dafny.Sequence, ProductValid bool, StockDelta _dafny.Int, Quantity _dafny.Int, Now _dafny.Int, ExpireAt _dafny.Int, PaymentSucceeded bool, Tracking _dafny.Sequence, Boxes _dafny.Int, Weight _dafny.Int, OrderId _dafny.Sequence, CouponCode _dafny.Sequence, CommandId _dafny.Sequence) CommandInput {
	return CommandInput{CommandInput_CommandInput{Kind, Role, Actor, Customer, NoteBytes, TrackingBytes, RequestedDateBytes, ExpectedVersion, Product, ProductValid, StockDelta, Quantity, Now, ExpireAt, PaymentSucceeded, Tracking, Boxes, Weight, OrderId, CouponCode, CommandId}}
}

func (_this CommandInput) Is_CommandInput() bool {
	_, ok := _this.Get_().(CommandInput_CommandInput)
	return ok
}

func (CompanionStruct_CommandInput_) Default() CommandInput {
	return Companion_CommandInput_.Create_CommandInput_(_dafny.EmptySeq, _dafny.EmptySeq, _dafny.EmptySeq, _dafny.EmptySeq, _dafny.Zero, _dafny.Zero, _dafny.Zero, _dafny.Zero, _dafny.EmptySeq, false, _dafny.Zero, _dafny.Zero, _dafny.Zero, _dafny.Zero, false, _dafny.EmptySeq, _dafny.Zero, _dafny.Zero, _dafny.EmptySeq, _dafny.EmptySeq, _dafny.EmptySeq)
}

func (_this CommandInput) Dtor_kind() _dafny.Sequence {
	return _this.Get_().(CommandInput_CommandInput).Kind
}

func (_this CommandInput) Dtor_role() _dafny.Sequence {
	return _this.Get_().(CommandInput_CommandInput).Role
}

func (_this CommandInput) Dtor_actor() _dafny.Sequence {
	return _this.Get_().(CommandInput_CommandInput).Actor
}

func (_this CommandInput) Dtor_customer() _dafny.Sequence {
	return _this.Get_().(CommandInput_CommandInput).Customer
}

func (_this CommandInput) Dtor_noteBytes() _dafny.Int {
	return _this.Get_().(CommandInput_CommandInput).NoteBytes
}

func (_this CommandInput) Dtor_trackingBytes() _dafny.Int {
	return _this.Get_().(CommandInput_CommandInput).TrackingBytes
}

func (_this CommandInput) Dtor_requestedDateBytes() _dafny.Int {
	return _this.Get_().(CommandInput_CommandInput).RequestedDateBytes
}

func (_this CommandInput) Dtor_expectedVersion() _dafny.Int {
	return _this.Get_().(CommandInput_CommandInput).ExpectedVersion
}

func (_this CommandInput) Dtor_product() _dafny.Sequence {
	return _this.Get_().(CommandInput_CommandInput).Product
}

func (_this CommandInput) Dtor_productValid() bool {
	return _this.Get_().(CommandInput_CommandInput).ProductValid
}

func (_this CommandInput) Dtor_stockDelta() _dafny.Int {
	return _this.Get_().(CommandInput_CommandInput).StockDelta
}

func (_this CommandInput) Dtor_quantity() _dafny.Int {
	return _this.Get_().(CommandInput_CommandInput).Quantity
}

func (_this CommandInput) Dtor_now() _dafny.Int {
	return _this.Get_().(CommandInput_CommandInput).Now
}

func (_this CommandInput) Dtor_expireAt() _dafny.Int {
	return _this.Get_().(CommandInput_CommandInput).ExpireAt
}

func (_this CommandInput) Dtor_paymentSucceeded() bool {
	return _this.Get_().(CommandInput_CommandInput).PaymentSucceeded
}

func (_this CommandInput) Dtor_tracking() _dafny.Sequence {
	return _this.Get_().(CommandInput_CommandInput).Tracking
}

func (_this CommandInput) Dtor_boxes() _dafny.Int {
	return _this.Get_().(CommandInput_CommandInput).Boxes
}

func (_this CommandInput) Dtor_weight() _dafny.Int {
	return _this.Get_().(CommandInput_CommandInput).Weight
}

func (_this CommandInput) Dtor_orderId() _dafny.Sequence {
	return _this.Get_().(CommandInput_CommandInput).OrderId
}

func (_this CommandInput) Dtor_couponCode() _dafny.Sequence {
	return _this.Get_().(CommandInput_CommandInput).CouponCode
}

func (_this CommandInput) Dtor_commandId() _dafny.Sequence {
	return _this.Get_().(CommandInput_CommandInput).CommandId
}

func (_this CommandInput) String() string {
	switch data := _this.Get_().(type) {
	case nil:
		return "null"
	case CommandInput_CommandInput:
		{
			return "CommandModel.CommandInput.CommandInput" + "(" + data.Kind.VerbatimString(true) + ", " + data.Role.VerbatimString(true) + ", " + data.Actor.VerbatimString(true) + ", " + data.Customer.VerbatimString(true) + ", " + _dafny.String(data.NoteBytes) + ", " + _dafny.String(data.TrackingBytes) + ", " + _dafny.String(data.RequestedDateBytes) + ", " + _dafny.String(data.ExpectedVersion) + ", " + data.Product.VerbatimString(true) + ", " + _dafny.String(data.ProductValid) + ", " + _dafny.String(data.StockDelta) + ", " + _dafny.String(data.Quantity) + ", " + _dafny.String(data.Now) + ", " + _dafny.String(data.ExpireAt) + ", " + _dafny.String(data.PaymentSucceeded) + ", " + data.Tracking.VerbatimString(true) + ", " + _dafny.String(data.Boxes) + ", " + _dafny.String(data.Weight) + ", " + data.OrderId.VerbatimString(true) + ", " + data.CouponCode.VerbatimString(true) + ", " + data.CommandId.VerbatimString(true) + ")"
		}
	default:
		{
			return "<unexpected>"
		}
	}
}

func (_this CommandInput) Equals(other CommandInput) bool {
	switch data1 := _this.Get_().(type) {
	case CommandInput_CommandInput:
		{
			data2, ok := other.Get_().(CommandInput_CommandInput)
			return ok && data1.Kind.Equals(data2.Kind) && data1.Role.Equals(data2.Role) && data1.Actor.Equals(data2.Actor) && data1.Customer.Equals(data2.Customer) && data1.NoteBytes.Cmp(data2.NoteBytes) == 0 && data1.TrackingBytes.Cmp(data2.TrackingBytes) == 0 && data1.RequestedDateBytes.Cmp(data2.RequestedDateBytes) == 0 && data1.ExpectedVersion.Cmp(data2.ExpectedVersion) == 0 && data1.Product.Equals(data2.Product) && data1.ProductValid == data2.ProductValid && data1.StockDelta.Cmp(data2.StockDelta) == 0 && data1.Quantity.Cmp(data2.Quantity) == 0 && data1.Now.Cmp(data2.Now) == 0 && data1.ExpireAt.Cmp(data2.ExpireAt) == 0 && data1.PaymentSucceeded == data2.PaymentSucceeded && data1.Tracking.Equals(data2.Tracking) && data1.Boxes.Cmp(data2.Boxes) == 0 && data1.Weight.Cmp(data2.Weight) == 0 && data1.OrderId.Equals(data2.OrderId) && data1.CouponCode.Equals(data2.CouponCode) && data1.CommandId.Equals(data2.CommandId)
		}
	default:
		{
			return false // unexpected
		}
	}
}

func (_this CommandInput) EqualsGeneric(other interface{}) bool {
	typed, ok := other.(CommandInput)
	return ok && _this.Equals(typed)
}

func Type_CommandInput_() _dafny.TypeDescriptor {
	return type_CommandInput_{}
}

type type_CommandInput_ struct {
}

func (_this type_CommandInput_) Default() interface{} {
	return Companion_CommandInput_.Default()
}

func (_this type_CommandInput_) String() string {
	return "CommandModel.CommandInput"
}
func (_this CommandInput) ParentTraits_() []*_dafny.TraitID {
	return [](*_dafny.TraitID){}
}

var _ _dafny.TraitOffspring = CommandInput{}

// End of datatype CommandInput

// Definition of datatype OrderState
type OrderState struct {
	Data_OrderState_
}

func (_this OrderState) Get_() Data_OrderState_ {
	return _this.Data_OrderState_
}

type Data_OrderState_ interface {
	isOrderState()
}

type CompanionStruct_OrderState_ struct {
}

var Companion_OrderState_ = CompanionStruct_OrderState_{}

type OrderState_OrderState struct {
	Customer _dafny.Sequence
	Worker   _dafny.Sequence
	Carrier  _dafny.Sequence
	Status   _dafny.Int
	AgeNanos _dafny.Int
	Points   _dafny.Int
}

func (OrderState_OrderState) isOrderState() {}

func (CompanionStruct_OrderState_) Create_OrderState_(Customer _dafny.Sequence, Worker _dafny.Sequence, Carrier _dafny.Sequence, Status _dafny.Int, AgeNanos _dafny.Int, Points _dafny.Int) OrderState {
	return OrderState{OrderState_OrderState{Customer, Worker, Carrier, Status, AgeNanos, Points}}
}

func (_this OrderState) Is_OrderState() bool {
	_, ok := _this.Get_().(OrderState_OrderState)
	return ok
}

func (CompanionStruct_OrderState_) Default() OrderState {
	return Companion_OrderState_.Create_OrderState_(_dafny.EmptySeq, _dafny.EmptySeq, _dafny.EmptySeq, _dafny.Zero, _dafny.Zero, _dafny.Zero)
}

func (_this OrderState) Dtor_customer() _dafny.Sequence {
	return _this.Get_().(OrderState_OrderState).Customer
}

func (_this OrderState) Dtor_worker() _dafny.Sequence {
	return _this.Get_().(OrderState_OrderState).Worker
}

func (_this OrderState) Dtor_carrier() _dafny.Sequence {
	return _this.Get_().(OrderState_OrderState).Carrier
}

func (_this OrderState) Dtor_status() _dafny.Int {
	return _this.Get_().(OrderState_OrderState).Status
}

func (_this OrderState) Dtor_ageNanos() _dafny.Int {
	return _this.Get_().(OrderState_OrderState).AgeNanos
}

func (_this OrderState) Dtor_points() _dafny.Int {
	return _this.Get_().(OrderState_OrderState).Points
}

func (_this OrderState) String() string {
	switch data := _this.Get_().(type) {
	case nil:
		return "null"
	case OrderState_OrderState:
		{
			return "CommandModel.OrderState.OrderState" + "(" + data.Customer.VerbatimString(true) + ", " + data.Worker.VerbatimString(true) + ", " + data.Carrier.VerbatimString(true) + ", " + _dafny.String(data.Status) + ", " + _dafny.String(data.AgeNanos) + ", " + _dafny.String(data.Points) + ")"
		}
	default:
		{
			return "<unexpected>"
		}
	}
}

func (_this OrderState) Equals(other OrderState) bool {
	switch data1 := _this.Get_().(type) {
	case OrderState_OrderState:
		{
			data2, ok := other.Get_().(OrderState_OrderState)
			return ok && data1.Customer.Equals(data2.Customer) && data1.Worker.Equals(data2.Worker) && data1.Carrier.Equals(data2.Carrier) && data1.Status.Cmp(data2.Status) == 0 && data1.AgeNanos.Cmp(data2.AgeNanos) == 0 && data1.Points.Cmp(data2.Points) == 0
		}
	default:
		{
			return false // unexpected
		}
	}
}

func (_this OrderState) EqualsGeneric(other interface{}) bool {
	typed, ok := other.(OrderState)
	return ok && _this.Equals(typed)
}

func Type_OrderState_() _dafny.TypeDescriptor {
	return type_OrderState_{}
}

type type_OrderState_ struct {
}

func (_this type_OrderState_) Default() interface{} {
	return Companion_OrderState_.Default()
}

func (_this type_OrderState_) String() string {
	return "CommandModel.OrderState"
}
func (_this OrderState) ParentTraits_() []*_dafny.TraitID {
	return [](*_dafny.TraitID){}
}

var _ _dafny.TraitOffspring = OrderState{}

// End of datatype OrderState

// Definition of datatype OrderDecision
type OrderDecision struct {
	Data_OrderDecision_
}

func (_this OrderDecision) Get_() Data_OrderDecision_ {
	return _this.Data_OrderDecision_
}

type Data_OrderDecision_ interface {
	isOrderDecision()
}

type CompanionStruct_OrderDecision_ struct {
}

var Companion_OrderDecision_ = CompanionStruct_OrderDecision_{}

type OrderDecision_Move struct {
	Target _dafny.Int
}

func (OrderDecision_Move) isOrderDecision() {}

func (CompanionStruct_OrderDecision_) Create_Move_(Target _dafny.Int) OrderDecision {
	return OrderDecision{OrderDecision_Move{Target}}
}

func (_this OrderDecision) Is_Move() bool {
	_, ok := _this.Get_().(OrderDecision_Move)
	return ok
}

type OrderDecision_NoChange struct {
}

func (OrderDecision_NoChange) isOrderDecision() {}

func (CompanionStruct_OrderDecision_) Create_NoChange_() OrderDecision {
	return OrderDecision{OrderDecision_NoChange{}}
}

func (_this OrderDecision) Is_NoChange() bool {
	_, ok := _this.Get_().(OrderDecision_NoChange)
	return ok
}

type OrderDecision_Denied struct {
	Message _dafny.Sequence
}

func (OrderDecision_Denied) isOrderDecision() {}

func (CompanionStruct_OrderDecision_) Create_Denied_(Message _dafny.Sequence) OrderDecision {
	return OrderDecision{OrderDecision_Denied{Message}}
}

func (_this OrderDecision) Is_Denied() bool {
	_, ok := _this.Get_().(OrderDecision_Denied)
	return ok
}

func (CompanionStruct_OrderDecision_) Default() OrderDecision {
	return Companion_OrderDecision_.Create_Move_(_dafny.Zero)
}

func (_this OrderDecision) Dtor_target() _dafny.Int {
	return _this.Get_().(OrderDecision_Move).Target
}

func (_this OrderDecision) Dtor_message() _dafny.Sequence {
	return _this.Get_().(OrderDecision_Denied).Message
}

func (_this OrderDecision) String() string {
	switch data := _this.Get_().(type) {
	case nil:
		return "null"
	case OrderDecision_Move:
		{
			return "CommandModel.OrderDecision.Move" + "(" + _dafny.String(data.Target) + ")"
		}
	case OrderDecision_NoChange:
		{
			return "CommandModel.OrderDecision.NoChange"
		}
	case OrderDecision_Denied:
		{
			return "CommandModel.OrderDecision.Denied" + "(" + data.Message.VerbatimString(true) + ")"
		}
	default:
		{
			return "<unexpected>"
		}
	}
}

func (_this OrderDecision) Equals(other OrderDecision) bool {
	switch data1 := _this.Get_().(type) {
	case OrderDecision_Move:
		{
			data2, ok := other.Get_().(OrderDecision_Move)
			return ok && data1.Target.Cmp(data2.Target) == 0
		}
	case OrderDecision_NoChange:
		{
			_, ok := other.Get_().(OrderDecision_NoChange)
			return ok
		}
	case OrderDecision_Denied:
		{
			data2, ok := other.Get_().(OrderDecision_Denied)
			return ok && data1.Message.Equals(data2.Message)
		}
	default:
		{
			return false // unexpected
		}
	}
}

func (_this OrderDecision) EqualsGeneric(other interface{}) bool {
	typed, ok := other.(OrderDecision)
	return ok && _this.Equals(typed)
}

func Type_OrderDecision_() _dafny.TypeDescriptor {
	return type_OrderDecision_{}
}

type type_OrderDecision_ struct {
}

func (_this type_OrderDecision_) Default() interface{} {
	return Companion_OrderDecision_.Default()
}

func (_this type_OrderDecision_) String() string {
	return "CommandModel.OrderDecision"
}
func (_this OrderDecision) ParentTraits_() []*_dafny.TraitID {
	return [](*_dafny.TraitID){}
}

var _ _dafny.TraitOffspring = OrderDecision{}

// End of datatype OrderDecision

// Definition of datatype Coupon
type Coupon struct {
	Data_Coupon_
}

func (_this Coupon) Get_() Data_Coupon_ {
	return _this.Data_Coupon_
}

type Data_Coupon_ interface {
	isCoupon()
}

type CompanionStruct_Coupon_ struct {
}

var Companion_Coupon_ = CompanionStruct_Coupon_{}

type Coupon_NoCoupon struct {
}

func (Coupon_NoCoupon) isCoupon() {}

func (CompanionStruct_Coupon_) Create_NoCoupon_() Coupon {
	return Coupon{Coupon_NoCoupon{}}
}

func (_this Coupon) Is_NoCoupon() bool {
	_, ok := _this.Get_().(Coupon_NoCoupon)
	return ok
}

type Coupon_UnknownCoupon struct {
}

func (Coupon_UnknownCoupon) isCoupon() {}

func (CompanionStruct_Coupon_) Create_UnknownCoupon_() Coupon {
	return Coupon{Coupon_UnknownCoupon{}}
}

func (_this Coupon) Is_UnknownCoupon() bool {
	_, ok := _this.Get_().(Coupon_UnknownCoupon)
	return ok
}

type Coupon_CouponRate struct {
	Percent _dafny.Int
}

func (Coupon_CouponRate) isCoupon() {}

func (CompanionStruct_Coupon_) Create_CouponRate_(Percent _dafny.Int) Coupon {
	return Coupon{Coupon_CouponRate{Percent}}
}

func (_this Coupon) Is_CouponRate() bool {
	_, ok := _this.Get_().(Coupon_CouponRate)
	return ok
}

func (CompanionStruct_Coupon_) Default() Coupon {
	return Companion_Coupon_.Create_NoCoupon_()
}

func (_this Coupon) Dtor_percent() _dafny.Int {
	return _this.Get_().(Coupon_CouponRate).Percent
}

func (_this Coupon) String() string {
	switch data := _this.Get_().(type) {
	case nil:
		return "null"
	case Coupon_NoCoupon:
		{
			return "CommandModel.Coupon.NoCoupon"
		}
	case Coupon_UnknownCoupon:
		{
			return "CommandModel.Coupon.UnknownCoupon"
		}
	case Coupon_CouponRate:
		{
			return "CommandModel.Coupon.CouponRate" + "(" + _dafny.String(data.Percent) + ")"
		}
	default:
		{
			return "<unexpected>"
		}
	}
}

func (_this Coupon) Equals(other Coupon) bool {
	switch data1 := _this.Get_().(type) {
	case Coupon_NoCoupon:
		{
			_, ok := other.Get_().(Coupon_NoCoupon)
			return ok
		}
	case Coupon_UnknownCoupon:
		{
			_, ok := other.Get_().(Coupon_UnknownCoupon)
			return ok
		}
	case Coupon_CouponRate:
		{
			data2, ok := other.Get_().(Coupon_CouponRate)
			return ok && data1.Percent.Cmp(data2.Percent) == 0
		}
	default:
		{
			return false // unexpected
		}
	}
}

func (_this Coupon) EqualsGeneric(other interface{}) bool {
	typed, ok := other.(Coupon)
	return ok && _this.Equals(typed)
}

func Type_Coupon_() _dafny.TypeDescriptor {
	return type_Coupon_{}
}

type type_Coupon_ struct {
}

func (_this type_Coupon_) Default() interface{} {
	return Companion_Coupon_.Default()
}

func (_this type_Coupon_) String() string {
	return "CommandModel.Coupon"
}
func (_this Coupon) ParentTraits_() []*_dafny.TraitID {
	return [](*_dafny.TraitID){}
}

var _ _dafny.TraitOffspring = Coupon{}

// End of datatype Coupon

// Definition of datatype RequestedDay
type RequestedDay struct {
	Data_RequestedDay_
}

func (_this RequestedDay) Get_() Data_RequestedDay_ {
	return _this.Data_RequestedDay_
}

type Data_RequestedDay_ interface {
	isRequestedDay()
}

type CompanionStruct_RequestedDay_ struct {
}

var Companion_RequestedDay_ = CompanionStruct_RequestedDay_{}

type RequestedDay_Unspecified struct {
}

func (RequestedDay_Unspecified) isRequestedDay() {}

func (CompanionStruct_RequestedDay_) Create_Unspecified_() RequestedDay {
	return RequestedDay{RequestedDay_Unspecified{}}
}

func (_this RequestedDay) Is_Unspecified() bool {
	_, ok := _this.Get_().(RequestedDay_Unspecified)
	return ok
}

type RequestedDay_InvalidDate struct {
}

func (RequestedDay_InvalidDate) isRequestedDay() {}

func (CompanionStruct_RequestedDay_) Create_InvalidDate_() RequestedDay {
	return RequestedDay{RequestedDay_InvalidDate{}}
}

func (_this RequestedDay) Is_InvalidDate() bool {
	_, ok := _this.Get_().(RequestedDay_InvalidDate)
	return ok
}

type RequestedDay_CalendarDay struct {
	Date _dafny.Int
}

func (RequestedDay_CalendarDay) isRequestedDay() {}

func (CompanionStruct_RequestedDay_) Create_CalendarDay_(Date _dafny.Int) RequestedDay {
	return RequestedDay{RequestedDay_CalendarDay{Date}}
}

func (_this RequestedDay) Is_CalendarDay() bool {
	_, ok := _this.Get_().(RequestedDay_CalendarDay)
	return ok
}

func (CompanionStruct_RequestedDay_) Default() RequestedDay {
	return Companion_RequestedDay_.Create_Unspecified_()
}

func (_this RequestedDay) Dtor_date() _dafny.Int {
	return _this.Get_().(RequestedDay_CalendarDay).Date
}

func (_this RequestedDay) String() string {
	switch data := _this.Get_().(type) {
	case nil:
		return "null"
	case RequestedDay_Unspecified:
		{
			return "CommandModel.RequestedDay.Unspecified"
		}
	case RequestedDay_InvalidDate:
		{
			return "CommandModel.RequestedDay.InvalidDate"
		}
	case RequestedDay_CalendarDay:
		{
			return "CommandModel.RequestedDay.CalendarDay" + "(" + _dafny.String(data.Date) + ")"
		}
	default:
		{
			return "<unexpected>"
		}
	}
}

func (_this RequestedDay) Equals(other RequestedDay) bool {
	switch data1 := _this.Get_().(type) {
	case RequestedDay_Unspecified:
		{
			_, ok := other.Get_().(RequestedDay_Unspecified)
			return ok
		}
	case RequestedDay_InvalidDate:
		{
			_, ok := other.Get_().(RequestedDay_InvalidDate)
			return ok
		}
	case RequestedDay_CalendarDay:
		{
			data2, ok := other.Get_().(RequestedDay_CalendarDay)
			return ok && data1.Date.Cmp(data2.Date) == 0
		}
	default:
		{
			return false // unexpected
		}
	}
}

func (_this RequestedDay) EqualsGeneric(other interface{}) bool {
	typed, ok := other.(RequestedDay)
	return ok && _this.Equals(typed)
}

func Type_RequestedDay_() _dafny.TypeDescriptor {
	return type_RequestedDay_{}
}

type type_RequestedDay_ struct {
}

func (_this type_RequestedDay_) Default() interface{} {
	return Companion_RequestedDay_.Default()
}

func (_this type_RequestedDay_) String() string {
	return "CommandModel.RequestedDay"
}
func (_this RequestedDay) ParentTraits_() []*_dafny.TraitID {
	return [](*_dafny.TraitID){}
}

var _ _dafny.TraitOffspring = RequestedDay{}

// End of datatype RequestedDay

// Definition of datatype CartState
type CartState struct {
	Data_CartState_
}

func (_this CartState) Get_() Data_CartState_ {
	return _this.Data_CartState_
}

type Data_CartState_ interface {
	isCartState()
}

type CompanionStruct_CartState_ struct {
}

var Companion_CartState_ = CompanionStruct_CartState_{}

type CartState_CartState struct {
	CustomerId   _dafny.Sequence
	Version      _dafny.Int
	Items        _dafny.Map
	Expires      _dafny.Int
	RequestedDay RequestedDay
	Today        _dafny.Int
	Coupon       Coupon
}

func (CartState_CartState) isCartState() {}

func (CompanionStruct_CartState_) Create_CartState_(CustomerId _dafny.Sequence, Version _dafny.Int, Items _dafny.Map, Expires _dafny.Int, RequestedDay RequestedDay, Today _dafny.Int, Coupon Coupon) CartState {
	return CartState{CartState_CartState{CustomerId, Version, Items, Expires, RequestedDay, Today, Coupon}}
}

func (_this CartState) Is_CartState() bool {
	_, ok := _this.Get_().(CartState_CartState)
	return ok
}

func (CompanionStruct_CartState_) Default() CartState {
	return Companion_CartState_.Create_CartState_(_dafny.EmptySeq, _dafny.Zero, _dafny.EmptyMap, _dafny.Zero, Companion_RequestedDay_.Default(), _dafny.Zero, Companion_Coupon_.Default())
}

func (_this CartState) Dtor_customerId() _dafny.Sequence {
	return _this.Get_().(CartState_CartState).CustomerId
}

func (_this CartState) Dtor_version() _dafny.Int {
	return _this.Get_().(CartState_CartState).Version
}

func (_this CartState) Dtor_items() _dafny.Map {
	return _this.Get_().(CartState_CartState).Items
}

func (_this CartState) Dtor_expires() _dafny.Int {
	return _this.Get_().(CartState_CartState).Expires
}

func (_this CartState) Dtor_requestedDay() RequestedDay {
	return _this.Get_().(CartState_CartState).RequestedDay
}

func (_this CartState) Dtor_today() _dafny.Int {
	return _this.Get_().(CartState_CartState).Today
}

func (_this CartState) Dtor_coupon() Coupon {
	return _this.Get_().(CartState_CartState).Coupon
}

func (_this CartState) String() string {
	switch data := _this.Get_().(type) {
	case nil:
		return "null"
	case CartState_CartState:
		{
			return "CommandModel.CartState.CartState" + "(" + data.CustomerId.VerbatimString(true) + ", " + _dafny.String(data.Version) + ", " + _dafny.String(data.Items) + ", " + _dafny.String(data.Expires) + ", " + _dafny.String(data.RequestedDay) + ", " + _dafny.String(data.Today) + ", " + _dafny.String(data.Coupon) + ")"
		}
	default:
		{
			return "<unexpected>"
		}
	}
}

func (_this CartState) Equals(other CartState) bool {
	switch data1 := _this.Get_().(type) {
	case CartState_CartState:
		{
			data2, ok := other.Get_().(CartState_CartState)
			return ok && data1.CustomerId.Equals(data2.CustomerId) && data1.Version.Cmp(data2.Version) == 0 && data1.Items.Equals(data2.Items) && data1.Expires.Cmp(data2.Expires) == 0 && data1.RequestedDay.Equals(data2.RequestedDay) && data1.Today.Cmp(data2.Today) == 0 && data1.Coupon.Equals(data2.Coupon)
		}
	default:
		{
			return false // unexpected
		}
	}
}

func (_this CartState) EqualsGeneric(other interface{}) bool {
	typed, ok := other.(CartState)
	return ok && _this.Equals(typed)
}

func Type_CartState_() _dafny.TypeDescriptor {
	return type_CartState_{}
}

type type_CartState_ struct {
}

func (_this type_CartState_) Default() interface{} {
	return Companion_CartState_.Default()
}

func (_this type_CartState_) String() string {
	return "CommandModel.CartState"
}
func (_this CartState) ParentTraits_() []*_dafny.TraitID {
	return [](*_dafny.TraitID){}
}

var _ _dafny.TraitOffspring = CartState{}

// End of datatype CartState

// Definition of datatype ReadRequest
type ReadRequest struct {
	Data_ReadRequest_
}

func (_this ReadRequest) Get_() Data_ReadRequest_ {
	return _this.Data_ReadRequest_
}

type Data_ReadRequest_ interface {
	isReadRequest()
}

type CompanionStruct_ReadRequest_ struct {
}

var Companion_ReadRequest_ = CompanionStruct_ReadRequest_{}

type ReadRequest_RejectRead struct {
	Reason _dafny.Sequence
}

func (ReadRequest_RejectRead) isReadRequest() {}

func (CompanionStruct_ReadRequest_) Create_RejectRead_(Reason _dafny.Sequence) ReadRequest {
	return ReadRequest{ReadRequest_RejectRead{Reason}}
}

func (_this ReadRequest) Is_RejectRead() bool {
	_, ok := _this.Get_().(ReadRequest_RejectRead)
	return ok
}

type ReadRequest_ReadCart struct {
	CustomerId _dafny.Sequence
}

func (ReadRequest_ReadCart) isReadRequest() {}

func (CompanionStruct_ReadRequest_) Create_ReadCart_(CustomerId _dafny.Sequence) ReadRequest {
	return ReadRequest{ReadRequest_ReadCart{CustomerId}}
}

func (_this ReadRequest) Is_ReadCart() bool {
	_, ok := _this.Get_().(ReadRequest_ReadCart)
	return ok
}

type ReadRequest_ReadStock struct {
	ProductId _dafny.Sequence
}

func (ReadRequest_ReadStock) isReadRequest() {}

func (CompanionStruct_ReadRequest_) Create_ReadStock_(ProductId _dafny.Sequence) ReadRequest {
	return ReadRequest{ReadRequest_ReadStock{ProductId}}
}

func (_this ReadRequest) Is_ReadStock() bool {
	_, ok := _this.Get_().(ReadRequest_ReadStock)
	return ok
}

type ReadRequest_ReadOrder struct {
	OrderId _dafny.Sequence
}

func (ReadRequest_ReadOrder) isReadRequest() {}

func (CompanionStruct_ReadRequest_) Create_ReadOrder_(OrderId _dafny.Sequence) ReadRequest {
	return ReadRequest{ReadRequest_ReadOrder{OrderId}}
}

func (_this ReadRequest) Is_ReadOrder() bool {
	_, ok := _this.Get_().(ReadRequest_ReadOrder)
	return ok
}

func (CompanionStruct_ReadRequest_) Default() ReadRequest {
	return Companion_ReadRequest_.Create_RejectRead_(_dafny.EmptySeq)
}

func (_this ReadRequest) Dtor_reason() _dafny.Sequence {
	return _this.Get_().(ReadRequest_RejectRead).Reason
}

func (_this ReadRequest) Dtor_customerId() _dafny.Sequence {
	return _this.Get_().(ReadRequest_ReadCart).CustomerId
}

func (_this ReadRequest) Dtor_productId() _dafny.Sequence {
	return _this.Get_().(ReadRequest_ReadStock).ProductId
}

func (_this ReadRequest) Dtor_orderId() _dafny.Sequence {
	return _this.Get_().(ReadRequest_ReadOrder).OrderId
}

func (_this ReadRequest) String() string {
	switch data := _this.Get_().(type) {
	case nil:
		return "null"
	case ReadRequest_RejectRead:
		{
			return "CommandModel.ReadRequest.RejectRead" + "(" + data.Reason.VerbatimString(true) + ")"
		}
	case ReadRequest_ReadCart:
		{
			return "CommandModel.ReadRequest.ReadCart" + "(" + data.CustomerId.VerbatimString(true) + ")"
		}
	case ReadRequest_ReadStock:
		{
			return "CommandModel.ReadRequest.ReadStock" + "(" + data.ProductId.VerbatimString(true) + ")"
		}
	case ReadRequest_ReadOrder:
		{
			return "CommandModel.ReadRequest.ReadOrder" + "(" + data.OrderId.VerbatimString(true) + ")"
		}
	default:
		{
			return "<unexpected>"
		}
	}
}

func (_this ReadRequest) Equals(other ReadRequest) bool {
	switch data1 := _this.Get_().(type) {
	case ReadRequest_RejectRead:
		{
			data2, ok := other.Get_().(ReadRequest_RejectRead)
			return ok && data1.Reason.Equals(data2.Reason)
		}
	case ReadRequest_ReadCart:
		{
			data2, ok := other.Get_().(ReadRequest_ReadCart)
			return ok && data1.CustomerId.Equals(data2.CustomerId)
		}
	case ReadRequest_ReadStock:
		{
			data2, ok := other.Get_().(ReadRequest_ReadStock)
			return ok && data1.ProductId.Equals(data2.ProductId)
		}
	case ReadRequest_ReadOrder:
		{
			data2, ok := other.Get_().(ReadRequest_ReadOrder)
			return ok && data1.OrderId.Equals(data2.OrderId)
		}
	default:
		{
			return false // unexpected
		}
	}
}

func (_this ReadRequest) EqualsGeneric(other interface{}) bool {
	typed, ok := other.(ReadRequest)
	return ok && _this.Equals(typed)
}

func Type_ReadRequest_() _dafny.TypeDescriptor {
	return type_ReadRequest_{}
}

type type_ReadRequest_ struct {
}

func (_this type_ReadRequest_) Default() interface{} {
	return Companion_ReadRequest_.Default()
}

func (_this type_ReadRequest_) String() string {
	return "CommandModel.ReadRequest"
}
func (_this ReadRequest) ParentTraits_() []*_dafny.TraitID {
	return [](*_dafny.TraitID){}
}

var _ _dafny.TraitOffspring = ReadRequest{}

// End of datatype ReadRequest

// Definition of datatype Snapshot
type Snapshot struct {
	Data_Snapshot_
}

func (_this Snapshot) Get_() Data_Snapshot_ {
	return _this.Data_Snapshot_
}

type Data_Snapshot_ interface {
	isSnapshot()
}

type CompanionStruct_Snapshot_ struct {
}

var Companion_Snapshot_ = CompanionStruct_Snapshot_{}

type Snapshot_Missing struct {
	Message _dafny.Sequence
}

func (Snapshot_Missing) isSnapshot() {}

func (CompanionStruct_Snapshot_) Create_Missing_(Message _dafny.Sequence) Snapshot {
	return Snapshot{Snapshot_Missing{Message}}
}

func (_this Snapshot) Is_Missing() bool {
	_, ok := _this.Get_().(Snapshot_Missing)
	return ok
}

type Snapshot_Cart struct {
	State CartState
}

func (Snapshot_Cart) isSnapshot() {}

func (CompanionStruct_Snapshot_) Create_Cart_(State CartState) Snapshot {
	return Snapshot{Snapshot_Cart{State}}
}

func (_this Snapshot) Is_Cart() bool {
	_, ok := _this.Get_().(Snapshot_Cart)
	return ok
}

type Snapshot_Stock struct {
	StockVersion _dafny.Int
}

func (Snapshot_Stock) isSnapshot() {}

func (CompanionStruct_Snapshot_) Create_Stock_(StockVersion _dafny.Int) Snapshot {
	return Snapshot{Snapshot_Stock{StockVersion}}
}

func (_this Snapshot) Is_Stock() bool {
	_, ok := _this.Get_().(Snapshot_Stock)
	return ok
}

type Snapshot_Order struct {
	OrderVersion _dafny.Int
	Order        OrderState
	Lines        _dafny.Map
}

func (Snapshot_Order) isSnapshot() {}

func (CompanionStruct_Snapshot_) Create_Order_(OrderVersion _dafny.Int, Order OrderState, Lines _dafny.Map) Snapshot {
	return Snapshot{Snapshot_Order{OrderVersion, Order, Lines}}
}

func (_this Snapshot) Is_Order() bool {
	_, ok := _this.Get_().(Snapshot_Order)
	return ok
}

func (CompanionStruct_Snapshot_) Default() Snapshot {
	return Companion_Snapshot_.Create_Missing_(_dafny.EmptySeq)
}

func (_this Snapshot) Dtor_message() _dafny.Sequence {
	return _this.Get_().(Snapshot_Missing).Message
}

func (_this Snapshot) Dtor_state() CartState {
	return _this.Get_().(Snapshot_Cart).State
}

func (_this Snapshot) Dtor_stockVersion() _dafny.Int {
	return _this.Get_().(Snapshot_Stock).StockVersion
}

func (_this Snapshot) Dtor_orderVersion() _dafny.Int {
	return _this.Get_().(Snapshot_Order).OrderVersion
}

func (_this Snapshot) Dtor_order() OrderState {
	return _this.Get_().(Snapshot_Order).Order
}

func (_this Snapshot) Dtor_lines() _dafny.Map {
	return _this.Get_().(Snapshot_Order).Lines
}

func (_this Snapshot) String() string {
	switch data := _this.Get_().(type) {
	case nil:
		return "null"
	case Snapshot_Missing:
		{
			return "CommandModel.Snapshot.Missing" + "(" + data.Message.VerbatimString(true) + ")"
		}
	case Snapshot_Cart:
		{
			return "CommandModel.Snapshot.Cart" + "(" + _dafny.String(data.State) + ")"
		}
	case Snapshot_Stock:
		{
			return "CommandModel.Snapshot.Stock" + "(" + _dafny.String(data.StockVersion) + ")"
		}
	case Snapshot_Order:
		{
			return "CommandModel.Snapshot.Order" + "(" + _dafny.String(data.OrderVersion) + ", " + _dafny.String(data.Order) + ", " + _dafny.String(data.Lines) + ")"
		}
	default:
		{
			return "<unexpected>"
		}
	}
}

func (_this Snapshot) Equals(other Snapshot) bool {
	switch data1 := _this.Get_().(type) {
	case Snapshot_Missing:
		{
			data2, ok := other.Get_().(Snapshot_Missing)
			return ok && data1.Message.Equals(data2.Message)
		}
	case Snapshot_Cart:
		{
			data2, ok := other.Get_().(Snapshot_Cart)
			return ok && data1.State.Equals(data2.State)
		}
	case Snapshot_Stock:
		{
			data2, ok := other.Get_().(Snapshot_Stock)
			return ok && data1.StockVersion.Cmp(data2.StockVersion) == 0
		}
	case Snapshot_Order:
		{
			data2, ok := other.Get_().(Snapshot_Order)
			return ok && data1.OrderVersion.Cmp(data2.OrderVersion) == 0 && data1.Order.Equals(data2.Order) && data1.Lines.Equals(data2.Lines)
		}
	default:
		{
			return false // unexpected
		}
	}
}

func (_this Snapshot) EqualsGeneric(other interface{}) bool {
	typed, ok := other.(Snapshot)
	return ok && _this.Equals(typed)
}

func Type_Snapshot_() _dafny.TypeDescriptor {
	return type_Snapshot_{}
}

type type_Snapshot_ struct {
}

func (_this type_Snapshot_) Default() interface{} {
	return Companion_Snapshot_.Default()
}

func (_this type_Snapshot_) String() string {
	return "CommandModel.Snapshot"
}
func (_this Snapshot) ParentTraits_() []*_dafny.TraitID {
	return [](*_dafny.TraitID){}
}

var _ _dafny.TraitOffspring = Snapshot{}

// End of datatype Snapshot

// Definition of datatype OrderPreparation
type OrderPreparation struct {
	Data_OrderPreparation_
}

func (_this OrderPreparation) Get_() Data_OrderPreparation_ {
	return _this.Data_OrderPreparation_
}

type Data_OrderPreparation_ interface {
	isOrderPreparation()
}

type CompanionStruct_OrderPreparation_ struct {
}

var Companion_OrderPreparation_ = CompanionStruct_OrderPreparation_{}

type OrderPreparation_Keep struct {
}

func (OrderPreparation_Keep) isOrderPreparation() {}

func (CompanionStruct_OrderPreparation_) Create_Keep_() OrderPreparation {
	return OrderPreparation{OrderPreparation_Keep{}}
}

func (_this OrderPreparation) Is_Keep() bool {
	_, ok := _this.Get_().(OrderPreparation_Keep)
	return ok
}

type OrderPreparation_AssignWorker struct {
	WorkerId _dafny.Sequence
}

func (OrderPreparation_AssignWorker) isOrderPreparation() {}

func (CompanionStruct_OrderPreparation_) Create_AssignWorker_(WorkerId _dafny.Sequence) OrderPreparation {
	return OrderPreparation{OrderPreparation_AssignWorker{WorkerId}}
}

func (_this OrderPreparation) Is_AssignWorker() bool {
	_, ok := _this.Get_().(OrderPreparation_AssignWorker)
	return ok
}

type OrderPreparation_Pack struct {
	BoxCount    _dafny.Int
	WeightGrams _dafny.Int
}

func (OrderPreparation_Pack) isOrderPreparation() {}

func (CompanionStruct_OrderPreparation_) Create_Pack_(BoxCount _dafny.Int, WeightGrams _dafny.Int) OrderPreparation {
	return OrderPreparation{OrderPreparation_Pack{BoxCount, WeightGrams}}
}

func (_this OrderPreparation) Is_Pack() bool {
	_, ok := _this.Get_().(OrderPreparation_Pack)
	return ok
}

type OrderPreparation_AssignCarrier struct {
	CarrierId      _dafny.Sequence
	TrackingNumber _dafny.Sequence
}

func (OrderPreparation_AssignCarrier) isOrderPreparation() {}

func (CompanionStruct_OrderPreparation_) Create_AssignCarrier_(CarrierId _dafny.Sequence, TrackingNumber _dafny.Sequence) OrderPreparation {
	return OrderPreparation{OrderPreparation_AssignCarrier{CarrierId, TrackingNumber}}
}

func (_this OrderPreparation) Is_AssignCarrier() bool {
	_, ok := _this.Get_().(OrderPreparation_AssignCarrier)
	return ok
}

func (CompanionStruct_OrderPreparation_) Default() OrderPreparation {
	return Companion_OrderPreparation_.Create_Keep_()
}

func (_this OrderPreparation) Dtor_workerId() _dafny.Sequence {
	return _this.Get_().(OrderPreparation_AssignWorker).WorkerId
}

func (_this OrderPreparation) Dtor_boxCount() _dafny.Int {
	return _this.Get_().(OrderPreparation_Pack).BoxCount
}

func (_this OrderPreparation) Dtor_weightGrams() _dafny.Int {
	return _this.Get_().(OrderPreparation_Pack).WeightGrams
}

func (_this OrderPreparation) Dtor_carrierId() _dafny.Sequence {
	return _this.Get_().(OrderPreparation_AssignCarrier).CarrierId
}

func (_this OrderPreparation) Dtor_trackingNumber() _dafny.Sequence {
	return _this.Get_().(OrderPreparation_AssignCarrier).TrackingNumber
}

func (_this OrderPreparation) String() string {
	switch data := _this.Get_().(type) {
	case nil:
		return "null"
	case OrderPreparation_Keep:
		{
			return "CommandModel.OrderPreparation.Keep"
		}
	case OrderPreparation_AssignWorker:
		{
			return "CommandModel.OrderPreparation.AssignWorker" + "(" + data.WorkerId.VerbatimString(true) + ")"
		}
	case OrderPreparation_Pack:
		{
			return "CommandModel.OrderPreparation.Pack" + "(" + _dafny.String(data.BoxCount) + ", " + _dafny.String(data.WeightGrams) + ")"
		}
	case OrderPreparation_AssignCarrier:
		{
			return "CommandModel.OrderPreparation.AssignCarrier" + "(" + data.CarrierId.VerbatimString(true) + ", " + data.TrackingNumber.VerbatimString(true) + ")"
		}
	default:
		{
			return "<unexpected>"
		}
	}
}

func (_this OrderPreparation) Equals(other OrderPreparation) bool {
	switch data1 := _this.Get_().(type) {
	case OrderPreparation_Keep:
		{
			_, ok := other.Get_().(OrderPreparation_Keep)
			return ok
		}
	case OrderPreparation_AssignWorker:
		{
			data2, ok := other.Get_().(OrderPreparation_AssignWorker)
			return ok && data1.WorkerId.Equals(data2.WorkerId)
		}
	case OrderPreparation_Pack:
		{
			data2, ok := other.Get_().(OrderPreparation_Pack)
			return ok && data1.BoxCount.Cmp(data2.BoxCount) == 0 && data1.WeightGrams.Cmp(data2.WeightGrams) == 0
		}
	case OrderPreparation_AssignCarrier:
		{
			data2, ok := other.Get_().(OrderPreparation_AssignCarrier)
			return ok && data1.CarrierId.Equals(data2.CarrierId) && data1.TrackingNumber.Equals(data2.TrackingNumber)
		}
	default:
		{
			return false // unexpected
		}
	}
}

func (_this OrderPreparation) EqualsGeneric(other interface{}) bool {
	typed, ok := other.(OrderPreparation)
	return ok && _this.Equals(typed)
}

func Type_OrderPreparation_() _dafny.TypeDescriptor {
	return type_OrderPreparation_{}
}

type type_OrderPreparation_ struct {
}

func (_this type_OrderPreparation_) Default() interface{} {
	return Companion_OrderPreparation_.Default()
}

func (_this type_OrderPreparation_) String() string {
	return "CommandModel.OrderPreparation"
}
func (_this OrderPreparation) ParentTraits_() []*_dafny.TraitID {
	return [](*_dafny.TraitID){}
}

var _ _dafny.TraitOffspring = OrderPreparation{}

// End of datatype OrderPreparation

// Definition of datatype StockChanges
type StockChanges struct {
	Data_StockChanges_
}

func (_this StockChanges) Get_() Data_StockChanges_ {
	return _this.Data_StockChanges_
}

type Data_StockChanges_ interface {
	isStockChanges()
}

type CompanionStruct_StockChanges_ struct {
}

var Companion_StockChanges_ = CompanionStruct_StockChanges_{}

type StockChanges_Reserve struct {
	Deltas _dafny.Map
}

func (StockChanges_Reserve) isStockChanges() {}

func (CompanionStruct_StockChanges_) Create_Reserve_(Deltas _dafny.Map) StockChanges {
	return StockChanges{StockChanges_Reserve{Deltas}}
}

func (_this StockChanges) Is_Reserve() bool {
	_, ok := _this.Get_().(StockChanges_Reserve)
	return ok
}

type StockChanges_Ship struct {
	Quantities _dafny.Map
}

func (StockChanges_Ship) isStockChanges() {}

func (CompanionStruct_StockChanges_) Create_Ship_(Quantities _dafny.Map) StockChanges {
	return StockChanges{StockChanges_Ship{Quantities}}
}

func (_this StockChanges) Is_Ship() bool {
	_, ok := _this.Get_().(StockChanges_Ship)
	return ok
}

func (CompanionStruct_StockChanges_) Default() StockChanges {
	return Companion_StockChanges_.Create_Reserve_(_dafny.EmptyMap)
}

func (_this StockChanges) Dtor_deltas() _dafny.Map {
	return _this.Get_().(StockChanges_Reserve).Deltas
}

func (_this StockChanges) Dtor_quantities() _dafny.Map {
	return _this.Get_().(StockChanges_Ship).Quantities
}

func (_this StockChanges) String() string {
	switch data := _this.Get_().(type) {
	case nil:
		return "null"
	case StockChanges_Reserve:
		{
			return "CommandModel.StockChanges.Reserve" + "(" + _dafny.String(data.Deltas) + ")"
		}
	case StockChanges_Ship:
		{
			return "CommandModel.StockChanges.Ship" + "(" + _dafny.String(data.Quantities) + ")"
		}
	default:
		{
			return "<unexpected>"
		}
	}
}

func (_this StockChanges) Equals(other StockChanges) bool {
	switch data1 := _this.Get_().(type) {
	case StockChanges_Reserve:
		{
			data2, ok := other.Get_().(StockChanges_Reserve)
			return ok && data1.Deltas.Equals(data2.Deltas)
		}
	case StockChanges_Ship:
		{
			data2, ok := other.Get_().(StockChanges_Ship)
			return ok && data1.Quantities.Equals(data2.Quantities)
		}
	default:
		{
			return false // unexpected
		}
	}
}

func (_this StockChanges) EqualsGeneric(other interface{}) bool {
	typed, ok := other.(StockChanges)
	return ok && _this.Equals(typed)
}

func Type_StockChanges_() _dafny.TypeDescriptor {
	return type_StockChanges_{}
}

type type_StockChanges_ struct {
}

func (_this type_StockChanges_) Default() interface{} {
	return Companion_StockChanges_.Default()
}

func (_this type_StockChanges_) String() string {
	return "CommandModel.StockChanges"
}
func (_this StockChanges) ParentTraits_() []*_dafny.TraitID {
	return [](*_dafny.TraitID){}
}

var _ _dafny.TraitOffspring = StockChanges{}

// End of datatype StockChanges

// Definition of datatype DeliveryObligation
type DeliveryObligation struct {
	Data_DeliveryObligation_
}

func (_this DeliveryObligation) Get_() Data_DeliveryObligation_ {
	return _this.Data_DeliveryObligation_
}

type Data_DeliveryObligation_ interface {
	isDeliveryObligation()
}

type CompanionStruct_DeliveryObligation_ struct {
}

var Companion_DeliveryObligation_ = CompanionStruct_DeliveryObligation_{}

type DeliveryObligation_AwardPoints struct {
	Points _dafny.Int
}

func (DeliveryObligation_AwardPoints) isDeliveryObligation() {}

func (CompanionStruct_DeliveryObligation_) Create_AwardPoints_(Points _dafny.Int) DeliveryObligation {
	return DeliveryObligation{DeliveryObligation_AwardPoints{Points}}
}

func (_this DeliveryObligation) Is_AwardPoints() bool {
	_, ok := _this.Get_().(DeliveryObligation_AwardPoints)
	return ok
}

type DeliveryObligation_CreateInvoice struct {
}

func (DeliveryObligation_CreateInvoice) isDeliveryObligation() {}

func (CompanionStruct_DeliveryObligation_) Create_CreateInvoice_() DeliveryObligation {
	return DeliveryObligation{DeliveryObligation_CreateInvoice{}}
}

func (_this DeliveryObligation) Is_CreateInvoice() bool {
	_, ok := _this.Get_().(DeliveryObligation_CreateInvoice)
	return ok
}

func (CompanionStruct_DeliveryObligation_) Default() DeliveryObligation {
	return Companion_DeliveryObligation_.Create_AwardPoints_(_dafny.Zero)
}

func (_this DeliveryObligation) Dtor_points() _dafny.Int {
	return _this.Get_().(DeliveryObligation_AwardPoints).Points
}

func (_this DeliveryObligation) String() string {
	switch data := _this.Get_().(type) {
	case nil:
		return "null"
	case DeliveryObligation_AwardPoints:
		{
			return "CommandModel.DeliveryObligation.AwardPoints" + "(" + _dafny.String(data.Points) + ")"
		}
	case DeliveryObligation_CreateInvoice:
		{
			return "CommandModel.DeliveryObligation.CreateInvoice"
		}
	default:
		{
			return "<unexpected>"
		}
	}
}

func (_this DeliveryObligation) Equals(other DeliveryObligation) bool {
	switch data1 := _this.Get_().(type) {
	case DeliveryObligation_AwardPoints:
		{
			data2, ok := other.Get_().(DeliveryObligation_AwardPoints)
			return ok && data1.Points.Cmp(data2.Points) == 0
		}
	case DeliveryObligation_CreateInvoice:
		{
			_, ok := other.Get_().(DeliveryObligation_CreateInvoice)
			return ok
		}
	default:
		{
			return false // unexpected
		}
	}
}

func (_this DeliveryObligation) EqualsGeneric(other interface{}) bool {
	typed, ok := other.(DeliveryObligation)
	return ok && _this.Equals(typed)
}

func Type_DeliveryObligation_() _dafny.TypeDescriptor {
	return type_DeliveryObligation_{}
}

type type_DeliveryObligation_ struct {
}

func (_this type_DeliveryObligation_) Default() interface{} {
	return Companion_DeliveryObligation_.Default()
}

func (_this type_DeliveryObligation_) String() string {
	return "CommandModel.DeliveryObligation"
}
func (_this DeliveryObligation) ParentTraits_() []*_dafny.TraitID {
	return [](*_dafny.TraitID){}
}

var _ _dafny.TraitOffspring = DeliveryObligation{}

// End of datatype DeliveryObligation

// Definition of datatype CommandPlan
type CommandPlan struct {
	Data_CommandPlan_
}

func (_this CommandPlan) Get_() Data_CommandPlan_ {
	return _this.Data_CommandPlan_
}

type Data_CommandPlan_ interface {
	isCommandPlan()
}

type CompanionStruct_CommandPlan_ struct {
}

var Companion_CommandPlan_ = CompanionStruct_CommandPlan_{}

type CommandPlan_Reject struct {
	Reason _dafny.Sequence
}

func (CommandPlan_Reject) isCommandPlan() {}

func (CompanionStruct_CommandPlan_) Create_Reject_(Reason _dafny.Sequence) CommandPlan {
	return CommandPlan{CommandPlan_Reject{Reason}}
}

func (_this CommandPlan) Is_Reject() bool {
	_, ok := _this.Get_().(CommandPlan_Reject)
	return ok
}

type CommandPlan_Unchanged struct {
}

func (CommandPlan_Unchanged) isCommandPlan() {}

func (CompanionStruct_CommandPlan_) Create_Unchanged_() CommandPlan {
	return CommandPlan{CommandPlan_Unchanged{}}
}

func (_this CommandPlan) Is_Unchanged() bool {
	_, ok := _this.Get_().(CommandPlan_Unchanged)
	return ok
}

type CommandPlan_UpdateCart struct {
	CustomerId   _dafny.Sequence
	NextItems    _dafny.Map
	Expiry       _dafny.Int
	Reservations _dafny.Map
	Expired      bool
}

func (CommandPlan_UpdateCart) isCommandPlan() {}

func (CompanionStruct_CommandPlan_) Create_UpdateCart_(CustomerId _dafny.Sequence, NextItems _dafny.Map, Expiry _dafny.Int, Reservations _dafny.Map, Expired bool) CommandPlan {
	return CommandPlan{CommandPlan_UpdateCart{CustomerId, NextItems, Expiry, Reservations, Expired}}
}

func (_this CommandPlan) Is_UpdateCart() bool {
	_, ok := _this.Get_().(CommandPlan_UpdateCart)
	return ok
}

type CommandPlan_Checkout struct {
	CheckoutCustomer _dafny.Sequence
	NewOrderId       _dafny.Sequence
	OrderLines       _dafny.Map
	RemainingItems   _dafny.Map
	CartExpiry       _dafny.Int
	Discount         _dafny.Int
}

func (CommandPlan_Checkout) isCommandPlan() {}

func (CompanionStruct_CommandPlan_) Create_Checkout_(CheckoutCustomer _dafny.Sequence, NewOrderId _dafny.Sequence, OrderLines _dafny.Map, RemainingItems _dafny.Map, CartExpiry _dafny.Int, Discount _dafny.Int) CommandPlan {
	return CommandPlan{CommandPlan_Checkout{CheckoutCustomer, NewOrderId, OrderLines, RemainingItems, CartExpiry, Discount}}
}

func (_this CommandPlan) Is_Checkout() bool {
	_, ok := _this.Get_().(CommandPlan_Checkout)
	return ok
}

type CommandPlan_Adjust struct {
	Product _dafny.Sequence
	Delta   _dafny.Int
}

func (CommandPlan_Adjust) isCommandPlan() {}

func (CompanionStruct_CommandPlan_) Create_Adjust_(Product _dafny.Sequence, Delta _dafny.Int) CommandPlan {
	return CommandPlan{CommandPlan_Adjust{Product, Delta}}
}

func (_this CommandPlan) Is_Adjust() bool {
	_, ok := _this.Get_().(CommandPlan_Adjust)
	return ok
}

type CommandPlan_UpdateOrder struct {
	Target        _dafny.Int
	OrderCustomer _dafny.Sequence
	Prepare       OrderPreparation
	Stock         StockChanges
	Obligations   _dafny.Set
}

func (CommandPlan_UpdateOrder) isCommandPlan() {}

func (CompanionStruct_CommandPlan_) Create_UpdateOrder_(Target _dafny.Int, OrderCustomer _dafny.Sequence, Prepare OrderPreparation, Stock StockChanges, Obligations _dafny.Set) CommandPlan {
	return CommandPlan{CommandPlan_UpdateOrder{Target, OrderCustomer, Prepare, Stock, Obligations}}
}

func (_this CommandPlan) Is_UpdateOrder() bool {
	_, ok := _this.Get_().(CommandPlan_UpdateOrder)
	return ok
}

func (CompanionStruct_CommandPlan_) Default() CommandPlan {
	return Companion_CommandPlan_.Create_Reject_(_dafny.EmptySeq)
}

func (_this CommandPlan) Dtor_reason() _dafny.Sequence {
	return _this.Get_().(CommandPlan_Reject).Reason
}

func (_this CommandPlan) Dtor_customerId() _dafny.Sequence {
	return _this.Get_().(CommandPlan_UpdateCart).CustomerId
}

func (_this CommandPlan) Dtor_nextItems() _dafny.Map {
	return _this.Get_().(CommandPlan_UpdateCart).NextItems
}

func (_this CommandPlan) Dtor_expiry() _dafny.Int {
	return _this.Get_().(CommandPlan_UpdateCart).Expiry
}

func (_this CommandPlan) Dtor_reservations() _dafny.Map {
	return _this.Get_().(CommandPlan_UpdateCart).Reservations
}

func (_this CommandPlan) Dtor_expired() bool {
	return _this.Get_().(CommandPlan_UpdateCart).Expired
}

func (_this CommandPlan) Dtor_checkoutCustomer() _dafny.Sequence {
	return _this.Get_().(CommandPlan_Checkout).CheckoutCustomer
}

func (_this CommandPlan) Dtor_newOrderId() _dafny.Sequence {
	return _this.Get_().(CommandPlan_Checkout).NewOrderId
}

func (_this CommandPlan) Dtor_orderLines() _dafny.Map {
	return _this.Get_().(CommandPlan_Checkout).OrderLines
}

func (_this CommandPlan) Dtor_remainingItems() _dafny.Map {
	return _this.Get_().(CommandPlan_Checkout).RemainingItems
}

func (_this CommandPlan) Dtor_cartExpiry() _dafny.Int {
	return _this.Get_().(CommandPlan_Checkout).CartExpiry
}

func (_this CommandPlan) Dtor_discount() _dafny.Int {
	return _this.Get_().(CommandPlan_Checkout).Discount
}

func (_this CommandPlan) Dtor_product() _dafny.Sequence {
	return _this.Get_().(CommandPlan_Adjust).Product
}

func (_this CommandPlan) Dtor_delta() _dafny.Int {
	return _this.Get_().(CommandPlan_Adjust).Delta
}

func (_this CommandPlan) Dtor_target() _dafny.Int {
	return _this.Get_().(CommandPlan_UpdateOrder).Target
}

func (_this CommandPlan) Dtor_orderCustomer() _dafny.Sequence {
	return _this.Get_().(CommandPlan_UpdateOrder).OrderCustomer
}

func (_this CommandPlan) Dtor_prepare() OrderPreparation {
	return _this.Get_().(CommandPlan_UpdateOrder).Prepare
}

func (_this CommandPlan) Dtor_stock() StockChanges {
	return _this.Get_().(CommandPlan_UpdateOrder).Stock
}

func (_this CommandPlan) Dtor_obligations() _dafny.Set {
	return _this.Get_().(CommandPlan_UpdateOrder).Obligations
}

func (_this CommandPlan) String() string {
	switch data := _this.Get_().(type) {
	case nil:
		return "null"
	case CommandPlan_Reject:
		{
			return "CommandModel.CommandPlan.Reject" + "(" + data.Reason.VerbatimString(true) + ")"
		}
	case CommandPlan_Unchanged:
		{
			return "CommandModel.CommandPlan.Unchanged"
		}
	case CommandPlan_UpdateCart:
		{
			return "CommandModel.CommandPlan.UpdateCart" + "(" + data.CustomerId.VerbatimString(true) + ", " + _dafny.String(data.NextItems) + ", " + _dafny.String(data.Expiry) + ", " + _dafny.String(data.Reservations) + ", " + _dafny.String(data.Expired) + ")"
		}
	case CommandPlan_Checkout:
		{
			return "CommandModel.CommandPlan.Checkout" + "(" + data.CheckoutCustomer.VerbatimString(true) + ", " + data.NewOrderId.VerbatimString(true) + ", " + _dafny.String(data.OrderLines) + ", " + _dafny.String(data.RemainingItems) + ", " + _dafny.String(data.CartExpiry) + ", " + _dafny.String(data.Discount) + ")"
		}
	case CommandPlan_Adjust:
		{
			return "CommandModel.CommandPlan.Adjust" + "(" + data.Product.VerbatimString(true) + ", " + _dafny.String(data.Delta) + ")"
		}
	case CommandPlan_UpdateOrder:
		{
			return "CommandModel.CommandPlan.UpdateOrder" + "(" + _dafny.String(data.Target) + ", " + data.OrderCustomer.VerbatimString(true) + ", " + _dafny.String(data.Prepare) + ", " + _dafny.String(data.Stock) + ", " + _dafny.String(data.Obligations) + ")"
		}
	default:
		{
			return "<unexpected>"
		}
	}
}

func (_this CommandPlan) Equals(other CommandPlan) bool {
	switch data1 := _this.Get_().(type) {
	case CommandPlan_Reject:
		{
			data2, ok := other.Get_().(CommandPlan_Reject)
			return ok && data1.Reason.Equals(data2.Reason)
		}
	case CommandPlan_Unchanged:
		{
			_, ok := other.Get_().(CommandPlan_Unchanged)
			return ok
		}
	case CommandPlan_UpdateCart:
		{
			data2, ok := other.Get_().(CommandPlan_UpdateCart)
			return ok && data1.CustomerId.Equals(data2.CustomerId) && data1.NextItems.Equals(data2.NextItems) && data1.Expiry.Cmp(data2.Expiry) == 0 && data1.Reservations.Equals(data2.Reservations) && data1.Expired == data2.Expired
		}
	case CommandPlan_Checkout:
		{
			data2, ok := other.Get_().(CommandPlan_Checkout)
			return ok && data1.CheckoutCustomer.Equals(data2.CheckoutCustomer) && data1.NewOrderId.Equals(data2.NewOrderId) && data1.OrderLines.Equals(data2.OrderLines) && data1.RemainingItems.Equals(data2.RemainingItems) && data1.CartExpiry.Cmp(data2.CartExpiry) == 0 && data1.Discount.Cmp(data2.Discount) == 0
		}
	case CommandPlan_Adjust:
		{
			data2, ok := other.Get_().(CommandPlan_Adjust)
			return ok && data1.Product.Equals(data2.Product) && data1.Delta.Cmp(data2.Delta) == 0
		}
	case CommandPlan_UpdateOrder:
		{
			data2, ok := other.Get_().(CommandPlan_UpdateOrder)
			return ok && data1.Target.Cmp(data2.Target) == 0 && data1.OrderCustomer.Equals(data2.OrderCustomer) && data1.Prepare.Equals(data2.Prepare) && data1.Stock.Equals(data2.Stock) && data1.Obligations.Equals(data2.Obligations)
		}
	default:
		{
			return false // unexpected
		}
	}
}

func (_this CommandPlan) EqualsGeneric(other interface{}) bool {
	typed, ok := other.(CommandPlan)
	return ok && _this.Equals(typed)
}

func Type_CommandPlan_() _dafny.TypeDescriptor {
	return type_CommandPlan_{}
}

type type_CommandPlan_ struct {
}

func (_this type_CommandPlan_) Default() interface{} {
	return Companion_CommandPlan_.Default()
}

func (_this type_CommandPlan_) String() string {
	return "CommandModel.CommandPlan"
}
func (_this CommandPlan) ParentTraits_() []*_dafny.TraitID {
	return [](*_dafny.TraitID){}
}

var _ _dafny.TraitOffspring = CommandPlan{}

// End of datatype CommandPlan

// Definition of datatype Reference
type Reference struct {
	Data_Reference_
}

func (_this Reference) Get_() Data_Reference_ {
	return _this.Data_Reference_
}

type Data_Reference_ interface {
	isReference()
}

type CompanionStruct_Reference_ struct {
}

var Companion_Reference_ = CompanionStruct_Reference_{}

type Reference_Absent struct {
}

func (Reference_Absent) isReference() {}

func (CompanionStruct_Reference_) Create_Absent_() Reference {
	return Reference{Reference_Absent{}}
}

func (_this Reference) Is_Absent() bool {
	_, ok := _this.Get_().(Reference_Absent)
	return ok
}

type Reference_Id struct {
	Value _dafny.Sequence
}

func (Reference_Id) isReference() {}

func (CompanionStruct_Reference_) Create_Id_(Value _dafny.Sequence) Reference {
	return Reference{Reference_Id{Value}}
}

func (_this Reference) Is_Id() bool {
	_, ok := _this.Get_().(Reference_Id)
	return ok
}

func (CompanionStruct_Reference_) Default() Reference {
	return Companion_Reference_.Create_Absent_()
}

func (_this Reference) Dtor_value() _dafny.Sequence {
	return _this.Get_().(Reference_Id).Value
}

func (_this Reference) String() string {
	switch data := _this.Get_().(type) {
	case nil:
		return "null"
	case Reference_Absent:
		{
			return "CommandModel.Reference.Absent"
		}
	case Reference_Id:
		{
			return "CommandModel.Reference.Id" + "(" + data.Value.VerbatimString(true) + ")"
		}
	default:
		{
			return "<unexpected>"
		}
	}
}

func (_this Reference) Equals(other Reference) bool {
	switch data1 := _this.Get_().(type) {
	case Reference_Absent:
		{
			_, ok := other.Get_().(Reference_Absent)
			return ok
		}
	case Reference_Id:
		{
			data2, ok := other.Get_().(Reference_Id)
			return ok && data1.Value.Equals(data2.Value)
		}
	default:
		{
			return false // unexpected
		}
	}
}

func (_this Reference) EqualsGeneric(other interface{}) bool {
	typed, ok := other.(Reference)
	return ok && _this.Equals(typed)
}

func Type_Reference_() _dafny.TypeDescriptor {
	return type_Reference_{}
}

type type_Reference_ struct {
}

func (_this type_Reference_) Default() interface{} {
	return Companion_Reference_.Default()
}

func (_this type_Reference_) String() string {
	return "CommandModel.Reference"
}
func (_this Reference) ParentTraits_() []*_dafny.TraitID {
	return [](*_dafny.TraitID){}
}

var _ _dafny.TraitOffspring = Reference{}

// End of datatype Reference

// Definition of datatype Dependencies
type Dependencies struct {
	Data_Dependencies_
}

func (_this Dependencies) Get_() Data_Dependencies_ {
	return _this.Data_Dependencies_
}

type Data_Dependencies_ interface {
	isDependencies()
}

type CompanionStruct_Dependencies_ struct {
}

var Companion_Dependencies_ = CompanionStruct_Dependencies_{}

type Dependencies_Dependencies struct {
	Products _dafny.Set
	Stocks   _dafny.Set
	Loyalty  Reference
	NewOrder Reference
}

func (Dependencies_Dependencies) isDependencies() {}

func (CompanionStruct_Dependencies_) Create_Dependencies_(Products _dafny.Set, Stocks _dafny.Set, Loyalty Reference, NewOrder Reference) Dependencies {
	return Dependencies{Dependencies_Dependencies{Products, Stocks, Loyalty, NewOrder}}
}

func (_this Dependencies) Is_Dependencies() bool {
	_, ok := _this.Get_().(Dependencies_Dependencies)
	return ok
}

func (CompanionStruct_Dependencies_) Default() Dependencies {
	return Companion_Dependencies_.Create_Dependencies_(_dafny.EmptySet, _dafny.EmptySet, Companion_Reference_.Default(), Companion_Reference_.Default())
}

func (_this Dependencies) Dtor_products() _dafny.Set {
	return _this.Get_().(Dependencies_Dependencies).Products
}

func (_this Dependencies) Dtor_stocks() _dafny.Set {
	return _this.Get_().(Dependencies_Dependencies).Stocks
}

func (_this Dependencies) Dtor_loyalty() Reference {
	return _this.Get_().(Dependencies_Dependencies).Loyalty
}

func (_this Dependencies) Dtor_newOrder() Reference {
	return _this.Get_().(Dependencies_Dependencies).NewOrder
}

func (_this Dependencies) String() string {
	switch data := _this.Get_().(type) {
	case nil:
		return "null"
	case Dependencies_Dependencies:
		{
			return "CommandModel.Dependencies.Dependencies" + "(" + _dafny.String(data.Products) + ", " + _dafny.String(data.Stocks) + ", " + _dafny.String(data.Loyalty) + ", " + _dafny.String(data.NewOrder) + ")"
		}
	default:
		{
			return "<unexpected>"
		}
	}
}

func (_this Dependencies) Equals(other Dependencies) bool {
	switch data1 := _this.Get_().(type) {
	case Dependencies_Dependencies:
		{
			data2, ok := other.Get_().(Dependencies_Dependencies)
			return ok && data1.Products.Equals(data2.Products) && data1.Stocks.Equals(data2.Stocks) && data1.Loyalty.Equals(data2.Loyalty) && data1.NewOrder.Equals(data2.NewOrder)
		}
	default:
		{
			return false // unexpected
		}
	}
}

func (_this Dependencies) EqualsGeneric(other interface{}) bool {
	typed, ok := other.(Dependencies)
	return ok && _this.Equals(typed)
}

func Type_Dependencies_() _dafny.TypeDescriptor {
	return type_Dependencies_{}
}

type type_Dependencies_ struct {
}

func (_this type_Dependencies_) Default() interface{} {
	return Companion_Dependencies_.Default()
}

func (_this type_Dependencies_) String() string {
	return "CommandModel.Dependencies"
}
func (_this Dependencies) ParentTraits_() []*_dafny.TraitID {
	return [](*_dafny.TraitID){}
}

var _ _dafny.TraitOffspring = Dependencies{}

// End of datatype Dependencies

// Definition of datatype Status
type Status struct {
	Data_Status_
}

func (_this Status) Get_() Data_Status_ {
	return _this.Data_Status_
}

type Data_Status_ interface {
	isStatus()
}

type CompanionStruct_Status_ struct {
}

var Companion_Status_ = CompanionStruct_Status_{}

type Status_Ok struct {
}

func (Status_Ok) isStatus() {}

func (CompanionStruct_Status_) Create_Ok_() Status {
	return Status{Status_Ok{}}
}

func (_this Status) Is_Ok() bool {
	_, ok := _this.Get_().(Status_Ok)
	return ok
}

type Status_Failed struct {
}

func (Status_Failed) isStatus() {}

func (CompanionStruct_Status_) Create_Failed_() Status {
	return Status{Status_Failed{}}
}

func (_this Status) Is_Failed() bool {
	_, ok := _this.Get_().(Status_Failed)
	return ok
}

type Status_Conflict struct {
}

func (Status_Conflict) isStatus() {}

func (CompanionStruct_Status_) Create_Conflict_() Status {
	return Status{Status_Conflict{}}
}

func (_this Status) Is_Conflict() bool {
	_, ok := _this.Get_().(Status_Conflict)
	return ok
}

type Status_Rejected struct {
}

func (Status_Rejected) isStatus() {}

func (CompanionStruct_Status_) Create_Rejected_() Status {
	return Status{Status_Rejected{}}
}

func (_this Status) Is_Rejected() bool {
	_, ok := _this.Get_().(Status_Rejected)
	return ok
}

type Status_Skipped struct {
}

func (Status_Skipped) isStatus() {}

func (CompanionStruct_Status_) Create_Skipped_() Status {
	return Status{Status_Skipped{}}
}

func (_this Status) Is_Skipped() bool {
	_, ok := _this.Get_().(Status_Skipped)
	return ok
}

func (CompanionStruct_Status_) Default() Status {
	return Companion_Status_.Create_Ok_()
}

func (_ CompanionStruct_Status_) AllSingletonConstructors() _dafny.Iterator {
	i := -1
	return func() (interface{}, bool) {
		i++
		switch i {
		case 0:
			return Companion_Status_.Create_Ok_(), true
		case 1:
			return Companion_Status_.Create_Failed_(), true
		case 2:
			return Companion_Status_.Create_Conflict_(), true
		case 3:
			return Companion_Status_.Create_Rejected_(), true
		case 4:
			return Companion_Status_.Create_Skipped_(), true
		default:
			return Status{}, false
		}
	}
}

func (_this Status) String() string {
	switch _this.Get_().(type) {
	case nil:
		return "null"
	case Status_Ok:
		{
			return "CommandModel.Status.Ok"
		}
	case Status_Failed:
		{
			return "CommandModel.Status.Failed"
		}
	case Status_Conflict:
		{
			return "CommandModel.Status.Conflict"
		}
	case Status_Rejected:
		{
			return "CommandModel.Status.Rejected"
		}
	case Status_Skipped:
		{
			return "CommandModel.Status.Skipped"
		}
	default:
		{
			return "<unexpected>"
		}
	}
}

func (_this Status) Equals(other Status) bool {
	switch _this.Get_().(type) {
	case Status_Ok:
		{
			_, ok := other.Get_().(Status_Ok)
			return ok
		}
	case Status_Failed:
		{
			_, ok := other.Get_().(Status_Failed)
			return ok
		}
	case Status_Conflict:
		{
			_, ok := other.Get_().(Status_Conflict)
			return ok
		}
	case Status_Rejected:
		{
			_, ok := other.Get_().(Status_Rejected)
			return ok
		}
	case Status_Skipped:
		{
			_, ok := other.Get_().(Status_Skipped)
			return ok
		}
	default:
		{
			return false // unexpected
		}
	}
}

func (_this Status) EqualsGeneric(other interface{}) bool {
	typed, ok := other.(Status)
	return ok && _this.Equals(typed)
}

func Type_Status_() _dafny.TypeDescriptor {
	return type_Status_{}
}

type type_Status_ struct {
}

func (_this type_Status_) Default() interface{} {
	return Companion_Status_.Default()
}

func (_this type_Status_) String() string {
	return "CommandModel.Status"
}
func (_this Status) ParentTraits_() []*_dafny.TraitID {
	return [](*_dafny.TraitID){}
}

var _ _dafny.TraitOffspring = Status{}

// End of datatype Status
