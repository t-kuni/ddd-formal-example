package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	core "wholesale.local/orbit/generated/Domain"
	d "wholesale.local/orbit/generated/dafny"
)

var ErrConflict = errors.New("更新が競合しました。再読込してください")

func Invalid(s string) error { return errors.New(s) }

type Stock struct {
	onHand   Quantity
	reserved Quantity
}
type StockSnapshot struct {
	OnHand   int64 `json:"onHand"`
	Reserved int64 `json:"reserved"`
}

func NewStock(h, r Quantity) (Stock, error) {
	s := Stock{h, r}
	if !s.Valid() {
		return Stock{}, Invalid("在庫不変条件違反")
	}
	return s, nil
}
func (s Stock) Snapshot() StockSnapshot { return StockSnapshot{s.onHand.Units(), s.reserved.Units()} }
func (s Stock) Valid() bool {
	return core.Companion_Default___.StockValid(d.IntOfInt64(s.onHand.Units()), d.IntOfInt64(s.reserved.Units()))
}
func (s Stock) MarshalJSON() ([]byte, error) { return json.Marshal(s.Snapshot()) }
func (s *Stock) UnmarshalJSON(b []byte) error {
	var x StockSnapshot
	if e := json.Unmarshal(b, &x); e != nil {
		return e
	}
	h, e := NewQuantity(x.OnHand)
	if e != nil {
		return e
	}
	r, e := NewQuantity(x.Reserved)
	if e != nil {
		return e
	}
	v, e := NewStock(h, r)
	if e == nil {
		*s = v
	}
	return e
}
func (s *Stock) Reserve(q ReservationDelta) error {
	if !s.Valid() {
		return Invalid("在庫不変条件違反")
	}
	ok, r := core.Companion_Default___.Reserve(d.IntOfInt64(s.onHand.Units()), d.IntOfInt64(s.reserved.Units()), d.IntOfInt64(q.Units()))
	if !ok {
		return Invalid("販売可能な在庫が不足しています")
	}
	s.reserved = Quantity{r.Int64()}
	return nil
}
func (s *Stock) Adjust(q StockAdjustment) error {
	if !s.Valid() {
		return Invalid("在庫不変条件違反")
	}
	ok, h := core.Companion_Default___.Adjust(d.IntOfInt64(s.onHand.Units()), d.IntOfInt64(s.reserved.Units()), d.IntOfInt64(q.Units()))
	if !ok {
		return Invalid("引当済み数量を下回る在庫調整はできません")
	}
	s.onHand = Quantity{h.Int64()}
	return nil
}
func (s *Stock) Ship(q OrderQuantity) error {
	if !s.Valid() {
		return Invalid("在庫不変条件違反")
	}
	ok, h, r := core.Companion_Default___.Ship(d.IntOfInt64(s.onHand.Units()), d.IntOfInt64(s.reserved.Units()), d.IntOfInt64(q.Units()))
	if !ok {
		return Invalid("出荷数量が引当を超えています")
	}
	s.onHand = Quantity{h.Int64()}
	s.reserved = Quantity{r.Int64()}
	return nil
}

type Cart struct {
	customer ActorID
	items    map[ProductID]OrderQuantity
	expires  int64
}
type CartSnapshot struct {
	CustomerID string           `json:"customerId"`
	Items      map[string]int64 `json:"items"`
	Expires    int64            `json:"expires"`
}

