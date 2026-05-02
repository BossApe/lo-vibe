# tools ディレクトリ

プロジェクトをスムーズに進めるためのツールスクリプトを格納するディレクトリです。

## 概要

開発・運用作業の効率化や自動化を目的としたスクリプト類を管理します。

## ディレクトリ構成

```text
tools/
├── README.md             # このファイル
├── save_promptinglog/    # Copilot プロンプトログ保存ツール (Go)
├── generate_issue_roadmap_vegalite/ # Issue難易度ロードマップ生成ツール (Shell/Python)
└── set_project_fields/   # GitHub Project フィールド一括設定ツール (Shell)
```

## ツール一覧

| ディレクトリ / ファイル名 | 種別 | 概要 |
| --- | --- | --- |
| `save_promptinglog/` | Go バッチ | Copilot Chat のトランスクリプトを Markdown 形式で保存する |
| `generate_issue_roadmap_vegalite/` | Shell/Python | `gh api` で取得した Issue 一覧から Vega-Lite 難易度ロードマップ Markdown を生成する |
| `set_project_fields/` | Shell script | GitHub Project のカスタムフィールドを一括設定する |

## 利用方法

各ツールのディレクトリ内の README を参照してください。

## 規約

- ツールはディレクトリ単位で管理する
- 各ツールディレクトリに README.md を作成し、概要・ビルド・使い方を記載する
- 新規ツールを追加した場合は本 README のツール一覧を更新する