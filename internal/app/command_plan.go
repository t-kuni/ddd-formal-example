package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	plan "wholesale.local/orbit/generated/CommandFlow"
	d "wholesale.local/orbit/generated/dafny"
	"wholesale.local/orbit/internal/domain"
	"wholesale.local/orbit/internal/store"
)

type commandAttempt struct {
	unit     *store.Unit
	command  Command
	actor    domain.Actor
	now      time.Time
	order    domain.Order
	stock    *domain.Stock
	products map[string]domain.Product
	stocks   map[string]*domain.Stock
	loyalty  domain.Loyalty
}

func textValue(value string) d.Sequence { return d.UnicodeSeqOfUtf8Bytes(value) }

func itemMap(items map[string]int64) d.Map {
	builder := d.NewMapBuilder()
	for id, qty := range items {
		builder.Add(textValue(id), d.IntOfInt64(qty))
	}
	return builder.ToMap()
}

func readItems(items d.Map) map[string]int64 {
	result := make(map[string]int64, items.CardinalityInt())
	next := items.Keys().Iterator()
	for key, ok := next(); ok; key, ok = next() {
		result[key.(d.Sequence).VerbatimString(false)] = items.Get(key).(d.Int).Int64()
	}
	return result
}

func (a *commandAttempt) input() plan.CommandInput {
	p := a.command.Payload
	customer, _ := domain.ParseID[domain.ActorTag](p.CustomerID)
	product, err := domain.ParseID[domain.ProductTag](p.ProductID)
	return plan.Companion_CommandInput_.Create_CommandInput_(
		textValue(a.command.Type), textValue(a.actor.Role), textValue(a.actor.ID), textValue(customer.String()),
		d.IntOf(len(p.Note)), d.IntOf(len(p.Tracking)), d.IntOf(len(p.RequestedDate)),
		d.IntOfInt64(p.ExpectedVersion), textValue(product.String()), err == nil,
		d.IntOfInt64(p.Delta), d.IntOfInt64(p.Quantity), d.IntOfInt64(a.now.Unix()), d.IntOfInt64(p.Expires),
		p.Success, textValue(p.Tracking), d.IntOf(p.BoxCount), d.IntOfInt64(p.WeightGrams),
		textValue(p.OrderID), textValue(p.Coupon), textValue(a.command.ID),
	)
}

func missing(message string) plan.Snapshot {
	return plan.Companion_Snapshot_.Create_Missing_(textValue(message))
}

func (a *commandAttempt) read(request plan.ReadRequest) (plan.Snapshot, error) {
	switch {
	case request.Is_ReadCart():
		return a.cartSnapshot(request.Dtor_customerId().VerbatimString(false))
	case request.Is_ReadStock():
		id := request.Dtor_productId().VerbatimString(false)
		var err error
		a.stock, err = loadStock(a.unit, id)
		var denied *Rejection
		if errors.As(err, &denied) {
			return missing(denied.Message), nil
		}
		return plan.Companion_Snapshot_.Create_Stock_(d.IntOfInt64(a.unit.Versions["stock"+id])), err
	case request.Is_ReadOrder():
		return a.orderSnapshot(request.Dtor_orderId().VerbatimString(false))
	default:
		return plan.Snapshot{}, errors.New("invalid read request")
	}
}

func (a *commandAttempt) cartSnapshot(customer string) (plan.Snapshot, error) {
	p := a.command.Payload
	id, err := domain.ParseID[domain.ActorTag](customer)
	if err != nil {
		return missing("顧客IDが不正です"), nil
	}
	customer = id.String()
	cart := domain.NewCart(id)
	if err := a.unit.Load("cart", customer, &cart); err != nil {
		return plan.Snapshot{}, err
	}
	coupon := plan.Companion_Coupon_.Create_NoCoupon_()
	day := plan.Companion_RequestedDay_.Create_Unspecified_()
	if p.RequestedDate != "" {
		date, err := time.Parse(time.DateOnly, p.RequestedDate)
		day = plan.Companion_RequestedDay_.Create_InvalidDate_()
		if err == nil {
			day = plan.Companion_RequestedDay_.Create_CalendarDay_(calendarDay(date))
		}
	}
	const jstOffset = 9 * 3600
	today := a.now.In(time.FixedZone("JST", jstOffset))
	snapshot := cart.Snapshot()
	state := plan.Companion_CartState_.Create_CartState_(
		textValue(customer), d.IntOfInt64(a.unit.Versions["cart"+customer]),
		itemMap(snapshot.Items), d.IntOfInt64(snapshot.Expires), day, calendarDay(today), coupon,
	)
	return plan.Companion_Snapshot_.Create_Cart_(state), nil
}

