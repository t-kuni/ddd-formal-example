package app

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	flow "wholesale.local/orbit/generated/CommandFlow"
	d "wholesale.local/orbit/generated/dafny"
	"wholesale.local/orbit/internal/domain"
	"wholesale.local/orbit/internal/store"
)

func (effect *commandEffects) LoadCart(id d.Sequence) (flow.Step, d.Int) {
	customer := id.VerbatimString(false)
	customerID, err := domain.ParseID[domain.ActorTag](customer)
	if err != nil {
		return effect.step(reject("顧客IDが不正です")), d.Zero
	}
	effect.cart = domain.NewCart(customerID)
	if err := effect.unit.Load("cart", customer, &effect.cart); err != nil {
		return effect.step(err), d.Zero
	}
	effect.customer = customer
	return effect.step(nil), d.IntOfInt64(effect.unit.Versions["cart"+customer])
}

func stockLines(items map[string]int64) d.Sequence {
	lines := make([]any, 0, len(items))
	// Retain a consistent lock acquisition order across transactions.
	for _, id := range sortedKeys(items) {
		lines = append(lines, flow.Companion_StockLine_.Create_StockLine_(d.UnicodeSeqOfUtf8Bytes(id), d.IntOfInt64(items[id])))
	}
	return d.SeqOf(lines...)
}

func (effect *commandEffects) ExpireCart() (flow.Step, d.Sequence) {
	delta, changed := effect.cart.Expire(effect.command.Payload.Expires, effect.now.Unix())
	if !changed {
		return flow.Companion_Step_.Create_Skip_(), d.EmptySeq
	}
	effect.result = map[string]any{"expired": true}
	return effect.step(nil), stockLines(delta)
}

func (effect *commandEffects) SetCart() (flow.Step, d.Sequence) {
	p := effect.command.Payload
	pid, err := domain.ParseID[domain.ProductTag](p.ProductID)
	if err != nil {
		return effect.step(checked(err)), d.EmptySeq
	}
	quantity, err := domain.NewCartQuantity(p.Quantity)
	if err != nil {
		return effect.step(checked(err)), d.EmptySeq
	}
	delta, changed, err := effect.cart.Change(pid, quantity, effect.now.Unix())
	if err != nil {
		return effect.step(checked(err)), d.EmptySeq
	}
	if !changed {
		return flow.Companion_Step_.Create_Skip_(), d.EmptySeq
	}
	effect.result = map[string]any{"customerId": effect.customer}
	return effect.step(nil), stockLines(delta)
}

func (effect *commandEffects) SaveCart() flow.Step {
	return effect.step(effect.unit.Save("cart", effect.customer, effect.cart))
}

func (effect *commandEffects) CheckoutCart() (flow.Step, d.Sequence) {
	items, err := effect.cart.Checkout(effect.now.Unix())
	effect.lines = nil
	return effect.step(checked(err)), stockLines(items)
}

func (effect *commandEffects) ReadShippingAddress() flow.Step {
	address, err := domain.NewShippingAddress(effect.command.Payload.Address)
	effect.address = address
	return effect.step(checked(err))
}

func (effect *commandEffects) RequestedDate() (bool, bool, d.Int, d.Int) {
	requested := effect.command.Payload.RequestedDate
	date, err := time.Parse(time.DateOnly, requested)
	today := effect.now.In(time.FixedZone("JST", 9*3600))
	dateKey := func(t time.Time) d.Int {
		year, month, day := t.Date()
		return d.IntOf(year*10000 + int(month)*100 + day)
	}
	return requested != "", err == nil, dateKey(date), dateKey(today)
}

func (effect *commandEffects) LoadCoupon() (flow.Step, d.Int) {
	var percent int64
	err := effect.tx.QueryRow(effect.ctx, "SELECT percent FROM coupons WHERE code=$1", effect.command.Payload.Coupon).Scan(&percent)
	if errors.Is(err, pgx.ErrNoRows) {
		err = reject("クーポンが存在しません")
	}
	return effect.step(err), d.IntOfInt64(percent)
}

func (effect *commandEffects) LoadProduct(product d.Sequence) flow.Step {
	var dto []byte
	if err := effect.tx.QueryRow(effect.ctx, store.ProductQuery+" WHERE id=$1", product.VerbatimString(false)).Scan(&dto); err != nil {
		return effect.step(err)
	}
	var value domain.Product
	err := json.Unmarshal(dto, &value)
	effect.product = value
	return effect.step(err)
}

