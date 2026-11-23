// Package CommandFlow
// Dafny module CommandFlow compiled into Go

package CommandFlow

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
	return "CommandFlow.Default__"
}
func (_this *Default__) ParentTraits_() []*_dafny.TraitID {
	return [](*_dafny.TraitID){}
}

var _ _dafny.TraitOffspring = &Default__{}

func (_static *CompanionStruct_Default___) DecideOrder(kind _dafny.Sequence, x OrderContext) OrderDecision {
	if (_dafny.Companion_Sequence_.Equal((x).Dtor_role(), _dafny.UnicodeSeqOfUtf8Bytes("customer"))) && (!_dafny.Companion_Sequence_.Equal((x).Dtor_actor(), (x).Dtor_customer())) {
		return Companion_OrderDecision_.Create_Denied_(_dafny.UnicodeSeqOfUtf8Bytes("他の顧客の受注は操作できません"))
	} else if _dafny.Companion_Sequence_.Equal(kind, _dafny.UnicodeSeqOfUtf8Bytes("order.pay")) {
		if !_dafny.Companion_Sequence_.Equal((x).Dtor_role(), _dafny.UnicodeSeqOfUtf8Bytes("customer")) {
			return Companion_OrderDecision_.Create_Denied_(_dafny.UnicodeSeqOfUtf8Bytes("この利用者には操作権限がありません"))
		} else {
			return Companion_OrderDecision_.Create_Move_((func() _dafny.Int {
				if (x).Dtor_paymentSucceeded() {
					return Companion_Default___.Accepted()
				}
				return Companion_Default___.PaymentFailed()
			})())
		}
	} else if _dafny.Companion_Sequence_.Equal(kind, _dafny.UnicodeSeqOfUtf8Bytes("order.cancel")) {
		if !(_dafny.SetOf(_dafny.UnicodeSeqOfUtf8Bytes("customer"), _dafny.UnicodeSeqOfUtf8Bytes("operator"), _dafny.UnicodeSeqOfUtf8Bytes("system"))).Contains((x).Dtor_role()) {
			return Companion_OrderDecision_.Create_Denied_(_dafny.UnicodeSeqOfUtf8Bytes("この利用者には操作権限がありません"))
		} else if (_dafny.Companion_Sequence_.Equal((x).Dtor_role(), _dafny.UnicodeSeqOfUtf8Bytes("system"))) && ((((x).Dtor_status()).Cmp(Companion_Default___.PaymentPending()) != 0) || (((x).Dtor_ageNanos()).Cmp(Companion_Default___.PaymentTimeoutNanos()) < 0)) {
			return Companion_OrderDecision_.Create_NoChange_()
		} else if (_dafny.Companion_Sequence_.Equal((x).Dtor_role(), _dafny.UnicodeSeqOfUtf8Bytes("customer"))) && (!(_dafny.SetOf(Companion_Default___.PaymentPending(), Companion_Default___.Accepted())).Contains((x).Dtor_status())) {
			return Companion_OrderDecision_.Create_Denied_(_dafny.UnicodeSeqOfUtf8Bytes("作業開始後のキャンセルは運用者に依頼してください"))
		} else {
			return Companion_OrderDecision_.Create_Move_(Companion_Default___.Cancelled())
		}
	} else if (_dafny.SetOf(_dafny.UnicodeSeqOfUtf8Bytes("order.pick"), _dafny.UnicodeSeqOfUtf8Bytes("order.block"), _dafny.UnicodeSeqOfUtf8Bytes("order.pack"))).Contains(kind) {
		if !_dafny.Companion_Sequence_.Equal((x).Dtor_role(), _dafny.UnicodeSeqOfUtf8Bytes("warehouse")) {
			return Companion_OrderDecision_.Create_Denied_(_dafny.UnicodeSeqOfUtf8Bytes("この利用者には操作権限がありません"))
		} else if _dafny.Companion_Sequence_.Equal(kind, _dafny.UnicodeSeqOfUtf8Bytes("order.pick")) {
			if !(m_Domain.Companion_Default___.MayAssignWorker(!_dafny.Companion_Sequence_.Equal((x).Dtor_worker(), _dafny.UnicodeSeqOfUtf8Bytes("")), _dafny.Companion_Sequence_.Equal((x).Dtor_worker(), (x).Dtor_actor()))) {
				return Companion_OrderDecision_.Create_Denied_(_dafny.UnicodeSeqOfUtf8Bytes("別の担当者が作業中です"))
			} else {
				return Companion_OrderDecision_.Create_Move_(Companion_Default___.Picking())
			}
		} else if !_dafny.Companion_Sequence_.Equal((x).Dtor_worker(), (x).Dtor_actor()) {
			return Companion_OrderDecision_.Create_Denied_(_dafny.UnicodeSeqOfUtf8Bytes("担当外の作業です"))
		} else if _dafny.Companion_Sequence_.Equal(kind, _dafny.UnicodeSeqOfUtf8Bytes("order.block")) {
			if !((x).Dtor_hasNote()) {
				return Companion_OrderDecision_.Create_Denied_(_dafny.UnicodeSeqOfUtf8Bytes("保留理由が必要です"))
			} else {
				return Companion_OrderDecision_.Create_Move_(Companion_Default___.PickBlocked())
			}
		} else if (!(m_Domain.Companion_Default___.BoxCountValid((x).Dtor_boxes()))) || (!(m_Domain.Companion_Default___.WeightValid((x).Dtor_weight()))) {
			return Companion_OrderDecision_.Create_Denied_(_dafny.UnicodeSeqOfUtf8Bytes("箱数は1〜100、1箱の重量は1〜500000gです"))
		} else {
			return Companion_OrderDecision_.Create_Move_(Companion_Default___.Packed())
		}
	} else if (_dafny.SetOf(_dafny.UnicodeSeqOfUtf8Bytes("order.dispatch"), _dafny.UnicodeSeqOfUtf8Bytes("order.delivery_failed"), _dafny.UnicodeSeqOfUtf8Bytes("order.redeliver"), _dafny.UnicodeSeqOfUtf8Bytes("order.deliver"), _dafny.UnicodeSeqOfUtf8Bytes("order.return"))).Contains(kind) {
		if !_dafny.Companion_Sequence_.Equal((x).Dtor_role(), _dafny.UnicodeSeqOfUtf8Bytes("carrier")) {
			return Companion_OrderDecision_.Create_Denied_(_dafny.UnicodeSeqOfUtf8Bytes("この利用者には操作権限がありません"))
		} else if _dafny.Companion_Sequence_.Equal(kind, _dafny.UnicodeSeqOfUtf8Bytes("order.dispatch")) {
			if !((x).Dtor_hasTracking()) {
				return Companion_OrderDecision_.Create_Denied_(_dafny.UnicodeSeqOfUtf8Bytes("追跡番号を入力してください"))
			} else if ((x).Dtor_status()).Cmp(Companion_Default___.Packed()) != 0 {
				return Companion_OrderDecision_.Create_Denied_(_dafny.UnicodeSeqOfUtf8Bytes("梱包完了後に出荷してください"))
			} else {
				return Companion_OrderDecision_.Create_Move_(Companion_Default___.Shipping())
			}
		} else if !_dafny.Companion_Sequence_.Equal((x).Dtor_carrier(), (x).Dtor_actor()) {
			return Companion_OrderDecision_.Create_Denied_(_dafny.UnicodeSeqOfUtf8Bytes("担当外の配送です"))
		} else if _dafny.Companion_Sequence_.Equal(kind, _dafny.UnicodeSeqOfUtf8Bytes("order.delivery_failed")) {
			if !((x).Dtor_hasNote()) {
				return Companion_OrderDecision_.Create_Denied_(_dafny.UnicodeSeqOfUtf8Bytes("配送失敗理由が必要です"))
			} else {
				return Companion_OrderDecision_.Create_Move_(Companion_Default___.DeliveryFailed())
			}
		} else if _dafny.Companion_Sequence_.Equal(kind, _dafny.UnicodeSeqOfUtf8Bytes("order.redeliver")) {
			if ((x).Dtor_status()).Cmp(Companion_Default___.DeliveryFailed()) != 0 {
				return Companion_OrderDecision_.Create_Denied_(_dafny.UnicodeSeqOfUtf8Bytes("再配送できる状態ではありません"))
			} else {
				return Companion_OrderDecision_.Create_Move_(Companion_Default___.Shipping())
			}
		} else if _dafny.Companion_Sequence_.Equal(kind, _dafny.UnicodeSeqOfUtf8Bytes("order.deliver")) {
			return Companion_OrderDecision_.Create_Move_(Companion_Default___.Delivered())
		} else if !((x).Dtor_hasNote()) {
			return Companion_OrderDecision_.Create_Denied_(_dafny.UnicodeSeqOfUtf8Bytes("返送理由が必要です"))
		} else {
			return Companion_OrderDecision_.Create_Move_(Companion_Default___.Returned())
		}
	} else {
		return Companion_OrderDecision_.Create_Denied_(_dafny.UnicodeSeqOfUtf8Bytes("未対応のCommandです"))
	}
}
func (_static *CompanionStruct_Default___) ChangeStock(e Effects, lines _dafny.Sequence, action StockAction) Step {
	var r Step = Companion_Step_.Default()
	_ = r
	var _hi0 _dafny.Int = _dafny.IntOfUint32((lines).Cardinality())
	_ = _hi0
	for _0_i := _dafny.Zero; _0_i.Cmp(_hi0) < 0; _0_i = _0_i.Plus(_dafny.One) {
		var _1_line StockLine
		_ = _1_line
		_1_line = (lines).Select((_0_i).Uint32()).(StockLine)
		var _out0 Step
		_ = _out0
		_out0 = (e).LoadStock((_1_line).Dtor_product())
		r = _out0
		if !((r).Is_Ok()) {
			return r
		}
		if (action).Is_Ship() {
			var _out1 Step
			_ = _out1
			_out1 = (e).ShipStock((_1_line).Dtor_quantity())
			r = _out1
		} else {
			var _2_delta _dafny.Int
			_ = _2_delta
			if (action).Is_Release() {
				_2_delta = (_dafny.Zero).Minus((_1_line).Dtor_quantity())
			} else {
				_2_delta = (_1_line).Dtor_quantity()
			}
			var _out2 Step
			_ = _out2
			_out2 = (e).ReserveStock(_2_delta)
			r = _out2
		}
		if !((r).Is_Ok()) {
			return r
		}
		var _out3 Step
		_ = _out3
		_out3 = (e).SaveStock((_1_line).Dtor_product())
		r = _out3
		if !((r).Is_Ok()) {
			return r
		}
	}
	r = Companion_Step_.Create_Ok_()
	return r
}
func (_static *CompanionStruct_Default___) Checkout(e Effects, hasCoupon bool) Step {
	var r Step = Companion_Step_.Default()
	_ = r
	var _0_lines _dafny.Sequence = _dafny.EmptySeq
	_ = _0_lines
	var _out0 Step
	_ = _out0
	var _out1 _dafny.Sequence
	_ = _out1
	_out0, _out1 = (e).CheckoutCart()
	r = _out0
	_0_lines = _out1
	if !((r).Is_Ok()) {
		return r
	}
	var _out2 Step
	_ = _out2
	_out2 = (e).ReadShippingAddress()
	r = _out2
	if !((r).Is_Ok()) {
		return r
	}
	var _1_provided bool
	_ = _1_provided
	var _2_parsed bool
	_ = _2_parsed
	var _3_day _dafny.Int
	_ = _3_day
	var _4_today _dafny.Int
	_ = _4_today
	var _out3 bool
	_ = _out3
	var _out4 bool
	_ = _out4
	var _out5 _dafny.Int
	_ = _out5
	var _out6 _dafny.Int
	_ = _out6
	_out3, _out4, _out5, _out6 = (e).RequestedDate()
	_1_provided = _out3
	_2_parsed = _out4
	_3_day = _out5
	_4_today = _out6
	if (_1_provided) && ((!(_2_parsed)) || ((_3_day).Cmp(_4_today) < 0)) {
		var _out7 Step
		_ = _out7
		_out7 = (e).Reject(_dafny.UnicodeSeqOfUtf8Bytes("希望納品日が不正です"))
		r = _out7
		return r
	}
	var _5_percent _dafny.Int
	_ = _5_percent
	_5_percent = _dafny.Zero
	if hasCoupon {
		var _out8 Step
		_ = _out8
		var _out9 _dafny.Int
		_ = _out9
		_out8, _out9 = (e).LoadCoupon()
		r = _out8
		_5_percent = _out9
		if !((r).Is_Ok()) {
			return r
		}
	}
	if !(m_Domain.Companion_Default___.DiscountRateValid(_5_percent)) {
		var _out10 Step
		_ = _out10
		_out10 = (e).Reject(_dafny.UnicodeSeqOfUtf8Bytes("割引率は0〜100です"))
		r = _out10
		return r
	}
	var _hi0 _dafny.Int = _dafny.IntOfUint32((_0_lines).Cardinality())
	_ = _hi0
	for _6_i := _dafny.Zero; _6_i.Cmp(_hi0) < 0; _6_i = _6_i.Plus(_dafny.One) {
		var _out11 Step
		_ = _out11
		_out11 = (e).LoadProduct(((_0_lines).Select((_6_i).Uint32()).(StockLine)).Dtor_product())
		r = _out11
		if !((r).Is_Ok()) {
			return r
		}
		var _out12 Step
		_ = _out12
		_out12 = (e).AddOrderLine((_0_lines).Select((_6_i).Uint32()).(StockLine))
		r = _out12
		if !((r).Is_Ok()) {
			return r
		}
	}
	var _out13 Step
	_ = _out13
	_out13 = (e).BuildOrder(_5_percent)
	r = _out13
	if !((r).Is_Ok()) {
		return r
	}
	var _out14 Step
	_ = _out14
	_out14 = (e).LoadCheckoutVersion()
	r = _out14
	if !((r).Is_Ok()) {
		return r
	}
	var _out15 Step
	_ = _out15
	_out15 = (e).SaveOrder()
	r = _out15
	if (r).Is_Ok() {
		var _out16 Step
		_ = _out16
		_out16 = (e).SaveCart()
		r = _out16
	}
	return r
}
func (_static *CompanionStruct_Default___) VersionMatches(expected _dafny.Int, actual _dafny.Int) bool {
	return ((expected).Sign() != 1) || ((expected).Cmp(actual) == 0)
}
func (_static *CompanionStruct_Default___) Apply(e Effects) Step {
	var r Step = Companion_Step_.Default()
	_ = r
	var _out0 Step
	_ = _out0
	_out0 = (e).LoadActor()
	r = _out0
	if !((r).Is_Ok()) {
		return r
	}
	var _out1 Step
	_ = _out1
	_out1 = (e).ReadClock()
	r = _out1
	if !((r).Is_Ok()) {
		return r
	}
	var _0_input CommandInput
	_ = _0_input
	var _out2 CommandInput
	_ = _out2
	_out2 = (e).Input()
	_0_input = _out2
	if ((((_0_input).Dtor_noteBytes()).Cmp(_dafny.IntOfInt64(1000)) > 0) || (((_0_input).Dtor_trackingBytes()).Cmp(_dafny.IntOfInt64(100)) > 0)) || (((_0_input).Dtor_requestedDateBytes()).Cmp(_dafny.IntOfInt64(10)) > 0) {
		var _out3 Step
		_ = _out3
		_out3 = (e).Reject(_dafny.UnicodeSeqOfUtf8Bytes("入力が長すぎます"))
		r = _out3
		return r
	}
	var _1_kind _dafny.Sequence
	_ = _1_kind
	_1_kind = (_0_input).Dtor_kind()
	if (_dafny.SetOf(_dafny.UnicodeSeqOfUtf8Bytes("cart.set"), _dafny.UnicodeSeqOfUtf8Bytes("cart.expire"), _dafny.UnicodeSeqOfUtf8Bytes("order.checkout"))).Contains(_1_kind) {
		var _2_requiredRole _dafny.Sequence
		_ = _2_requiredRole
		if _dafny.Companion_Sequence_.Equal(_1_kind, _dafny.UnicodeSeqOfUtf8Bytes("cart.expire")) {
			_2_requiredRole = _dafny.UnicodeSeqOfUtf8Bytes("system")
		} else {
			_2_requiredRole = _dafny.UnicodeSeqOfUtf8Bytes("customer")
		}
		if !_dafny.Companion_Sequence_.Equal((_0_input).Dtor_role(), _2_requiredRole) {
			var _out4 Step
			_ = _out4
			_out4 = (e).Reject(_dafny.UnicodeSeqOfUtf8Bytes("この利用者には操作権限がありません"))
			r = _out4
			return r
		}
		var _3_customer _dafny.Sequence
		_ = _3_customer
		if _dafny.Companion_Sequence_.Equal(_1_kind, _dafny.UnicodeSeqOfUtf8Bytes("cart.expire")) {
			_3_customer = (_0_input).Dtor_customer()
		} else {
			_3_customer = (_0_input).Dtor_actor()
		}
		var _4_version _dafny.Int = _dafny.Zero
		_ = _4_version
		var _out5 Step
		_ = _out5
		var _out6 _dafny.Int
		_ = _out6
		_out5, _out6 = (e).LoadCart(_3_customer)
		r = _out5
		_4_version = _out6
		if !((r).Is_Ok()) {
			return r
		}
		if !(Companion_Default___.VersionMatches((_0_input).Dtor_expectedVersion(), _4_version)) {
			var _out7 Step
			_ = _out7
			_out7 = (e).Reject(Companion_Default___.StaleVersion())
			r = _out7
			return r
		}
		var _5_lines _dafny.Sequence = _dafny.EmptySeq
		_ = _5_lines
		if _dafny.Companion_Sequence_.Equal(_1_kind, _dafny.UnicodeSeqOfUtf8Bytes("order.checkout")) {
			var _out8 Step
			_ = _out8
			_out8 = Companion_Default___.Checkout(e, (_0_input).Dtor_hasCoupon())
			r = _out8
		} else {
			if _dafny.Companion_Sequence_.Equal(_1_kind, _dafny.UnicodeSeqOfUtf8Bytes("cart.expire")) {
				var _out9 Step
				_ = _out9
				var _out10 _dafny.Sequence
				_ = _out10
				_out9, _out10 = (e).ExpireCart()
				r = _out9
				_5_lines = _out10
			} else {
				var _out11 Step
				_ = _out11
				var _out12 _dafny.Sequence
				_ = _out12
				_out11, _out12 = (e).SetCart()
				r = _out11
				_5_lines = _out12
			}
			if (r).Is_Skip() {
				(e).Noop()
				r = Companion_Step_.Create_Ok_()
				return r
			}
			if !((r).Is_Ok()) {
				return r
			}
			var _out13 Step
			_ = _out13
			_out13 = Companion_Default___.ChangeStock(e, _5_lines, Companion_StockAction_.Create_Reserve_())
			r = _out13
			if (r).Is_Ok() {
				var _out14 Step
				_ = _out14
				_out14 = (e).SaveCart()
				r = _out14
			}
		}
	} else if _dafny.Companion_Sequence_.Equal(_1_kind, _dafny.UnicodeSeqOfUtf8Bytes("stock.adjust")) {
		if !_dafny.Companion_Sequence_.Equal((_0_input).Dtor_role(), _dafny.UnicodeSeqOfUtf8Bytes("operator")) {
			var _out15 Step
			_ = _out15
			_out15 = (e).Reject(_dafny.UnicodeSeqOfUtf8Bytes("この利用者には操作権限がありません"))
			r = _out15
			return r
		}
		if ((!((_0_input).Dtor_productValid())) || (!(m_Domain.Companion_Default___.StockDeltaValid((_0_input).Dtor_stockDelta())))) || (((_0_input).Dtor_noteBytes()).Sign() == 0) {
			var _out16 Step
			_ = _out16
			_out16 = (e).Reject(_dafny.UnicodeSeqOfUtf8Bytes("商品・調整数・理由を入力してください"))
			r = _out16
			return r
		}
		var _6_version _dafny.Int = _dafny.Zero
		_ = _6_version
		var _out17 Step
		_ = _out17
		var _out18 _dafny.Int
		_ = _out18
		_out17, _out18 = (e).LoadAdjustment()
		r = _out17
		_6_version = _out18
		if !((r).Is_Ok()) {
			return r
		}
		if !(Companion_Default___.VersionMatches((_0_input).Dtor_expectedVersion(), _6_version)) {
			var _out19 Step
			_ = _out19
			_out19 = (e).Reject(Companion_Default___.StaleVersion())
			r = _out19
			return r
		}
		var _out20 Step
		_ = _out20
		_out20 = (e).AdjustStock()
		r = _out20
		if (r).Is_Ok() {
			var _out21 Step
			_ = _out21
			_out21 = (e).SaveAdjustment()
			r = _out21
		}
	} else {
		var _7_x OrderContext = Companion_OrderContext_.Default()
		_ = _7_x
		var _8_version _dafny.Int = _dafny.Zero
		_ = _8_version
		var _out22 Step
		_ = _out22
		var _out23 OrderContext
		_ = _out23
		var _out24 _dafny.Int
		_ = _out24
		_out22, _out23, _out24 = (e).LoadOrder()
		r = _out22
		_7_x = _out23
		_8_version = _out24
		if !((r).Is_Ok()) {
			return r
		}
		if !(Companion_Default___.VersionMatches((_0_input).Dtor_expectedVersion(), _8_version)) {
			var _out25 Step
			_ = _out25
			_out25 = (e).Reject(Companion_Default___.StaleVersion())
			r = _out25
			return r
		}
		var _9_decision OrderDecision
		_ = _9_decision
		_9_decision = Companion_Default___.DecideOrder(_1_kind, _7_x)
		if (_9_decision).Is_NoChange() {
			(e).Noop()
			return r
		}
		if (_9_decision).Is_Denied() {
			var _out26 Step
			_ = _out26
			_out26 = (e).Reject((_9_decision).Dtor_message())
			r = _out26
			return r
		}
		var _10_target _dafny.Int
		_ = _10_target
		_10_target = (_9_decision).Dtor_target()
		if _dafny.Companion_Sequence_.Equal(_1_kind, _dafny.UnicodeSeqOfUtf8Bytes("order.pick")) {
			var _out27 Step
			_ = _out27
			_out27 = (e).AssignWorker()
			r = _out27
		} else if _dafny.Companion_Sequence_.Equal(_1_kind, _dafny.UnicodeSeqOfUtf8Bytes("order.pack")) {
			var _out28 Step
			_ = _out28
			_out28 = (e).PackOrder()
			r = _out28
		} else if _dafny.Companion_Sequence_.Equal(_1_kind, _dafny.UnicodeSeqOfUtf8Bytes("order.dispatch")) {
			var _out29 Step
			_ = _out29
			_out29 = (e).AssignCarrier()
			r = _out29
		}
		if !((r).Is_Ok()) {
			return r
		}
		var _11_lines _dafny.Sequence = _dafny.EmptySeq
		_ = _11_lines
		var _out30 Step
		_ = _out30
		var _out31 _dafny.Sequence
		_ = _out31
		_out30, _out31 = (e).MoveOrder(_10_target)
		r = _out30
		_11_lines = _out31
		if !((r).Is_Ok()) {
			return r
		}
		if ((_10_target).Cmp(Companion_Default___.Cancelled()) == 0) || ((_10_target).Cmp(Companion_Default___.PaymentFailed()) == 0) {
			var _out32 Step
			_ = _out32
			_out32 = Companion_Default___.ChangeStock(e, _11_lines, Companion_StockAction_.Create_Release_())
			r = _out32
		} else if _dafny.Companion_Sequence_.Equal(_1_kind, _dafny.UnicodeSeqOfUtf8Bytes("order.dispatch")) {
			var _out33 Step
			_ = _out33
			_out33 = Companion_Default___.ChangeStock(e, _11_lines, Companion_StockAction_.Create_Ship_())
			r = _out33
		}
		if !((r).Is_Ok()) {
			return r
		}
		var _out34 Step
		_ = _out34
		_out34 = (e).SaveOrder()
		r = _out34
		if (!((r).Is_Ok())) || ((_10_target).Cmp(Companion_Default___.Delivered()) != 0) {
			return r
		}
		var _out35 Step
		_ = _out35
		_out35 = (e).InsertPointAward()
		r = _out35
		if !((r).Is_Ok()) {
			return r
		}
		var _out36 Step
		_ = _out36
		_out36 = (e).LoadLoyalty()
		r = _out36
		if !((r).Is_Ok()) {
			return r
		}
		var _out37 Step
		_ = _out37
		_out37 = (e).CreditLoyalty()
		r = _out37
		if !((r).Is_Ok()) {
			return r
		}
		var _out38 Step
		_ = _out38
		_out38 = (e).SaveLoyalty()
		r = _out38
		if !((r).Is_Ok()) {
			return r
		}
		var _out39 Step
		_ = _out39
		_out39 = (e).RenderInvoice()
		r = _out39
		if (r).Is_Ok() {
			var _out40 Step
			_ = _out40
			_out40 = (e).InsertInvoice()
			r = _out40
		}
	}
	return r
}
func (_static *CompanionStruct_Default___) Attempt(e Effects) Step {
	var r Step = Companion_Step_.Default()
	_ = r
	var _out0 Step
	_ = _out0
	_out0 = (e).Begin()
	r = _out0
	if !((r).Is_Ok()) {
		return r
	}
	var _0_status _dafny.Sequence = _dafny.EmptySeq
	_ = _0_status
	var _out1 Step
	_ = _out1
	var _out2 _dafny.Sequence
	_ = _out2
	_out1, _out2 = (e).LoadCommand()
	r = _out1
	_0_status = _out2
	if (r).Is_Ok() {
		if _dafny.Companion_Sequence_.Equal(_0_status, _dafny.UnicodeSeqOfUtf8Bytes("queued")) {
			var _out3 Step
			_ = _out3
			_out3 = (e).DecodePayload()
			r = _out3
		} else {
			r = Companion_Step_.Create_Skip_()
		}
	}
	if (r).Is_Ok() {
		var _out4 Step
		_ = _out4
		_out4 = Companion_Default___.Apply(e)
		r = _out4
	}
	if (r).Is_Ok() {
		var _out5 Step
		_ = _out5
		_out5 = (e).SaveCompletion()
		r = _out5
		if (r).Is_Ok() {
			var _out6 Step
			_ = _out6
			_out6 = (e).Commit()
			r = _out6
			return r
		}
	}
	var _1_rollback Step
	_ = _1_rollback
	var _out7 Step
	_ = _out7
	_out7 = (e).Rollback()
	_1_rollback = _out7
	if (r).Is_Rejected() {
		if (_1_rollback).Is_Ok() {
			var _out8 Step
			_ = _out8
			_out8 = (e).SaveRejection()
			r = _out8
			if (r).Is_Ok() {
				r = Companion_Step_.Create_Skip_()
			}
		} else {
			r = _1_rollback
		}
	}
	return r
}
func (_static *CompanionStruct_Default___) ExecuteCommand(e Effects) Step {
	var r Step = Companion_Step_.Default()
	_ = r
	var _0_attempt _dafny.Int
	_ = _0_attempt
	_0_attempt = _dafny.Zero
	for (_0_attempt).Cmp(_dafny.IntOfInt64(5)) < 0 {
		var _out0 Step
		_ = _out0
		_out0 = Companion_Default___.Attempt(e)
		r = _out0
		if !((r).Is_Conflict()) {
			if (r).Is_Failed() {
				(e).RecordFailure(_dafny.IntOfInt64(8))
			}
			return r
		}
		_0_attempt = (_0_attempt).Plus(_dafny.One)
	}
	r = Companion_Step_.Create_Conflict_()
	return r
}
func (_static *CompanionStruct_Default___) Delivered() _dafny.Int {
	return _dafny.IntOfInt64(7)
}
func (_static *CompanionStruct_Default___) PaymentPending() _dafny.Int {
	return _dafny.Zero
}
func (_static *CompanionStruct_Default___) PaymentTimeoutNanos() _dafny.Int {
	return ((_dafny.IntOfInt64(15)).Times(_dafny.IntOfInt64(60))).Times(_dafny.IntOfInt64(1000000000))
}
func (_static *CompanionStruct_Default___) Accepted() _dafny.Int {
	return _dafny.One
}
func (_static *CompanionStruct_Default___) PaymentFailed() _dafny.Int {
	return _dafny.IntOfInt64(9)
}
func (_static *CompanionStruct_Default___) Cancelled() _dafny.Int {
	return _dafny.IntOfInt64(8)
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
func (_static *CompanionStruct_Default___) Returned() _dafny.Int {
	return _dafny.IntOfInt64(10)
}
func (_static *CompanionStruct_Default___) StaleVersion() _dafny.Sequence {
	return _dafny.UnicodeSeqOfUtf8Bytes("更新が競合しました。再読込してください")
}

// End of class Default__

// Definition of datatype OrderContext
type OrderContext struct {
	Data_OrderContext_
}

func (_this OrderContext) Get_() Data_OrderContext_ {
	return _this.Data_OrderContext_
}

type Data_OrderContext_ interface {
	isOrderContext()
}

type CompanionStruct_OrderContext_ struct {
}

var Companion_OrderContext_ = CompanionStruct_OrderContext_{}

type OrderContext_OrderContext struct {
	Role             _dafny.Sequence
	Actor            _dafny.Sequence
	Customer         _dafny.Sequence
	Worker           _dafny.Sequence
	Carrier          _dafny.Sequence
	Status           _dafny.Int
	AgeNanos         _dafny.Int
	PaymentSucceeded bool
	HasNote          bool
	HasTracking      bool
	Boxes            _dafny.Int
	Weight           _dafny.Int
}

func (OrderContext_OrderContext) isOrderContext() {}

func (CompanionStruct_OrderContext_) Create_OrderContext_(Role _dafny.Sequence, Actor _dafny.Sequence, Customer _dafny.Sequence, Worker _dafny.Sequence, Carrier _dafny.Sequence, Status _dafny.Int, AgeNanos _dafny.Int, PaymentSucceeded bool, HasNote bool, HasTracking bool, Boxes _dafny.Int, Weight _dafny.Int) OrderContext {
	return OrderContext{OrderContext_OrderContext{Role, Actor, Customer, Worker, Carrier, Status, AgeNanos, PaymentSucceeded, HasNote, HasTracking, Boxes, Weight}}
}

func (_this OrderContext) Is_OrderContext() bool {
	_, ok := _this.Get_().(OrderContext_OrderContext)
	return ok
}

func (CompanionStruct_OrderContext_) Default() OrderContext {
	return Companion_OrderContext_.Create_OrderContext_(_dafny.EmptySeq, _dafny.EmptySeq, _dafny.EmptySeq, _dafny.EmptySeq, _dafny.EmptySeq, _dafny.Zero, _dafny.Zero, false, false, false, _dafny.Zero, _dafny.Zero)
}

func (_this OrderContext) Dtor_role() _dafny.Sequence {
	return _this.Get_().(OrderContext_OrderContext).Role
}

func (_this OrderContext) Dtor_actor() _dafny.Sequence {
	return _this.Get_().(OrderContext_OrderContext).Actor
}

func (_this OrderContext) Dtor_customer() _dafny.Sequence {
	return _this.Get_().(OrderContext_OrderContext).Customer
}

func (_this OrderContext) Dtor_worker() _dafny.Sequence {
	return _this.Get_().(OrderContext_OrderContext).Worker
}

func (_this OrderContext) Dtor_carrier() _dafny.Sequence {
	return _this.Get_().(OrderContext_OrderContext).Carrier
}

func (_this OrderContext) Dtor_status() _dafny.Int {
	return _this.Get_().(OrderContext_OrderContext).Status
}

func (_this OrderContext) Dtor_ageNanos() _dafny.Int {
	return _this.Get_().(OrderContext_OrderContext).AgeNanos
}

func (_this OrderContext) Dtor_paymentSucceeded() bool {
	return _this.Get_().(OrderContext_OrderContext).PaymentSucceeded
}

func (_this OrderContext) Dtor_hasNote() bool {
	return _this.Get_().(OrderContext_OrderContext).HasNote
}

func (_this OrderContext) Dtor_hasTracking() bool {
	return _this.Get_().(OrderContext_OrderContext).HasTracking
}

func (_this OrderContext) Dtor_boxes() _dafny.Int {
	return _this.Get_().(OrderContext_OrderContext).Boxes
}

func (_this OrderContext) Dtor_weight() _dafny.Int {
	return _this.Get_().(OrderContext_OrderContext).Weight
}

func (_this OrderContext) String() string {
	switch data := _this.Get_().(type) {
	case nil:
		return "null"
	case OrderContext_OrderContext:
		{
			return "CommandFlow.OrderContext.OrderContext" + "(" + data.Role.VerbatimString(true) + ", " + data.Actor.VerbatimString(true) + ", " + data.Customer.VerbatimString(true) + ", " + data.Worker.VerbatimString(true) + ", " + data.Carrier.VerbatimString(true) + ", " + _dafny.String(data.Status) + ", " + _dafny.String(data.AgeNanos) + ", " + _dafny.String(data.PaymentSucceeded) + ", " + _dafny.String(data.HasNote) + ", " + _dafny.String(data.HasTracking) + ", " + _dafny.String(data.Boxes) + ", " + _dafny.String(data.Weight) + ")"
		}
	default:
		{
			return "<unexpected>"
		}
	}
}

func (_this OrderContext) Equals(other OrderContext) bool {
	switch data1 := _this.Get_().(type) {
	case OrderContext_OrderContext:
		{
			data2, ok := other.Get_().(OrderContext_OrderContext)
			return ok && data1.Role.Equals(data2.Role) && data1.Actor.Equals(data2.Actor) && data1.Customer.Equals(data2.Customer) && data1.Worker.Equals(data2.Worker) && data1.Carrier.Equals(data2.Carrier) && data1.Status.Cmp(data2.Status) == 0 && data1.AgeNanos.Cmp(data2.AgeNanos) == 0 && data1.PaymentSucceeded == data2.PaymentSucceeded && data1.HasNote == data2.HasNote && data1.HasTracking == data2.HasTracking && data1.Boxes.Cmp(data2.Boxes) == 0 && data1.Weight.Cmp(data2.Weight) == 0
		}
	default:
		{
			return false // unexpected
		}
	}
}

func (_this OrderContext) EqualsGeneric(other interface{}) bool {
	typed, ok := other.(OrderContext)
	return ok && _this.Equals(typed)
}

func Type_OrderContext_() _dafny.TypeDescriptor {
	return type_OrderContext_{}
}

type type_OrderContext_ struct {
}

func (_this type_OrderContext_) Default() interface{} {
	return Companion_OrderContext_.Default()
}

func (_this type_OrderContext_) String() string {
	return "CommandFlow.OrderContext"
}
func (_this OrderContext) ParentTraits_() []*_dafny.TraitID {
	return [](*_dafny.TraitID){}
}

var _ _dafny.TraitOffspring = OrderContext{}

// End of datatype OrderContext

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
			return "CommandFlow.OrderDecision.Move" + "(" + _dafny.String(data.Target) + ")"
		}
	case OrderDecision_NoChange:
		{
			return "CommandFlow.OrderDecision.NoChange"
		}
	case OrderDecision_Denied:
		{
			return "CommandFlow.OrderDecision.Denied" + "(" + data.Message.VerbatimString(true) + ")"
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
	return "CommandFlow.OrderDecision"
}
func (_this OrderDecision) ParentTraits_() []*_dafny.TraitID {
	return [](*_dafny.TraitID){}
}

