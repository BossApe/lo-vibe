# generate_issue_roadmap_vegalite

GitHub Projects の Issue 一覧を `gh api` で取得し、Vega-Lite 形式の難易度ロードマップ HTML を自動生成するツールです。

## 目的

- GitHub Projects のタスクを、日付軸ではなく難易度軸で可視化する
- チケットを並行実行せず、親子関係を維持したまま実施順に帯が横へ伸びる形で可視化する
- `Estimate` フィールドを Difficulty point に変換し、実施順の累積進捗を横棒で比較可能にする

## 前提条件

- GitHub CLI (`gh`) が利用可能であること
- `gh auth login` 済みで、対象 Project を参照できること
- `python3` が利用可能であること
- `npm` が利用可能であること

## インストール

```bash
cd tools/generate_issue_roadmap_vegalite
npm install
chmod +x generate_issue_roadmap_vegalite.sh
```

`npm install` により、Vega-Lite 依存がこのディレクトリへインストールされる。

## 使用方法

```bash
cd tools/generate_issue_roadmap_vegalite
./generate_issue_roadmap_vegalite.sh \
  --owner BossApe \
  --repo Musuhi \
  --project-number 2 \
  --output output/issue_difficulty_roadmap.html
```

既定値を使う場合:

```bash
cd tools/generate_issue_roadmap_vegalite
./generate_issue_roadmap_vegalite.sh
```

## 入出力

- 入力: GitHub Project v2 アイテム（Issue）
- 出力: Vega-Lite をブラウザ描画する HTML
  - 既定出力先: `tools/generate_issue_roadmap_vegalite/output/issue_difficulty_roadmap.html`

描画ルール:

- `Type=Ticket` のみを対象に描画する
- レーンは `Parent` + `Phase/Iteration` で分ける
- 同一レーン内は `TK-x-n` の `n` 昇順で実施順を決める
- 帯は `start` から `end`（Difficulty 累積）で連結され、並行帯は作らない

## 難易度マッピング

- `XS` = 1
- `S` = 2
- `M` = 3
- `L` = 5
- `XL` = 8

`Estimate` が未設定の Issue は `1` 扱いで出力する。

## 注意事項

- GraphQL は Project 内から指定リポジトリの Issue を抽出する
- 100件を超えるアイテムを扱う場合は、クエリにページング追加が必要
- 生成される HTML はブラウザで直接表示できる
