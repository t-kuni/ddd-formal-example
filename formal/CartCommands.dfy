module CartCommands {
  import Domain
  import opened CommandModel

  function SetCart(input: CommandInput, cart: CartState): (plan: CommandPlan)
    ensures plan.Reject? || plan.Unchanged? || plan.UpdateCart?
    ensures plan.UpdateCart? ==> (!plan.expired && plan.customerId == cart.customerId &&
      plan.expiry == input.now + 3600 &&
      plan.nextItems == WithQuantity(if input.now >= cart.expires then map[] else cart.items,
        input.product, input.quantity) &&
      plan.reservations ==
        (if input.now >= cart.expires then Release(cart.items) else map[])
          [input.product := input.quantity - (if input.product in cart.items then cart.items[input.product] else 0)])
    ensures input.productValid &&
      VersionMatches(input.expectedVersion, cart.version) &&
      Domain.CartShapeValid(|cart.items|, cart.expires) && Domain.TimestampValid(input.now) &&
      input.product in cart.items && Domain.CartQuantityValid(cart.items[input.product]) &&
      input.quantity == cart.items[input.product] && input.now < cart.expires
      ==> plan.Unchanged?
  {
    if !VersionMatches(input.expectedVersion, cart.version) then Reject(StaleVersion)
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

  function ExpireCart(input: CommandInput, cart: CartState): (plan: CommandPlan)
    ensures plan.Reject? || plan.Unchanged? || plan.UpdateCart?
    ensures plan.UpdateCart? ==> (plan.expired && plan.customerId == cart.customerId &&
      plan.nextItems == map[] && plan.expiry == cart.expires && plan.reservations == Release(cart.items) &&
      Domain.MayExpire(input.now, cart.expires, input.expireAt, |cart.items| > 0))
  {
    if !VersionMatches(input.expectedVersion, cart.version) then Reject(StaleVersion)
    else if Domain.MayExpire(input.now, cart.expires, input.expireAt, |cart.items| > 0) then
      UpdateCart(cart.customerId, map[], cart.expires, Release(cart.items), true)
    else Unchanged
  }

  function CheckoutOrder(input: CommandInput, cart: CartState): (plan: CommandPlan)
    ensures plan.Reject? || plan.Checkout?
    ensures plan.Checkout? ==> (plan.checkoutCustomer == cart.customerId &&
      plan.orderLines == cart.items && plan.remainingItems == map[] && plan.cartExpiry == cart.expires &&
      Domain.MayCheckout(input.now, cart.expires, |cart.items| > 0))
  {
    if !VersionMatches(input.expectedVersion, cart.version) then Reject(StaleVersion)
    else if !Domain.MayCheckout(input.now, cart.expires, |cart.items| > 0) then
      Reject("カートが空、または有効期限切れです")
    else if !ValidRequestedDay(cart) then Reject("希望納品日が不正です")
    else match cart.coupon
      case UnknownCoupon => Reject("クーポンが存在しません")
      case NoCoupon => Checkout(cart.customerId, input.commandId, cart.items, map[], cart.expires, 0)
      case CouponRate(percent) =>
        if Domain.DiscountRateValid(percent) then Checkout(cart.customerId, input.commandId, cart.items, map[], cart.expires, percent)
        else Reject("割引率は0〜100です")
  }

  predicate ValidRequestedDay(cart: CartState) {
    cart.requestedDay.Unspecified? ||
    (cart.requestedDay.CalendarDay? && cart.requestedDay.date >= cart.today)
  }

  function NeedsCoupon(input: CommandInput, snapshot: Snapshot): bool {
    input.kind == "order.checkout" && input.couponCode != "" && snapshot.Cart? &&
    VersionMatches(input.expectedVersion, snapshot.state.version) &&
    Domain.MayCheckout(input.now, snapshot.state.expires, |snapshot.state.items| > 0) &&
    ValidRequestedDay(snapshot.state)
  }
}
