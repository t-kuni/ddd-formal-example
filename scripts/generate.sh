#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
DAFNY=${DAFNY:-dafny}
mkdir -p .build
"$DAFNY" translate go formal/Domain.dfy formal/CommandModel.dfy formal/CartCommands.dfy formal/OrderCommands.dfy formal/StockCommands.dfy formal/CommandFlow.dfy --include-runtime --go-module-name wholesale.local/orbit/generated --output .build/Domain
# Dafny 4.11 leaves System_ runtime imports unqualified; normalize imports only.
find .build/Domain-go -name '*.go' -exec sed -i 's|"dafny"|"wholesale.local/orbit/generated/dafny"|g' {} +
for name in Domain CommandModel CartCommands OrderCommands StockCommands CommandFlow System_ dafny; do
 mkdir -p "generated/$name"
 cp ".build/Domain-go/$name/"*.go "generated/$name/"
done