var _ _dafny.TraitOffspring = OrderDecision{}

// End of datatype OrderDecision

// Definition of datatype Step
type Step struct {
	Data_Step_
}

func (_this Step) Get_() Data_Step_ {
	return _this.Data_Step_
}

type Data_Step_ interface {
	isStep()
}

type CompanionStruct_Step_ struct {
}

var Companion_Step_ = CompanionStruct_Step_{}

type Step_Ok struct {
}

func (Step_Ok) isStep() {}

func (CompanionStruct_Step_) Create_Ok_() Step {
	return Step{Step_Ok{}}
}

func (_this Step) Is_Ok() bool {
	_, ok := _this.Get_().(Step_Ok)
	return ok
}

type Step_Failed struct {
}

func (Step_Failed) isStep() {}

func (CompanionStruct_Step_) Create_Failed_() Step {
	return Step{Step_Failed{}}
}

func (_this Step) Is_Failed() bool {
	_, ok := _this.Get_().(Step_Failed)
	return ok
}

type Step_Conflict struct {
}

func (Step_Conflict) isStep() {}

func (CompanionStruct_Step_) Create_Conflict_() Step {
	return Step{Step_Conflict{}}
}

func (_this Step) Is_Conflict() bool {
	_, ok := _this.Get_().(Step_Conflict)
	return ok
}

