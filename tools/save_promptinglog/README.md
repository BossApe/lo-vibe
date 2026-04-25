# save_promptinglog

GitHub Copilot Chat のトランスクリプトを Markdown 形式のプロンプトログとして保存するバッチツールです。

## 概要

VS Code が保存している Copilot Chat のトランスクリプト JSONL ファイルを読み込み、
ユーザーとアシスタントの会話を整形して
`lo-vibe/_document/promptinglog/promptinglog_<yyyymmddhhmmss>.md`
として保存します。

## ファイル構成

```text
save_promptinglog/
├── main.go              # CLI エントリポイント
├── go.mod               # Go モジュール定義
├── README.md            # このファイル
├── save_promptinglog    # ビルド済みバイナリ (.gitignore 対象)
└── source/
    ├── app.go           # 実装本体
    └── app_test.go      # 単体テスト
```

## ビルド

```bash
cd tools/save_promptinglog
go build -o save_promptinglog .
```

## テスト

```bash
cd tools/save_promptinglog
go test ./...
```

## 使い方

### 最新セッションを自動検出して保存（推奨）

```bash
./save_promptinglog
```

### セッション ID を指定して保存

```bash
./save_promptinglog -session-id <セッションID>
```

### オプション一覧

| フラグ | デフォルト | 説明 |
| --- | --- | --- |
| `-session-id` | （なし = 最新） | 保存するセッション ID (UUID) |
| `-output-dir` | `../../_document/promptinglog` | 出力先ディレクトリ |
| `-storage-dir` | `~/Library/Application Support/Code/User/workspaceStorage` | VS Code の workspaceStorage ルートパス |

## トランスクリプトファイルの場所

VS Code が管理する Copilot Chat のトランスクリプトは以下に保存されています。

```text
~/Library/Application Support/Code/User/workspaceStorage/
  └── <workspaceId>/
        └── GitHub.copilot-chat/
              └── transcripts/
                    └── <sessionId>.jsonl
```

本ツールは `-storage-dir` 以下を再帰検索し、最も更新日時が新しい `.jsonl` ファイルを自動選択します。

## 出力フォーマット

出力される Markdown ファイルには以下の情報が含まれます。

- セッション情報（セッション ID・開始日時・バージョン）
- 会話ログ（ターン番号・ユーザー発言・アシスタント応答・タイムスタンプ）

## 実行例

```bash
$ ./save_promptinglog
トランスクリプト: /Users/m.nohara/.../transcripts/cce9a725-....jsonl
保存完了: ../../_document/promptinglog/promptinglog_20260425183000.md
ターン数: 5
```