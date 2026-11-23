# 設計・技術選定

ローカル専用の卸受注・配送検証アプリ「ORBIT Wholesale」。

## 決定

| 領域 | 採用 | 理由 |
|---|---|---|
| 形式検証 | Dafny 4.11.0 + 同梱Z3 | 契約と実行ロジックを同じソースに置き、検証した処理をGoに生成する |
| バックエンド | Go、標準net/http、pgx v5、GORM、AWS SDK v2 | 永続化と外部サービスのアダプターを明確に分離 |
| DB | PostgreSQL 18 | 正規化した業務テーブル、トランザクション、一意制約、楽観ロック |
| DB閲覧 | Adminer | ブラウザでテーブル定義と保存データを確認。ホスト側の公開先は127.0.0.1 |
| 非同期 | AWS EmulatorのSQS Standard、Go Lambda、EventBridge Scheduler | 少なくとも1回の配信を前提に、DBで冪等性を保証する |
| 画像 | AWS EmulatorのS3 | 公式エミュレーターを実際のSDKで使用。LocalStackは使わない |
| UI | HTML/CSS/JavaScript ES modules、Goによる配信 | Nodeのビルド環境なしで起動でき、画面・業務フローの検証に集中する |
| 請求書 | GoのPDF生成、SMTP、Mailpit | ローカルでPDF添付メールの内容を閲覧できる |
| 起動 | Docker Compose | DB、AWS、API、各worker、メールを分離 |

