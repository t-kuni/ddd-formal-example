package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"gorm.io/gorm"
	"wholesale.local/orbit/internal/domain"
)

type orderRow struct {
	ID              string
	Version         int64
	Number          string
	CustomerID      string
	Status          int
	RequestedDate   string
	Note            string
	Coupon          string
	DiscountPercent int64
	Subtotal        int64
	Discount        int64
	Net             int64
	Tax             int64
	Total           int64
	Points          int64
	WorkerID        string
	CreatedAt       time.Time

	Address  addressRow   `gorm:"foreignKey:OrderID;references:ID"`
	Items    []itemRow    `gorm:"foreignKey:OrderID;references:ID"`
	Shipment shipmentRow  `gorm:"foreignKey:OrderID;references:ID"`
	History  []historyRow `gorm:"foreignKey:OrderID;references:ID"`
}

func (orderRow) TableName() string { return "orders" }

type addressRow struct {
	OrderID   string
	Company   string
	Recipient string
	Postal    string
	Address   string
	Phone     string
}

func (addressRow) TableName() string { return "order_addresses" }

type itemRow struct {
	OrderID   string
	Position  int
	ProductID string
	SKU       string
	Name      string
	Unit      string
	Quantity  int64
	UnitPrice int64
}

func (itemRow) TableName() string { return "order_items" }

type shipmentRow struct {
	OrderID   string
	CarrierID string
	Tracking  string
	Boxes     []boxRow `gorm:"foreignKey:OrderID;references:OrderID"`
}

func (shipmentRow) TableName() string { return "shipments" }

type boxRow struct {
	ID          string
	OrderID     string
	Position    int
	Label       string
	WeightGrams int64
}

func (boxRow) TableName() string { return "shipment_boxes" }

type historyRow struct {
	OrderID    string
	Sequence   int
	OccurredAt time.Time
	ActorID    string
	StatusFrom int
	StatusTo   int
	Note       string
}

func (historyRow) TableName() string { return "order_status_history" }

type InvoiceStatus struct {
	ID    string
	Sent  bool
	Error string
}

// ReadOrderDetails reads the root and all requested relations from one DB snapshot.
func (s *Store) ReadOrderDetails(ctx context.Context, id, actorID, role string) (domain.OrderSnapshot, int64, *InvoiceStatus, error) {
	var row orderRow
	var invoice *InvoiceStatus
	err := s.orm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		query := tx.Where("orders.id = ?", id)
		if role == "customer" {
			query = query.Where("orders.customer_id = ?", actorID)
		}
		if err := query.Preload("Address").
			Preload("Items", func(q *gorm.DB) *gorm.DB { return q.Order("position") }).
			Preload("Shipment.Boxes", func(q *gorm.DB) *gorm.DB { return q.Order("position") }).
			Preload("History", func(q *gorm.DB) *gorm.DB { return q.Order("sequence") }).
			Take(&row).Error; err != nil {
			return err
		}
		if role == "customer" || role == "operator" {
			var status InvoiceStatus
			result := tx.Table("invoices").Select("id, sent_at IS NOT NULL AS sent, last_error AS error").
				Where("order_id = ?", id).Scan(&status)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 1 {
				invoice = &status
			}
		}
		return nil
	}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = pgx.ErrNoRows
	}
	if err != nil {
		return domain.OrderSnapshot{}, 0, nil, err
	}
	return row.snapshot(), row.Version, invoice, nil
}

func (row orderRow) snapshot() domain.OrderSnapshot {
	result := domain.OrderSnapshot{
		ID: row.ID, Number: row.Number, CustomerID: row.CustomerID, Status: row.Status,
		Address: domain.Address{Company: row.Address.Company, Recipient: row.Address.Recipient,
			Postal: row.Address.Postal, Text: row.Address.Address, Phone: row.Address.Phone},
		RequestedDate: row.RequestedDate, Note: row.Note, Coupon: row.Coupon, DiscountPercent: row.DiscountPercent,
		Subtotal: row.Subtotal, Discount: row.Discount, Net: row.Net, Tax: row.Tax, Total: row.Total, Points: row.Points,
		WorkerID: row.WorkerID, CarrierID: row.Shipment.CarrierID, Tracking: row.Shipment.Tracking, CreatedAt: row.CreatedAt,
		Lines:   make([]domain.LineSnapshot, 0, len(row.Items)),
		Boxes:   make([]domain.Box, 0, len(row.Shipment.Boxes)),
		History: make([]domain.History, 0, len(row.History)),
	}
	for _, item := range row.Items {
		result.Lines = append(result.Lines, domain.LineSnapshot{
			ProductID: item.ProductID, SKU: item.SKU, Name: item.Name, Unit: item.Unit,
			Quantity: item.Quantity, Price: item.UnitPrice,
		})
	}
	for _, box := range row.Shipment.Boxes {
		result.Boxes = append(result.Boxes, domain.Box{ID: box.ID, Label: box.Label, Weight: box.WeightGrams})
	}
	for _, event := range row.History {
		result.History = append(result.History, domain.History{
			At: event.OccurredAt, Actor: event.ActorID, From: event.StatusFrom, To: event.StatusTo, Note: event.Note,
		})
	}
	return result
}
