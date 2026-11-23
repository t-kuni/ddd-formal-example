module CommandFlow {
  import Domain

  import opened CommandModel
  import CartCommands
  import OrderCommands
  import StockCommands

  function DecideCommand(input: CommandInput, snapshot: Snapshot): (plan: CommandPlan)
    ensures plan.UpdateOrder? && plan.target == Delivered ==>
      (snapshot.Order? && input.role == "carrier" && input.actor == snapshot.order.carrier &&
        plan.obligations == {AwardPoints(snapshot.order.points), CreateInvoice})
    ensures InputSizeValid(input) && snapshot.Order? && OrderCommands.EligibleDelivery(input, snapshot.orderVersion, snapshot.order)
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
        else if input.kind == "cart.set" then CartCommands.SetCart(input, snapshot.state)
        else if input.kind == "cart.expire" then CartCommands.ExpireCart(input, snapshot.state)
        else CartCommands.CheckoutOrder(input, snapshot.state)
      case ReadStock(product) =>
        if snapshot.Missing? then Reject(snapshot.message)
        else if !snapshot.Stock? then Reject("商品が存在しません")
        else StockCommands.AdjustStock(input, product, snapshot.stockVersion)
      case ReadOrder(_) =>
        if snapshot.Missing? then Reject(snapshot.message)
        else if !snapshot.Order? then Reject("受注が存在しません")
        else OrderCommands.PlanOrder(input, snapshot.orderVersion, snapshot.order, snapshot.lines)
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
    if CartCommands.NeedsCoupon(input, snapshot) {
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
