// Package CommandFlow
// Dafny module CommandFlow compiled into Go

package CommandFlow

import (
	os "os"

	m_CartCommands "wholesale.local/orbit/generated/CartCommands"
	m_CommandModel "wholesale.local/orbit/generated/CommandModel"
	m_Domain "wholesale.local/orbit/generated/Domain"
	m_OrderCommands "wholesale.local/orbit/generated/OrderCommands"
	m_StockCommands "wholesale.local/orbit/generated/StockCommands"
	m__System "wholesale.local/orbit/generated/System_"
	_dafny "wholesale.local/orbit/generated/dafny"
)

var _ = os.Args
var _ _dafny.Dummy__
var _ m__System.Dummy__
var _ m_Domain.Dummy__
var _ m_CommandModel.Dummy__
var _ m_CartCommands.Dummy__
var _ m_OrderCommands.Dummy__
var _ m_StockCommands.Dummy__

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

func (_static *CompanionStruct_Default___) DecideCommand(input m_CommandModel.CommandInput, snapshot m_CommandModel.Snapshot) m_CommandModel.CommandPlan {
	var _source0 m_CommandModel.ReadRequest = Companion_Default___.RequestState(input)
	_ = _source0
	{
		if _source0.Is_RejectRead() {
			var _0_reason _dafny.Sequence = _source0.Get_().(m_CommandModel.ReadRequest_RejectRead).Reason
			_ = _0_reason
			return m_CommandModel.Companion_CommandPlan_.Create_Reject_(_0_reason)
		}
	}
	{
		if _source0.Is_ReadCart() {
			var _1_customer _dafny.Sequence = _source0.Get_().(m_CommandModel.ReadRequest_ReadCart).CustomerId
			_ = _1_customer
			if (snapshot).Is_Missing() {
				return m_CommandModel.Companion_CommandPlan_.Create_Reject_((snapshot).Dtor_message())
			} else if !((snapshot).Is_Cart()) {
				return m_CommandModel.Companion_CommandPlan_.Create_Reject_(_dafny.UnicodeSeqOfUtf8Bytes("カートが存在しません"))
			} else if !_dafny.Companion_Sequence_.Equal(((snapshot).Dtor_state()).Dtor_customerId(), _1_customer) {
				return m_CommandModel.Companion_CommandPlan_.Create_Reject_(_dafny.UnicodeSeqOfUtf8Bytes("カート所有者が一致しません"))
			} else if _dafny.Companion_Sequence_.Equal((input).Dtor_kind(), _dafny.UnicodeSeqOfUtf8Bytes("cart.set")) {
				return m_CartCommands.Companion_Default___.SetCart(input, (snapshot).Dtor_state())
			} else if _dafny.Companion_Sequence_.Equal((input).Dtor_kind(), _dafny.UnicodeSeqOfUtf8Bytes("cart.expire")) {
				return m_CartCommands.Companion_Default___.ExpireCart(input, (snapshot).Dtor_state())
			} else {
				return m_CartCommands.Companion_Default___.CheckoutOrder(input, (snapshot).Dtor_state())
			}
		}
	}
	{
		if _source0.Is_ReadStock() {
			var _2_product _dafny.Sequence = _source0.Get_().(m_CommandModel.ReadRequest_ReadStock).ProductId
			_ = _2_product
			if (snapshot).Is_Missing() {
				return m_CommandModel.Companion_CommandPlan_.Create_Reject_((snapshot).Dtor_message())
			} else if !((snapshot).Is_Stock()) {
				return m_CommandModel.Companion_CommandPlan_.Create_Reject_(_dafny.UnicodeSeqOfUtf8Bytes("商品が存在しません"))
			} else {
				return m_StockCommands.Companion_Default___.AdjustStock(input, _2_product, (snapshot).Dtor_stockVersion())
			}
		}
	}
	{
		if (snapshot).Is_Missing() {
			return m_CommandModel.Companion_CommandPlan_.Create_Reject_((snapshot).Dtor_message())
		} else if !((snapshot).Is_Order()) {
			return m_CommandModel.Companion_CommandPlan_.Create_Reject_(_dafny.UnicodeSeqOfUtf8Bytes("受注が存在しません"))
		} else {
			return m_OrderCommands.Companion_Default___.PlanOrder(input, (snapshot).Dtor_orderVersion(), (snapshot).Dtor_order(), (snapshot).Dtor_lines())
		}
	}
}
func (_static *CompanionStruct_Default___) RequestState(input m_CommandModel.CommandInput) m_CommandModel.ReadRequest {
	if !(m_CommandModel.Companion_Default___.InputSizeValid(input)) {
		return m_CommandModel.Companion_ReadRequest_.Create_RejectRead_(_dafny.UnicodeSeqOfUtf8Bytes("入力が長すぎます"))
	} else if (_dafny.SetOf(_dafny.UnicodeSeqOfUtf8Bytes("cart.set"), _dafny.UnicodeSeqOfUtf8Bytes("cart.expire"), _dafny.UnicodeSeqOfUtf8Bytes("order.checkout"))).Contains((input).Dtor_kind()) {
		var _0_role _dafny.Sequence = (func() _dafny.Sequence {
			if _dafny.Companion_Sequence_.Equal((input).Dtor_kind(), _dafny.UnicodeSeqOfUtf8Bytes("cart.expire")) {
				return _dafny.UnicodeSeqOfUtf8Bytes("system")
			}
			return _dafny.UnicodeSeqOfUtf8Bytes("customer")
		})()
		_ = _0_role
		if !_dafny.Companion_Sequence_.Equal((input).Dtor_role(), _0_role) {
			return m_CommandModel.Companion_ReadRequest_.Create_RejectRead_(_dafny.UnicodeSeqOfUtf8Bytes("この利用者には操作権限がありません"))
		} else {
			return m_CommandModel.Companion_ReadRequest_.Create_ReadCart_((func() _dafny.Sequence {
				if _dafny.Companion_Sequence_.Equal((input).Dtor_kind(), _dafny.UnicodeSeqOfUtf8Bytes("cart.expire")) {
					return (input).Dtor_customer()
				}
				return (input).Dtor_actor()
			})())
		}
	} else if _dafny.Companion_Sequence_.Equal((input).Dtor_kind(), _dafny.UnicodeSeqOfUtf8Bytes("stock.adjust")) {
		if !_dafny.Companion_Sequence_.Equal((input).Dtor_role(), _dafny.UnicodeSeqOfUtf8Bytes("operator")) {
			return m_CommandModel.Companion_ReadRequest_.Create_RejectRead_(_dafny.UnicodeSeqOfUtf8Bytes("この利用者には操作権限がありません"))
		} else if ((!((input).Dtor_productValid())) || (!(m_Domain.Companion_Default___.StockDeltaValid((input).Dtor_stockDelta())))) || (((input).Dtor_noteBytes()).Sign() == 0) {
			return m_CommandModel.Companion_ReadRequest_.Create_RejectRead_(_dafny.UnicodeSeqOfUtf8Bytes("商品・調整数・理由を入力してください"))
		} else {
			return m_CommandModel.Companion_ReadRequest_.Create_ReadStock_((input).Dtor_product())
		}
	} else if (_dafny.SetOf(_dafny.UnicodeSeqOfUtf8Bytes("order.pay"), _dafny.UnicodeSeqOfUtf8Bytes("order.cancel"), _dafny.UnicodeSeqOfUtf8Bytes("order.pick"), _dafny.UnicodeSeqOfUtf8Bytes("order.block"), _dafny.UnicodeSeqOfUtf8Bytes("order.pack"), _dafny.UnicodeSeqOfUtf8Bytes("order.dispatch"), _dafny.UnicodeSeqOfUtf8Bytes("order.delivery_failed"), _dafny.UnicodeSeqOfUtf8Bytes("order.redeliver"), _dafny.UnicodeSeqOfUtf8Bytes("order.deliver"), _dafny.UnicodeSeqOfUtf8Bytes("order.return"))).Contains((input).Dtor_kind()) {
		return m_CommandModel.Companion_ReadRequest_.Create_ReadOrder_((input).Dtor_orderId())
	} else {
		return m_CommandModel.Companion_ReadRequest_.Create_RejectRead_(_dafny.UnicodeSeqOfUtf8Bytes("未対応のCommandです"))
	}
}
func (_static *CompanionStruct_Default___) DependenciesFor(plan m_CommandModel.CommandPlan) m_CommandModel.Dependencies {
	var _source0 m_CommandModel.CommandPlan = plan
	_ = _source0
	{
		if _source0.Is_Checkout() {
			var _0_orderId _dafny.Sequence = _source0.Get_().(m_CommandModel.CommandPlan_Checkout).NewOrderId
			_ = _0_orderId
			var _1_lines _dafny.Map = _source0.Get_().(m_CommandModel.CommandPlan_Checkout).OrderLines
			_ = _1_lines
			return m_CommandModel.Companion_Dependencies_.Create_Dependencies_((_1_lines).Keys(), _dafny.SetOf(), m_CommandModel.Companion_Reference_.Create_Absent_(), m_CommandModel.Companion_Reference_.Create_Id_(_0_orderId))
		}
	}
	{
		if _source0.Is_UpdateCart() {
			var _2_delta _dafny.Map = _source0.Get_().(m_CommandModel.CommandPlan_UpdateCart).Reservations
			_ = _2_delta
			return m_CommandModel.Companion_Dependencies_.Create_Dependencies_(_dafny.SetOf(), (_2_delta).Keys(), m_CommandModel.Companion_Reference_.Create_Absent_(), m_CommandModel.Companion_Reference_.Create_Absent_())
		}
	}
	{
		if _source0.Is_UpdateOrder() {
			var _3_target _dafny.Int = _source0.Get_().(m_CommandModel.CommandPlan_UpdateOrder).Target
			_ = _3_target
			var _4_customer _dafny.Sequence = _source0.Get_().(m_CommandModel.CommandPlan_UpdateOrder).OrderCustomer
			_ = _4_customer
			var _5_stock m_CommandModel.StockChanges = _source0.Get_().(m_CommandModel.CommandPlan_UpdateOrder).Stock
			_ = _5_stock
			var _6_ids _dafny.Set = func() _dafny.Set {
				var _source1 m_CommandModel.StockChanges = _5_stock
				_ = _source1
				{
					if _source1.Is_Reserve() {
						var _7_delta _dafny.Map = _source1.Get_().(m_CommandModel.StockChanges_Reserve).Deltas
						_ = _7_delta
						return (_7_delta).Keys()
					}
				}
				{
					var _8_quantities _dafny.Map = _source1.Get_().(m_CommandModel.StockChanges_Ship).Quantities
					_ = _8_quantities
					return (_8_quantities).Keys()
				}
			}()
			_ = _6_ids
			return m_CommandModel.Companion_Dependencies_.Create_Dependencies_(_dafny.SetOf(), _6_ids, (func() m_CommandModel.Reference {
				if (_3_target).Cmp(m_CommandModel.Companion_Default___.Delivered()) == 0 {
					return m_CommandModel.Companion_Reference_.Create_Id_(_4_customer)
				}
				return m_CommandModel.Companion_Reference_.Create_Absent_()
			})(), m_CommandModel.Companion_Reference_.Create_Absent_())
		}
	}
	{
		return m_CommandModel.Companion_Dependencies_.Create_Dependencies_(_dafny.SetOf(), _dafny.SetOf(), m_CommandModel.Companion_Reference_.Create_Absent_(), m_CommandModel.Companion_Reference_.Create_Absent_())
	}
}
func (_static *CompanionStruct_Default___) DecideAndApply(e Effects) m_CommandModel.Status {
	var status m_CommandModel.Status = m_CommandModel.Companion_Status_.Default()
	_ = status
	var _0_input m_CommandModel.CommandInput = m_CommandModel.Companion_CommandInput_.Default()
	_ = _0_input
	var _out0 m_CommandModel.Status
	_ = _out0
	var _out1 m_CommandModel.CommandInput
	_ = _out1
	_out0, _out1 = (e).ReadInput()
	status = _out0
	_0_input = _out1
	if !((status).Is_Ok()) {
		return status
	}
	var _1_request m_CommandModel.ReadRequest
	_ = _1_request
	_1_request = Companion_Default___.RequestState(_0_input)
	if (_1_request).Is_RejectRead() {
		(e).SetRejection((_1_request).Dtor_reason())
		status = m_CommandModel.Companion_Status_.Create_Rejected_()
		return status
	}
	var _2_snapshot m_CommandModel.Snapshot = m_CommandModel.Companion_Snapshot_.Default()
	_ = _2_snapshot
	var _out2 m_CommandModel.Status
	_ = _out2
	var _out3 m_CommandModel.Snapshot
	_ = _out3
	_out2, _out3 = (e).Read(_1_request)
	status = _out2
	_2_snapshot = _out3
	if !((status).Is_Ok()) {
		return status
	}
	if m_CartCommands.Companion_Default___.NeedsCoupon(_0_input, _2_snapshot) {
		var _3_coupon m_CommandModel.Coupon = m_CommandModel.Companion_Coupon_.Default()
		_ = _3_coupon
		var _out4 m_CommandModel.Status
		_ = _out4
		var _out5 m_CommandModel.Coupon
		_ = _out5
		_out4, _out5 = (e).ReadCoupon((_0_input).Dtor_couponCode())
		status = _out4
		_3_coupon = _out5
		if !((status).Is_Ok()) {
			return status
		}
		var _pat_let_tv0 = _3_coupon
		_ = _pat_let_tv0
		_2_snapshot = m_CommandModel.Companion_Snapshot_.Create_Cart_(func(_pat_let0_0 m_CommandModel.CartState) m_CommandModel.CartState {
			return func(_4_dt__update__tmp_h0 m_CommandModel.CartState) m_CommandModel.CartState {
				return func(_pat_let1_0 m_CommandModel.Coupon) m_CommandModel.CartState {
					return func(_5_dt__update_hcoupon_h0 m_CommandModel.Coupon) m_CommandModel.CartState {
						return m_CommandModel.Companion_CartState_.Create_CartState_((_4_dt__update__tmp_h0).Dtor_customerId(), (_4_dt__update__tmp_h0).Dtor_version(), (_4_dt__update__tmp_h0).Dtor_items(), (_4_dt__update__tmp_h0).Dtor_expires(), (_4_dt__update__tmp_h0).Dtor_requestedDay(), (_4_dt__update__tmp_h0).Dtor_today(), _5_dt__update_hcoupon_h0)
					}(_pat_let1_0)
				}(_pat_let_tv0)
			}(_pat_let0_0)
		}((_2_snapshot).Dtor_state()))
	}
	var _6_plan m_CommandModel.CommandPlan
	_ = _6_plan
	_6_plan = Companion_Default___.DecideCommand(_0_input, _2_snapshot)
	if (_6_plan).Is_Reject() {
		(e).SetRejection((_6_plan).Dtor_reason())
		status = m_CommandModel.Companion_Status_.Create_Rejected_()
		return status
	}
	var _7_dependencies m_CommandModel.Dependencies
	_ = _7_dependencies
	_7_dependencies = Companion_Default___.DependenciesFor(_6_plan)
	if !(_7_dependencies).Equals(m_CommandModel.Companion_Dependencies_.Create_Dependencies_(_dafny.SetOf(), _dafny.SetOf(), m_CommandModel.Companion_Reference_.Create_Absent_(), m_CommandModel.Companion_Reference_.Create_Absent_())) {
		var _out6 m_CommandModel.Status
		_ = _out6
		_out6 = (e).ReadDependencies(_7_dependencies)
		status = _out6
		if !((status).Is_Ok()) {
			return status
		}
	}
	var _out7 m_CommandModel.Status
	_ = _out7
	_out7 = (e).Apply(_6_plan)
	status = _out7
	return status
}
func (_static *CompanionStruct_Default___) ExecuteAttempt(e Effects) m_CommandModel.Status {
	var status m_CommandModel.Status = m_CommandModel.Companion_Status_.Default()
	_ = status
	var _out0 m_CommandModel.Status
	_ = _out0
	_out0 = (e).Begin()
	status = _out0
	if !((status).Is_Ok()) {
		return status
	}
	var _0_state _dafny.Sequence = _dafny.EmptySeq
	_ = _0_state
	var _out1 m_CommandModel.Status
	_ = _out1
	var _out2 _dafny.Sequence
	_ = _out2
	_out1, _out2 = (e).LockCommand()
	status = _out1
	_0_state = _out2
	if (status).Is_Ok() {
		if _dafny.Companion_Sequence_.Equal(_0_state, _dafny.UnicodeSeqOfUtf8Bytes("queued")) {
			status = m_CommandModel.Companion_Status_.Create_Ok_()
		} else {
			status = m_CommandModel.Companion_Status_.Create_Skipped_()
		}
		if (status).Is_Ok() {
			var _out3 m_CommandModel.Status
			_ = _out3
			_out3 = Companion_Default___.DecideAndApply(e)
			status = _out3
		}
		if (status).Is_Ok() {
			var _out4 m_CommandModel.Status
			_ = _out4
			_out4 = (e).RecordSuccess()
			status = _out4
		} else if (status).Is_Rejected() {
			var _out5 m_CommandModel.Status
			_ = _out5
			_out5 = (e).RecordRejection()
			status = _out5
		}
		if (status).Is_Ok() {
			var _out6 m_CommandModel.Status
			_ = _out6
			_out6 = (e).Commit()
			status = _out6
			return status
		}
	}
	(e).Rollback()
	return status
}
func (_static *CompanionStruct_Default___) ExecuteCommand(e Effects) m_CommandModel.Status {
	var status m_CommandModel.Status = m_CommandModel.Companion_Status_.Default()
	_ = status
	var _hi0 _dafny.Int = _dafny.IntOfInt64(5)
	_ = _hi0
	for _0_attempt := _dafny.Zero; _0_attempt.Cmp(_hi0) < 0; _0_attempt = _0_attempt.Plus(_dafny.One) {
		var _out0 m_CommandModel.Status
		_ = _out0
		_out0 = Companion_Default___.ExecuteAttempt(e)
		status = _out0
		if !((status).Is_Conflict()) {
			if (status).Is_Failed() {
				(e).RecordFailure(_dafny.IntOfInt64(8))
			}
			return status
		}
	}
	status = m_CommandModel.Companion_Status_.Create_Conflict_()
	return status
}

// End of class Default__

// Definition of trait Effects
type Effects interface {
	String() string
	Begin() m_CommandModel.Status
	LockCommand() (m_CommandModel.Status, _dafny.Sequence)
	ReadInput() (m_CommandModel.Status, m_CommandModel.CommandInput)
	Read(request m_CommandModel.ReadRequest) (m_CommandModel.Status, m_CommandModel.Snapshot)
	ReadCoupon(code _dafny.Sequence) (m_CommandModel.Status, m_CommandModel.Coupon)
	ReadDependencies(dependencies m_CommandModel.Dependencies) m_CommandModel.Status
	Apply(plan m_CommandModel.CommandPlan) m_CommandModel.Status
	RecordSuccess() m_CommandModel.Status
	RecordRejection() m_CommandModel.Status
	SetRejection(reason _dafny.Sequence)
	Commit() m_CommandModel.Status
	Rollback()
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
