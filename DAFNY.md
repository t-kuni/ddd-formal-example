# Domain.dfy の読み方

[formal/Domain.dfy](formal/Domain.dfy) は、在庫・カート・受注・配送のルールを Dafny 4.11.0 で記述したファイルです。まず `Reserve` で「呼べる条件・返り値の保証・実際の処理」の区別をつかみ、次に `PlanCartChange` と `CheckHistory` を読むと、ファイル全体で使う構文を追えます。

## 1. `Reserve` で契約と処理を読む

```dafny
predicate StockValid(h: int, r: int) { 0 <= r <= h <= StockLimit }

method Reserve(h: int, r: int, delta: int) returns (ok: bool, next: int)
  requires StockValid(h,r)
  ensures ok <==> 0 <= r + delta <= h
  ensures ok ==> next == r + delta && StockValid(h,next)
  ensures !ok ==> next == r
{
  ok := 0 <= r + delta <= h;
  next := if ok then r + delta else r;
}
```

[原本の在庫処理](formal/Domain.dfy#L28-L60)

- `h: int` は手持ち数、`r: int` は引当数、`delta: int` は引当の増減です。`returns (...)` は名前付きの出力を宣言します。ここでは成否の `ok` と変更後の引当数 `next` を返します。
- `requires` は呼び出し側が満たす事前条件です。この処理は、引当数が非負で、引当数 ≤ 手持ち数 ≤ 上限である場合に呼べます。Dafny 内から呼ぶときは、呼び出し側でこの条件を証明します。
- `ensures` は処理側が満たす事後条件です。複数あればすべて成立する必要があります。1 行目は「成功するのは変更後の引当数が範囲内の場合に限り、その場合は必ず成功する」、2 行目は「成功なら変更後の引当数が計算どおりで在庫の条件も保つ」、3 行目は「失敗なら元の引当数を返す」と読みます。
- `{ ... }` の本体では、`:=` で出力変数に代入します。`if ... then ... else ...` は値を選ぶ式です。出力を代入して本体の末尾に達すると、その値が返ります。

たとえば `Reserve(10, 4, 3)` は `(true, 7)`、`Reserve(10, 4, 7)` は `(false, 4)` を返します。`Reserve(10, 11, -1)` は事前条件の範囲外です。上の契約で説明している失敗は、事前条件を満たした入力に対する業務上の拒否です。

## 2. 宣言・式・証明の早見表

| 構文 | 読み方と原本での用途 |
| --- | --- |
| `module Domain { ... }` | 宣言をまとめるモジュール。 |
| `const StockLimit: int := 1000000000` | 型付き定数。 |
| `int`、`bool`、`string` | 任意精度の整数、真偽値、文字列。時刻を Go の符号付き64ビット整数に収める上限は、`PlanCartChange` に明示されています。 |
| `predicate StockValid(...)` | 真偽値を返す関数。契約でも実行処理でも使えます。 |
| `function LineAmount(...): int { ... }` | 副作用のない関数。末尾の型が結果型で、本体の式が結果です。`ReleasedReservation` の `ensures` にある `ReleasedReservation(quantity)` は関数の結果を指します。 |
| `method ... returns (...)` | 代入や分岐などの文で処理を記述します。`var expiry := Touch(now);` はメソッドを呼び、結果を型推論つきのローカル変数で受け取ります。 |
| `if 条件 { ... } else if 条件 { ... } else { ... }` | 文の分岐。先に成立した分岐を実行します。`return 値;` で出力に値を渡して途中で終了する書き方もあります。 |
| `==`、`!=`、`0 <= r <= h` | 等しい、異なる、範囲の連鎖比較。`0 <= r <= h` は `0 <= r && r <= h` と同じ条件です。 |
| `!`、`&&`、`\|\|` | 否定、かつ、または。`&&` と `\|\|` は左側だけで結果が決まれば右側を評価しません。 |
| `P ==> Q`、`P <==> Q` | 「P なら Q」と「P と Q は同値」。`==>` で P が偽なら、Q に追加の制約はかかりません。 |
| `set<string>`、`{product}`、`in`、`!in` | 文字列の有限集合、1 要素の集合、所属、非所属。`seen + {product}` は集合の和で、元の `seen` は変えません。`RegisterProduct` が重複を判定するのに使います。 |
| `assert 条件;` | その地点で条件が成立することを検証器に証明させます。証明した条件は後続の推論に使えます。 |
| `// ...` | 行コメント。`Allowed` の直前に状態番号と名称の対応があります。 |

[公式 Dafny リファレンス](https://dafny.org/v4.11.0/DafnyRef/DafnyRef) · [公式クイックリファレンス](https://dafny.org/latest/QuickReference)

`Price` の `assert 0 <= subtotal * percent <= subtotal * 100;` は、割引計算に使う積の範囲を示す証明の途中経過です。非負の整数を正の整数で割ると小数部分が切り捨てられます。税抜小計が 1,235 円で割引率が 10% なら、割引 123 円、割引後税抜 1,112 円、税 111 円、合計 1,223 円、11 ポイントです。[`Price` の契約と計算](formal/Domain.dfy#L103-L118)

## 3. `PlanCartChange` で結果の種類を読む

```dafny
datatype CartChange = CartFull | CartUnchanged | CartTimeOutOfRange |
  CartChanged(reset: bool, delta: int, expiry: int)
```

`datatype` は、いくつかの形のうち一つを取る値型です。`|` は候補の区切りで、各候補の名前はコンストラクターです。`CartFull` は追加データを持たず、`CartChanged(true, 2, 7200)` はリセットの要否、引当差分、次の期限を持ちます。

- `plan.CartChanged?` は「結果が `CartChanged` か」を調べます。`plan.reset` などのフィールドは、それを持つコンストラクターに対して読みます。`plan.CartChanged? ==> ...` は、保証とフィールド参照の両方を該当する結果に絞ります。
- 本体の分岐順は「有効期限内の満杯カートへの新規追加」「変更なし」「期限を表せる時刻範囲の超過」「変更成立」です。たとえば、有効期限内で数量を 2 から 5 に変え、時刻上限にも達していなければ、`reset` は偽、`delta` は 3、`expiry` は `now + 3600` です。
- `now >= expires` の場合、変更が成立すると `reset` は真です。`delta` の定義は常に `requested - previousQuantity` です。期限切れ時の既存引当の解放は Go 側の処理と合わせて読みます。

[カート変更の定義と契約](formal/Domain.dfy#L131-L155) · [Go 側の呼び出し](internal/domain/types.go#L163)

## 4. `CheckHistory` で列と全称量化を読む

```dafny
predicate HistoryStepsValid(edges: seq<HistoryEdge>) {
  forall i | 1 <= i < |edges| ::
    edges[i].from == edges[i-1].to && Allowed(edges[i].from, edges[i].to)
}
```

- `seq<HistoryEdge>` は順番を保持する不変の列です。`|edges|` は長さ、`edges[i]` は 0 始まりの添字です。`edges[i].from` はその要素の開始状態を読みます。
- `forall i | 範囲 :: 条件` は「範囲内のすべての `i` で条件が成立する」です。ここでは 2 件目から末尾まで、直前の終了状態と現在の開始状態が一致し、現在の遷移が `Allowed` に含まれることを求めます。要素が 0 件または 1 件なら対象の添字がないため、この述語は真です。
- `CheckHistory` はさらに、列が空でなく、先頭が `-1 → 0` で、末尾の終了状態が現在の受注状態に一致することを求めます。`ensures` の `|edges| > 0 && ...` は、空の列に添字でアクセスしないためのガードにもなります。

たとえば `[-1 → 0, 0 → 1, 1 → 2]` と現在状態 `2` の組み合わせは `ValidHistory` です。同じ履歴で現在状態が `1` なら `StatusMismatch` です。[`Allowed`](formal/Domain.dfy#L62-L81) · [履歴の定義と判定](formal/Domain.dfy#L184-L204)

## 5. 残りのルールを探す

| 知りたいこと | 読む場所 |
| --- | --- |
| 数量・金額・箱数・重量などの入力範囲、ポイント残高の更新 | [値の範囲と Credit](formal/Domain.dfy#L2-L26) |
| 引当・在庫調整・出荷 | [Reserve、Adjust、Ship](formal/Domain.dfy#L28-L60)。`Ship` の `nh-nr == h-r` は出荷前後で販売可能数が同じという保証です。 |
| 受注で許される状態遷移 | [Allowed と Transition](formal/Domain.dfy#L62-L81)。`StatusValid` は番号の範囲、`Allowed` は状態の組み合わせを判定します。 |
| カートの延長・失効・発注の境界 | [Touch、Expired、MayExpire、MayCheckout](formal/Domain.dfy#L83-L101)。期限ちょうどでは発注不可で、失効要求は現在の期限と一致する必要があります。 |
| 明細の妥当性・商品重複・小計の上限 | [LineValid、RegisterProduct、AddLineAmount](formal/Domain.dfy#L157-L182) |
| 作業担当、箱、配送担当・追跡番号の条件 | [作業・配送に関する述語](formal/Domain.dfy#L206-L224) |

## 検証の範囲

Dafny は事前条件を満たす入力について、処理が契約を満たすことや、添字アクセスなどが安全であることを静的に検証します。実行用の述語・関数・メソッドは Go に生成され、`requires`・`ensures`・`assert` は静的検証に使われます。Go 側の入力検査、DB の保存、トランザクション、外部サービスとの連携は、別途テストで確認する範囲です。[設計](docs/design.md) · [検証結果](docs/verification.md) · [検証の実行方法](README.md#形式検証テスト)
