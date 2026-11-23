# Codex向け引継ぎ

## このリポジトリの目的

企業向け卸業者の受注から配送完了を、DDD・型による保護・形式手法で検証するローカル専用アプリ。ユーザーの環境はUbuntu、Docker Compose利用可能。本番対応、実決済、実認証、実AWSへのデプロイは対象外。説明・資料・応答は日本語で簡潔にし、結論を先に書く。ASCII図は使わず、図にはMermaidを用いる。

このファイルは、以前の会話がなくても保守できるようにまとめたもの。実装詳細はコードを確認し、仕様変更時は関連資料も更新する。

## 最初に読む資料

1. `README.md`: 起動、利用者、操作、停止、テスト、制約。
2. `docs/design.md`: 集約、整合性、型、技術選定の理由。
3. `docs/workflow.md`: 正常・異常系の状態遷移。
4. `docs/verification.md`: 実施済みの検証と未検証事項。過去の成功を今回の成功として報告しない。

## 維持する設計上の条件

- Go、PostgreSQL、公式`aws/aws-emulator`を使う。LocalStackに置換しない。
- Commandは永続受付後にSQS経由で非同期実行し、冪等にする。Queryは結果整合の読み取りモデル。
- 業務Commandの非同期化にアプリケーションのgoroutineを使わない。標準ライブラリ・DBドライバー・SDK内部のgoroutineは許容する。
- 集約を取得し、ドメイン操作と不変条件検査を行い、version比較による楽観ロックで保存する。処理効率より整合性を優先。
- 複数集約をまたぐ引当・受注・ポイントの変更は、必要に応じ同一DBトランザクションに含める。
- リソースIDはUUID。SKU、追跡番号、受注番号は別の業務属性。
- ID、金額、数量等は専用型に変換し、集約フィールドを公開しない。DTOを集約へ直接保存しない。Snapshotはコピーとする。
- Goのゼロ値・enumは完全には型で防げない。入力境界・集約操作・DB復元の検査を維持する。
- 形式検証対象のロジックは`formal/Domain.dfy`が原本。`generated/`を手で修正しない。手書きGoで同じ計算を複製しない。

## 業務ルールの要点

- カートは他の顧客が購入できる在庫を実際に減らす引当。成立した追加・削除・数量変更からカート全体を1時間延長。閲覧、同数、拒否、重複では延長しない。
- 期限到達時点で発注不可。解放はSchedulerによる定期処理。古い期限の失効Commandで更新済みカートを消してはいけない。
- 発注でカートから受注に引当を移管。決済失敗・取消で解放。出荷で手持ちと引当を両方減算。返送は自動入庫しない。
- 価格・住所・割引は発注時に固定。JPY整数、割引後税抜に税10%、円未満切捨て。配送完了時に割引後税抜100円につき1ポイント、1回だけ付与。
- 単一倉庫、1受注1配送、複数箱。作業担当・配送担当の権限と所有者をAPI側で検査。
- 請求書レコードとポイントの重複はDBで防ぐ。SMTP受付とDB更新は原子的ではなく、メール重複はあり得る。

## コードの入口

| 修正内容 | 主な場所 |
|---|---|
| 在庫・期限・遷移・金額の形式仕様 | `formal/Domain.dfy` |
| 値オブジェクト・集約 | `internal/domain/values.go`, `types.go`, `aggregate.go` |
| Command受付・冪等性・業務処理 | `internal/app/commands.go` |
| 請求書・SMTP | `internal/app/invoice.go` |
| 永続化・CAS・投影 | `internal/store/store.go`, `schema.sql` |
| 初期マスタ | `internal/store/seed.go` |
| AWS SDK・SQS・S3・Scheduler初期化 | `internal/cloud/cloud.go` |
| API・投影完了の判定 | `internal/web/api.go` |
| UI | `internal/web/static/` |
| Lambda・ローカルアダプター | `cmd/app/main.go` |
| 実行構成 | `compose.yaml`, `deploy/` |

集約スナップショットが正本であり、Event Sourcingではない。Commandの完了、集約更新、projection_eventsの追加は同じトランザクション。SQS送信に失敗してもDBの未処理Commandから再送する。投影はversionを比較し、古いイベントで上書きしない。

