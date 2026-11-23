#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
mkdir -p .build/typecheck
cat > .build/typecheck/misuse.go <<'GO'
package misuse
import "wholesale.local/orbit/internal/domain"
func invalid(stock *domain.Stock,product domain.ProductID,order domain.OrderID,money domain.Money) {
 order=product // ProductID is not OrderID.
 stock.Reserve(money) // Money is not ReservationDelta.
 stock.onHand=0 // Aggregate state cannot be directly written.
 _=order
}
GO
if go test ./.build/typecheck > .build/typecheck-result.txt 2>&1; then
 echo 'ERROR: invalid typed operations compiled' >&2; exit 1
fi
grep -q 'cannot use product' .build/typecheck-result.txt
grep -q 'cannot use money' .build/typecheck-result.txt
grep -q 'unexported field onHand' .build/typecheck-result.txt
printf 'Typed IDs, unit separation, aggregate encapsulation: PASS\n'