type Step_Rejected struct {
}

func (Step_Rejected) isStep() {}

func (CompanionStruct_Step_) Create_Rejected_() Step {
	return Step{Step_Rejected{}}
}

func (_this Step) Is_Rejected() bool {
	_, ok := _this.Get_().(Step_Rejected)
	return ok
}

type Step_Skip struct {
}

func (Step_Skip) isStep() {}

func (CompanionStruct_Step_) Create_Skip_() Step {
	return Step{Step_Skip{}}
}

func (_this Step) Is_Skip() bool {
	_, ok := _this.Get_().(Step_Skip)
	return ok
}

func (CompanionStruct_Step_) Default() Step {
	return Companion_Step_.Create_Ok_()
}

func (_ CompanionStruct_Step_) AllSingletonConstructors() _dafny.Iterator {
	i := -1
	return func() (interface{}, bool) {
		i++
		switch i {
		case 0:
			return Companion_Step_.Create_Ok_(), true
		case 1:
			return Companion_Step_.Create_Failed_(), true
		case 2:
			return Companion_Step_.Create_Conflict_(), true
		case 3:
			return Companion_Step_.Create_Rejected_(), true
		case 4:
			return Companion_Step_.Create_Skip_(), true
		default:
			return Step{}, false
		}
	}
}