func calendarDay(t time.Time) d.Int {
	year, month, day := t.Date()
	return d.IntOf(year*10000 + int(month)*100 + day)
}

func (a *commandAttempt) orderSnapshot(orderID string) (plan.Snapshot, error) {
	id, err := domain.ParseID[domain.OrderTag](orderID)
	if err != nil {
		return missing("受注IDが不正です"), nil
	}
	if err := a.unit.Load("order", id.String(), &a.order); err != nil {
		return plan.Snapshot{}, err
	}
	if !a.order.Exists() {
		return missing("受注が存在しません"), nil
	}
	x := a.order.Snapshot()
	order := plan.Companion_OrderState_.Create_OrderState_(
		textValue(x.CustomerID), textValue(x.WorkerID), textValue(x.CarrierID),
		d.IntOf(x.Status), d.IntOfInt64(a.now.Sub(x.CreatedAt).Nanoseconds()), d.IntOfInt64(x.Points),
	)
	lines := make(map[string]int64, len(x.Lines))
	for _, line := range x.Lines {
		lines[line.ProductID] = line.Quantity
	}
	return plan.Companion_Snapshot_.Create_Order_(d.IntOfInt64(a.unit.Versions["order"+id.String()]), order, itemMap(lines)), nil
}

func (a *commandAttempt) apply(p plan.CommandPlan) (map[string]any, error) {
	switch {
	case p.Is_Reject():
		return nil, reject(p.Dtor_reason().VerbatimString(false))
	case p.Is_Unchanged():
		return map[string]any{"noop": true}, nil
	case p.Is_UpdateCart():
		customer := p.Dtor_customerId().VerbatimString(false)
		if err := a.changeStock(plan.Companion_StockChanges_.Create_Reserve_(p.Dtor_reservations())); err != nil {
			return nil, err
		}
		result := map[string]any{"customerId": customer}
		if p.Dtor_expired() {
			result = map[string]any{"expired": true}
		}
		return result, a.saveCart(customer, readItems(p.Dtor_nextItems()), p.Dtor_expiry().Int64())
	case p.Is_Checkout():
		return a.checkout(p)
	case p.Is_Adjust():
		delta, err := domain.NewStockAdjustment(p.Dtor_delta().Int64())
		if err != nil {
			return nil, checked(err)
		}
		if err := a.stock.Adjust(delta); err != nil {
			return nil, checked(err)
		}
		id := p.Dtor_product().VerbatimString(false)
		return map[string]any{"productId": id}, a.unit.Save("stock", id, *a.stock)
	case p.Is_UpdateOrder():
		return a.updateOrder(p)
	default:
		return nil, errors.New("unknown command plan")
	}
}

func (a *commandAttempt) saveCart(customer string, items map[string]int64, expires int64) error {
	cart, err := domain.RestoreCart(domain.CartSnapshot{CustomerID: customer, Items: items, Expires: expires})
	if err != nil {
		return checked(err)
	}
	return a.unit.Save("cart", customer, cart)
}

