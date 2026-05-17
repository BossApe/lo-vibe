# generate_issue_roadmap

GitHub Projects の Issue 一覧を `gh api` で取得し、難易度ロードマップ HTML を自動生成する Go ツールです。

## 概要

- `gh api graphql` で Project v2 の Issue とフィールド情報を取得
- `Estimate` を難易度ポイントに変換し、親レーン単位で実行順に累積表示
- 出力先ディレクトリを指定して HTML を生成

## ファイル構成

```text
generate_issue_roadmap/
├── main.go            # CLI エントリポイント
├── go.mod             # Go モジュール定義
├── README.md          # このファイル
├── generate_issue_roadmap  # ビルド済みバイナリ (.gitignore 対象)
└── internal/
  └── roadmap/
    ├── roadmap.go       # 実行フロー
    ├── config.go        # 設定/引数処理
    ├── fetch.go         # GitHub Project 取得
    ├── lane.go          # レーン構築/集計
    ├── render.go        # HTML描画
    ├── model.go         # 型定義
    └── roadmap_test.go  # 単体テスト
```

## 難易度マッピング

- `XS` = 1
- `S` = 2
- `M` = 3
- `L` = 5
- `XL` = 8

`Estimate` が未設定または不正な場合は `1` として扱います。

## ビルド

```bash
cd tools/generate_issue_roadmap
go build -o generate_issue_roadmap .
```

## テスト

```bash
cd tools/generate_issue_roadmap
go test ./...
```

## 使い方

```bash
cd tools/generate_issue_roadmap
./generate_issue_roadmap \
  -owner BossApe \
  -repo Musuhi \
  -project-number 2 \
  -output-dir ../../_document/003.設計・開発・テストフェーズ/002.開発進捗
```

## オプション一覧

| フラグ | デフォルト | 説明 |
| --- | --- | --- |
| `-owner` | `BossApe` | GitHub owner (user または organization) |
| `-repo` | `Musuhi` | 対象リポジトリ名 |
| `-project-number` | `2` | GitHub Project (v2) 番号 |
| `-output-dir` | `../../_document/003.設計・開発・テストフェーズ/002.開発進捗` | 出力先ディレクトリ |

## 出力ファイル名

生成ファイル名は以下の形式です。

- `<プロジェクト名>開発進捗_<yyyymmddhhmmss>.html`

例:

- `Musuhi開発進捗_20260503183045.html`

## 前提条件

- `gh` コマンドが利用可能であること
- `gh auth login` 済みで対象 Project を参照できること
- 対象 Project の item 数が 100 件を超える場合はページング拡張が必要
