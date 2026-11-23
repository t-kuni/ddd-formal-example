module CommandFlow {
  import Domain

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

  function DecideOrder(input: CommandInput, x: OrderState): (decision: OrderDecision)
    ensures decision.Move? && input.role == "customer" ==> input.actor == x.customer
    ensures decision.Move? && input.kind == "order.deliver" ==>
      input.role == "carrier" && input.actor == x.carrier && decision.target == Delivered
    ensures input.kind == "order.deliver" && input.role == "carrier" && input.actor == x.carrier ==>
      decision == Move(Delivered)
    ensures decision.Move? && input.kind == "order.cancel" && input.role == "system" ==>
      x.status == PaymentPending && x.ageNanos >= PaymentTimeoutNanos
  {
    var kind := input.kind;
    if input.role == "customer" && input.actor != x.customer then
      Denied("他の顧客の受注は操作できません")
    else if kind == "order.pay" then
      if input.role != "customer" then Denied("この利用者には操作権限がありません")
      else Move(if input.paymentSucceeded then Accepted else PaymentFailed)
    else if kind == "order.cancel" then
      if input.role !in {"customer", "operator", "system"} then Denied("この利用者には操作権限がありません")
      else if input.role == "system" && (x.status != PaymentPending || x.ageNanos < PaymentTimeoutNanos) then NoChange
      else if input.role == "customer" && x.status !in {PaymentPending, Accepted} then
        Denied("作業開始後のキャンセルは運用者に依頼してください")
      else Move(Cancelled)
    else if kind in {"order.pick", "order.block", "order.pack"} then
      if input.role != "warehouse" then Denied("この利用者には操作権限がありません")
      else if kind == "order.pick" then
        if !Domain.MayAssignWorker(x.worker != "", x.worker == input.actor) then Denied("別の担当者が作業中です")
        else Move(Picking)
      else if x.worker != input.actor then Denied("担当外の作業です")
      else if kind == "order.block" then
        if input.noteBytes == 0 then Denied("保留理由が必要です") else Move(PickBlocked)
      else if !Domain.BoxCountValid(input.boxes) || !Domain.WeightValid(input.weight) then
        Denied("箱数は1〜100、1箱の重量は1〜500000gです")
      else Move(Packed)
    else if kind in {"order.dispatch", "order.delivery_failed", "order.redeliver", "order.deliver", "order.return"} then
      if input.role != "carrier" then Denied("この利用者には操作権限がありません")
      else if kind == "order.dispatch" then
        if input.tracking == "" then Denied("追跡番号を入力してください")
        else if x.status != Packed then Denied("梱包完了後に出荷してください")
        else Move(Shipping)
      else if x.carrier != input.actor then Denied("担当外の配送です")
      else if kind == "order.delivery_failed" then
        if input.noteBytes == 0 then Denied("配送失敗理由が必要です") else Move(DeliveryFailed)
      else if kind == "order.redeliver" then
        if x.status != DeliveryFailed then Denied("再配送できる状態ではありません") else Move(Shipping)
      else if kind == "order.deliver" then Move(Delivered)
      else if input.noteBytes == 0 then Denied("返送理由が必要です") else Move(Returned)
    else Denied("未対応のCommandです")
  }

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

  function PlanCart(input: CommandInput, cart: CartState): (plan: CommandPlan)
    ensures plan.Checkout? ==> (input.kind == "order.checkout" && plan.checkoutCustomer == cart.customerId &&
      plan.orderLines == cart.items && plan.remainingItems == map[] && plan.cartExpiry == cart.expires &&
      Domain.MayCheckout(input.now, cart.expires, |cart.items| > 0))
    ensures plan.UpdateCart? && plan.expired ==> (input.kind == "cart.expire" && plan.customerId == cart.customerId &&
      plan.nextItems == map[] &&
      plan.expiry == cart.expires && plan.reservations == Release(cart.items) &&
      Domain.MayExpire(input.now, cart.expires, input.expireAt, |cart.items| > 0))
    ensures plan.UpdateCart? && !plan.expired ==> (input.kind == "cart.set" && plan.customerId == cart.customerId &&
      plan.expiry == input.now + 3600 &&
      plan.nextItems == WithQuantity(if input.now >= cart.expires then map[] else cart.items,
        input.product, input.quantity) &&
      plan.reservations ==
        (if input.now >= cart.expires then Release(cart.items) else map[])
          [input.product := input.quantity - (if input.product in cart.items then cart.items[input.product] else 0)])
    ensures input.kind == "cart.set" && input.productValid &&
      VersionMatches(input.expectedVersion, cart.version) &&
      Domain.CartShapeValid(|cart.items|, cart.expires) && Domain.TimestampValid(input.now) &&
      input.product in cart.items && Domain.CartQuantityValid(cart.items[input.product]) &&
      input.quantity == cart.items[input.product] && input.now < cart.expires
      ==> plan.Unchanged?
  {
    if !VersionMatches(input.expectedVersion, cart.version) then Reject(StaleVersion)
    else if input.kind == "order.checkout" then
      if !Domain.MayCheckout(input.now, cart.expires, |cart.items| > 0) then
        Reject("カートが空、または有効期限切れです")
      else if !ValidRequestedDay(cart) then Reject("希望納品日が不正です")
      else match cart.coupon
        case UnknownCoupon => Reject("クーポンが存在しません")
        case NoCoupon => Checkout(cart.customerId, input.commandId, cart.items, map[], cart.expires, 0)
        case CouponRate(percent) =>
          if Domain.DiscountRateValid(percent) then Checkout(cart.customerId, input.commandId, cart.items, map[], cart.expires, percent)
          else Reject("割引率は0〜100です")
    else if input.kind == "cart.expire" then
      if Domain.MayExpire(input.now, cart.expires, input.expireAt, |cart.items| > 0) then
        UpdateCart(cart.customerId, map[], cart.expires, Release(cart.items), true)
      else Unchanged
    else if input.kind != "cart.set" then Reject("未対応のCommandです")
    else if !input.productValid then Reject("商品IDが不正です")
    else if !Domain.CartQuantityValid(input.quantity) then Reject("カート数量が範囲外です")
    else
      var previous := if input.product in cart.items then cart.items[input.product] else 0;
      if !Domain.CartShapeValid(|cart.items|, cart.expires) || !Domain.TimestampValid(input.now) ||
        !Domain.CartQuantityValid(previous) then Reject("カート不変条件違反")
      else match Domain.PlanCartChange(|cart.items|, previous, input.quantity, input.now, cart.expires)
        case CartFull => Reject("カートは100商品までです")
        case CartUnchanged => Unchanged
        case CartTimeOutOfRange => Reject("カートの有効期限が範囲外です")
        case CartChanged(reset, delta, expiry) =>
          var retained := if reset then map[] else cart.items;
          var releases := if reset then Release(cart.items) else map[];
          UpdateCart(cart.customerId, WithQuantity(retained, input.product, input.quantity), expiry,
            releases[input.product := delta], false)
  }

  function PlanOrder(input: CommandInput, version: int, order: OrderState, lines: map<string, int>):
    (plan: CommandPlan)
    ensures plan.UpdateOrder? ==> Domain.Allowed(order.status, plan.target)
    ensures plan.UpdateOrder? ==> (plan.obligations ==
      (if plan.target == Delivered then {AwardPoints(order.points), CreateInvoice} else {}))
    ensures plan.UpdateOrder? && plan.target == Delivered ==>
      input.role == "carrier" && input.actor == order.carrier
    ensures EligibleDelivery(input, version, order) ==> plan.UpdateOrder? && plan.target == Delivered
    ensures plan.UpdateOrder? && plan.target in {Cancelled, PaymentFailed} ==>
      plan.stock == Reserve(Release(lines))
    ensures plan.UpdateOrder? && input.kind == "order.dispatch" ==> plan.stock == Ship(lines)
  {
    if !VersionMatches(input.expectedVersion, version) then Reject(StaleVersion)
    else match DecideOrder(input, order)
      case NoChange => Unchanged
      case Denied(message) => Reject(message)
      case Move(target) =>
        if !Domain.Allowed(order.status, target) then Reject("許可されない状態遷移です")
        else
          var prepare := if input.kind == "order.pick" then AssignWorker(input.actor)
            else if input.kind == "order.pack" then Pack(input.boxes, input.weight)
            else if input.kind == "order.dispatch" then AssignCarrier(input.actor, input.tracking) else Keep;
          var stock := if target in {Cancelled, PaymentFailed} then Reserve(Release(lines))
            else if input.kind == "order.dispatch" then Ship(lines) else Reserve(map[]);
          var obligations := if target == Delivered then {AwardPoints(order.points), CreateInvoice} else {};
          UpdateOrder(target, order.customer, prepare, stock, obligations)
  }

  predicate InputSizeValid(input: CommandInput) {
    input.noteBytes <= 1000 && input.trackingBytes <= 100 && input.requestedDateBytes <= 10
  }

  predicate EligibleDelivery(input: CommandInput, version: int, order: OrderState) {
    input.kind == "order.deliver" && input.role == "carrier" && input.actor == order.carrier &&
      order.status == Shipping && VersionMatches(input.expectedVersion, version)
  }

  function DecideCommand(input: CommandInput, snapshot: Snapshot): (plan: CommandPlan)
    ensures plan.UpdateOrder? && plan.target == Delivered ==>
      (snapshot.Order? && input.role == "carrier" && input.actor == snapshot.order.carrier &&
        plan.obligations == {AwardPoints(snapshot.order.points), CreateInvoice})
    ensures InputSizeValid(input) && snapshot.Order? && EligibleDelivery(input, snapshot.orderVersion, snapshot.order)
      ==> plan.UpdateOrder? && plan.target == Delivered
    ensures plan.UpdateOrder? && plan.target in {Cancelled, PaymentFailed} ==>
      snapshot.Order? && plan.stock == Reserve(Release(snapshot.lines))
    ensures plan.Checkout? ==> (snapshot.Cart? && input.role == "customer" &&
      plan.checkoutCustomer == input.actor && plan.orderLines == snapshot.state.items && plan.remainingItems == map[])
    ensures plan.UpdateCart? ==> (snapshot.Cart? &&
      (if plan.expired then input.role == "system" && plan.customerId == input.customer
       else input.role == "customer" && plan.customerId == input.actor))
    ensures plan.Adjust? ==> input.role == "operator" && Domain.StockDeltaValid(plan.delta)
  {
    match RequestState(input)
      case RejectRead(reason) => Reject(reason)
      case ReadCart(customer) =>
        if snapshot.Missing? then Reject(snapshot.message)
        else if !snapshot.Cart? then Reject("カートが存在しません")
        else if snapshot.state.customerId != customer then Reject("カート所有者が一致しません")
        else PlanCart(input, snapshot.state)
      case ReadStock(product) =>
        if snapshot.Missing? then Reject(snapshot.message)
        else if !snapshot.Stock? then Reject("商品が存在しません")
        else if !VersionMatches(input.expectedVersion, snapshot.stockVersion) then Reject(StaleVersion)
        else Adjust(product, input.stockDelta)
      case ReadOrder(_) =>
        if snapshot.Missing? then Reject(snapshot.message)
        else if !snapshot.Order? then Reject("受注が存在しません")
        else PlanOrder(input, snapshot.orderVersion, snapshot.order, snapshot.lines)
  }

  function RequestState(input: CommandInput): (request: ReadRequest)
    ensures request.ReadCart? ==>
      (if input.kind == "cart.expire" then input.role == "system" && request.customerId == input.customer
       else input.role == "customer" && request.customerId == input.actor)
    ensures (InputSizeValid(input) && input.kind in {"cart.set", "cart.expire", "order.checkout"} &&
      input.role == (if input.kind == "cart.expire" then "system" else "customer")) ==> request.ReadCart?
    ensures request.ReadStock? ==> input.role == "operator" && Domain.StockDeltaValid(input.stockDelta)
    ensures InputSizeValid(input) && input.kind == "order.deliver" ==> request.ReadOrder?
  {
    if !InputSizeValid(input) then RejectRead("入力が長すぎます")
    else if input.kind in {"cart.set", "cart.expire", "order.checkout"} then
      var role := if input.kind == "cart.expire" then "system" else "customer";
      if input.role != role then RejectRead("この利用者には操作権限がありません")
      else ReadCart(if input.kind == "cart.expire" then input.customer else input.actor)
    else if input.kind == "stock.adjust" then
      if input.role != "operator" then RejectRead("この利用者には操作権限がありません")
      else if !input.productValid || !Domain.StockDeltaValid(input.stockDelta) || input.noteBytes == 0 then
        RejectRead("商品・調整数・理由を入力してください")
      else ReadStock(input.product)
    else if input.kind in {"order.pay", "order.cancel", "order.pick", "order.block", "order.pack",
      "order.dispatch", "order.delivery_failed", "order.redeliver", "order.deliver", "order.return"} then ReadOrder(input.orderId)
    else RejectRead("未対応のCommandです")
  }

  function NeedsCoupon(input: CommandInput, snapshot: Snapshot): bool {
    input.kind == "order.checkout" && input.couponCode != "" && snapshot.Cart? &&
    VersionMatches(input.expectedVersion, snapshot.state.version) &&
    Domain.MayCheckout(input.now, snapshot.state.expires, |snapshot.state.items| > 0) &&
    ValidRequestedDay(snapshot.state)
  }

  predicate ValidRequestedDay(cart: CartState) {
    cart.requestedDay.Unspecified? ||
    (cart.requestedDay.CalendarDay? && cart.requestedDay.date >= cart.today)
  }

  datatype Reference = Absent | Id(value: string)
  datatype Dependencies = Dependencies(products: set<string>, stocks: set<string>,
    loyalty: Reference, newOrder: Reference)

  function DependenciesFor(plan: CommandPlan): (dependencies: Dependencies)
    ensures plan.Checkout? ==> dependencies.products == plan.orderLines.Keys && dependencies.newOrder == Id(plan.newOrderId)
    ensures plan.UpdateCart? ==> dependencies.stocks == plan.reservations.Keys
    ensures !plan.Checkout? ==> dependencies.products == {} && dependencies.newOrder.Absent?
    ensures plan.UpdateOrder? ==> (dependencies.stocks ==
      (if plan.stock.Reserve? then plan.stock.deltas.Keys else plan.stock.quantities.Keys))
    ensures dependencies.loyalty.Id? <==> plan.UpdateOrder? && plan.target == Delivered
    ensures plan.UpdateOrder? && plan.target == Delivered ==> dependencies.loyalty == Id(plan.orderCustomer)
  {
    match plan
      case Checkout(_, orderId, lines, _, _, _) => Dependencies(lines.Keys, {}, Absent, Id(orderId))
      case UpdateCart(_, _, _, delta, _) => Dependencies({}, delta.Keys, Absent, Absent)
      case UpdateOrder(target, customer, _, stock, _) =>
        var ids := match stock
          case Reserve(delta) => delta.Keys
          case Ship(quantities) => quantities.Keys;
        Dependencies({}, ids, if target == Delivered then Id(customer) else Absent, Absent)
      case _ => Dependencies({}, {}, Absent, Absent)
  }

  datatype Status = Ok | Failed | Conflict | Rejected | Skipped

  // Go provides these operations on one attempt's transaction.
  trait Effects {
    ghost var active: bool
    ghost var applied: bool
    ghost var recorded: bool
    ghost var published: bool

    method Begin() returns (status: Status)
      modifies this
      requires !active && !applied && !recorded && !published
      ensures active == status.Ok? && !applied && !recorded && !published

    method LockCommand() returns (status: Status, state: string)
      requires active
    method ReadInput() returns (status: Status, input: CommandInput)
      requires active
    method Read(request: ReadRequest) returns (status: Status, snapshot: Snapshot)
      requires active && !request.RejectRead?
      ensures status.Ok? && request.ReadCart? ==> (snapshot.Missing? ||
        (snapshot.Cart? && snapshot.state.customerId == request.customerId))
    method ReadCoupon(code: string) returns (status: Status, coupon: Coupon)
      requires active && code != ""
    method ReadDependencies(dependencies: Dependencies) returns (status: Status)
      requires active

    method Apply(plan: CommandPlan) returns (status: Status)
      requires active && !recorded && !plan.Reject?

    method RecordSuccess() returns (status: Status)
      modifies this
      requires active && applied && !recorded
      ensures active && recorded == status.Ok? && !published

    method RecordRejection() returns (status: Status)
      modifies this
      requires active && !applied && !recorded
      ensures active && recorded == status.Ok? && !published

    method SetRejection(reason: string)

    method Commit() returns (status: Status)
      modifies this
      requires active && recorded && !published
      ensures !active && !applied && !recorded && published == status.Ok?

    method Rollback()
      modifies this
      requires active && !published
      ensures !active && !applied && !recorded && !published

    method RecordFailure(maxAttempts: int)
      requires !active && !applied && !recorded && !published
  }

  method DecideAndApply(e: Effects) returns (status: Status)
    requires e.active && !e.recorded
  {
    var input;
    status, input := e.ReadInput();
    if !status.Ok? { return; }
    var request := RequestState(input);
    if request.RejectRead? {
      e.SetRejection(request.reason);
      return Rejected;
    }
    var snapshot;
    status, snapshot := e.Read(request);
    if !status.Ok? { return; }
    if NeedsCoupon(input, snapshot) {
      var coupon;
      status, coupon := e.ReadCoupon(input.couponCode);
      if !status.Ok? { return; }
      snapshot := Cart(snapshot.state.(coupon := coupon));
    }
    var plan := DecideCommand(input, snapshot);
    if plan.Reject? {
      e.SetRejection(plan.reason);
      return Rejected;
    }
    var dependencies := DependenciesFor(plan);
    if dependencies != Dependencies({}, {}, Absent, Absent) {
      status := e.ReadDependencies(dependencies);
      if !status.Ok? { return; }
    }
    status := e.Apply(plan);
  }

  method ExecuteAttempt(e: Effects) returns (status: Status)
    modifies e
    requires !e.active && !e.applied && !e.recorded && !e.published
    ensures !e.active && !e.applied && !e.recorded
    ensures e.published ==> status.Ok?
  {
    status := e.Begin();
    if !status.Ok? { return; }
    var state;
    status, state := e.LockCommand();
    if status.Ok? {
      status := if state == "queued" then Ok else Skipped;
      if status.Ok? {
        status := DecideAndApply(e);
        e.applied := status.Ok?;
      }
      if status.Ok? { status := e.RecordSuccess(); }
      else if status.Rejected? { status := e.RecordRejection(); }
      if status.Ok? {
        status := e.Commit();
        return;
      }
    }
    e.Rollback();
  }

  method ExecuteCommand(e: Effects) returns (status: Status)
    modifies e
    requires !e.active && !e.applied && !e.recorded && !e.published
    ensures !e.active && !e.applied && !e.recorded
    ensures e.published ==> status.Ok?
  {
    for attempt := 0 to 5
      invariant !e.active && !e.applied && !e.recorded && !e.published
    {
      status := ExecuteAttempt(e);
      if !status.Conflict? {
        if status.Failed? { e.RecordFailure(8); }
        return;
      }
    }
    status := Conflict;
  }
}