画面はCommandの`succeeded`だけでなく`projected`も待つ。ここを省くと直後の操作が古いexpected versionで拒否される（初期検証で修正した不具合）。

## ローカルAWSの注意

AWS Emulator 1.0.580ではGoの`provided.al2023`呼出しが失敗したため、Go LambdaだけAWS公式RIEで実行。API GatewayとSQS Event Source Mapping相当の薄いローカルアダプターを置く。業務ロジックをこれらに移さない。

RIEの同時Invokeで異常終了を観測したため、gatewayとbridgeが共有ファイルロックでInvokeを直列化している。これはローカル実行上の制約であり、集約の楽観ロックを代替しない。共有`invoke-lock`ボリュームを外さない。

ComposeではAWSコンテナのネットワーク名前空間を共有し、ローカルHTTPSのCAと生成資格情報を共有する。TLS検査を無効にして起動問題を回避しない。実AWSの資格情報は不要。

エミュレーター再起動でS3/SQSは消える。DBに保持した画像とCommandから再初期化・再送する。`restart aws`単体ではなくREADMEの全体down/upを使う。LTS/latestは変動するため、不具合時は実際の版を記録する。

## 修正と検証

まず不具合を再現し、関係する最小の回帰テストを加えて修正する。既存のテストや形式条件を弱めて通さない。

```bash
# 通常起動とログ
docker compose up --build -d
docker compose ps
docker compose logs --tail=150 init aws lambda gateway event-source-mapping

# 形式検証・生成一致・変異検出・型誤用拒否・Goテスト・vet
docker compose --profile test run --build --rm verify

# 実PostgreSQLの結合テスト。通常アプリとは別DB
docker compose --profile test run --build --rm integration

# 起動済み構成へのAPI E2E。サンプル受注とメールを作成する
python3 tests/e2e.py
```

仕様変更ではDafnyを変更後、Goを再生成して生成物もコミットする。ホストの準備はREADME参照。Dockerだけで再生成する場合:

```bash
docker compose --profile test run --build --rm -v "$PWD:/src" verify ./scripts/generate.sh
# 再生成後にverifyを再ビルド・実行して生成物の一致を検証
```

バインドマウントで生成したファイルはroot所有になる場合がある。必要なら対象の生成物だけ所有権を戻す。

`go test ./...`だけでは、`TEST_DATABASE_URL`未設定時に結合テストはskipされる。verify成功を結合テスト成功とみなさない。verifyの型誤用テストは「コンパイル失敗が成功」という負のテスト。

画面や業務フローを変更した場合は、利用可能なブラウザ操作機能で顧客→倉庫→運送→顧客を確認する。税込額・ポイント・履歴・PDFとMailpitも確認。不可能なら未実施と明記する。

## 引継ぎ時の未確認事項・復旧

- 2026-11-23時点ではDockerデーモンがない環境で作成したため、Composeのイメージビルドと起動そのものは未検証。各実サービスを個別プロセスで起動したE2E、認証有効なエミュレーター接続、ブラウザ検証は成功。手元では最初にCompose起動を確認する。
- arm64の実行は未検証。verifyはamd64版Dafnyを使用する。
- スキーマは起動時SQLで初期化する簡易構成。既存DBを更新する変更では明示的な移行SQLを用意する。CREATE TABLE IF NOT EXISTSだけで既存列は変更されない。
- JSONの構造や状態コードを変更する場合、既存集約と投影の復元互換性を考慮する。
- 起動不具合を理由にデータを勝手に削除しない。`docker compose down -v`は破壊的リセット。
- 初期開発環境固有の`.tools`やネイティブ起動スクリプトは配布物に含まない。ここに記載したComposeとリポジトリ内のスクリプトを使う。
- 形式証明は業務コアのみ。DB・ネットワーク・手書きGo・UIまで数学的に証明済みとは説明しない。

修正後の報告には変更点、実行した検証、未検証事項を含める。設計や起動方法が変わればREADME・docs・このファイルを更新する。