func (effect *commandEffects) AddOrderLine(line flow.StockLine) flow.Step {
	pid, err := domain.ParseID[domain.ProductTag](line.Dtor_product().VerbatimString(false))
	if err != nil {
		return effect.step(err)
	}
	qty, err := domain.NewOrderQuantity(line.Dtor_quantity().Int64())
	if err != nil {
		return effect.step(checked(err))
	}
	product := effect.product
	price, err := domain.NewUnitPrice(product.Price)
	if err != nil {
		return effect.step(checked(err))
	}
	value, err := domain.NewLine(pid, product.SKU, product.Name, product.Unit, qty, price)
	if err != nil {
		return effect.step(checked(err))
	}
	effect.lines = append(effect.lines, value)
	return effect.step(nil)
}

func (effect *commandEffects) BuildOrder(percent d.Int) flow.Step {
	c := effect.command
	p := c.Payload
	customer, err := domain.ParseID[domain.ActorTag](effect.customer)
	if err != nil {
		return effect.step(checked(err))
	}
	order, err := domain.ParseID[domain.OrderTag](c.ID)
	if err != nil {
		return effect.step(checked(err))
	}
	rate, err := domain.NewDiscountRate(percent.Int64())
	if err != nil {
		return effect.step(checked(err))
	}
	effect.order, err = domain.NewOrder(order, customer, effect.lines, effect.address, rate, p.Coupon, p.RequestedDate, p.Note, effect.now)
	effect.result = map[string]any{"orderId": c.ID}
	return effect.step(checked(err))
}

func (effect *commandEffects) LoadCheckoutVersion() flow.Step {
	var existing domain.Order
	return effect.step(effect.unit.Load("order", effect.command.ID, &existing))
}

func (effect *commandEffects) LoadAdjustment() (flow.Step, d.Int) {
	id := effect.command.Payload.ProductID
	stock, err := loadStock(effect.unit, id)
	effect.stock = stock
	return effect.step(err), d.IntOfInt64(effect.unit.Versions["stock"+id])
}

func (effect *commandEffects) AdjustStock() flow.Step {
	delta, err := domain.NewStockAdjustment(effect.command.Payload.Delta)
	if err != nil {
		return effect.step(checked(err))
	}
	return effect.step(checked(effect.stock.Adjust(delta)))
}

func (effect *commandEffects) SaveAdjustment() flow.Step {
	id := effect.command.Payload.ProductID
	effect.result = map[string]any{"productId": id}
	return effect.SaveStock(d.UnicodeSeqOfUtf8Bytes(id))
}

func (effect *commandEffects) LoadOrder() (flow.Step, flow.OrderContext, d.Int) {
	p := effect.command.Payload
	var empty flow.OrderContext
	if !validID(p.OrderID) {
		return effect.step(reject("受注IDが不正です")), empty, d.Zero
	}
	var order domain.Order
	if err := effect.unit.Load("order", p.OrderID, &order); err != nil {
		return effect.step(err), empty, d.Zero
	}
	if !order.Exists() {
		return effect.step(reject("受注が存在しません")), empty, d.Zero
	}
	x := order.Snapshot()
	effect.order = order
	a := effect.actor
	return effect.step(nil), flow.Companion_OrderContext_.Create_OrderContext_(
		d.UnicodeSeqOfUtf8Bytes(a.Role), d.UnicodeSeqOfUtf8Bytes(a.ID),
		d.UnicodeSeqOfUtf8Bytes(x.CustomerID), d.UnicodeSeqOfUtf8Bytes(x.WorkerID),
		d.UnicodeSeqOfUtf8Bytes(x.CarrierID), d.IntOf(order.Status().Code()),
		d.IntOfInt64(effect.now.Sub(x.CreatedAt).Nanoseconds()), p.Success, p.Note != "", p.Tracking != "",
		d.IntOf(p.BoxCount), d.IntOfInt64(p.WeightGrams),
	), d.IntOfInt64(effect.unit.Versions["order"+x.ID])
}

func (effect *commandEffects) Noop() {
	effect.result = map[string]any{"noop": true}
}

func (effect *commandEffects) Reject(message d.Sequence) flow.Step {
	return effect.step(reject(message.VerbatimString(false)))
}