func NewCart(id ActorID) Cart { return Cart{customer: id, items: map[ProductID]OrderQuantity{}} }
func (c Cart) Snapshot() CartSnapshot {
	x := CartSnapshot{CustomerID: c.customer.String(), Items: map[string]int64{}, Expires: c.expires}
	for id, q := range c.items {
		x.Items[id.String()] = q.Units()
	}
	return x
}
func (c Cart) MarshalJSON() ([]byte, error) {
	if !c.customer.Valid() {
		return nil, Invalid("未初期化のカート")
	}
	return json.Marshal(c.Snapshot())
}
func (c *Cart) UnmarshalJSON(b []byte) error {
	var x CartSnapshot
	if e := json.Unmarshal(b, &x); e != nil {
		return e
	}
	id, e := ParseID[ActorTag](x.CustomerID)
	if e != nil {
		return e
	}
	v := NewCart(id)
	if !core.Companion_Default___.CartShapeValid(d.IntOf(len(x.Items)), d.IntOfInt64(x.Expires)) {
		return Invalid("カート不変条件違反")
	}
	v.expires = x.Expires
	for key, q := range x.Items {
		pid, e := ParseID[ProductTag](key)
		if e != nil {
			return e
		}
		qty, e := NewOrderQuantity(q)
		if e != nil {
			return e
		}
		v.items[pid] = qty
	}
	*c = v
	return nil
}
func (c Cart) Expired(now int64) bool {
	return core.Companion_Default___.Expired(d.IntOfInt64(now), d.IntOfInt64(c.expires))
}
func (c *Cart) Change(id ProductID, requested CartQuantity, now int64) (map[string]int64, bool, error) {
	quantity := requested.Units()
	if !c.customer.Valid() || !id.Valid() || !core.Companion_Default___.TimestampValid(d.IntOfInt64(now)) {
		return nil, false, Invalid("未初期化のカート・ID・時刻")
	}
	var qty OrderQuantity
	var e error
	if quantity != 0 {
		qty, e = NewOrderQuantity(quantity)
		if e != nil {
			return nil, false, e
		}
	}
	plan := core.Companion_Default___.PlanCartChange(d.IntOf(len(c.items)),
		d.IntOfInt64(c.items[id].Units()), d.IntOfInt64(quantity), d.IntOfInt64(now), d.IntOfInt64(c.expires))
	switch {
	case plan.Is_CartFull():
		return nil, false, Invalid("カートは100商品までです")
	case plan.Is_CartUnchanged():
		return nil, false, nil
	case plan.Is_CartTimeOutOfRange():
		return nil, false, Invalid("カートの有効期限が範囲外です")
	}
	delta := map[string]int64{}
	if plan.Dtor_reset() {
		for pid, q := range c.items {
			delta[pid.String()] = core.Companion_Default___.ReleasedReservation(d.IntOfInt64(q.Units())).Int64()
		}
		c.items = map[ProductID]OrderQuantity{}
	}
	delta[id.String()] = plan.Dtor_delta().Int64()
	if quantity == 0 {
		delete(c.items, id)
	} else {
		c.items[id] = qty
	}
	c.expires = plan.Dtor_expiry().Int64()
	return delta, true, nil
}
func (c *Cart) Expire(requested, now int64) (map[string]int64, bool) {
	if !core.Companion_Default___.MayExpire(d.IntOfInt64(now), d.IntOfInt64(c.expires), d.IntOfInt64(requested), len(c.items) > 0) {
		return nil, false
	}
	delta := map[string]int64{}
	for id, q := range c.items {
		delta[id.String()] = core.Companion_Default___.ReleasedReservation(d.IntOfInt64(q.Units())).Int64()
	}
	c.items = map[ProductID]OrderQuantity{}
	return delta, true
}
func (c *Cart) Checkout(now int64) (map[string]int64, error) {
	if !core.Companion_Default___.MayCheckout(d.IntOfInt64(now), d.IntOfInt64(c.expires), len(c.items) > 0) {
		return nil, Invalid("カートが空、または有効期限切れです")
	}
	items := c.Snapshot().Items
	c.items = map[ProductID]OrderQuantity{}
	return items, nil
}

var StatusNames = []string{"決済待ち", "受注確定", "集品中", "作業保留", "梱包済み", "配送中", "配送失敗", "配送完了", "キャンセル", "決済失敗", "倉庫返送"}

type Line struct {
	product         ProductID
	sku, name, unit string
	quantity        OrderQuantity
	price           UnitPrice
}
type LineSnapshot struct {
	ProductID string `json:"productId"`
	SKU       string `json:"sku"`
	Name      string `json:"name"`
	Unit      string `json:"unit"`
	Quantity  int64  `json:"quantity"`
	Price     int64  `json:"price"`
}

func NewLine(id ProductID, sku, name, unit string, q OrderQuantity, p UnitPrice) (Line, error) {
	if !core.Companion_Default___.LineValid(id.Valid(), name != "", d.IntOfInt64(q.Units()), d.IntOfInt64(p.Yen())) {
		return Line{}, Invalid("商品明細が不正です")
	}
	return Line{id, sku, name, unit, q, p}, nil
}
func (l Line) Snapshot() LineSnapshot {
	return LineSnapshot{l.product.String(), l.sku, l.name, l.unit, l.quantity.Units(), l.price.Yen()}
}

