# 第三者ソフトウェア

- `generated/dafny/`はDafny 4.11.0のGo runtime。MIT License。著作権表示は各ソース、ライセンス本文は`third_party/Dafny-LICENSE.txt`。
- `internal/app/fonts/BIZUDGothic-Regular.ttf`はBIZ UDGothic。SIL Open Font License 1.1。ライセンス本文は同ディレクトリの`OFL.txt`。
- Go依存の版とハッシュは`go.mod`と`go.sum`に固定。各依存のライセンスは各配布元に従う。
- AWS Emulator、AWS Lambda RIE、PostgreSQL、Mailpit、Go、Dafny実行環境はDockerビルド時に配布元から取得する。バイナリはこのリポジトリに同梱していない。

配布元:

- https://github.com/dafny-lang/dafny
- https://github.com/googlefonts/morisawa-biz-ud-gothic
- https://github.com/aws/aws-emulator
- https://github.com/aws/aws-lambda-runtime-interface-emulator
- https://www.postgresql.org/
- https://github.com/axllent/mailpit
