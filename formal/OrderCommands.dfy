module OrderCommands {
  import Domain
  import opened CommandModel

  predicate EligibleDelivery(input: CommandInput, version: int, order: OrderState) {
    input.kind == "order.deliver" && input.role == "carrier" && input.actor == order.carrier &&
      order.status == Shipping && VersionMatches(input.expectedVersion, version)
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
    if input.kind == "order.pay" then PayOrder(input, version, order, lines)
    else if input.kind == "order.cancel" then CancelOrder(input, version, order, lines)
    else if input.kind == "order.pick" then PickOrder(input, version, order, lines)
    else if input.kind == "order.block" then BlockOrder(input, version, order, lines)
    else if input.kind == "order.pack" then PackOrder(input, version, order, lines)
    else if input.kind == "order.dispatch" then DispatchOrder(input, version, order, lines)
    else if input.kind == "order.delivery_failed" then FailDelivery(input, version, order, lines)
    else if input.kind == "order.redeliver" then RedeliverOrder(input, version, order, lines)
    else if input.kind == "order.deliver" then DeliverOrder(input, version, order, lines)
    else if input.kind == "order.return" then ReturnOrder(input, version, order, lines)
    else Reject("未対応のCommandです")
  }

  function PayOrder(input: CommandInput, version: int, order: OrderState, lines: map<string, int>): CommandPlan {
    OrderChange(input, version, order, lines,
      ForRole(input, "customer", Move(if input.paymentSucceeded then Accepted else PaymentFailed)), Keep, false)
  }

  function CancelOrder(input: CommandInput, version: int, order: OrderState, lines: map<string, int>):
    (plan: CommandPlan)
    ensures plan.UpdateOrder? && input.role == "system" ==>
      order.status == PaymentPending && order.ageNanos >= PaymentTimeoutNanos
  {
    var decision :=
      if input.role !in {"customer", "operator", "system"} then Denied(RoleDenied)
      else if input.role == "system" && (order.status != PaymentPending || order.ageNanos < PaymentTimeoutNanos) then NoChange
      else if input.role == "customer" && order.status !in {PaymentPending, Accepted} then
        Denied("作業開始後のキャンセルは運用者に依頼してください")
      else Move(Cancelled);
    OrderChange(input, version, order, lines, decision, Keep, false)
  }

  function PickOrder(input: CommandInput, version: int, order: OrderState, lines: map<string, int>): CommandPlan {
    var decision := if Domain.MayAssignWorker(order.worker != "", order.worker == input.actor) then Move(Picking)
      else Denied("別の担当者が作業中です");
    OrderChange(input, version, order, lines, ForRole(input, "warehouse", decision), AssignWorker(input.actor), false)
  }

  function BlockOrder(input: CommandInput, version: int, order: OrderState, lines: map<string, int>): CommandPlan {
    var decision := if input.noteBytes == 0 then Denied("保留理由が必要です") else Move(PickBlocked);
    OrderChange(input, version, order, lines, ForWorker(input, order, decision), Keep, false)
  }

  function PackOrder(input: CommandInput, version: int, order: OrderState, lines: map<string, int>): CommandPlan {
    var decision := if Domain.BoxCountValid(input.boxes) && Domain.WeightValid(input.weight) then Move(Packed)
      else Denied("箱数は1〜100、1箱の重量は1〜500000gです");
    OrderChange(input, version, order, lines, ForWorker(input, order, decision), Pack(input.boxes, input.weight), false)
  }

  function DispatchOrder(input: CommandInput, version: int, order: OrderState, lines: map<string, int>): CommandPlan {
    var decision := if input.tracking == "" then Denied("追跡番号を入力してください")
      else if order.status != Packed then Denied("梱包完了後に出荷してください")
      else Move(Shipping);
    OrderChange(input, version, order, lines,
      ForRole(input, "carrier", decision), AssignCarrier(input.actor, input.tracking), true)
  }

  function FailDelivery(input: CommandInput, version: int, order: OrderState, lines: map<string, int>): CommandPlan {
    var decision := if input.noteBytes == 0 then Denied("配送失敗理由が必要です") else Move(DeliveryFailed);
    OrderChange(input, version, order, lines, ForCarrier(input, order, decision), Keep, false)
  }

  function RedeliverOrder(input: CommandInput, version: int, order: OrderState, lines: map<string, int>): CommandPlan {
    var decision := if order.status != DeliveryFailed then Denied("再配送できる状態ではありません") else Move(Shipping);
    OrderChange(input, version, order, lines, ForCarrier(input, order, decision), Keep, false)
  }

  function DeliverOrder(input: CommandInput, version: int, order: OrderState, lines: map<string, int>):
    (plan: CommandPlan)
    ensures plan.UpdateOrder? ==> (input.role == "carrier" && input.actor == order.carrier &&
      plan.target == Delivered && plan.obligations == {AwardPoints(order.points), CreateInvoice})
    ensures (input.role == "carrier" && input.actor == order.carrier &&
      order.status == Shipping && VersionMatches(input.expectedVersion, version)) ==> plan.UpdateOrder?
  {
    OrderChange(input, version, order, lines, ForCarrier(input, order, Move(Delivered)), Keep, false)
  }

  function ReturnOrder(input: CommandInput, version: int, order: OrderState, lines: map<string, int>): CommandPlan {
    var decision := if input.noteBytes == 0 then Denied("返送理由が必要です") else Move(Returned);
    OrderChange(input, version, order, lines, ForCarrier(input, order, decision), Keep, false)
  }

  const RoleDenied := "この利用者には操作権限がありません"

  function ForRole(input: CommandInput, role: string, decision: OrderDecision): OrderDecision {
    if input.role == role then decision else Denied(RoleDenied)
  }

  function ForWorker(input: CommandInput, order: OrderState, decision: OrderDecision): OrderDecision {
    ForRole(input, "warehouse", if order.worker == input.actor then decision else Denied("担当外の作業です"))
  }

  function ForCarrier(input: CommandInput, order: OrderState, decision: OrderDecision): OrderDecision {
    ForRole(input, "carrier", if order.carrier == input.actor then decision else Denied("担当外の配送です"))
  }

  function OrderChange(input: CommandInput, version: int, order: OrderState, lines: map<string, int>,
    decision: OrderDecision, prepare: OrderPreparation, ship: bool): (plan: CommandPlan)
    ensures plan.Reject? || plan.Unchanged? || plan.UpdateOrder?
    ensures plan.UpdateOrder? ==> decision.Move? && plan.target == decision.target && Domain.Allowed(order.status, plan.target)
    ensures plan.UpdateOrder? ==> (plan.obligations ==
      (if plan.target == Delivered then {AwardPoints(order.points), CreateInvoice} else {}))
    ensures plan.UpdateOrder? && plan.target in {Cancelled, PaymentFailed} ==>
      plan.stock == Reserve(Release(lines))
    ensures plan.UpdateOrder? && ship && plan.target !in {Cancelled, PaymentFailed} ==> plan.stock == Ship(lines)
    ensures plan.UpdateOrder? && input.role == "customer" ==> input.actor == order.customer
    ensures (VersionMatches(input.expectedVersion, version) && input.role != "customer" &&
      decision.Move? && Domain.Allowed(order.status, decision.target)) ==> plan.UpdateOrder?
  {
    if !VersionMatches(input.expectedVersion, version) then Reject(StaleVersion)
    else if input.role == "customer" && input.actor != order.customer then Reject("他の顧客の受注は操作できません")
    else match decision
      case NoChange => Unchanged
      case Denied(message) => Reject(message)
      case Move(target) =>
        if !Domain.Allowed(order.status, target) then Reject("許可されない状態遷移です")
        else
          var stock := if target in {Cancelled, PaymentFailed} then Reserve(Release(lines))
            else if ship then Ship(lines) else Reserve(map[]);
          var obligations := if target == Delivered then {AwardPoints(order.points), CreateInvoice} else {};
          UpdateOrder(target, order.customer, prepare, stock, obligations)
  }
}
