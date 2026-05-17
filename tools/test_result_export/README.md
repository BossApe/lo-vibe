# test_result_export

テスト結果ファイルを規約どおりの固定ファイル名で出力するシェルスクリプトです。

## できること

- テスト種別ごとの保存先ディレクトリを自動判定する
- 対象ID規則をチェックする
- `FR-001単体テスト結果.txt` のような固定ファイル名で結果を出力する
- 外部結合テスト・受け入れテストなど、未実行テスト用の雛形ファイルを生成する

## 使い方

```bash
# 単体テスト結果を出力
tools/test_result_export/export_test_result.sh \
  unit FR-001 \
  go test ./internal/handler ./internal/service -run SystemOverview -v

# 内部結合テスト結果を出力
tools/test_result_export/export_test_result.sh \
  internal FR-001 \
  go test ./test/integration -run 'SystemOverview|ヘルスチェック' -v

# 外部結合テスト結果の雛形を生成
tools/test_result_export/export_test_result.sh --template external FR-001

# 受け入れテスト結果の雛形を生成
tools/test_result_export/export_test_result.sh --template acceptance FR-001
```

## 対象ID規則

- `unit`, `internal`, `external`, `acceptance`, `api`: `FR-001`
- `e2e`, `security`: `SPR-1-1`
- `performance`: `PH-1`
- `regression`: `MRG-123`

## 出力先

- `001.単体テスト`
- `002.内部結合テスト`
- `003.外部結合テスト`
- `004.受け入れテスト`
- `005.APIテスト`
- `006.E2Eテスト`
- `007.セキュリティテスト`
- `008.性能テスト`
- `009.回帰テスト`
