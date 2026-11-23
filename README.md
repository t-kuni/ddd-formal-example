# ORBIT Wholesale

企業向け卸業者の、受注から配送完了までを追跡するローカル検証アプリです。

Go / DDD・値オブジェクトによる型保護 / Dafnyによる形式検証 / PostgreSQL / 非同期Command / 楽観ロック / CQRSを組み合わせています。商品検索、在庫引当付きカート、発注、擬似決済、集品、梱包、配送、異常時の保留・再配送・返送、ポイント、PDF請求書メールを扱います。

検証結果と未検証事項は [docs/verification.md](docs/verification.md) を参照してください。Docker Compose起動自体はこの環境では未実行で、同じ構成要素を個別プロセスとして接続して検証しています。

## 起動

UbuntuにDocker EngineとDocker Compose v2が入っていれば、GoやNode.jsをホストに入れる必要はありません。初回ビルドではコンテナ・Go依存・AWS Emulatorのダウンロードにインターネット接続が必要です。

```bash
cd wholesale
# compose.yaml があるディレクトリで実行
docker compose up --build -d
docker compose ps
```

`gateway`がhealthyになったら開きます。

- アプリ: http://localhost:8088
- APIリファレンス（Scalar）: http://localhost:8088/docs
- メール確認（Mailpit）: http://localhost:8025
- [DB閲覧（Adminer）](http://127.0.0.1:8089/?pgsql=postgres)

初期化には少し時間がかかります。起動中の502は数秒待って再読み込みしてください。初期化に失敗した場合は以下で確認します。

```bash
docker compose logs init aws lambda gateway event-source-mapping
```

AWSアカウント・アクセスキー・課金は不要です。実AWSには接続しません。APIとメール画面のポートは127.0.0.1にだけ公開しています。

## 構成

| 要素 | ローカルで動くもの |
|---|---|
| データベース | PostgreSQL 18.3 |
| S3、SQS、IAM、EventBridge Scheduler | 公式 [aws/aws-emulator](https://github.com/aws/aws-emulator) |
| Go Lambda | AWS公式 Runtime Interface Emulator (RIE) |
| API Gateway相当 | HTTPリクエストをLambdaイベントに変換する薄いGoアダプター |
| SQS Event Source Mapping相当 | SQS受信、Lambda呼出し、成功時の削除を行うローカルアダプター |
| 期限・再送・読み取り反映・メール送信 | SchedulerからSQS経由で起動するmaintenance処理 |
| SMTPと受信箱 | Mailpit 1.34.0 |
| 画面 | Goに埋め込んだHTML/CSS/JavaScript |

AWS Emulator 1.0.580で、Go用`provided.al2023`のInvokeが`InvalidRuntimeException`になることを実測しました。そのためLambdaの実行だけをRIEで補います。RIEはAWS公式ですが、Lambdaサービス全体の再現ではありません。

RIEは同時Invokeで異常終了することがあるため、ローカルアダプターだけでファイルロックを共有し、Invokeを直列化しています。DBの業務更新はversionによる楽観ロックです。このファイルロックはDBの悲観ロックではありません。

Step Functionsは採用していません。人が操作する受注状態をPostgreSQL集約で管理し、別の状態機械と二重管理しないためです。RDS管理APIは使わず、PostgreSQL実体をComposeで起動します。

AWS EmulatorのDockerイメージは公式の配布パッケージからビルドします。既定はLTS/latestで、取得時点の版が使われます。アプリ側の依存はgo.mod/go.sum、Dafnyは4.11.0に固定しています。エミュレーターの版は以下で確認できます。

```bash
docker compose exec aws aws-emulator --version
```

## 最初の操作

画面右上の「利用者を切替」でロールを選びます。パスワードはありません。サンプルの会社・氏名・住所は架空です。

| 利用者 | 主な操作 |
|---|---|
| 青葉商事 / 佐藤、港フード / 鈴木 | 商品検索、カート、発注、擬似決済、受注追跡、請求書 |
| 東京倉庫 / 田中、高橋 | 未着手の受注確認、作業開始、欠品保留、再開、梱包完了 |
| 東都配送 / 伊藤 | 箱と配送先の確認、運送開始、配送失敗、再配送、配送完了、返送 |
| 運用 / 山本 | 受注確認、作業中の取消、在庫調整、失敗Command再投入 |

1. 顧客を選び、商品をカートへ追加します。
2. カートで数量と配送先を確認します。クーポンは`WELCOME10`（10%）または`PARTNER5`（5%）です。
3. 発注後、受注詳細で「決済成功」を選びます。
4. 倉庫作業者へ切替え、対象受注で「作業開始」、箱数と重量を入力して「梱包完了」を選びます。
5. 運送業者へ切替え、追跡番号を入力し「運送開始」、最後に「配送完了」を選びます。
6. 顧客に戻るとポイントと請求書PDFが確認できます。MailpitにPDF添付メールが届きます。

操作は非同期です。「受付済み」だけでは業務操作は確定していません。画面は約1.5秒おきに処理結果を確認します。読み取りモデルへの反映が少し遅れる場合があります。

## 業務上の前提

- 単一倉庫、1受注1配送、複数箱。部分出荷・分割請求は対象外。
- カートは実在庫を引き当てます。正常な追加・削除・数量変更のたび、カート全体の期限が1時間後になります。閲覧や同一数量への変更では延長しません。
- 期限到達後は発注できません。引当解放は定期処理まで最大約1分＋処理待ちの遅れがあります。
- 発注時に価格・税・割引・住所を固定します。JPY整数、割引後税抜額に消費税10%、円未満切捨て。
- 配送完了時、割引後税抜100円につき1ポイント。ポイント使用は対象外。
- 決済待ちは15分で取消。決済失敗・取消は在庫引当を解放。
- 出荷時に手持ちと引当を減らします。倉庫返送では自動入庫せず、検品後に在庫調整します。
- 商品画像は初期商品をS3へ登録します。画像アップロードの管理画面は対象外。
- SMTP送信の厳密なexactly-onceは保証しません。送信受付直後のクラッシュでは、同一Message-IDのメールが重複し得ます。

状態遷移と異常系の詳細は [docs/workflow.md](docs/workflow.md)、設計は [docs/design.md](docs/design.md) を参照してください。Mermaid図はGitHub上で表示できます。

## 停止・再起動・リセット

```bash
# 停止。PostgreSQLのデータは保持
docker compose down

# 再起動。AWSリソース・商品画像を再初期化
docker compose up -d

# ソース変更後
docker compose up --build -d
```

AWS Emulatorは停止するとS3/SQSなどの内容を失います。画像の元データと未処理CommandはPostgreSQLに保持しており、初期化と再送で復旧します。`docker compose restart aws`だけではなく、全体をdown/upしてください。

すべての検証データを初期状態へ戻す場合:

```bash
docker compose down -v
docker compose up --build -d
```

`down -v`は受注・在庫変更・Command・請求書を削除します。Mailpitの受信箱もコンテナ再作成で消えます。

## 形式検証・テスト

Dafnyの原本は [formal/Domain.dfy](formal/Domain.dfy)。在庫、カートの引当差分と期限、受注履歴、金額とポイントの計算を形式検証し、入力範囲や配送条件の判定も同じ原本からGoへ生成しています。生成された処理を実際の集約が呼びます。仕様だけを検証して別の手書き実装を使う構成ではありません。

```bash
# Dafny検証、生成コードとの一致、故意のバグ検出、型誤用のコンパイル拒否、Goテスト、静的検査
docker compose --profile test run --build --rm verify

# 実PostgreSQLを使う結合テスト（通常アプリとは別DB）
docker compose --profile test run --build --rm integration

# 起動済みアプリに対するS3/SQS/Lambda/SMTPを含むAPI E2E
# ホストにPython 3が必要。サンプル受注とメールが追加されます。
python3 tests/e2e.py
```

形式検証用コンテナはamd64版Dafnyを使用します。arm64ホストではDockerのamd64エミュレーションが必要です。アプリ本体はGoのビルド対象アーキテクチャに従います。

ホストにGo 1.27.4とDafny 4.11.0を入れた場合:

```bash
go install golang.org/x/tools/cmd/goimports@v0.43.0
export PATH="$(go env GOPATH)/bin:$PATH"
./scripts/verify.sh

# 仕様変更後に生成コードを更新
./scripts/generate.sh
```

`generated/`は直接編集しません。Dafny 4.11が出力する一部のruntime importをスクリプトが機械的に補正します。業務ロジックを生成後に書き換えてはいません。

IDは商品・受注・利用者で別の型です。金額・数量・割引率・ポイント・箱数・重量も専用型とし、集約内部の状態を非公開にしています。DBからの復元時にも不変条件を検査します。型の保証範囲は[設計資料](docs/design.md#型による保護)を参照してください。

Goテストのみなら`go test ./...`。`TEST_DATABASE_URL`が未設定のとき、PostgreSQL結合テストはskipされます。`verify`成功と結合テスト成功を混同しないでください。

## 異常系を試す

- 在庫より多い数量へカートを更新する: 拒否され、元の引当が維持されます。
- 決済失敗を選ぶ: 受注が終端となり、引当が解放されます。
- 倉庫で理由を入力し「欠品・破損で保留」: 同じ担当者が作業を再開できます。
- 配送失敗後「再配送開始」または「倉庫へ返送」: 履歴に残ります。
- 引当より少なくなる在庫調整: 拒否されます。
- 運用者の処理履歴「期限・再送・メール処理を実行」: 1分の定期実行を待たずmaintenanceを投入します。
- 期限切れの手動検証: [scripts/expire-cart.sql](scripts/expire-cart.sql) のコメントに従い、検証用カートの期限を過去にします。

## API

[Scalar APIリファレンス](http://localhost:8088/docs)で8操作のパラメーター、入出力のスキーマ、応答コードを確認できます。「Test Request」でローカルAPIに送信できます。最初に `GET /api/bootstrap` で利用者IDを取得し、利用者の指定が必要な操作では `X-Actor-ID` に入力してください。Commandの送信には `Idempotency-Key` にもUUIDを指定します。

[OpenAPIのYAML](http://localhost:8088/openapi.yaml)とScalar（v1.83.0）はGoアプリが配信します。APIリファレンスの閲覧とリクエスト送信に外部CDNは不要です。

全8操作の契約は [api/openapi.yaml](api/openapi.yaml)（OpenAPI 3.1）です。契約を変更したら `go generate ./api` でGoのルーティング、入出力型、ハンドラーインターフェースを再生成します。生成器は `go.mod` の `tool` で固定し、`verify` で生成物との一致を確認します。ホストにGoを入れずに生成する場合は以下を実行します。

```bash
docker compose --profile test run --build --rm -v "$PWD:/src" verify go generate ./api
```

| メソッド・パス | 用途 |
|---|---|
| GET `/api/health` | DB接続確認 |
| GET `/api/bootstrap` | 初期利用者・商品 |
| GET `/api/state` | Queryモデル、処理結果、請求書情報 |
| POST `/api/commands` | Commandの永続受付。`Idempotency-Key: UUID`必須 |
| POST `/api/commands/{id}/retry` | 運用者によるdead Command再投入 |
| POST `/api/maintenance` | 運用者による定期処理の即時投入 |
| GET `/api/images/{id}` | S3画像 |
| GET `/api/invoices/{id}` | 所有者・運用者のPDF取得 |

利用者は`X-Actor-ID: UUID`で指定します。擬似ログインであり、本番認証には使えません。公開Commandの名前と入力項目はOpenAPIに定義し、形が合わないリクエストは受付前に拒否します。権限・在庫・遷移などの業務条件は非同期実行時に評価します。実例は `tests/e2e.py` にあります。同じキー・同じ内容は同じCommand、異なる内容は409です。

## Codexへの引継ぎ

リポジトリ直下の [AGENTS.md](AGENTS.md) に、維持する設計条件、コードの入口、修正・検証手順、既知の制約を記載しています。手元のCodexはこのディレクトリで起動してください。

## リポジトリ

| ディレクトリ | 内容 |
|---|---|
| `api` | OpenAPI契約とGo生成設定 |
| `formal` | Dafny仕様と検証対象ロジック |
| `generated` | 生成GoとDafny runtime |
| `internal/domain` | 集約と生成コードへの境界 |
| `internal/app` | ユースケース・Command処理・請求 |
| `internal/store` | PostgreSQLリポジトリ・スキーマ・seed |
| `internal/cloud` | AWS SDKアダプター・初期化 |
| `internal/web` | API・画面 |
| `cmd/app` | Lambda、ローカルgateway、配信アダプター |
| `deploy` | Dockerfile |
| `tests` | 実サービス接続E2E |
| `docs` | 設計・状態遷移・検証記録 |

GitHubに置く場合は、このディレクトリを通常のGitリポジトリとして追加できます。秘密情報・実AWS資格情報は不要です。`.build`、ローカル環境ファイルは`.gitignore`で除外しています。
