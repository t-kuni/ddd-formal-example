module StockCommands {
  import opened CommandModel

  function AdjustStock(input: CommandInput, product: string, version: int): (plan: CommandPlan)
    ensures plan.Reject? || plan.Adjust?
    ensures plan.Adjust? ==> plan.product == product && plan.delta == input.stockDelta
  {
    if !VersionMatches(input.expectedVersion, version) then Reject(StaleVersion)
    else Adjust(product, input.stockDelta)
  }
}
