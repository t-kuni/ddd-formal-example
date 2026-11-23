# 現行構成の検証

検証環境はUbuntu amd64、Docker Compose、PostgreSQL 18.3、AWS Emulator 1.0.580、AWS Lambda RIE、Mailpit 1.34.0。

| 検証 | 結果・範囲 |
|---|---|
| Compose `verify` | Dafny 27件、Dafny・OpenAPI生成物との差分なし、在庫とCommand完了順序の変異検出、型誤用のコンパイル拒否、Goテスト、vet |
| Compose `integration` | PostgreSQL結合テスト6件。カート失効と古い期限の拒否、複数商品の引当変更の原子性、楽観ロック、Command重複、障害時の試行回数とdead遷移、配送完了・ポイント・PDF |
| Compose起動 | AWS Emulator・Lambda・API・SQS workerが稼働。PostgreSQL結合テストは空の専用DBから初期化 |
| API E2E | S3画像、OpenAPI入力検証、SQS処理、在庫引当、倉庫・配送の例外系、Command完了直後のQuery、顧客間の所有権、請求書PDF、SMTP・Mailpit |
| 受注一覧 | 実DBに同一発注日時の検証用行を用意し、25件のページング、UUIDによる順序、ステータス絞り込み、集計、顧客をまたいだカーソルの扱いを確認 |
| カートの連続追加 | Chromeで前の処理結果が返る前に異なる3商品と同一商品を続けて押し、4回すべてが数量に反映されることを確認 |
| Chrome | 顧客→倉庫→運送→顧客の操作で、税込2,376円、21ポイントの増分、追跡番号・状態履歴、請求書PDF、Mailpitの添付を確認 |

arm64は未検証。業務コアの不変条件と、Goの操作契約に基づくCommandフローの権限・完了順序を形式検証する。Goの永続化処理、DB、外部サービス、ブラウザの動作は結合テストとE2Eで確認する。