type Address struct {
	Company   string `json:"company"`
	Recipient string `json:"recipient"`
	Postal    string `json:"postal"`
	Text      string `json:"text"`
	Phone     string `json:"phone"`
}
type Box struct {
	ID     string `json:"id"`
	Label  string `json:"label"`
	Weight int64  `json:"weightGrams"`
}
type History struct {
	At    time.Time `json:"at"`
	Actor string    `json:"actor"`
	From  int       `json:"from"`
	To    int       `json:"to"`
	Note  string    `json:"note"`
}
type OrderSnapshot struct {
	ID              string         `json:"id"`
	Number          string         `json:"number"`
	CustomerID      string         `json:"customerId"`
	Status          int            `json:"status"`
	Lines           []LineSnapshot `json:"lines"`
	Address         Address        `json:"address"`
	RequestedDate   string         `json:"requestedDate"`
	Note            string         `json:"note"`
	Coupon          string         `json:"coupon"`
	DiscountPercent int64          `json:"discountPercent"`
	Subtotal        int64          `json:"subtotal"`
	Discount        int64          `json:"discount"`
	Net             int64          `json:"net"`
	Tax             int64          `json:"tax"`
	Total           int64          `json:"total"`
	Points          int64          `json:"points"`
	WorkerID        string         `json:"workerId"`
	CarrierID       string         `json:"carrierId"`
	Tracking        string         `json:"tracking"`
	Boxes           []Box          `json:"boxes"`
	History         []History      `json:"history"`
	CreatedAt       time.Time      `json:"createdAt"`
}
type Order struct {
	id                                  OrderID
	customer                            ActorID
	status                              OrderStatus
	lines                               []Line
	address                             ShippingAddress
	number, requestedDate, note, coupon string
	rate                                DiscountRate
	subtotal, discount, net, tax, total Money
	points                              Points
	worker, carrier                     ActorID
	tracking                            string
	boxes                               []Box
	history                             []History
	createdAt                           time.Time
}