func (_this Step) String() string {
	switch _this.Get_().(type) {
	case nil:
		return "null"
	case Step_Ok:
		{
			return "CommandFlow.Step.Ok"
		}
	case Step_Failed:
		{
			return "CommandFlow.Step.Failed"
		}
	case Step_Conflict:
		{
			return "CommandFlow.Step.Conflict"
		}
	case Step_Rejected:
		{
			return "CommandFlow.Step.Rejected"
		}
	case Step_Skip:
		{
			return "CommandFlow.Step.Skip"
		}
	default:
		{
			return "<unexpected>"
		}
	}
}

func (_this Step) Equals(other Step) bool {
	switch _this.Get_().(type) {
	case Step_Ok:
		{
			_, ok := other.Get_().(Step_Ok)
			return ok
		}
	case Step_Failed:
		{
			_, ok := other.Get_().(Step_Failed)
			return ok
		}
	case Step_Conflict:
		{
			_, ok := other.Get_().(Step_Conflict)
			return ok
		}
	case Step_Rejected:
		{
			_, ok := other.Get_().(Step_Rejected)
			return ok
		}
	case Step_Skip:
		{
			_, ok := other.Get_().(Step_Skip)
			return ok
		}
	default:
		{
			return false // unexpected
		}
	}
}

func (_this Step) EqualsGeneric(other interface{}) bool {
	typed, ok := other.(Step)
	return ok && _this.Equals(typed)
}

