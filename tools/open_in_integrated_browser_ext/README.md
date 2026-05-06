# Open in Integrated Browser

Explorer の右クリックメニューに「統合ブラウザで開く」を追加する VS Code 拡張です。

## 目的

HTML ファイルを OS 既定ブラウザではなく、VS Code の統合ブラウザ（Simple Browser）で開くために使用します。

## 前提条件

- VS Code 1.90.0 以上
- Node.js / npm（拡張の依存解決と実行に必要）
- 対象ファイルがローカルファイル（`file` スキーム）であること

## ディレクトリ構成

```text
tools/open_in_integrated_browser_ext/
  README.md
  package.json
  extension.js
```

## 使い方

### 1. 依存をインストール

```bash
cd tools/open_in_integrated_browser_ext
npm install
```

### 2. 拡張を起動（開発モード）

1. VS Code で `tools/open_in_integrated_browser_ext` を開く
2. `F5` キーで Extension Development Host を起動
3. 開いた別ウィンドウ側で確認する

### 3. HTML を統合ブラウザで開く

1. Explorer で `.html` または `.htm` ファイルを右クリック
2. `統合ブラウザで開く` を選択
3. VS Code 内のタブで Simple Browser が開く

## メニュー表示条件

コンテキストメニューは以下の条件で表示されます。

- `resourceScheme == file`
- 拡張子が `.html` または `.htm`

## 動作仕様

- コマンド ID: `openInIntegratedBrowser.open`
- `vscode.open` に `simpleBrowser.show` を指定して表示
- URL ではなくファイルパスを Simple Browser で表示

## トラブルシュート

- メニューが表示されない
  - 対象が `.html` / `.htm` か確認
  - リモート環境や仮想スキーム上のファイルではないか確認
- F5 で起動できない
  - `npm install` 実行済みか確認
  - `package.json` の `engines.vscode` が利用中バージョンに合っているか確認

## 備考

- 本拡張は Musuhi リポジトリ内の補助ツールです。
- 将来的に Markdown / JSON など対象拡張子の追加が可能です（`package.json` の `menus.explorer/context.when` を更新）。