Dafnyを原本とするのは値の範囲、在庫とカートの引当、金額・ポイント計算、受注の状態遷移・履歴と配送条件。CommandワーカーはGoの操作実装を渡して `ExecuteCommand` を1回呼ぶ。Dafnyの `RequestState` が最初の集約とIDを決め、`DecideCommand` が更新内容を返し、`DependenciesFor` が更新に必要な追加データを示す。カート・受注の明細は `map<string, int>` で受け渡す。Commandの個別判断はカート・発注、受注・配送、在庫調整のモジュールに分け、各Commandが名前付きの純粋関数を持つ。 [Commandごとの関数一覧](../DAFNY.md#commandごとの入口)

Goは `Effects` を通じて、指定されたIDのSQL読み込み、型付き集約への更新適用、PDF生成、トランザクションを実装する。在庫行は商品ID順に保存する。Command行をロックし、業務上の拒否時はセーブポイントまで戻して、同じロックの下で拒否を記録する。成功時は業務変更とCommand完了を同じトランザクションでコミットする。Dafnyが楽観ロック競合からの再試行を指示し、その試行のトランザクションで読み直す。生成Goは直接編集しない。

Lambdaを採用する。Step Functionsは、今回の人手による業務状態をDBと二重管理しないため採用しない。非同期処理にアプリケーションのgoroutineを使用しない。HTTP・DB・SDK内部のgoroutineは許容する。

## 境界と集約

- Catalog: 商品マスタ（SKU、名称、カテゴリー、単位、価格、画像）。
- Inventory: 商品別在庫集約。手持ち・引当数。販売可能数=手持ち−引当。
- Cart: 顧客ごとのカート集約。商品明細・数量・最終変更時刻・期限。
- Ordering/Fulfillment: 受注集約。価格・税・住所のスナップショット、箱、作業担当、配送追跡、履歴を一体管理。
- Loyalty: 顧客ポイント集約。配送完了時に付与。使用は今回対象外。
- Billing: 請求書とメール送信ジョブ。受注と別の外部副作用。

受注〜配送を主眼に、単一倉庫・1受注1配送（複数箱）とする。部分出荷・複数倉庫間の移動・分割請求は対象外。配送不能から再配送、倉庫返送を扱う。返送品は検品前なので在庫に自動復帰しない。オペレーターの在庫調整で戻す。

整合性優先のモジュラーモノリスとする。カートと複数SKU、注文確定と引当移管など、強い整合性が必要な変更は複数集約を1つのDBトランザクションで保存する。DDDの「1トランザクション1集約」を機械的には適用しない。集約の責務・リポジトリは分ける。

## カートと引当

正常に成立した商品追加・数量変更・削除で、カート全体をDB時刻から1時間延長。閲覧、同じ数量への変更、重複Command、拒否された操作では延長しない。空のカートに引当はない。期限到達と同時に発注不可。定期処理の遅延中も期限切れカートは発注できない。解放が遅れる間、販売可能数は保守的に少なく見える。

期限切れCommandは対象の期限を含む。更新済みの期限と一致しない古いCommandは何もしない。期限切れカートへの変更は古い全引当を解放してから新しい内容を同一トランザクションで引き当てる。発注時はカート引当を受注引当に移管（在庫の引当総数は変わらない）。決済失敗・キャンセルで解放し、出荷時に手持ちと引当の両方を減らす。

## 金額・顧客

JPY整数、価格は税抜。商品合計にクーポン割引率を適用して円未満切捨て、その後10%税を円未満切捨て。クーポンは受注につき1枚、マスタに有効なものを使用。価格と割引は発注時に固定。配送完了時、割引後税抜100円につき1ポイント、1回のみ付与。決済は顧客が成功/失敗を選ぶ擬似決済。請求書は配送完了後に発行。税込請求額、明細、住所、割引、税を含む。制度上の適格請求書対応を目的としない。

顧客・倉庫作業者・運送業者・オペレーターはseed済み。画面で利用者を切替。パスワードはないが、APIでロールと所有者を検査する。リソースIDはUUID。SKU・伝票番号などの業務コードは別属性。

## HTTP契約

`api/openapi.yaml` をHTTP契約の原本とする。固定版の `oapi-codegen` が `net/http` のルート、型付きハンドラー、入出力型を生成する。入力はミドルウェアで検証し、Commandは `type` ごとに必要な `payload` を `oneOf` で指定する。Queryのレスポンスは各画面の用途に合わせて定義する。HTTPのDTOは `internal/web` で既存のアプリケーション入力へ変換し、業務ルールはドメイン層で評価する。

## CommandとQuery

POST /api/commands はUUIDのIdempotency-Key、actor、type、payloadをDBに永続化して202を返す。同一キー・同一内容は同じ結果を返し、内容違いは409。Commandの受付時点では業務操作は未確定。画面はCommand完了を確認して、表示中のQueryを再取得する。

dispatcherは未完了CommandのIDをSQSに周期的に送信。DBが正本なので、送信とDB記録間のクラッシュは再送で回復できる。workerは同一トランザクションで集約取得、ドメイン操作・検査、version一致のUPDATE、明細等の更新、Command完了を行う。業務違反はrejected、インフラ障害は再試行、上限超過はdead。deadはオペレーターが再投入できる。DBコミット後にSQSを削除。重複受信は完了済みCommandを確認して無処理。

### PostgreSQLの正本

| 最新の事実 | テーブル |
|---|---|
| 利用者、顧客住所、商品 | `actors`、`customer_addresses`、`products` |
| 手持ち・引当、カートと明細 | `inventory`、`carts`、`cart_items` |
| 受注の現在状態、確定した配送先と明細 | `orders`、`order_addresses`、`order_items` |
| 配送と箱 | `shipments`、`shipment_boxes` |
| 受注の状態遷移を追跡する履歴 | `order_status_history` |
| 現在のポイント残高と受注別の付与実績 | `loyalty_accounts`、`order_point_awards` |
| 受付結果と請求書・メール送信状態 | `commands`、`invoices` |

受注の状態変更と追跡履歴の追記をまとめてコミットする。確定明細の価格・商品表示情報・配送先は発注時の値を保持する。受注とポイント付与実績、請求書は一意制約で結び付ける。

### Query

用途別のAPIが正規化テーブルを直接参照する。受注一覧は顧客・キュー・ステータスをSQLで絞り、発注日時とUUIDをカーソルキーとして25件ずつ取得する。件数と保有ポイントは同じDBスナップショットで集計する。受注詳細だけが明細・箱・状態履歴を読む。

pgxでCommandの楽観ロック・トランザクションと一覧の集計SQLを扱う。受注詳細ではGORMの `Preload` で配送先、明細、配送と箱、状態履歴を読み、明細と履歴の順序を指定する。関連取得は読み取り専用の `REPEATABLE READ` で行う。GORMはpgxと接続プールを共有し、テーブル定義は `schema.sql` を原本とする。ライブラリの依存は永続化アダプターに置き、業務ロジックとHTTP契約を独立させる。GORMの本体とPostgreSQLドライバーは同じ公開組織で保守され、版は `go.mod` に固定する。 [GORMのリリース](https://github.com/go-gorm/gorm/releases) [PostgreSQLドライバー](https://github.com/go-gorm/postgres)

楽観ロック競合は最新状態からトランザクション全体をやり直す。利用者が期待versionを指定した操作は、変化していれば業務上の競合として拒否する。同一Commandの並行実行はCommand行のロックで直列化し、異なるCommand間の集約更新は楽観ロックで競合を検出する。

SMTP送信とDBコミットは原子的にできない。請求書の作成・ポイント付与は冪等だが、SMTP受付後のクラッシュでは同じMessage-IDのメールが重複し得る。exactly-once送信とは主張しない。

## 外部プロセス

Go LambdaにHTTPリクエスト、Command用SQSイベント、maintenance用SQSイベントの入口を用意する。EventBridge Schedulerが1分ごとにmaintenanceキューへ送信。maintenanceで期限切れCommand作成、未配信Commandの再送、請求メール送信を行う。

ローカルのAWS Emulator 1.0.580では、Goのprovided.al2023をInvokeするとInvalidRuntimeExceptionになることを確認した。この部分はAWS公式Lambda Runtime Interface Emulator (RIE)でGo Lambdaを実行する。API Gateway相当のHTTPイベント変換と、SQS Event Source Mapping相当の受信・Lambda呼出し・成功時削除のみを薄いローカルアダプターで補う。アダプターに業務処理は置かない。RIEはLambdaサービス全体を再現するものではない。

S3・SQS・IAM・EventBridge Schedulerは公式AWS Emulatorの実API。RDBはPostgreSQL実体、メールはMailpit。RDSの管理APIを追加してもDBの検証には寄与しないため、ローカルではPostgreSQLコンテナを使用する。SQSのdead-letter queueとDBのdead状態を設ける。DBが正本なので、キュー消失からの再送・運用画面からの再投入が可能。

AWS Emulatorは再起動でS3/SQSを失う。商品画像の元データはPostgreSQLにも保持し、再初期化でS3へ復元。未処理CommandはDBから再送。初期化は冪等。

## 検証の境界

1. Dafny: 在庫と引当の不変条件、カートの差分・期限更新と失効、正常な受注履歴、明細小計・割引・税・ポイントの事後条件。Commandの事後条件は、カート読取先の権限、カートの次の明細と引当差分、受注取消時の解放、配送完了時の付与ポイント数と請求書作成を保証する。操作契約の下で、成功記録を更新適用後に書き、成功か拒否の記録後にコミットする。
2. Go: 生成コードを呼ぶ集約操作、UUIDとJSONの変換、保存値と再計算結果の照合、決済・配送完了の冪等性、ロール・所有権の実行結果。
3. PostgreSQL: 古いversionの書込み失敗、同じキーの競合、複数集約とCommand完了の原子性、拒否時の部分更新の取り消し。
4. AWS Emulator: S3画像、SQS送受信/重複/再送、プロセス終了後の再開。
5. UI/E2E: 顧客から倉庫・運送へ、配送完了、請求PDF、ポイント。

コミット応答を受け取れない場合は、再配信時にDBのCommand記録を読み直す。外部I/Oの成功は実行環境に依存する。最終的な処理完了にはDB/AWSの回復とworkerの継続稼働が必要。

## 型による保護

`internal/domain/values.go`に値オブジェクトを定義する。HTTP payload、DB復元用の値、Queryのレスポンスは境界のDTOとしてプリミティブ値を用い、業務処理へ入る際に検証して専用型へ変換する。

| 専用型・構造 | 防ぐ誤り |
|---|---|
| `ProductID` / `OrderID` / `ActorID` | 同じUUIDでもリソース種別を取り違えて渡すこと |
| `Quantity` / `OrderQuantity` / `CartQuantity` | 在庫数、正の受注明細数量、削除を含むカート数量の混同 |
| `ReservationDelta` / `StockAdjustment` | 引当増減と実在庫調整の取り違え |
| `Money` / `UnitPrice` / `DiscountRate` / `Points` | 金額・単価・率・ポイントの混同、範囲外の生成 |
| `BoxCount` / `WeightGrams` | 箱数と重量の取り違え |
| `ShippingAddress` | 空の必須項目を持つ配送先の生成 |
| `OrderStatus` | 状態を通常の数値と取り違えること。未知の値・不許可遷移は実行時にも拒否 |
| 非公開の集約フィールド | 集約メソッドを迂回する状態の書換え |
| sealed `Aggregate` / `RestorableAggregate` | DTOや任意のmapを集約としてリポジトリへ保存・復元すること |

Goには完全な閉じたenumや、ゼロ値の生成を禁止する機構はない。型だけですべてを保証できるとはしない。未初期化のID・正数量・配送先・受注はコンストラクターと集約操作で拒否する。DB復元時にも、不正な在庫、金額の不一致、履歴にない状態、担当者・箱・追跡番号の不足を検査する。

`Snapshot()`は独立したDTOコピーを返す。スライスやmapを変更しても集約内部は変化しない。生成Goはアプリケーション層から直接呼ばず、ドメイン層で値オブジェクトの境界を通す。

`./scripts/typecheck.sh`は、商品IDを受注IDに代入する、金額を引当増減に渡す、在庫内部を直接書換えるコードを実際にコンパイルし、失敗することを確認する。