func (a *commandAttempt) changeStock(changes plan.StockChanges) error {
	var items map[string]int64
	if changes.Is_Ship() {
		items = readItems(changes.Dtor_quantities())
	} else {
		items = readItems(changes.Dtor_deltas())
	}
	// Acquire stock locks in the same order across transactions.
	for _, id := range sortedKeys(items) {
		stock, ok := a.stocks[id]
		if !ok {
			return fmt.Errorf("stock %s not loaded", id)
		}
		if changes.Is_Ship() {
			qty, err := domain.NewOrderQuantity(items[id])
			if err != nil {
				return checked(err)
			}
			err = stock.Ship(qty)
			if err != nil {
				return checked(err)
			}
		} else {
			delta, err := domain.NewReservationDelta(items[id])
			if err != nil {
				return checked(err)
			}
			if err := stock.Reserve(delta); err != nil {
				return checked(err)
			}
		}
		if err := a.unit.Save("stock", id, *stock); err != nil {
			return err
		}
	}
	return nil
}

func (a *commandAttempt) checkout(p plan.CommandPlan) (map[string]any, error) {
	command := a.command
	customer := p.Dtor_checkoutCustomer().VerbatimString(false)
	customerID, err := domain.ParseID[domain.ActorTag](customer)
	if err != nil {
		return nil, checked(err)
	}
	newOrderID := p.Dtor_newOrderId().VerbatimString(false)
	orderID, err := domain.ParseID[domain.OrderTag](newOrderID)
	if err != nil {
		return nil, checked(err)
	}
	address, err := domain.NewShippingAddress(command.Payload.Address)
	if err != nil {
		return nil, checked(err)
	}
	rate, err := domain.NewDiscountRate(p.Dtor_discount().Int64())
	if err != nil {
		return nil, checked(err)
	}
	items := readItems(p.Dtor_orderLines())
	lines := make([]domain.Line, 0, len(items))
	for _, id := range sortedKeys(items) {
		line, err := a.orderLine(id, items[id])
		if err != nil {
			return nil, err
		}
		lines = append(lines, line)
	}
	payload := command.Payload
	order, err := domain.NewOrder(orderID, customerID, lines, address, rate, payload.Coupon, payload.RequestedDate, payload.Note, a.now)
	if err != nil {
		return nil, checked(err)
	}
	if err := a.unit.Save("order", newOrderID, order); err != nil {
		return nil, err
	}
	return map[string]any{"orderId": newOrderID}, a.saveCart(customer, readItems(p.Dtor_remainingItems()), p.Dtor_cartExpiry().Int64())
}

func (a *commandAttempt) readDependencies(reads plan.Dependencies) error {
	if err := a.readProducts(reads.Dtor_products()); err != nil {
		return err
	}
	a.stocks = make(map[string]*domain.Stock)
	next := reads.Dtor_stocks().Iterator()
	for value, ok := next(); ok; value, ok = next() {
		id := value.(d.Sequence).VerbatimString(false)
		stock, err := loadStock(a.unit, id)
		if err != nil {
			return err
		}
		a.stocks[id] = stock
	}
	if loyalty := reads.Dtor_loyalty(); loyalty.Is_Id() {
		if err := a.unit.Load("loyalty", loyalty.Dtor_value().VerbatimString(false), &a.loyalty); err != nil {
			return err
		}
	}
	if order := reads.Dtor_newOrder(); order.Is_Id() {
		var existing domain.Order
		return a.unit.Load("order", order.Dtor_value().VerbatimString(false), &existing)
	}
	return nil
}

func (a *commandAttempt) readProducts(products d.Set) error {
	a.products = make(map[string]domain.Product)
	next := products.Iterator()
	for value, ok := next(); ok; value, ok = next() {
		id := value.(d.Sequence).VerbatimString(false)
		var raw []byte
		if err := a.unit.Tx.QueryRow(a.unit.Ctx, store.ProductQuery+" WHERE id=$1", id).Scan(&raw); err != nil {
			return err
		}
		var product domain.Product
		if err := json.Unmarshal(raw, &product); err != nil {
			return err
		}
		a.products[id] = product
	}
	return nil
}

