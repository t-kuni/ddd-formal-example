// Package OrderCommands
// Dafny module OrderCommands compiled into Go

package OrderCommands

import (
	os "os"

	m_CartCommands "wholesale.local/orbit/generated/CartCommands"
	m_CommandModel "wholesale.local/orbit/generated/CommandModel"
	m_Domain "wholesale.local/orbit/generated/Domain"
	m__System "wholesale.local/orbit/generated/System_"
	_dafny "wholesale.local/orbit/generated/dafny"
)

var _ = os.Args
var _ _dafny.Dummy__
var _ m__System.Dummy__
var _ m_Domain.Dummy__
var _ m_CommandModel.Dummy__
var _ m_CartCommands.Dummy__

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
	return "OrderCommands.Default__"
}
func (_this *Default__) ParentTraits_() []*_dafny.TraitID {
	return [](*_dafny.TraitID){}
}

var _ _dafny.TraitOffspring = &Default__{}

func (_static *CompanionStruct_Default___) EligibleDelivery(input m_CommandModel.CommandInput, version _dafny.Int, order m_CommandModel.OrderState) bool {
	return ((((_dafny.Companion_Sequence_.Equal((input).Dtor_kind(), _dafny.UnicodeSeqOfUtf8Bytes("order.deliver"))) && (_dafny.Companion_Sequence_.Equal((input).Dtor_role(), _dafny.UnicodeSeqOfUtf8Bytes("carrier")))) && (_dafny.Companion_Sequence_.Equal((input).Dtor_actor(), (order).Dtor_carrier()))) && (((order).Dtor_status()).Cmp(m_CommandModel.Companion_Default___.Shipping()) == 0)) && (m_CommandModel.Companion_Default___.VersionMatches((input).Dtor_expectedVersion(), version))
}
func (_static *CompanionStruct_Default___) PlanOrder(input m_CommandModel.CommandInput, version _dafny.Int, order m_CommandModel.OrderState, lines _dafny.Map) m_CommandModel.CommandPlan {
	if _dafny.Companion_Sequence_.Equal((input).Dtor_kind(), _dafny.UnicodeSeqOfUtf8Bytes("order.pay")) {
		return Companion_Default___.PayOrder(input, version, order, lines)
	} else if _dafny.Companion_Sequence_.Equal((input).Dtor_kind(), _dafny.UnicodeSeqOfUtf8Bytes("order.cancel")) {
		return Companion_Default___.CancelOrder(input, version, order, lines)
	} else if _dafny.Companion_Sequence_.Equal((input).Dtor_kind(), _dafny.UnicodeSeqOfUtf8Bytes("order.pick")) {
		return Companion_Default___.PickOrder(input, version, order, lines)
	} else if _dafny.Companion_Sequence_.Equal((input).Dtor_kind(), _dafny.UnicodeSeqOfUtf8Bytes("order.block")) {
		return Companion_Default___.BlockOrder(input, version, order, lines)
	} else if _dafny.Companion_Sequence_.Equal((input).Dtor_kind(), _dafny.UnicodeSeqOfUtf8Bytes("order.pack")) {
		return Companion_Default___.PackOrder(input, version, order, lines)
	} else if _dafny.Companion_Sequence_.Equal((input).Dtor_kind(), _dafny.UnicodeSeqOfUtf8Bytes("order.dispatch")) {
		return Companion_Default___.DispatchOrder(input, version, order, lines)
	} else if _dafny.Companion_Sequence_.Equal((input).Dtor_kind(), _dafny.UnicodeSeqOfUtf8Bytes("order.delivery_failed")) {
		return Companion_Default___.FailDelivery(input, version, order, lines)
	} else if _dafny.Companion_Sequence_.Equal((input).Dtor_kind(), _dafny.UnicodeSeqOfUtf8Bytes("order.redeliver")) {
		return Companion_Default___.RedeliverOrder(input, version, order, lines)
	} else if _dafny.Companion_Sequence_.Equal((input).Dtor_kind(), _dafny.UnicodeSeqOfUtf8Bytes("order.deliver")) {
		return Companion_Default___.DeliverOrder(input, version, order, lines)
	} else if _dafny.Companion_Sequence_.Equal((input).Dtor_kind(), _dafny.UnicodeSeqOfUtf8Bytes("order.return")) {
		return Companion_Default___.ReturnOrder(input, version, order, lines)
	} else {
		return m_CommandModel.Companion_CommandPlan_.Create_Reject_(_dafny.UnicodeSeqOfUtf8Bytes("未対応のCommandです"))
	}
}
func (_static *CompanionStruct_Default___) PayOrder(input m_CommandModel.CommandInput, version _dafny.Int, order m_CommandModel.OrderState, lines _dafny.Map) m_CommandModel.CommandPlan {
	return Companion_Default___.OrderChange(input, version, order, lines, Companion_Default___.ForRole(input, _dafny.UnicodeSeqOfUtf8Bytes("customer"), m_CommandModel.Companion_OrderDecision_.Create_Move_((func() _dafny.Int {
		if (input).Dtor_paymentSucceeded() {
			return m_CommandModel.Companion_Default___.Accepted()
		}
		return m_CommandModel.Companion_Default___.PaymentFailed()
	})())), m_CommandModel.Companion_OrderPreparation_.Create_Keep_(), false)
}
func (_static *CompanionStruct_Default___) CancelOrder(input m_CommandModel.CommandInput, version _dafny.Int, order m_CommandModel.OrderState, lines _dafny.Map) m_CommandModel.CommandPlan {
	var _0_decision m_CommandModel.OrderDecision = (func() m_CommandModel.OrderDecision {
		if !(_dafny.SetOf(_dafny.UnicodeSeqOfUtf8Bytes("customer"), _dafny.UnicodeSeqOfUtf8Bytes("operator"), _dafny.UnicodeSeqOfUtf8Bytes("system"))).Contains((input).Dtor_role()) {
			return m_CommandModel.Companion_OrderDecision_.Create_Denied_(Companion_Default___.RoleDenied())
		}
		return (func() m_CommandModel.OrderDecision {
			if (_dafny.Companion_Sequence_.Equal((input).Dtor_role(), _dafny.UnicodeSeqOfUtf8Bytes("system"))) && ((((order).Dtor_status()).Cmp(m_CommandModel.Companion_Default___.PaymentPending()) != 0) || (((order).Dtor_ageNanos()).Cmp(m_CommandModel.Companion_Default___.PaymentTimeoutNanos()) < 0)) {
				return m_CommandModel.Companion_OrderDecision_.Create_NoChange_()
			}
			return (func() m_CommandModel.OrderDecision {
				if (_dafny.Companion_Sequence_.Equal((input).Dtor_role(), _dafny.UnicodeSeqOfUtf8Bytes("customer"))) && (!(_dafny.SetOf(m_CommandModel.Companion_Default___.PaymentPending(), m_CommandModel.Companion_Default___.Accepted())).Contains((order).Dtor_status())) {
					return m_CommandModel.Companion_OrderDecision_.Create_Denied_(_dafny.UnicodeSeqOfUtf8Bytes("作業開始後のキャンセルは運用者に依頼してください"))
				}
				return m_CommandModel.Companion_OrderDecision_.Create_Move_(m_CommandModel.Companion_Default___.Cancelled())
			})()
		})()
	})()
	_ = _0_decision
	return Companion_Default___.OrderChange(input, version, order, lines, _0_decision, m_CommandModel.Companion_OrderPreparation_.Create_Keep_(), false)
}
func (_static *CompanionStruct_Default___) PickOrder(input m_CommandModel.CommandInput, version _dafny.Int, order m_CommandModel.OrderState, lines _dafny.Map) m_CommandModel.CommandPlan {
	var _0_decision m_CommandModel.OrderDecision = (func() m_CommandModel.OrderDecision {
		if m_Domain.Companion_Default___.MayAssignWorker(!_dafny.Companion_Sequence_.Equal((order).Dtor_worker(), _dafny.UnicodeSeqOfUtf8Bytes("")), _dafny.Companion_Sequence_.Equal((order).Dtor_worker(), (input).Dtor_actor())) {
			return m_CommandModel.Companion_OrderDecision_.Create_Move_(m_CommandModel.Companion_Default___.Picking())
		}
		return m_CommandModel.Companion_OrderDecision_.Create_Denied_(_dafny.UnicodeSeqOfUtf8Bytes("別の担当者が作業中です"))
	})()
	_ = _0_decision
	return Companion_Default___.OrderChange(input, version, order, lines, Companion_Default___.ForRole(input, _dafny.UnicodeSeqOfUtf8Bytes("warehouse"), _0_decision), m_CommandModel.Companion_OrderPreparation_.Create_AssignWorker_((input).Dtor_actor()), false)
}
func (_static *CompanionStruct_Default___) BlockOrder(input m_CommandModel.CommandInput, version _dafny.Int, order m_CommandModel.OrderState, lines _dafny.Map) m_CommandModel.CommandPlan {
	var _0_decision m_CommandModel.OrderDecision = (func() m_CommandModel.OrderDecision {
		if ((input).Dtor_noteBytes()).Sign() == 0 {
			return m_CommandModel.Companion_OrderDecision_.Create_Denied_(_dafny.UnicodeSeqOfUtf8Bytes("保留理由が必要です"))
		}
		return m_CommandModel.Companion_OrderDecision_.Create_Move_(m_CommandModel.Companion_Default___.PickBlocked())
	})()
	_ = _0_decision
	return Companion_Default___.OrderChange(input, version, order, lines, Companion_Default___.ForWorker(input, order, _0_decision), m_CommandModel.Companion_OrderPreparation_.Create_Keep_(), false)
}
func (_static *CompanionStruct_Default___) PackOrder(input m_CommandModel.CommandInput, version _dafny.Int, order m_CommandModel.OrderState, lines _dafny.Map) m_CommandModel.CommandPlan {
	var _0_decision m_CommandModel.OrderDecision = (func() m_CommandModel.OrderDecision {
		if (m_Domain.Companion_Default___.BoxCountValid((input).Dtor_boxes())) && (m_Domain.Companion_Default___.WeightValid((input).Dtor_weight())) {
			return m_CommandModel.Companion_OrderDecision_.Create_Move_(m_CommandModel.Companion_Default___.Packed())
		}
		return m_CommandModel.Companion_OrderDecision_.Create_Denied_(_dafny.UnicodeSeqOfUtf8Bytes("箱数は1〜100、1箱の重量は1〜500000gです"))
	})()
	_ = _0_decision
	return Companion_Default___.OrderChange(input, version, order, lines, Companion_Default___.ForWorker(input, order, _0_decision), m_CommandModel.Companion_OrderPreparation_.Create_Pack_((input).Dtor_boxes(), (input).Dtor_weight()), false)
}
func (_static *CompanionStruct_Default___) DispatchOrder(input m_CommandModel.CommandInput, version _dafny.Int, order m_CommandModel.OrderState, lines _dafny.Map) m_CommandModel.CommandPlan {
	var _0_decision m_CommandModel.OrderDecision = (func() m_CommandModel.OrderDecision {
		if _dafny.Companion_Sequence_.Equal((input).Dtor_tracking(), _dafny.UnicodeSeqOfUtf8Bytes("")) {
			return m_CommandModel.Companion_OrderDecision_.Create_Denied_(_dafny.UnicodeSeqOfUtf8Bytes("追跡番号を入力してください"))
		}
		return (func() m_CommandModel.OrderDecision {
			if ((order).Dtor_status()).Cmp(m_CommandModel.Companion_Default___.Packed()) != 0 {
				return m_CommandModel.Companion_OrderDecision_.Create_Denied_(_dafny.UnicodeSeqOfUtf8Bytes("梱包完了後に出荷してください"))
			}
			return m_CommandModel.Companion_OrderDecision_.Create_Move_(m_CommandModel.Companion_Default___.Shipping())
		})()
	})()
	_ = _0_decision
	return Companion_Default___.OrderChange(input, version, order, lines, Companion_Default___.ForRole(input, _dafny.UnicodeSeqOfUtf8Bytes("carrier"), _0_decision), m_CommandModel.Companion_OrderPreparation_.Create_AssignCarrier_((input).Dtor_actor(), (input).Dtor_tracking()), true)
}
func (_static *CompanionStruct_Default___) FailDelivery(input m_CommandModel.CommandInput, version _dafny.Int, order m_CommandModel.OrderState, lines _dafny.Map) m_CommandModel.CommandPlan {
	var _0_decision m_CommandModel.OrderDecision = (func() m_CommandModel.OrderDecision {
		if ((input).Dtor_noteBytes()).Sign() == 0 {
			return m_CommandModel.Companion_OrderDecision_.Create_Denied_(_dafny.UnicodeSeqOfUtf8Bytes("配送失敗理由が必要です"))
		}
		return m_CommandModel.Companion_OrderDecision_.Create_Move_(m_CommandModel.Companion_Default___.DeliveryFailed())
	})()
	_ = _0_decision
	return Companion_Default___.OrderChange(input, version, order, lines, Companion_Default___.ForCarrier(input, order, _0_decision), m_CommandModel.Companion_OrderPreparation_.Create_Keep_(), false)
}
func (_static *CompanionStruct_Default___) RedeliverOrder(input m_CommandModel.CommandInput, version _dafny.Int, order m_CommandModel.OrderState, lines _dafny.Map) m_CommandModel.CommandPlan {
	var _0_decision m_CommandModel.OrderDecision = (func() m_CommandModel.OrderDecision {
		if ((order).Dtor_status()).Cmp(m_CommandModel.Companion_Default___.DeliveryFailed()) != 0 {
			return m_CommandModel.Companion_OrderDecision_.Create_Denied_(_dafny.UnicodeSeqOfUtf8Bytes("再配送できる状態ではありません"))
		}
		return m_CommandModel.Companion_OrderDecision_.Create_Move_(m_CommandModel.Companion_Default___.Shipping())
	})()
	_ = _0_decision
	return Companion_Default___.OrderChange(input, version, order, lines, Companion_Default___.ForCarrier(input, order, _0_decision), m_CommandModel.Companion_OrderPreparation_.Create_Keep_(), false)
}
func (_static *CompanionStruct_Default___) DeliverOrder(input m_CommandModel.CommandInput, version _dafny.Int, order m_CommandModel.OrderState, lines _dafny.Map) m_CommandModel.CommandPlan {
	return Companion_Default___.OrderChange(input, version, order, lines, Companion_Default___.ForCarrier(input, order, m_CommandModel.Companion_OrderDecision_.Create_Move_(m_CommandModel.Companion_Default___.Delivered())), m_CommandModel.Companion_OrderPreparation_.Create_Keep_(), false)
}
func (_static *CompanionStruct_Default___) ReturnOrder(input m_CommandModel.CommandInput, version _dafny.Int, order m_CommandModel.OrderState, lines _dafny.Map) m_CommandModel.CommandPlan {
	var _0_decision m_CommandModel.OrderDecision = (func() m_CommandModel.OrderDecision {
		if ((input).Dtor_noteBytes()).Sign() == 0 {
			return m_CommandModel.Companion_OrderDecision_.Create_Denied_(_dafny.UnicodeSeqOfUtf8Bytes("返送理由が必要です"))
		}
		return m_CommandModel.Companion_OrderDecision_.Create_Move_(m_CommandModel.Companion_Default___.Returned())
	})()
	_ = _0_decision
	return Companion_Default___.OrderChange(input, version, order, lines, Companion_Default___.ForCarrier(input, order, _0_decision), m_CommandModel.Companion_OrderPreparation_.Create_Keep_(), false)
}
func (_static *CompanionStruct_Default___) ForRole(input m_CommandModel.CommandInput, role _dafny.Sequence, decision m_CommandModel.OrderDecision) m_CommandModel.OrderDecision {
	if _dafny.Companion_Sequence_.Equal((input).Dtor_role(), role) {
		return decision
	} else {
		return m_CommandModel.Companion_OrderDecision_.Create_Denied_(Companion_Default___.RoleDenied())
	}
}
func (_static *CompanionStruct_Default___) ForWorker(input m_CommandModel.CommandInput, order m_CommandModel.OrderState, decision m_CommandModel.OrderDecision) m_CommandModel.OrderDecision {
	return Companion_Default___.ForRole(input, _dafny.UnicodeSeqOfUtf8Bytes("warehouse"), (func() m_CommandModel.OrderDecision {
		if _dafny.Companion_Sequence_.Equal((order).Dtor_worker(), (input).Dtor_actor()) {
			return decision
		}
		return m_CommandModel.Companion_OrderDecision_.Create_Denied_(_dafny.UnicodeSeqOfUtf8Bytes("担当外の作業です"))
	})())
}
func (_static *CompanionStruct_Default___) ForCarrier(input m_CommandModel.CommandInput, order m_CommandModel.OrderState, decision m_CommandModel.OrderDecision) m_CommandModel.OrderDecision {
	return Companion_Default___.ForRole(input, _dafny.UnicodeSeqOfUtf8Bytes("carrier"), (func() m_CommandModel.OrderDecision {
		if _dafny.Companion_Sequence_.Equal((order).Dtor_carrier(), (input).Dtor_actor()) {
			return decision
		}
		return m_CommandModel.Companion_OrderDecision_.Create_Denied_(_dafny.UnicodeSeqOfUtf8Bytes("担当外の配送です"))
	})())
}
func (_static *CompanionStruct_Default___) OrderChange(input m_CommandModel.CommandInput, version _dafny.Int, order m_CommandModel.OrderState, lines _dafny.Map, decision m_CommandModel.OrderDecision, prepare m_CommandModel.OrderPreparation, ship bool) m_CommandModel.CommandPlan {
	if !(m_CommandModel.Companion_Default___.VersionMatches((input).Dtor_expectedVersion(), version)) {
		return m_CommandModel.Companion_CommandPlan_.Create_Reject_(m_CommandModel.Companion_Default___.StaleVersion())
	} else if (_dafny.Companion_Sequence_.Equal((input).Dtor_role(), _dafny.UnicodeSeqOfUtf8Bytes("customer"))) && (!_dafny.Companion_Sequence_.Equal((input).Dtor_actor(), (order).Dtor_customer())) {
		return m_CommandModel.Companion_CommandPlan_.Create_Reject_(_dafny.UnicodeSeqOfUtf8Bytes("他の顧客の受注は操作できません"))
	} else {
		var _source0 m_CommandModel.OrderDecision = decision
		_ = _source0
		{
			if _source0.Is_NoChange() {
				return m_CommandModel.Companion_CommandPlan_.Create_Unchanged_()
			}
		}
		{
			if _source0.Is_Denied() {
				var _0_message _dafny.Sequence = _source0.Get_().(m_CommandModel.OrderDecision_Denied).Message
				_ = _0_message
				return m_CommandModel.Companion_CommandPlan_.Create_Reject_(_0_message)
			}
		}
		{
			var _1_target _dafny.Int = _source0.Get_().(m_CommandModel.OrderDecision_Move).Target
			_ = _1_target
			if !(m_Domain.Companion_Default___.Allowed((order).Dtor_status(), _1_target)) {
				return m_CommandModel.Companion_CommandPlan_.Create_Reject_(_dafny.UnicodeSeqOfUtf8Bytes("許可されない状態遷移です"))
			} else {
				var _2_stock m_CommandModel.StockChanges = (func() m_CommandModel.StockChanges {
					if (_dafny.SetOf(m_CommandModel.Companion_Default___.Cancelled(), m_CommandModel.Companion_Default___.PaymentFailed())).Contains(_1_target) {
						return m_CommandModel.Companion_StockChanges_.Create_Reserve_(m_CommandModel.Companion_Default___.Release(lines))
					}
					return (func() m_CommandModel.StockChanges {
						if ship {
							return m_CommandModel.Companion_StockChanges_.Create_Ship_(lines)
						}
						return m_CommandModel.Companion_StockChanges_.Create_Reserve_(_dafny.NewMapBuilder().ToMap())
					})()
				})()
				_ = _2_stock
				var _3_obligations _dafny.Set = (func() _dafny.Set {
					if (_1_target).Cmp(m_CommandModel.Companion_Default___.Delivered()) == 0 {
						return _dafny.SetOf(m_CommandModel.Companion_DeliveryObligation_.Create_AwardPoints_((order).Dtor_points()), m_CommandModel.Companion_DeliveryObligation_.Create_CreateInvoice_())
					}
					return _dafny.SetOf()
				})()
				_ = _3_obligations
				return m_CommandModel.Companion_CommandPlan_.Create_UpdateOrder_(_1_target, (order).Dtor_customer(), prepare, _2_stock, _3_obligations)
			}
		}
	}
}
func (_static *CompanionStruct_Default___) RoleDenied() _dafny.Sequence {
	return _dafny.UnicodeSeqOfUtf8Bytes("この利用者には操作権限がありません")
}

// End of class Default__
