package domain

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"

	core "wholesale.local/orbit/generated/Domain"
	d "wholesale.local/orbit/generated/dafny"
)

// Tags make resource identities non-interchangeable at compile time.
type ProductTag struct{}
type OrderTag struct{}
type ActorTag struct{}
type ResourceID[T any] struct{ value uuid.UUID }
type ProductID = ResourceID[ProductTag]
type OrderID = ResourceID[OrderTag]
type ActorID = ResourceID[ActorTag]

func ParseID[T any](s string) (ResourceID[T], error) {
	id, e := uuid.Parse(s)
	if e != nil || id == uuid.Nil {
		return ResourceID[T]{}, Invalid("IDは空でないUUIDにしてください")
	}
	return ResourceID[T]{id}, nil
}
func (id ResourceID[T]) String() string { return id.value.String() }
func (id ResourceID[T]) Valid() bool    { return id.value != uuid.Nil }
func (id ResourceID[T]) MarshalJSON() ([]byte, error) {
	if !id.Valid() {
		return nil, Invalid("未初期化のID")
	}
	return json.Marshal(id.String())
}
func (id *ResourceID[T]) UnmarshalJSON(b []byte) error {
	var s string
	if e := json.Unmarshal(b, &s); e != nil {
		return e
	}
	v, e := ParseID[T](s)
	if e == nil {
		*id = v
	}
	return e
}

// Each numeric value has a separate type and an unexported representation.
type Quantity struct{ value int64 }
type OrderQuantity struct{ value int64 }
type ReservationDelta struct{ value int64 }
type StockAdjustment struct{ value int64 }
type Money struct{ value int64 }
type UnitPrice struct{ value int64 }
type DiscountRate struct{ value int64 }

func NewQuantity(v int64) (Quantity, error) {
	if !core.Companion_Default___.QuantityValid(d.IntOfInt64(v)) {
		return Quantity{}, Invalid("在庫数量が範囲外です")
	}
	return Quantity{v}, nil
}
func NewOrderQuantity(v int64) (OrderQuantity, error) {
	if !core.Companion_Default___.OrderQuantityValid(d.IntOfInt64(v)) {
		return OrderQuantity{}, Invalid("明細数量は1〜10000です")
	}
	return OrderQuantity{v}, nil
}
func NewReservationDelta(v int64) (ReservationDelta, error) {
	if !core.Companion_Default___.StockDeltaValid(d.IntOfInt64(v)) {
		return ReservationDelta{}, Invalid("引当増減数が範囲外です")
	}
	return ReservationDelta{v}, nil
}
func NewStockAdjustment(v int64) (StockAdjustment, error) {
	if !core.Companion_Default___.StockDeltaValid(d.IntOfInt64(v)) {
		return StockAdjustment{}, Invalid("在庫調整数が範囲外です")
	}
	return StockAdjustment{v}, nil
}
func NewMoney(v int64) (Money, error) {
	if !core.Companion_Default___.MoneyValid(d.IntOfInt64(v)) {
		return Money{}, Invalid("金額が範囲外です")
	}
	return Money{v}, nil
}
func NewUnitPrice(v int64) (UnitPrice, error) {
	if !core.Companion_Default___.UnitPriceValid(d.IntOfInt64(v)) {
		return UnitPrice{}, Invalid("単価が範囲外です")
	}
	return UnitPrice{v}, nil
}
func NewDiscountRate(v int64) (DiscountRate, error) {
	if !core.Companion_Default___.DiscountRateValid(d.IntOfInt64(v)) {
		return DiscountRate{}, Invalid("割引率は0〜100です")
	}
	return DiscountRate{v}, nil
}
func (q Quantity) Units() int64         { return q.value }
func (q OrderQuantity) Units() int64    { return q.value }
func (q ReservationDelta) Units() int64 { return q.value }
func (q StockAdjustment) Units() int64  { return q.value }
func (m Money) Yen() int64              { return m.value }
func (m UnitPrice) Yen() int64          { return m.value }
func (r DiscountRate) Percent() int64   { return r.value }

// Go enums are named types, not closed sum types. Parse and Move validate the boundary.
type OrderStatus uint8

const (
	PaymentPending OrderStatus = iota
	Accepted
	Picking
	PickBlocked
	Packed
	Shipping
	DeliveryFailed
	Delivered
	Cancelled
	PaymentFailed
	Returned
)

func ParseOrderStatus(v int) (OrderStatus, error) {
	if !core.Companion_Default___.StatusValid(d.IntOf(v)) {
		return 0, Invalid("未知の受注状態")
	}
	return OrderStatus(v), nil
}
func (s OrderStatus) Code() int { return int(s) }

// No arithmetic methods on IDs, status, money, quantity, or rates.

type ShippingAddress struct{ value Address }

func NewShippingAddress(a Address) (ShippingAddress, error) {
	for name, v := range map[string]string{"会社名": a.Company, "担当者": a.Recipient, "郵便番号": a.Postal, "住所": a.Text, "電話番号": a.Phone} {
		if !core.Companion_Default___.TextFieldValid(strings.TrimSpace(v) != "", d.IntOf(len(v))) {
			return ShippingAddress{}, Invalid(fmt.Sprintf("%sを入力してください（250文字程度まで）", name))
		}
	}
	return ShippingAddress{a}, nil
}
func (a ShippingAddress) Snapshot() Address { return a.value }

type Points struct{ value int64 }

func NewPoints(v int64) (Points, error) {
	if !core.Companion_Default___.PointsValid(d.IntOfInt64(v)) {
		return Points{}, Invalid("ポイントが範囲外です")
	}
	return Points{v}, nil
}
func (p Points) Value() int64 { return p.value }

type BoxCount struct{ value int }

func NewBoxCount(v int) (BoxCount, error) {
	if !core.Companion_Default___.BoxCountValid(d.IntOf(v)) {
		return BoxCount{}, Invalid("箱数は1〜100です")
	}
	return BoxCount{v}, nil
}

type WeightGrams struct{ value int64 }

func NewWeightGrams(v int64) (WeightGrams, error) {
	if !core.Companion_Default___.WeightValid(d.IntOfInt64(v)) {
		return WeightGrams{}, Invalid("重量は1〜500000gです")
	}
	return WeightGrams{v}, nil
}