func (effect *commandEffects) AssignWorker() flow.Step {
	actor, err := domain.ParseID[domain.ActorTag](effect.actor.ID)
	if err != nil {
		return effect.step(checked(err))
	}
	return effect.step(checked(effect.order.AssignWorker(actor)))
}

func (effect *commandEffects) PackOrder() flow.Step {
	p := effect.command.Payload
	boxCount, err := domain.NewBoxCount(p.BoxCount)
	if err != nil {
		return effect.step(checked(err))
	}
	weight, err := domain.NewWeightGrams(p.WeightGrams)
	if err != nil {
		return effect.step(checked(err))
	}
	return effect.step(checked(effect.order.Pack(boxCount, weight)))
}

func (effect *commandEffects) AssignCarrier() flow.Step {
	actor, err := domain.ParseID[domain.ActorTag](effect.actor.ID)
	if err != nil {
		return effect.step(checked(err))
	}
	return effect.step(checked(effect.order.AssignCarrier(actor, effect.command.Payload.Tracking)))
}

func (effect *commandEffects) MoveOrder(status d.Int) (flow.Step, d.Sequence) {
	target, err := domain.ParseOrderStatus(status.Int())
	if err != nil {
		return effect.step(checked(err)), d.EmptySeq
	}
	actor, err := domain.ParseID[domain.ActorTag](effect.actor.ID)
	if err != nil {
		return effect.step(checked(err)), d.EmptySeq
	}
	if err := effect.order.Move(target, actor, effect.command.Payload.Note, effect.now); err != nil {
		return effect.step(checked(err)), d.EmptySeq
	}
	x := effect.order.Snapshot()
	items := make(map[string]int64, len(x.Lines))
	for _, line := range x.Lines {
		items[line.ProductID] = line.Quantity
	}
	effect.result = map[string]any{"orderId": x.ID, "status": target.Code()}
	return effect.step(nil), stockLines(items)
}

func (effect *commandEffects) SaveOrder() flow.Step {
	return effect.step(effect.unit.Save("order", effect.order.Snapshot().ID, effect.order))
}

func (effect *commandEffects) LoadStock(product d.Sequence) flow.Step {
	stock, err := loadStock(effect.unit, product.VerbatimString(false))
	effect.stock = stock
	return effect.step(err)
}

func (effect *commandEffects) ReserveStock(quantity d.Int) flow.Step {
	delta, err := domain.NewReservationDelta(quantity.Int64())
	if err != nil {
		return effect.step(checked(err))
	}
	return effect.step(checked(effect.stock.Reserve(delta)))
}

func (effect *commandEffects) ShipStock(quantity d.Int) flow.Step {
	units, err := domain.NewOrderQuantity(quantity.Int64())
	if err != nil {
		return effect.step(checked(err))
	}
	return effect.step(checked(effect.stock.Ship(units)))
}

func (effect *commandEffects) SaveStock(product d.Sequence) flow.Step {
	return effect.step(effect.unit.Save("stock", product.VerbatimString(false), *effect.stock))
}

func (effect *commandEffects) InsertPointAward() flow.Step {
	_, err := effect.tx.Exec(effect.ctx, "INSERT INTO order_point_awards(order_id) VALUES($1)", effect.order.Snapshot().ID)
	return effect.step(err)
}

func (effect *commandEffects) LoadLoyalty() flow.Step {
	return effect.step(effect.unit.Load("loyalty", effect.order.Snapshot().CustomerID, &effect.loyalty))
}

func (effect *commandEffects) CreditLoyalty() flow.Step {
	return effect.step(effect.loyalty.Credit(effect.order.EarnedPoints()))
}

func (effect *commandEffects) SaveLoyalty() flow.Step {
	return effect.step(effect.unit.Save("loyalty", effect.order.Snapshot().CustomerID, effect.loyalty))
}

func (effect *commandEffects) RenderInvoice() flow.Step {
	pdf, err := InvoicePDF(effect.order)
	effect.pdf = pdf
	return effect.step(err)
}

func (effect *commandEffects) InsertInvoice() flow.Step {
	id := effect.order.Snapshot().ID
	_, err := effect.tx.Exec(effect.ctx, "INSERT INTO invoices(id,order_id,pdf) VALUES($1,$2,$3) ON CONFLICT(order_id) DO NOTHING", store.ID("invoice/"+id), id, effect.pdf)
	return effect.step(err)
}