func NewOrder(id OrderID, customer ActorID, lines []Line, address ShippingAddress, rate DiscountRate, coupon, requested, note string, now time.Time) (Order, error) {
	if !id.Valid() || !customer.Valid() || !core.Companion_Default___.OrderLineCountValid(d.IntOf(len(lines))) || !core.Companion_Default___.TimestampValid(d.IntOfInt64(now.Unix())) {
		return Order{}, Invalid("受注の必須項目が不正です")
	}
	if _, e := NewShippingAddress(address.Snapshot()); e != nil {
		return Order{}, e
	}
	o := Order{id: id, customer: customer, status: PaymentPending, lines: append([]Line(nil), lines...), address: address, rate: rate, coupon: coupon, requestedDate: requested, note: note, createdAt: now, number: "OR-" + now.Format("060102") + "-" + id.String()[:8], boxes: []Box{}, history: []History{{At: now, Actor: customer.String(), From: -1, To: 0, Note: "発注・引当をカートから移管"}}}
	if e := o.calculate(); e != nil {
		return Order{}, e
	}
	return o, nil
}
func (o *Order) calculate() error {
	sum := d.Zero
	seen := d.SetOf()
	for _, l := range o.lines {
		var ok bool
		ok, seen = core.Companion_Default___.RegisterProduct(seen, d.SeqOfString(l.product.String()))
		if !ok {
			return Invalid("同一商品が明細に重複しています")
		}
		quantity, unitPrice := d.IntOfInt64(l.quantity.Units()), d.IntOfInt64(l.price.Yen())
		if !core.Companion_Default___.LineValid(l.product.Valid(), l.name != "", quantity, unitPrice) {
			return Invalid("明細が不正です")
		}
		ok, sum = core.Companion_Default___.AddLineAmount(sum, quantity, unitPrice)
		if !ok {
			return Invalid("受注金額の上限を超えています")
		}
	}
	discount, net, tax, total, points := core.Companion_Default___.Price(sum, d.IntOfInt64(o.rate.Percent()))
	o.subtotal = Money{sum.Int64()}
	o.discount = Money{discount.Int64()}
	o.net = Money{net.Int64()}
	o.tax = Money{tax.Int64()}
	o.total = Money{total.Int64()}
	o.points = Points{points.Int64()}
	return nil
}
func (o Order) EarnedPoints() Points { return o.points }
func (o Order) Status() OrderStatus  { return o.status }
func (o Order) Exists() bool         { return o.id.Valid() }
func (o Order) Snapshot() OrderSnapshot {
	x := OrderSnapshot{ID: o.id.String(), Number: o.number, CustomerID: o.customer.String(), Status: o.status.Code(), Address: o.address.Snapshot(), RequestedDate: o.requestedDate, Note: o.note, Coupon: o.coupon, DiscountPercent: o.rate.Percent(), Subtotal: o.subtotal.Yen(), Discount: o.discount.Yen(), Net: o.net.Yen(), Tax: o.tax.Yen(), Total: o.total.Yen(), Points: o.points.Value(), Tracking: o.tracking, Boxes: append([]Box{}, o.boxes...), History: append([]History{}, o.history...), CreatedAt: o.createdAt, Lines: []LineSnapshot{}}
	if o.worker.Valid() {
		x.WorkerID = o.worker.String()
	}
	if o.carrier.Valid() {
		x.CarrierID = o.carrier.String()
	}
	for _, l := range o.lines {
		x.Lines = append(x.Lines, l.Snapshot())
	}
	return x
}
func (o Order) MarshalJSON() ([]byte, error) {
	if !o.Exists() {
		return nil, Invalid("未初期化の受注")
	}
	if e := o.validateFulfillment(o.status); e != nil {
		return nil, e
	}
	return json.Marshal(o.Snapshot())
}
func (o *Order) UnmarshalJSON(b []byte) error {
	var x OrderSnapshot
	if e := json.Unmarshal(b, &x); e != nil {
		return e
	}
	id, e := ParseID[OrderTag](x.ID)
	if e != nil {
		return e
	}
	customer, e := ParseID[ActorTag](x.CustomerID)
	if e != nil {
		return e
	}
	address, e := NewShippingAddress(x.Address)
	if e != nil {
		return e
	}
	rate, e := NewDiscountRate(x.DiscountPercent)
	if e != nil {
		return e
	}
	status, e := ParseOrderStatus(x.Status)
	if e != nil {
		return e
	}
	lines := []Line{}
	for _, l := range x.Lines {
		pid, e := ParseID[ProductTag](l.ProductID)
		if e != nil {
			return e
		}
		qty, e := NewOrderQuantity(l.Quantity)
		if e != nil {
			return e
		}
		price, e := NewUnitPrice(l.Price)
		if e != nil {
			return e
		}
		line, e := NewLine(pid, l.SKU, l.Name, l.Unit, qty, price)
		if e != nil {
			return e
		}
		lines = append(lines, line)
	}
	v, e := NewOrder(id, customer, lines, address, rate, x.Coupon, x.RequestedDate, x.Note, x.CreatedAt)
	if e != nil {
		return e
	}
	if v.total.Yen() != x.Total || v.subtotal.Yen() != x.Subtotal || v.discount.Yen() != x.Discount || v.net.Yen() != x.Net || v.tax.Yen() != x.Tax || v.points.Value() != x.Points {
		return Invalid("保存された受注金額が不変条件に違反しています")
	}
	edges := make([]interface{}, len(x.History))
	for i, h := range x.History {
		edges[i] = core.Companion_HistoryEdge_.Create_HistoryEdge_(d.IntOf(h.From), d.IntOf(h.To))
	}
	check := core.Companion_Default___.CheckHistory(d.SeqOf(edges...), d.IntOf(x.Status))
	switch {
	case check.Is_InvalidStart():
		return Invalid("受注履歴の初期状態が不正です")
	case check.Is_InvalidStep():
		return Invalid("受注履歴の遷移が不正です")
	case check.Is_StatusMismatch():
		return Invalid("受注状態と履歴が一致しません")
	}
	for _, h := range x.History {
		if _, e := ParseID[ActorTag](h.Actor); e != nil {
			return e
		}
	}
	v.status = status
	v.number = x.Number
	v.tracking = x.Tracking
	v.boxes = append([]Box{}, x.Boxes...)
	v.history = append([]History{}, x.History...)
	if x.WorkerID != "" {
		v.worker, e = ParseID[ActorTag](x.WorkerID)
		if e != nil {
			return e
		}
	}
	if x.CarrierID != "" {
		v.carrier, e = ParseID[ActorTag](x.CarrierID)
		if e != nil {
			return e
		}
	}
	if e = v.validateFulfillment(status); e != nil {
		return e
	}
	*o = v
	return nil
}
func (o Order) validateFulfillment(next OrderStatus) error {
	status := d.IntOf(next.Code())
	if !core.Companion_Default___.WorkerReady(status, o.worker.Valid()) {
		return Invalid("作業担当者が未設定です")
	}
	if !core.Companion_Default___.BoxesReady(status, d.IntOf(len(o.boxes))) {
		return Invalid("箱情報が未設定です")
	}
	if core.Companion_Default___.NeedsBoxes(status) {
		for _, b := range o.boxes {
			_, e := uuid.Parse(b.ID)
			if !core.Companion_Default___.BoxValid(e == nil, d.IntOfInt64(b.Weight)) {
				return Invalid("箱情報が不正です")
			}
		}
	}
	if !core.Companion_Default___.CarrierReady(status, o.carrier.Valid(), o.tracking != "") {
		return Invalid("配送情報が未設定です")
	}
	return nil
}
func (o *Order) AssignWorker(actor ActorID) error {
	if !core.Companion_Default___.MayStartWork(d.IntOf(o.status.Code()), actor.Valid()) {
		return Invalid("作業開始できない状態です")
	}
	if !core.Companion_Default___.MayAssignWorker(o.worker.Valid(), o.worker == actor) {
		return Invalid("別の担当者が作業中です")
	}
	o.worker = actor
	return nil
}
func (o *Order) Pack(boxes BoxCount, grams WeightGrams) error {
	count, weight := boxes.value, grams.value
	if !core.Companion_Default___.MayPack(d.IntOf(o.status.Code()), d.IntOf(count), d.IntOfInt64(weight)) {
		return Invalid("梱包状態・箱数・重量が不正です")
	}
	o.boxes = []Box{}
	for i := 0; i < count; i++ {
		o.boxes = append(o.boxes, Box{uuid.NewString(), fmt.Sprintf("%s / %d", o.number, i+1), weight})
	}
	return nil
}
func (o *Order) AssignCarrier(actor ActorID, tracking string) error {
	if !core.Companion_Default___.MayAssignCarrier(d.IntOf(o.status.Code()), actor.Valid(), d.IntOf(len(tracking))) {
		return Invalid("配送情報が不正です")
	}
	o.carrier = actor
	o.tracking = tracking
	return nil
}
func (o *Order) Move(next OrderStatus, actor ActorID, note string, now time.Time) error {
	if !o.Exists() || !actor.Valid() {
		return Invalid("未初期化の受注・担当者")
	}
	ok, n := core.Companion_Default___.Transition(d.IntOfInt64(int64(o.status.Code())), d.IntOfInt64(int64(next.Code())))
	if !ok {
		return Invalid(fmt.Sprintf("%s からの操作は許可されていません", StatusNames[o.status.Code()]))
	}
	if e := o.validateFulfillment(next); e != nil {
		return e
	}
	o.history = append(o.history, History{now, actor.String(), o.status.Code(), next.Code(), note})
	o.status = OrderStatus(n.Int64())
	return nil
}

type Loyalty struct{ balance Points }

func (l Loyalty) Balance() Points { return l.balance }
func (l *Loyalty) Credit(p Points) error {
	ok, next := core.Companion_Default___.Credit(d.IntOfInt64(l.balance.Value()), d.IntOfInt64(p.Value()))
	if !ok {
		return Invalid("ポイントが範囲外です")
	}
	l.balance = Points{next.Int64()}
	return nil
}
func (l Loyalty) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Points int64 `json:"points"`
	}{l.balance.Value()})
}
func (l *Loyalty) UnmarshalJSON(b []byte) error {
	var x struct {
		Points int64 `json:"points"`
	}
	if e := json.Unmarshal(b, &x); e != nil {
		return e
	}
	v, e := NewPoints(x.Points)
	if e == nil {
		l.balance = v
	}
	return e
}

type Actor struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	Role    string  `json:"role"`
	Email   string  `json:"email"`
	Address Address `json:"address"`
}
type Product struct {
	ID       string `json:"id"`
	SKU      string `json:"sku"`
	Name     string `json:"name"`
	Category string `json:"category"`
	Unit     string `json:"unit"`
	Price    int64  `json:"price"`
	Image    string `json:"image"`
}
