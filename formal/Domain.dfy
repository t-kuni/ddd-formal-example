module Domain {
  const StockLimit: int := 1000000000
  const LineCountLimit: int := 100
  const LineQuantityLimit: int := 10000
  const SubtotalLimit: int := 1000000000000

  predicate QuantityValid(v: int) { 0 <= v <= StockLimit }
  predicate StockDeltaValid(v: int) { -StockLimit <= v <= StockLimit }
  predicate OrderQuantityValid(v: int) { 1 <= v <= LineQuantityLimit }
  predicate CartQuantityValid(v: int) { 0 <= v <= LineQuantityLimit }
  predicate MoneyValid(v: int) { 0 <= v <= 1100000000000 }
  predicate UnitPriceValid(v: int) { 0 <= v <= 10000000 }
  predicate DiscountRateValid(v: int) { 0 <= v <= 100 }
  predicate PointsValid(v: int) { 0 <= v <= 1000000000000000 }
  predicate BoxCountValid(v: int) { 1 <= v <= 100 }
  predicate WeightValid(v: int) { 1 <= v <= 500000 }
  predicate TextFieldValid(nonBlank: bool, byteCount: int) { nonBlank && 1 <= byteCount <= 750 }

  method Credit(balance: int, award: int) returns (ok: bool, next: int)
    ensures ok <==> PointsValid(balance + award)
    ensures ok ==> next == balance + award && PointsValid(next)
    ensures !ok ==> next == balance
  {
    ok := PointsValid(balance + award);
    next := if ok then balance + award else balance;
  }

  predicate StockValid(h: int, r: int) { 0 <= r <= h <= StockLimit }

  method Reserve(h: int, r: int, delta: int) returns (ok: bool, next: int)
    requires StockValid(h,r)
    ensures ok <==> 0 <= r + delta <= h
    ensures ok ==> next == r + delta && StockValid(h,next)
    ensures !ok ==> next == r
  {
    ok := 0 <= r + delta <= h;
    next := if ok then r + delta else r;
  }

  method Adjust(h: int, r: int, delta: int) returns (ok: bool, next: int)
    requires StockValid(h,r)
    ensures ok <==> r <= h + delta <= StockLimit
    ensures ok ==> next == h + delta && StockValid(next,r)
    ensures !ok ==> next == h
  {
    ok := r <= h + delta <= StockLimit;
    next := if ok then h + delta else h;
  }

  method Ship(h: int, r: int, qty: int) returns (ok: bool, nh: int, nr: int)
    requires StockValid(h,r)
    ensures ok <==> 0 < qty <= r
    ensures ok ==> nh == h-qty && nr == r-qty && StockValid(nh,nr)
    ensures ok ==> nh-nr == h-r
    ensures !ok ==> nh == h && nr == r
  {
    ok := 0 < qty <= r;
    nh := if ok then h-qty else h;
    nr := if ok then r-qty else r;
  }

  // 0 payment, 1 accepted, 2 picking, 3 blocked, 4 packed,
  // 5 shipping, 6 delivery_failed, 7 delivered, 8 cancelled,
  // 9 payment_failed, 10 returned.
  predicate StatusValid(s: int) { 0 <= s <= 10 }
  predicate Allowed(s: int, t: int) {
    (s == 0 && (t == 1 || t == 8 || t == 9)) ||
    (s == 1 && (t == 2 || t == 8)) ||
    (s == 2 && (t == 3 || t == 4 || t == 8)) ||
    (s == 3 && (t == 2 || t == 8)) ||
    (s == 4 && t == 5) ||
    (s == 5 && (t == 6 || t == 7)) ||
    (s == 6 && (t == 5 || t == 10))
  }
  method Transition(s: int, t: int) returns (ok: bool, next: int)
    requires StatusValid(s)
    ensures ok <==> Allowed(s,t)
    ensures next == (if ok then t else s)
    ensures StatusValid(next)
    ensures s in {7,8,9,10} ==> !ok && next == s
  { ok := Allowed(s,t); next := if ok then t else s; }

  function Touch(now: int): (expires: int)
    requires TimestampValid(now)
    ensures expires == now + 3600
    ensures expires > now
  { now + 3600 }

  predicate Expired(now: int, expires: int) { now >= expires }

  function MayExpire(now: int, current: int, requested: int, hasItems: bool): (expire: bool)
    ensures expire <==> now >= current && current == requested && hasItems
    ensures current != requested ==> !expire
  { Expired(now, current) && current == requested && hasItems }

  function MayCheckout(now: int, expires: int, hasItems: bool): (allowed: bool)
    ensures allowed <==> now < expires && hasItems
    ensures now >= expires ==> !allowed
  { !Expired(now, expires) && hasItems }

  method Price(subtotal: int, percent: int) returns (discount: int, net: int, tax: int, total: int, points: int)
    requires 0 <= subtotal <= SubtotalLimit
    requires DiscountRateValid(percent)
    ensures discount == subtotal * percent / 100
    ensures net == subtotal - discount && 0 <= net <= subtotal
    ensures tax == net / 10 && total == net + tax
    ensures MoneyValid(total)
    ensures points == net / 100 && PointsValid(points)
  {
    discount := subtotal * percent / 100;
    assert 0 <= subtotal * percent <= subtotal * 100;
    net := subtotal - discount;
    tax := net / 10;
    total := net + tax;
    points := net / 100;
  }

  predicate TimestampValid(seconds: int) { seconds >= 0 }

  predicate CartShapeValid(count: int, expires: int) {
    0 <= count <= LineCountLimit && TimestampValid(expires)
  }

  datatype CartChange = CartFull | CartUnchanged | CartTimeOutOfRange |
    CartChanged(reset: bool, delta: int, expiry: int)

  function PlanCartChange(count: int, previousQuantity: int, requested: int, now: int, expires: int): (plan: CartChange)
    requires CartShapeValid(count, expires) && TimestampValid(now)
    requires CartQuantityValid(previousQuantity) && CartQuantityValid(requested)
    requires count == 0 ==> previousQuantity == 0
    ensures plan.CartFull? <==> count == LineCountLimit && previousQuantity == 0 && requested > 0 && now < expires
    ensures plan.CartUnchanged? <==> (count == 0 && requested == 0) || (previousQuantity == requested && now < expires)
    ensures plan.CartChanged? ==> (plan.reset == (now >= expires) &&
      plan.delta == requested - previousQuantity && StockDeltaValid(plan.delta) &&
      plan.expiry == now + 3600 && now < plan.expiry <= 9223372036854775807)
    ensures plan.CartChanged? && !plan.reset && previousQuantity == 0 && requested > 0 ==> count < LineCountLimit
  {
    if count == LineCountLimit && previousQuantity == 0 && requested > 0 && now < expires then CartFull
    else if (count == 0 && requested == 0) || (previousQuantity == requested && now < expires) then CartUnchanged
    else if now > 9223372036854775807 - 3600 then CartTimeOutOfRange
    else CartChanged(Expired(now, expires), requested - previousQuantity, Touch(now))
  }

  predicate OrderLineCountValid(count: int) { 1 <= count <= LineCountLimit }
  predicate LineValid(hasProduct: bool, hasName: bool, quantity: int, unitPrice: int) {
    hasProduct && hasName && OrderQuantityValid(quantity) && UnitPriceValid(unitPrice)
  }

  method RegisterProduct(seen: set<string>, product: string) returns (ok: bool, next: set<string>)
    ensures ok <==> product !in seen
    ensures next == (if ok then seen + {product} else seen)
  {
    ok := product !in seen;
    next := if ok then seen + {product} else seen;
  }

  function LineAmount(quantity: int, unitPrice: int): int { quantity * unitPrice }

  method AddLineAmount(subtotal: int, quantity: int, unitPrice: int) returns (ok: bool, next: int)
    requires 0 <= subtotal <= SubtotalLimit
    requires OrderQuantityValid(quantity) && UnitPriceValid(unitPrice)
    ensures ok <==> subtotal + LineAmount(quantity, unitPrice) <= SubtotalLimit
    ensures ok ==> next == subtotal + LineAmount(quantity, unitPrice) && 0 <= next <= SubtotalLimit
    ensures !ok ==> next == subtotal
  {
    var candidate := subtotal + LineAmount(quantity, unitPrice);
    ok := candidate <= SubtotalLimit;
    next := if ok then candidate else subtotal;
  }

  predicate InitialHistory(from: int, to: int) { from == -1 && to == 0 }
  datatype HistoryEdge = HistoryEdge(from: int, to: int)
  datatype HistoryCheck = InvalidStart | InvalidStep | StatusMismatch | ValidHistory

  predicate HistoryStepsValid(edges: seq<HistoryEdge>) {
    forall i | 1 <= i < |edges| ::
      edges[i].from == edges[i-1].to && Allowed(edges[i].from, edges[i].to)
  }

  method CheckHistory(edges: seq<HistoryEdge>, status: int) returns (result: HistoryCheck)
    ensures result.ValidHistory? <==> (|edges| > 0 && InitialHistory(edges[0].from, edges[0].to) &&
      HistoryStepsValid(edges) && edges[|edges|-1].to == status)
  {
    if |edges| == 0 || !InitialHistory(edges[0].from, edges[0].to) {
      return InvalidStart;
    }
    if !HistoryStepsValid(edges) {
      return InvalidStep;
    }
    return if edges[|edges|-1].to == status then ValidHistory else StatusMismatch;
  }

  predicate NeedsWorker(status: int) { status in {2,3,4,5,6,7,10} }
  predicate NeedsBoxes(status: int) { status in {4,5,6,7,10} }
  predicate NeedsCarrier(status: int) { status in {5,6,7,10} }

  predicate WorkerReady(status: int, hasAssignee: bool) { !NeedsWorker(status) || hasAssignee }
  predicate BoxesReady(status: int, count: int) { !NeedsBoxes(status) || BoxCountValid(count) }
  predicate BoxValid(parsedID: bool, weight: int) { parsedID && WeightValid(weight) }
  predicate CarrierReady(status: int, hasAssignee: bool, hasTracking: bool) {
    !NeedsCarrier(status) || (hasAssignee && hasTracking)
  }

  predicate MayStartWork(status: int, validActor: bool) { validActor && Allowed(status, 2) }
  predicate MayAssignWorker(hasAssignee: bool, sameActor: bool) { !hasAssignee || sameActor }
  predicate MayPack(status: int, count: int, weight: int) {
    Allowed(status, 4) && BoxCountValid(count) && WeightValid(weight)
  }
  predicate MayAssignCarrier(status: int, validActor: bool, trackingBytes: int) {
    status == 4 && validActor && 1 <= trackingBytes <= 100
  }
}
