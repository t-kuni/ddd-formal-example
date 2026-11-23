// Package CartCommands
// Dafny module CartCommands compiled into Go

package CartCommands

import (
	os "os"

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
	return "CartCommands.Default__"
}
func (_this *Default__) ParentTraits_() []*_dafny.TraitID {
	return [](*_dafny.TraitID){}
}

var _ _dafny.TraitOffspring = &Default__{}

func (_static *CompanionStruct_Default___) SetCart(input m_CommandModel.CommandInput, cart m_CommandModel.CartState) m_CommandModel.CommandPlan {
	if !(m_CommandModel.Companion_Default___.VersionMatches((input).Dtor_expectedVersion(), (cart).Dtor_version())) {
		return m_CommandModel.Companion_CommandPlan_.Create_Reject_(m_CommandModel.Companion_Default___.StaleVersion())
	} else if !((input).Dtor_productValid()) {
		return m_CommandModel.Companion_CommandPlan_.Create_Reject_(_dafny.UnicodeSeqOfUtf8Bytes("商品IDが不正です"))
	} else if !(m_Domain.Companion_Default___.CartQuantityValid((input).Dtor_quantity())) {
		return m_CommandModel.Companion_CommandPlan_.Create_Reject_(_dafny.UnicodeSeqOfUtf8Bytes("カート数量が範囲外です"))
	} else {
		var _0_previous _dafny.Int = (func() _dafny.Int {
			if ((cart).Dtor_items()).Contains((input).Dtor_product()) {
				return ((cart).Dtor_items()).Get((input).Dtor_product()).(_dafny.Int)
			}
			return _dafny.Zero
		})()
		_ = _0_previous
		if ((!(m_Domain.Companion_Default___.CartShapeValid(((cart).Dtor_items()).Cardinality(), (cart).Dtor_expires()))) || (!(m_Domain.Companion_Default___.TimestampValid((input).Dtor_now())))) || (!(m_Domain.Companion_Default___.CartQuantityValid(_0_previous))) {
			return m_CommandModel.Companion_CommandPlan_.Create_Reject_(_dafny.UnicodeSeqOfUtf8Bytes("カート不変条件違反"))
		} else {
			var _source0 m_Domain.CartChange = m_Domain.Companion_Default___.PlanCartChange(((cart).Dtor_items()).Cardinality(), _0_previous, (input).Dtor_quantity(), (input).Dtor_now(), (cart).Dtor_expires())
			_ = _source0
			{
				if _source0.Is_CartFull() {
					return m_CommandModel.Companion_CommandPlan_.Create_Reject_(_dafny.UnicodeSeqOfUtf8Bytes("カートは100商品までです"))
				}
			}
			{
				if _source0.Is_CartUnchanged() {
					return m_CommandModel.Companion_CommandPlan_.Create_Unchanged_()
				}
			}
			{
				if _source0.Is_CartTimeOutOfRange() {
					return m_CommandModel.Companion_CommandPlan_.Create_Reject_(_dafny.UnicodeSeqOfUtf8Bytes("カートの有効期限が範囲外です"))
				}
			}
			{
				var _1_reset bool = _source0.Get_().(m_Domain.CartChange_CartChanged).Reset
				_ = _1_reset
				var _2_delta _dafny.Int = _source0.Get_().(m_Domain.CartChange_CartChanged).Delta
				_ = _2_delta
				var _3_expiry _dafny.Int = _source0.Get_().(m_Domain.CartChange_CartChanged).Expiry
				_ = _3_expiry
				var _4_retained _dafny.Map = (func() _dafny.Map {
					if _1_reset {
						return _dafny.NewMapBuilder().ToMap()
					}
					return (cart).Dtor_items()
				})()
				_ = _4_retained
				var _5_releases _dafny.Map = (func() _dafny.Map {
					if _1_reset {
						return m_CommandModel.Companion_Default___.Release((cart).Dtor_items())
					}
					return _dafny.NewMapBuilder().ToMap()
				})()
				_ = _5_releases
				return m_CommandModel.Companion_CommandPlan_.Create_UpdateCart_((cart).Dtor_customerId(), m_CommandModel.Companion_Default___.WithQuantity(_4_retained, (input).Dtor_product(), (input).Dtor_quantity()), _3_expiry, (_5_releases).Update((input).Dtor_product(), _2_delta), false)
			}
		}
	}
}
func (_static *CompanionStruct_Default___) ExpireCart(input m_CommandModel.CommandInput, cart m_CommandModel.CartState) m_CommandModel.CommandPlan {
	if !(m_CommandModel.Companion_Default___.VersionMatches((input).Dtor_expectedVersion(), (cart).Dtor_version())) {
		return m_CommandModel.Companion_CommandPlan_.Create_Reject_(m_CommandModel.Companion_Default___.StaleVersion())
	} else if m_Domain.Companion_Default___.MayExpire((input).Dtor_now(), (cart).Dtor_expires(), (input).Dtor_expireAt(), (((cart).Dtor_items()).Cardinality()).Sign() == 1) {
		return m_CommandModel.Companion_CommandPlan_.Create_UpdateCart_((cart).Dtor_customerId(), _dafny.NewMapBuilder().ToMap(), (cart).Dtor_expires(), m_CommandModel.Companion_Default___.Release((cart).Dtor_items()), true)
	} else {
		return m_CommandModel.Companion_CommandPlan_.Create_Unchanged_()
	}
}
func (_static *CompanionStruct_Default___) CheckoutOrder(input m_CommandModel.CommandInput, cart m_CommandModel.CartState) m_CommandModel.CommandPlan {
	if !(m_CommandModel.Companion_Default___.VersionMatches((input).Dtor_expectedVersion(), (cart).Dtor_version())) {
		return m_CommandModel.Companion_CommandPlan_.Create_Reject_(m_CommandModel.Companion_Default___.StaleVersion())
	} else if !(m_Domain.Companion_Default___.MayCheckout((input).Dtor_now(), (cart).Dtor_expires(), (((cart).Dtor_items()).Cardinality()).Sign() == 1)) {
		return m_CommandModel.Companion_CommandPlan_.Create_Reject_(_dafny.UnicodeSeqOfUtf8Bytes("カートが空、または有効期限切れです"))
	} else if !(Companion_Default___.ValidRequestedDay(cart)) {
		return m_CommandModel.Companion_CommandPlan_.Create_Reject_(_dafny.UnicodeSeqOfUtf8Bytes("希望納品日が不正です"))
	} else {
		var _source0 m_CommandModel.Coupon = (cart).Dtor_coupon()
		_ = _source0
		{
			if _source0.Is_UnknownCoupon() {
				return m_CommandModel.Companion_CommandPlan_.Create_Reject_(_dafny.UnicodeSeqOfUtf8Bytes("クーポンが存在しません"))
			}
		}
		{
			if _source0.Is_NoCoupon() {
				return m_CommandModel.Companion_CommandPlan_.Create_Checkout_((cart).Dtor_customerId(), (input).Dtor_commandId(), (cart).Dtor_items(), _dafny.NewMapBuilder().ToMap(), (cart).Dtor_expires(), _dafny.Zero)
			}
		}
		{
			var _0_percent _dafny.Int = _source0.Get_().(m_CommandModel.Coupon_CouponRate).Percent
			_ = _0_percent
			if m_Domain.Companion_Default___.DiscountRateValid(_0_percent) {
				return m_CommandModel.Companion_CommandPlan_.Create_Checkout_((cart).Dtor_customerId(), (input).Dtor_commandId(), (cart).Dtor_items(), _dafny.NewMapBuilder().ToMap(), (cart).Dtor_expires(), _0_percent)
			} else {
				return m_CommandModel.Companion_CommandPlan_.Create_Reject_(_dafny.UnicodeSeqOfUtf8Bytes("割引率は0〜100です"))
			}
		}
	}
}
func (_static *CompanionStruct_Default___) ValidRequestedDay(cart m_CommandModel.CartState) bool {
	return (((cart).Dtor_requestedDay()).Is_Unspecified()) || ((((cart).Dtor_requestedDay()).Is_CalendarDay()) && ((((cart).Dtor_requestedDay()).Dtor_date()).Cmp((cart).Dtor_today()) >= 0))
}
func (_static *CompanionStruct_Default___) NeedsCoupon(input m_CommandModel.CommandInput, snapshot m_CommandModel.Snapshot) bool {
	return (((((_dafny.Companion_Sequence_.Equal((input).Dtor_kind(), _dafny.UnicodeSeqOfUtf8Bytes("order.checkout"))) && (!_dafny.Companion_Sequence_.Equal((input).Dtor_couponCode(), _dafny.UnicodeSeqOfUtf8Bytes("")))) && ((snapshot).Is_Cart())) && (m_CommandModel.Companion_Default___.VersionMatches((input).Dtor_expectedVersion(), ((snapshot).Dtor_state()).Dtor_version()))) && (m_Domain.Companion_Default___.MayCheckout((input).Dtor_now(), ((snapshot).Dtor_state()).Dtor_expires(), ((((snapshot).Dtor_state()).Dtor_items()).Cardinality()).Sign() == 1))) && (Companion_Default___.ValidRequestedDay((snapshot).Dtor_state()))
}

// End of class Default__
