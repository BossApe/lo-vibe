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

### 最新セッションを自動検出して差分保存（推奨）

```bash
./save_promptinglog
```

### セッション ID を指定して保存

```bash
./save_promptinglog -session-id <セッションID>
```

### 何日前のログを保存

```bash
./save_promptinglog -d2
./save_promptinglog -d=2
```

`-d2` と `-d=2` はどちらも、2 日前 (JST) の会話ログのみを保存します。

受け付ける形式は `-d2` / `-d=2` / `-d 2` です。

### オプション一覧

| フラグ | デフォルト | 説明 |
| --- | --- | --- |
| `-session-id` | （なし = 最新） | 保存するセッション ID (UUID)。省略時は前回保存分の更新時刻以降のみを保存 |
| `-d` | （未指定） | 何日前のログを保存するか。例: `-d2` / `-d=2` は2日前 (JST)。`0` 以上のみ指定可能 |
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

`-session-id` を省略した場合は、`-output-dir` にある最新の `promptinglog_*.md` の更新時刻を境界に、
それ以降の会話ログのみを保存します。

`-d` を指定した場合は差分保存よりも `-d` の条件を優先し、対象日 (JST) の会話ログのみを保存します。

`-d` 指定時に対象ログが 0 件だった場合は、Markdown ファイルを作成せず `指定日(YYYY-MM-DD)のログはありません` を表示します。

差分対象のログが 0 件だった場合は、新しい Markdown ファイルは作成せず `新規ログはありません` を表示して終了します。

`-session-id` を指定した場合は差分保存を行わず、指定セッションの会話ログ全体を保存します。

## 条件の優先順位

複数条件が同時に指定された場合は、以下の順に評価されます。

1. `-d` が指定されていれば、対象日 (JST) フィルタを適用
2. `-d` 未指定で `-session-id` が指定されていれば、指定セッション全体を保存
3. どちらも未指定なら、前回保存分以降の差分を保存

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