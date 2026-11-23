package domain

import "encoding/json"

type AggregateKind string

const (
	StockKind   AggregateKind = "stock"
	CartKind    AggregateKind = "cart"
	OrderKind   AggregateKind = "order"
	LoyaltyKind AggregateKind = "loyalty"
)

// The unexported marker seals the persistence contract to domain aggregates.
type Aggregate interface {
	json.Marshaler
	Kind() AggregateKind
	aggregate()
}
type RestorableAggregate interface {
	Aggregate
	json.Unmarshaler
}

func (Stock) aggregate()            {}
func (Cart) aggregate()             {}
func (Order) aggregate()            {}
func (Loyalty) aggregate()          {}
func (Stock) Kind() AggregateKind   { return StockKind }
func (Cart) Kind() AggregateKind    { return CartKind }
func (Order) Kind() AggregateKind   { return OrderKind }
func (Loyalty) Kind() AggregateKind { return LoyaltyKind }
