#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
DAFNY=${DAFNY:-dafny}
mkdir -p .build
"$DAFNY" verify formal/Domain.dfy formal/CommandFlow.dfy --verification-time-limit 30
"$DAFNY" translate go formal/Domain.dfy formal/CommandFlow.dfy --include-runtime --go-module-name wholesale.local/orbit/generated --output .build/Domain
# Dafny 4.11 leaves System_ runtime imports unqualified; normalize imports only.
find .build/Domain-go -name '*.go' -exec sed -i 's|"dafny"|"wholesale.local/orbit/generated/dafny"|g' {} +
for name in Domain CommandFlow System_ dafny; do diff -r "generated/$name" ".build/Domain-go/$name"; done
# A deliberately broken invariant must be detected; this prevents a no-op verifier.
sed 's/next := if ok then r + delta else r;/next := if ok then r + delta + 1 else r;/' formal/Domain.dfy > .build/Mutant.dfy
if "$DAFNY" verify .build/Mutant.dfy > .build/mutation-result.txt 2>&1; then
 echo 'ERROR: verifier accepted the broken reservation' >&2; exit 1
fi
grep -q 'could not be proved' .build/mutation-result.txt
# A delivery plan without its invoice must violate the postcondition.
sed '/var obligations :=/s/, CreateInvoice//' formal/CommandFlow.dfy > .build/CommandFlowMutant.dfy
if "$DAFNY" verify formal/Domain.dfy .build/CommandFlowMutant.dfy > .build/command-mutation-result.txt 2>&1; then
 echo 'ERROR: verifier accepted delivery without an invoice' >&2; exit 1
fi
grep -q 'postcondition could not be proved' .build/command-mutation-result.txt
printf 'Formal verification, generated-code consistency, mutation check: PASS\n'
(
 cd api
 go tool oapi-codegen -config oapi-codegen.yaml -o ../.build/openapi.gen.go openapi.yaml
)
diff -u internal/web/contract/api.gen.go .build/openapi.gen.go
printf 'OpenAPI generated-code consistency: PASS\n'
./scripts/typecheck.sh
go test ./...
go vet ./internal/... ./cmd/...
