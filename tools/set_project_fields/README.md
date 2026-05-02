# set_project_fields

GitHub Projects のアイテムに対して、カスタムフィールドを一括設定するためのシェルスクリプトです。

## 目的

Project の再作成後や初期投入後に、Type、Phase、Iteration、Service、Priority、Estimate、Parent、Depends on をまとめて反映する。

## 前提条件

- GitHub CLI (`gh`) が利用可能であること
- `gh auth login` 済みで、対象 Project を編集できる権限があること
- スクリプト内の Project ID、Field ID、Option ID、Item ID が対象 Project の現行値と一致していること

## 使用方法

```bash
cd tools/set_project_fields
./set_project_fields.sh
```

構文チェックのみ行う場合:

```bash
cd tools/set_project_fields
sh -n set_project_fields.sh
```

## 入出力

- 入力: スクリプト内部に定義された Project ID、Field ID、Option ID、Item ID
- 出力: `gh project item-edit` の実行結果を標準出力へ出す
- 影響: 対象 Project の各アイテムのカスタムフィールド値を更新する

## 注意事項

- 対象の Project を作り直した場合、ID が変わるため値の更新が必要
- 誤った ID のまま実行すると、意図しない Project に更新をかけるか、コマンド失敗になる
- 破壊的な削除処理は含まないが、既存のフィールド値は上書きされる