func Type_Step_() _dafny.TypeDescriptor {
	return type_Step_{}
}

type type_Step_ struct {
}

func (_this type_Step_) Default() interface{} {
	return Companion_Step_.Default()
}

func (_this type_Step_) String() string {
	return "CommandFlow.Step"
}
func (_this Step) ParentTraits_() []*_dafny.TraitID {
	return [](*_dafny.TraitID){}
}

var _ _dafny.TraitOffspring = Step{}

// End of datatype Step

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
	HasCoupon          bool
	ExpectedVersion    _dafny.Int
	ProductValid       bool
	StockDelta         _dafny.Int
}

func (CommandInput_CommandInput) isCommandInput() {}

func (CompanionStruct_CommandInput_) Create_CommandInput_(Kind _dafny.Sequence, Role _dafny.Sequence, Actor _dafny.Sequence, Customer _dafny.Sequence, NoteBytes _dafny.Int, TrackingBytes _dafny.Int, RequestedDateBytes _dafny.Int, HasCoupon bool, ExpectedVersion _dafny.Int, ProductValid bool, StockDelta _dafny.Int) CommandInput {
	return CommandInput{CommandInput_CommandInput{Kind, Role, Actor, Customer, NoteBytes, TrackingBytes, RequestedDateBytes, HasCoupon, ExpectedVersion, ProductValid, StockDelta}}
}

