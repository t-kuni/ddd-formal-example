module CommandModel {
  const PaymentPending := 0
  const Accepted := 1
  const Picking := 2
  const PickBlocked := 3
  const Packed := 4
  const Shipping := 5
  const DeliveryFailed := 6
  const Delivered := 7
  const Cancelled := 8
  const PaymentFailed := 9
  const Returned := 10
  const PaymentTimeoutNanos := 15 * 60 * 1000000000
  const StaleVersion := "更新が競合しました。再読込してください"

  datatype CommandInput = CommandInput(kind: string, role: string, actor: string, customer: string,
    noteBytes: int, trackingBytes: int, requestedDateBytes: int, expectedVersion: int,
    product: string, productValid: bool, stockDelta: int, quantity: int,
    now: int, expireAt: int, paymentSucceeded: bool, tracking: string, boxes: int, weight: int,
    orderId: string, couponCode: string, commandId: string)

  datatype OrderState = OrderState(customer: string, worker: string, carrier: string,
    status: int, ageNanos: int, points: int)
  datatype OrderDecision = Move(target: int) | NoChange | Denied(message: string)

  datatype Coupon = NoCoupon | UnknownCoupon | CouponRate(percent: int)
  datatype RequestedDay = Unspecified | InvalidDate | CalendarDay(date: int)
  datatype CartState = CartState(customerId: string, version: int, items: map<string, int>,
    expires: int, requestedDay: RequestedDay, today: int, coupon: Coupon)

  datatype ReadRequest = RejectRead(reason: string) | ReadCart(customerId: string) |
    ReadStock(productId: string) | ReadOrder(orderId: string)

  datatype Snapshot = Missing(message: string) | Cart(state: CartState) |
    Stock(stockVersion: int) | Order(orderVersion: int, order: OrderState, lines: map<string, int>)

  datatype OrderPreparation = Keep | AssignWorker(workerId: string) |
    Pack(boxCount: int, weightGrams: int) | AssignCarrier(carrierId: string, trackingNumber: string)
  datatype StockChanges = Reserve(deltas: map<string, int>) | Ship(quantities: map<string, int>)
  datatype DeliveryObligation = AwardPoints(points: int) | CreateInvoice

  datatype CommandPlan = Reject(reason: string) | Unchanged |
    UpdateCart(customerId: string, nextItems: map<string, int>, expiry: int,
      reservations: map<string, int>, expired: bool) |
    Checkout(checkoutCustomer: string, newOrderId: string, orderLines: map<string, int>,
      remainingItems: map<string, int>, cartExpiry: int, discount: int) |
    Adjust(product: string, delta: int) |
    UpdateOrder(target: int, orderCustomer: string, prepare: OrderPreparation, stock: StockChanges,
      obligations: set<DeliveryObligation>)

  predicate VersionMatches(expected: int, actual: int) {
    expected <= 0 || expected == actual
  }

  function Release(lines: map<string, int>): (deltas: map<string, int>)
    ensures deltas.Keys == lines.Keys
    ensures forall id | id in lines :: deltas[id] == -lines[id]
  {
    map id | id in lines :: -lines[id]
  }

  function WithQuantity(items: map<string, int>, product: string, quantity: int): map<string, int> {
    if quantity == 0 then items - {product} else items[product := quantity]
  }

  predicate InputSizeValid(input: CommandInput) {
    input.noteBytes <= 1000 && input.trackingBytes <= 100 && input.requestedDateBytes <= 10
  }

  datatype Reference = Absent | Id(value: string)
  datatype Dependencies = Dependencies(products: set<string>, stocks: set<string>,
    loyalty: Reference, newOrder: Reference)

  datatype Status = Ok | Failed | Conflict | Rejected | Skipped

}