func (a *commandAttempt) orderLine(id string, quantity int64) (domain.Line, error) {
	product, ok := a.products[id]
	if !ok {
		return domain.Line{}, fmt.Errorf("product %s not loaded", id)
	}
	pid, err := domain.ParseID[domain.ProductTag](id)
	if err != nil {
		return domain.Line{}, err
	}
	qty, err := domain.NewOrderQuantity(quantity)
	if err != nil {
		return domain.Line{}, checked(err)
	}
	price, err := domain.NewUnitPrice(product.Price)
	if err != nil {
		return domain.Line{}, checked(err)
	}
	line, err := domain.NewLine(pid, product.SKU, product.Name, product.Unit, qty, price)
	return line, checked(err)
}

func (a *commandAttempt) updateOrder(p plan.CommandPlan) (map[string]any, error) {
	target, err := domain.ParseOrderStatus(p.Dtor_target().Int())
	if err != nil {
		return nil, checked(err)
	}
	actor, err := domain.ParseID[domain.ActorTag](a.actor.ID)
	if err != nil {
		return nil, checked(err)
	}
	if err := a.prepareOrder(p.Dtor_prepare()); err != nil {
		return nil, checked(err)
	}
	if err := a.order.Move(target, actor, a.command.Payload.Note, a.now); err != nil {
		return nil, checked(err)
	}
	if err := a.changeStock(p.Dtor_stock()); err != nil {
		return nil, err
	}
	order := a.order.Snapshot()
	if err := a.unit.Save("order", order.ID, a.order); err != nil {
		return nil, err
	}
	next := p.Dtor_obligations().Iterator()
	for value, ok := next(); ok; value, ok = next() {
		obligation := value.(plan.DeliveryObligation)
		switch {
		case obligation.Is_AwardPoints():
			if err := a.awardPoints(order, obligation.Dtor_points().Int64()); err != nil {
				return nil, err
			}
		case obligation.Is_CreateInvoice():
			if err := a.createInvoice(order.ID); err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("unknown delivery obligation: %v", obligation)
		}
	}
	return map[string]any{"orderId": order.ID, "status": target.Code()}, nil
}

func (a *commandAttempt) prepareOrder(preparation plan.OrderPreparation) error {
	switch {
	case preparation.Is_Keep():
		return nil
	case preparation.Is_AssignWorker():
		worker, err := domain.ParseID[domain.ActorTag](preparation.Dtor_workerId().VerbatimString(false))
		if err != nil {
			return err
		}
		return a.order.AssignWorker(worker)
	case preparation.Is_AssignCarrier():
		carrier, err := domain.ParseID[domain.ActorTag](preparation.Dtor_carrierId().VerbatimString(false))
		if err != nil {
			return err
		}
		return a.order.AssignCarrier(carrier, preparation.Dtor_trackingNumber().VerbatimString(false))
	case preparation.Is_Pack():
		boxes, err := domain.NewBoxCount(preparation.Dtor_boxCount().Int())
		if err != nil {
			return err
		}
		weight, err := domain.NewWeightGrams(preparation.Dtor_weightGrams().Int64())
		if err != nil {
			return err
		}
		return a.order.Pack(boxes, weight)
	default:
		return fmt.Errorf("unknown order preparation: %v", preparation)
	}
}

func (a *commandAttempt) awardPoints(order domain.OrderSnapshot, amount int64) error {
	points, err := domain.NewPoints(amount)
	if err != nil {
		return checked(err)
	}
	if _, err := a.unit.Tx.Exec(a.unit.Ctx, "INSERT INTO order_point_awards(order_id) VALUES($1)", order.ID); err != nil {
		return err
	}
	if err := a.loyalty.Credit(points); err != nil {
		return err
	}
	return a.unit.Save("loyalty", order.CustomerID, a.loyalty)
}

func (a *commandAttempt) createInvoice(id string) error {
	pdf, err := InvoicePDF(a.order)
	if err != nil {
		return err
	}
	_, err = a.unit.Tx.Exec(a.unit.Ctx, "INSERT INTO invoices(id,order_id,pdf) VALUES($1,$2,$3) ON CONFLICT(order_id) DO NOTHING", store.ID("invoice/"+id), id, pdf)
	return err
}
