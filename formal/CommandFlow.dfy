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

  datatype OrderContext = OrderContext(role: string, actor: string, customer: string,
    worker: string, carrier: string, status: int, ageNanos: int,
    paymentSucceeded: bool, hasNote: bool, hasTracking: bool, boxes: int, weight: int)
  datatype OrderDecision = Move(target: int) | NoChange | Denied(message: string)

  function DecideOrder(kind: string, x: OrderContext): (decision: OrderDecision)
    ensures decision.Move? && x.role == "customer" ==> x.actor == x.customer
    ensures decision.Move? && kind == "order.deliver" ==>
      x.role == "carrier" && x.actor == x.carrier && decision.target == Delivered
    ensures decision.Move? && kind == "order.cancel" && x.role == "system" ==>
      x.status == PaymentPending && x.ageNanos >= PaymentTimeoutNanos
  {
    if x.role == "customer" && x.actor != x.customer then
      Denied("他の顧客の受注は操作できません")
    else if kind == "order.pay" then
      if x.role != "customer" then Denied("この利用者には操作権限がありません")
      else Move(if x.paymentSucceeded then Accepted else PaymentFailed)
    else if kind == "order.cancel" then
      if x.role !in {"customer", "operator", "system"} then Denied("この利用者には操作権限がありません")
      else if x.role == "system" && (x.status != PaymentPending || x.ageNanos < PaymentTimeoutNanos) then NoChange
      else if x.role == "customer" && x.status !in {PaymentPending, Accepted} then
        Denied("作業開始後のキャンセルは運用者に依頼してください")
      else Move(Cancelled)
    else if kind in {"order.pick", "order.block", "order.pack"} then
      if x.role != "warehouse" then Denied("この利用者には操作権限がありません")
      else if kind == "order.pick" then
        if !Domain.MayAssignWorker(x.worker != "", x.worker == x.actor) then Denied("別の担当者が作業中です")
        else Move(Picking)
      else if x.worker != x.actor then Denied("担当外の作業です")
      else if kind == "order.block" then
        if !x.hasNote then Denied("保留理由が必要です") else Move(PickBlocked)
      else if !Domain.BoxCountValid(x.boxes) || !Domain.WeightValid(x.weight) then
        Denied("箱数は1〜100、1箱の重量は1〜500000gです")
      else Move(Packed)
    else if kind in {"order.dispatch", "order.delivery_failed", "order.redeliver", "order.deliver", "order.return"} then
      if x.role != "carrier" then Denied("この利用者には操作権限がありません")
      else if kind == "order.dispatch" then
        if !x.hasTracking then Denied("追跡番号を入力してください")
        else if x.status != Packed then Denied("梱包完了後に出荷してください")
        else Move(Shipping)
      else if x.carrier != x.actor then Denied("担当外の配送です")
      else if kind == "order.delivery_failed" then
        if !x.hasNote then Denied("配送失敗理由が必要です") else Move(DeliveryFailed)
      else if kind == "order.redeliver" then
        if x.status != DeliveryFailed then Denied("再配送できる状態ではありません") else Move(Shipping)
      else if kind == "order.deliver" then Move(Delivered)
      else if !x.hasNote then Denied("返送理由が必要です") else Move(Returned)
    else Denied("未対応のCommandです")
  }
  datatype Step = Ok | Failed | Conflict | Rejected | Skip

  datatype CommandInput = CommandInput(kind: string, role: string, actor: string, customer: string,
    noteBytes: int, trackingBytes: int, requestedDateBytes: int, hasCoupon: bool,
    expectedVersion: int,
    productValid: bool, stockDelta: int)
  datatype StockLine = StockLine(product: string, quantity: int)
  datatype StockAction = Reserve | Release | Ship

  // The Go implementation owns the connection and carries the original error.
  // These ghost fields model one attempt and an acknowledged commit. The business
  // ports below preserve these fields. Their actual DB writes belong to the
  // transaction opened by Begin, a boundary checked in PostgreSQL tests.
  trait Effects {
    ghost var active: bool
    ghost var workReady: bool
    ghost var staged: bool
    ghost var published: bool

    method Begin() returns (r: Step)
      modifies this
      requires !active && !published
      ensures active == r.Ok? && !workReady && !staged && !published

    method LoadCommand() returns (r: Step, status: string)
      modifies this
      requires active && !workReady && !staged && !published
      ensures active && !workReady && !staged && !published

    method DecodePayload() returns (r: Step)
    method LoadActor() returns (r: Step)
    method ReadClock() returns (r: Step)
    method Input() returns (input: CommandInput)
    method LoadCart(customer: string) returns (r: Step, version: int)
    method ExpireCart() returns (r: Step, lines: seq<StockLine>)
    method SetCart() returns (r: Step, lines: seq<StockLine>)
    method SaveCart() returns (r: Step)
    method CheckoutCart() returns (r: Step, lines: seq<StockLine>)
    method ReadShippingAddress() returns (r: Step)
    method RequestedDate() returns (provided: bool, parsed: bool, day: int, today: int)
    method LoadCoupon() returns (r: Step, percent: int)
    method LoadProduct(product: string) returns (r: Step)
    method AddOrderLine(line: StockLine) returns (r: Step)
    method BuildOrder(percent: int) returns (r: Step)
    method LoadCheckoutVersion() returns (r: Step)
    method LoadAdjustment() returns (r: Step, version: int)
    method AdjustStock() returns (r: Step)
    method SaveAdjustment() returns (r: Step)
    method LoadOrder() returns (r: Step, x: OrderContext, version: int)
    method Noop()
    method Reject(message: string) returns (r: Step)
    method AssignWorker() returns (r: Step)
    method PackOrder() returns (r: Step)
    method AssignCarrier() returns (r: Step)
    method MoveOrder(target: int) returns (r: Step, lines: seq<StockLine>)
    method SaveOrder() returns (r: Step)
    method LoadStock(product: string) returns (r: Step)
    method ReserveStock(delta: int) returns (r: Step)
    method ShipStock(quantity: int) returns (r: Step)
    method SaveStock(product: string) returns (r: Step)
    method InsertPointAward() returns (r: Step)
    method LoadLoyalty() returns (r: Step)
    method CreditLoyalty() returns (r: Step)
    method SaveLoyalty() returns (r: Step)
    method RenderInvoice() returns (r: Step)
    method InsertInvoice() returns (r: Step)

    method SaveCompletion() returns (r: Step)
      modifies this
      requires active && workReady && !staged && !published
      ensures active && workReady && staged == r.Ok? && !published

    method Commit() returns (r: Step)
      modifies this
      requires active && workReady && staged && !published
      ensures !active && !workReady && !staged && published == r.Ok?

    method Rollback() returns (r: Step)
      modifies this
      requires active && !published
      ensures !active && !workReady && !staged && !published

    method SaveRejection() returns (r: Step)
      modifies this
      requires !active && !staged && !published
      ensures !active && !workReady && !staged && !published

    method RecordFailure(maxAttempts: int)
      modifies this
      requires !active && !staged && !published
      ensures !active && !workReady && !staged && !published
  }

  method ChangeStock(e: Effects, lines: seq<StockLine>, action: StockAction)
    returns (r: Step)
  {
    for i := 0 to |lines|
      invariant 0 <= i <= |lines|
    {
      var line := lines[i];
      r := e.LoadStock(line.product);
      if !r.Ok? { return; }
      if action.Ship? {
        r := e.ShipStock(line.quantity);
      } else {
        var delta := if action.Release? then -line.quantity else line.quantity;
        r := e.ReserveStock(delta);
      }
      if !r.Ok? { return; }
      r := e.SaveStock(line.product);
      if !r.Ok? { return; }
    }
    r := Ok;
  }

  method Checkout(e: Effects, hasCoupon: bool) returns (r: Step) {
    var lines;
    r, lines := e.CheckoutCart();
    if !r.Ok? { return; }
    r := e.ReadShippingAddress();
    if !r.Ok? { return; }
    var provided, parsed, day, today := e.RequestedDate();
    if provided && (!parsed || day < today) {
      r := e.Reject("希望納品日が不正です");
      return;
    }
    var percent := 0;
    if hasCoupon {
      r, percent := e.LoadCoupon();
      if !r.Ok? { return; }
    }
    if !Domain.DiscountRateValid(percent) {
      r := e.Reject("割引率は0〜100です");
      return;
    }
    for i := 0 to |lines| {
      r := e.LoadProduct(lines[i].product);
      if !r.Ok? { return; }
      r := e.AddOrderLine(lines[i]);
      if !r.Ok? { return; }
    }
    r := e.BuildOrder(percent);
    if !r.Ok? { return; }
    r := e.LoadCheckoutVersion();
    if !r.Ok? { return; }
    r := e.SaveOrder();
    if r.Ok? { r := e.SaveCart(); }
  }

  predicate VersionMatches(expected: int, actual: int) {
    expected <= 0 || expected == actual
  }

  method Apply(e: Effects) returns (r: Step) {
    r := e.LoadActor();
    if !r.Ok? { return; }
    r := e.ReadClock();
    if !r.Ok? { return; }
    var input := e.Input();
    if input.noteBytes > 1000 || input.trackingBytes > 100 || input.requestedDateBytes > 10 {
      r := e.Reject("入力が長すぎます");
      return;
    }
    var kind := input.kind;
    if kind in {"cart.set", "cart.expire", "order.checkout"} {
      var requiredRole := if kind == "cart.expire" then "system" else "customer";
      if input.role != requiredRole {
        r := e.Reject("この利用者には操作権限がありません");
        return;
      }
      var customer := if kind == "cart.expire" then input.customer else input.actor;
      var version;
      r, version := e.LoadCart(customer);
      if !r.Ok? { return; }
      if !VersionMatches(input.expectedVersion, version) {
        r := e.Reject(StaleVersion);
        return;
      }
      var lines;
      if kind == "order.checkout" {
        r := Checkout(e, input.hasCoupon);
      } else {
        if kind == "cart.expire" { r, lines := e.ExpireCart(); }
        else { r, lines := e.SetCart(); }
        if r.Skip? { e.Noop(); r := Ok; return; }
        if !r.Ok? { return; }
        r := ChangeStock(e, lines, Reserve);
        if r.Ok? { r := e.SaveCart(); }
      }
    } else if kind == "stock.adjust" {
      if input.role != "operator" {
        r := e.Reject("この利用者には操作権限がありません");
        return;
      }
      if !input.productValid || !Domain.StockDeltaValid(input.stockDelta) || input.noteBytes == 0 {
        r := e.Reject("商品・調整数・理由を入力してください");
        return;
      }
      var version;
      r, version := e.LoadAdjustment();
      if !r.Ok? { return; }
      if !VersionMatches(input.expectedVersion, version) {
        r := e.Reject(StaleVersion);
        return;
      }
      r := e.AdjustStock();
      if r.Ok? { r := e.SaveAdjustment(); }
    } else {
      var x, version;
      r, x, version := e.LoadOrder();
      if !r.Ok? { return; }
      if !VersionMatches(input.expectedVersion, version) {
        r := e.Reject(StaleVersion);
        return;
      }
      var decision := DecideOrder(kind, x);
      if decision.NoChange? { e.Noop(); return; }
      if decision.Denied? { r := e.Reject(decision.message); return; }
      var target := decision.target;
      if kind == "order.pick" { r := e.AssignWorker(); }
      else if kind == "order.pack" { r := e.PackOrder(); }
      else if kind == "order.dispatch" { r := e.AssignCarrier(); }
      if !r.Ok? { return; }
      var lines;
      r, lines := e.MoveOrder(target);
      if !r.Ok? { return; }
      if target == Cancelled || target == PaymentFailed {
        r := ChangeStock(e, lines, Release);
      } else if kind == "order.dispatch" {
        r := ChangeStock(e, lines, Ship);
      }
      if !r.Ok? { return; }
      r := e.SaveOrder();
      if !r.Ok? || target != Delivered { return; }
      r := e.InsertPointAward();
      if !r.Ok? { return; }
      r := e.LoadLoyalty();
      if !r.Ok? { return; }
      r := e.CreditLoyalty();
      if !r.Ok? { return; }
      r := e.SaveLoyalty();
      if !r.Ok? { return; }
      r := e.RenderInvoice();
      if r.Ok? { r := e.InsertInvoice(); }
    }
  }

  method Attempt(e: Effects) returns (r: Step)
    modifies e
    requires !e.active && !e.workReady && !e.staged && !e.published
    ensures !e.active && !e.workReady && !e.staged
    ensures e.published <==> r.Ok?
  {
    r := e.Begin();
    if !r.Ok? { return; }
    var status;
    r, status := e.LoadCommand();
    if r.Ok? {
      if status == "queued" { r := e.DecodePayload(); }
      else { r := Skip; }
    }
    if r.Ok? {
      r := Apply(e);
      e.workReady := r.Ok?;
    }
    if r.Ok? {
      r := e.SaveCompletion();
      if r.Ok? {
        r := e.Commit();
        return;
      }
    }
    var rollback := e.Rollback();
    if r.Rejected? {
      if rollback.Ok? {
        r := e.SaveRejection();
        if r.Ok? { r := Skip; }
      } else {
        r := rollback;
      }
    }
  }

  method ExecuteCommand(e: Effects) returns (r: Step)
    modifies e
    requires !e.active && !e.workReady && !e.staged && !e.published
    ensures !e.active && !e.workReady && !e.staged
    ensures e.published <==> r.Ok?
  {
    var attempt := 0;
    while attempt < 5
      invariant !e.active && !e.workReady && !e.staged && !e.published
      invariant 0 <= attempt <= 5
    {
      r := Attempt(e);
      if !r.Conflict? {
        if r.Failed? { e.RecordFailure(8); }
        return;
      }
      attempt := attempt + 1;
    }
    r := Conflict;
  }
}