func (_this CommandInput) Is_CommandInput() bool {
	_, ok := _this.Get_().(CommandInput_CommandInput)
	return ok
}

func (CompanionStruct_CommandInput_) Default() CommandInput {
	return Companion_CommandInput_.Create_CommandInput_(_dafny.EmptySeq, _dafny.EmptySeq, _dafny.EmptySeq, _dafny.EmptySeq, _dafny.Zero, _dafny.Zero, _dafny.Zero, false, _dafny.Zero, false, _dafny.Zero)
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

func (_this CommandInput) Dtor_hasCoupon() bool {
	return _this.Get_().(CommandInput_CommandInput).HasCoupon
}

func (_this CommandInput) Dtor_expectedVersion() _dafny.Int {
	return _this.Get_().(CommandInput_CommandInput).ExpectedVersion
}

func (_this CommandInput) Dtor_productValid() bool {
	return _this.Get_().(CommandInput_CommandInput).ProductValid
}

func (_this CommandInput) Dtor_stockDelta() _dafny.Int {
	return _this.Get_().(CommandInput_CommandInput).StockDelta
}

func (_this CommandInput) String() string {
	switch data := _this.Get_().(type) {
	case nil:
		return "null"
	case CommandInput_CommandInput:
		{
			return "CommandFlow.CommandInput.CommandInput" + "(" + data.Kind.VerbatimString(true) + ", " + data.Role.VerbatimString(true) + ", " + data.Actor.VerbatimString(true) + ", " + data.Customer.VerbatimString(true) + ", " + _dafny.String(data.NoteBytes) + ", " + _dafny.String(data.TrackingBytes) + ", " + _dafny.String(data.RequestedDateBytes) + ", " + _dafny.String(data.HasCoupon) + ", " + _dafny.String(data.ExpectedVersion) + ", " + _dafny.String(data.ProductValid) + ", " + _dafny.String(data.StockDelta) + ")"
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
			return ok && data1.Kind.Equals(data2.Kind) && data1.Role.Equals(data2.Role) && data1.Actor.Equals(data2.Actor) && data1.Customer.Equals(data2.Customer) && data1.NoteBytes.Cmp(data2.NoteBytes) == 0 && data1.TrackingBytes.Cmp(data2.TrackingBytes) == 0 && data1.RequestedDateBytes.Cmp(data2.RequestedDateBytes) == 0 && data1.HasCoupon == data2.HasCoupon && data1.ExpectedVersion.Cmp(data2.ExpectedVersion) == 0 && data1.ProductValid == data2.ProductValid && data1.StockDelta.Cmp(data2.StockDelta) == 0
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
	return "CommandFlow.CommandInput"
}
func (_this CommandInput) ParentTraits_() []*_dafny.TraitID {
	return [](*_dafny.TraitID){}
}

var _ _dafny.TraitOffspring = CommandInput{}

// End of datatype CommandInput

// Definition of datatype StockLine
type StockLine struct {
	Data_StockLine_
}

func (_this StockLine) Get_() Data_StockLine_ {
	return _this.Data_StockLine_
}

type Data_StockLine_ interface {
	isStockLine()
}

type CompanionStruct_StockLine_ struct {
}

var Companion_StockLine_ = CompanionStruct_StockLine_{}

type StockLine_StockLine struct {
	Product  _dafny.Sequence
	Quantity _dafny.Int
}

func (StockLine_StockLine) isStockLine() {}

func (CompanionStruct_StockLine_) Create_StockLine_(Product _dafny.Sequence, Quantity _dafny.Int) StockLine {
	return StockLine{StockLine_StockLine{Product, Quantity}}
}

func (_this StockLine) Is_StockLine() bool {
	_, ok := _this.Get_().(StockLine_StockLine)
	return ok
}

func (CompanionStruct_StockLine_) Default() StockLine {
	return Companion_StockLine_.Create_StockLine_(_dafny.EmptySeq, _dafny.Zero)
}

func (_this StockLine) Dtor_product() _dafny.Sequence {
	return _this.Get_().(StockLine_StockLine).Product
}

func (_this StockLine) Dtor_quantity() _dafny.Int {
	return _this.Get_().(StockLine_StockLine).Quantity
}

func (_this StockLine) String() string {
	switch data := _this.Get_().(type) {
	case nil:
		return "null"
	case StockLine_StockLine:
		{
			return "CommandFlow.StockLine.StockLine" + "(" + data.Product.VerbatimString(true) + ", " + _dafny.String(data.Quantity) + ")"
		}
	default:
		{
			return "<unexpected>"
		}
	}
}

func (_this StockLine) Equals(other StockLine) bool {
	switch data1 := _this.Get_().(type) {
	case StockLine_StockLine:
		{
			data2, ok := other.Get_().(StockLine_StockLine)
			return ok && data1.Product.Equals(data2.Product) && data1.Quantity.Cmp(data2.Quantity) == 0
		}
	default:
		{
			return false // unexpected
		}
	}
}

func (_this StockLine) EqualsGeneric(other interface{}) bool {
	typed, ok := other.(StockLine)
	return ok && _this.Equals(typed)
}

func Type_StockLine_() _dafny.TypeDescriptor {
	return type_StockLine_{}
}

type type_StockLine_ struct {
}

func (_this type_StockLine_) Default() interface{} {
	return Companion_StockLine_.Default()
}

func (_this type_StockLine_) String() string {
	return "CommandFlow.StockLine"
}
func (_this StockLine) ParentTraits_() []*_dafny.TraitID {
	return [](*_dafny.TraitID){}
}

var _ _dafny.TraitOffspring = StockLine{}

// End of datatype StockLine

// Definition of datatype StockAction
type StockAction struct {
	Data_StockAction_
}

func (_this StockAction) Get_() Data_StockAction_ {
	return _this.Data_StockAction_
}

type Data_StockAction_ interface {
	isStockAction()
}

type CompanionStruct_StockAction_ struct {
}

var Companion_StockAction_ = CompanionStruct_StockAction_{}

type StockAction_Reserve struct {
}

func (StockAction_Reserve) isStockAction() {}

func (CompanionStruct_StockAction_) Create_Reserve_() StockAction {
	return StockAction{StockAction_Reserve{}}
}

func (_this StockAction) Is_Reserve() bool {
	_, ok := _this.Get_().(StockAction_Reserve)
	return ok
}

type StockAction_Release struct {
}

func (StockAction_Release) isStockAction() {}

func (CompanionStruct_StockAction_) Create_Release_() StockAction {
	return StockAction{StockAction_Release{}}
}

func (_this StockAction) Is_Release() bool {
	_, ok := _this.Get_().(StockAction_Release)
	return ok
}

type StockAction_Ship struct {
}

func (StockAction_Ship) isStockAction() {}

func (CompanionStruct_StockAction_) Create_Ship_() StockAction {
	return StockAction{StockAction_Ship{}}
}

func (_this StockAction) Is_Ship() bool {
	_, ok := _this.Get_().(StockAction_Ship)
	return ok
}

func (CompanionStruct_StockAction_) Default() StockAction {
	return Companion_StockAction_.Create_Reserve_()
}

func (_ CompanionStruct_StockAction_) AllSingletonConstructors() _dafny.Iterator {
	i := -1
	return func() (interface{}, bool) {
		i++
		switch i {
		case 0:
			return Companion_StockAction_.Create_Reserve_(), true
		case 1:
			return Companion_StockAction_.Create_Release_(), true
		case 2:
			return Companion_StockAction_.Create_Ship_(), true
		default:
			return StockAction{}, false
		}
	}
}

func (_this StockAction) String() string {
	switch _this.Get_().(type) {
	case nil:
		return "null"
	case StockAction_Reserve:
		{
			return "CommandFlow.StockAction.Reserve"
		}
	case StockAction_Release:
		{
			return "CommandFlow.StockAction.Release"
		}
	case StockAction_Ship:
		{
			return "CommandFlow.StockAction.Ship"
		}
	default:
		{
			return "<unexpected>"
		}
	}
}

func (_this StockAction) Equals(other StockAction) bool {
	switch _this.Get_().(type) {
	case StockAction_Reserve:
		{
			_, ok := other.Get_().(StockAction_Reserve)
			return ok
		}
	case StockAction_Release:
		{
			_, ok := other.Get_().(StockAction_Release)
			return ok
		}
	case StockAction_Ship:
		{
			_, ok := other.Get_().(StockAction_Ship)
			return ok
		}
	default:
		{
			return false // unexpected
		}
	}
}

func (_this StockAction) EqualsGeneric(other interface{}) bool {
	typed, ok := other.(StockAction)
	return ok && _this.Equals(typed)
}

func Type_StockAction_() _dafny.TypeDescriptor {
	return type_StockAction_{}
}

type type_StockAction_ struct {
}

func (_this type_StockAction_) Default() interface{} {
	return Companion_StockAction_.Default()
}

func (_this type_StockAction_) String() string {
	return "CommandFlow.StockAction"
}
func (_this StockAction) ParentTraits_() []*_dafny.TraitID {
	return [](*_dafny.TraitID){}
}

var _ _dafny.TraitOffspring = StockAction{}

// End of datatype StockAction

// Definition of trait Effects
type Effects interface {
	String() string
	Begin() Step
	LoadCommand() (Step, _dafny.Sequence)
	DecodePayload() Step
	LoadActor() Step
	ReadClock() Step
	Input() CommandInput
	LoadCart(customer _dafny.Sequence) (Step, _dafny.Int)
	ExpireCart() (Step, _dafny.Sequence)
	SetCart() (Step, _dafny.Sequence)
	SaveCart() Step
	CheckoutCart() (Step, _dafny.Sequence)
	ReadShippingAddress() Step
	RequestedDate() (bool, bool, _dafny.Int, _dafny.Int)
	LoadCoupon() (Step, _dafny.Int)
	LoadProduct(product _dafny.Sequence) Step
	AddOrderLine(line StockLine) Step
	BuildOrder(percent _dafny.Int) Step
	LoadCheckoutVersion() Step
	LoadAdjustment() (Step, _dafny.Int)
	AdjustStock() Step
	SaveAdjustment() Step
	LoadOrder() (Step, OrderContext, _dafny.Int)
	Noop()
	Reject(message _dafny.Sequence) Step
	AssignWorker() Step
	PackOrder() Step
	AssignCarrier() Step
	MoveOrder(target _dafny.Int) (Step, _dafny.Sequence)
	SaveOrder() Step
	LoadStock(product _dafny.Sequence) Step
	ReserveStock(delta _dafny.Int) Step
	ShipStock(quantity _dafny.Int) Step
	SaveStock(product _dafny.Sequence) Step
	InsertPointAward() Step
	LoadLoyalty() Step
	CreditLoyalty() Step
	SaveLoyalty() Step
	RenderInvoice() Step
	InsertInvoice() Step
	SaveCompletion() Step
	Commit() Step
	Rollback() Step
	SaveRejection() Step
	RecordFailure(maxAttempts _dafny.Int)
}
type CompanionStruct_Effects_ struct {
	TraitID_ *_dafny.TraitID
}

var Companion_Effects_ = CompanionStruct_Effects_{
	TraitID_: &_dafny.TraitID{},
}

func (CompanionStruct_Effects_) CastTo_(x interface{}) Effects {
	var t Effects
	t, _ = x.(Effects)
	return t
}

// End of trait Effects
