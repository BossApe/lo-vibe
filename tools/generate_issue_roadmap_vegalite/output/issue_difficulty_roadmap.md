# Issue Difficulty Roadmap (Vega-Lite)

- Generated at: 2026-05-02T23:59:59+09:00
- Source: GitHub Project #2 (BossApe/Musuhi)
- Items: 16

Estimate to difficulty point mapping:
- XS=1, S=2, M=3, L=5, XL=8

```vega-lite
{
  "$schema": "https://vega.github.io/schema/vega-lite/v5.json",
  "title": {
    "text": "Musuhi Task Difficulty Roadmap",
    "subtitle": "Horizontal axis is difficulty points, not date"
  },
  "width": 1000,
  "height": {
    "step": 28
  },
  "data": {
    "values": [
      {
        "id": "#10",
        "title": "IT0-1: [FR-001] ユーザー登録・認証基盤構築",
        "issue": "#10 IT0-1: [FR-001] ユーザー登録・認証基盤構築",
        "url": "https://github.com/BossApe/Musuhi/issues/10",
        "status": "Todo",
        "type": "Iteration",
        "phase": "Phase 0",
        "iteration": "1",
        "estimate": "L",
        "difficulty": 5
      },
      {
        "id": "#16",
        "title": "TK-1-1: [api] ユーザーモデル・DBスキーマ設計",
        "issue": "#16 TK-1-1: [api] ユーザーモデル・DBスキーマ設計",
        "url": "https://github.com/BossApe/Musuhi/issues/16",
        "status": "Todo",
        "type": "Ticket",
        "phase": "Phase 0",
        "iteration": "1",
        "estimate": "L",
        "difficulty": 5
      },
      {
        "id": "#22",
        "title": "TK-1-7: [ui] ログイン・登録画面実装（SvelteKit）",
        "issue": "#22 TK-1-7: [ui] ログイン・登録画面実装（SvelteKit）",
        "url": "https://github.com/BossApe/Musuhi/issues/22",
        "status": "Todo",
        "type": "Ticket",
        "phase": "Phase 0",
        "iteration": "1",
        "estimate": "L",
        "difficulty": 5
      },
      {
        "id": "#11",
        "title": "IT0-2: [FR-002] 書籍登録・蔵書管理（基本CRUD）",
        "issue": "#11 IT0-2: [FR-002] 書籍登録・蔵書管理（基本CRUD）",
        "url": "https://github.com/BossApe/Musuhi/issues/11",
        "status": "Todo",
        "type": "Iteration",
        "phase": "Phase 0",
        "iteration": "2",
        "estimate": "L",
        "difficulty": 5
      },
      {
        "id": "#12",
        "title": "IT0-3: [FR-003] 検索・フィルタ機能（基本）",
        "issue": "#12 IT0-3: [FR-003] 検索・フィルタ機能（基本）",
        "url": "https://github.com/BossApe/Musuhi/issues/12",
        "status": "Todo",
        "type": "Iteration",
        "phase": "Phase 0",
        "iteration": "3",
        "estimate": "L",
        "difficulty": 5
      },
      {
        "id": "#7",
        "title": "PH0: Phase 0 (MVP)",
        "issue": "#7 PH0: Phase 0 (MVP)",
        "url": "https://github.com/BossApe/Musuhi/issues/7",
        "status": "Todo",
        "type": "Phase",
        "phase": "Phase 0",
        "iteration": "Unknown",
        "estimate": "L",
        "difficulty": 5
      },
      {
        "id": "#13",
        "title": "IT1-1: [FR-004] 貸出・返却管理",
        "issue": "#13 IT1-1: [FR-004] 貸出・返却管理",
        "url": "https://github.com/BossApe/Musuhi/issues/13",
        "status": "Todo",
        "type": "Iteration",
        "phase": "Phase 1",
        "iteration": "4",
        "estimate": "L",
        "difficulty": 5
      },
      {
        "id": "#14",
        "title": "IT1-2: [FR-005] 通知・リマインダー機能",
        "issue": "#14 IT1-2: [FR-005] 通知・リマインダー機能",
        "url": "https://github.com/BossApe/Musuhi/issues/14",
        "status": "Todo",
        "type": "Iteration",
        "phase": "Phase 1",
        "iteration": "5",
        "estimate": "L",
        "difficulty": 5
      },
      {
        "id": "#15",
        "title": "IT1-3: [FR-006] レポート・統計ダッシュボード",
        "issue": "#15 IT1-3: [FR-006] レポート・統計ダッシュボード",
        "url": "https://github.com/BossApe/Musuhi/issues/15",
        "status": "Todo",
        "type": "Iteration",
        "phase": "Phase 1",
        "iteration": "6",
        "estimate": "L",
        "difficulty": 5
      },
      {
        "id": "#8",
        "title": "PH1: Phase 1 (コア拡張)",
        "issue": "#8 PH1: Phase 1 (コア拡張)",
        "url": "https://github.com/BossApe/Musuhi/issues/8",
        "status": "Todo",
        "type": "Phase",
        "phase": "Phase 1",
        "iteration": "Unknown",
        "estimate": "L",
        "difficulty": 5
      },
      {
        "id": "#9",
        "title": "PH2: Phase 2 (統合・最適化)",
        "issue": "#9 PH2: Phase 2 (統合・最適化)",
        "url": "https://github.com/BossApe/Musuhi/issues/9",
        "status": "Todo",
        "type": "Phase",
        "phase": "Phase 2",
        "iteration": "Unknown",
        "estimate": "L",
        "difficulty": 5
      },
      {
        "id": "#17",
        "title": "TK-1-2: [api] 認証エンドポイント実装（登録/ログイン/ログアウト）",
        "issue": "#17 TK-1-2: [api] 認証エンドポイント実装（登録/ログイン/ログアウト）",
        "url": "https://github.com/BossApe/Musuhi/issues/17",
        "status": "Todo",
        "type": "Ticket",
        "phase": "Phase 0",
        "iteration": "1",
        "estimate": "M",
        "difficulty": 3
      },
      {
        "id": "#18",
        "title": "TK-1-3: [api] JWTトークン管理・ミドルウェア実装",
        "issue": "#18 TK-1-3: [api] JWTトークン管理・ミドルウェア実装",
        "url": "https://github.com/BossApe/Musuhi/issues/18",
        "status": "Todo",
        "type": "Ticket",
        "phase": "Phase 0",
        "iteration": "1",
        "estimate": "M",
        "difficulty": 3
      },
      {
        "id": "#19",
        "title": "TK-1-4: [infra] PostgreSQL初期セットアップ・マイグレーション基盤",
        "issue": "#19 TK-1-4: [infra] PostgreSQL初期セットアップ・マイグレーション基盤",
        "url": "https://github.com/BossApe/Musuhi/issues/19",
        "status": "Todo",
        "type": "Ticket",
        "phase": "Phase 0",
        "iteration": "1",
        "estimate": "M",
        "difficulty": 3
      },
      {
        "id": "#20",
        "title": "TK-1-5: [infra] Docker Compose環境整備（api/db/cache）",
        "issue": "#20 TK-1-5: [infra] Docker Compose環境整備（api/db/cache）",
        "url": "https://github.com/BossApe/Musuhi/issues/20",
        "status": "Todo",
        "type": "Ticket",
        "phase": "Phase 0",
        "iteration": "1",
        "estimate": "M",
        "difficulty": 3
      },
      {
        "id": "#21",
        "title": "TK-1-6: [test] 認証APIの単体・統合テスト実装",
        "issue": "#21 TK-1-6: [test] 認証APIの単体・統合テスト実装",
        "url": "https://github.com/BossApe/Musuhi/issues/21",
        "status": "Todo",
        "type": "Ticket",
        "phase": "Phase 0",
        "iteration": "1",
        "estimate": "M",
        "difficulty": 3
      }
    ]
  },
  "transform": [
    {
      "calculate": "datum.phase + ' / ' + datum.iteration",
      "as": "lane"
    }
  ],
  "mark": {
    "type": "bar",
    "cornerRadiusEnd": 4
  },
  "encoding": {
    "y": {
      "field": "issue",
      "type": "ordinal",
      "sort": "-x",
      "title": "Issue"
    },
    "x": {
      "field": "difficulty",
      "type": "quantitative",
      "title": "Difficulty (points)",
      "scale": {
        "zero": true
      }
    },
    "color": {
      "field": "lane",
      "type": "nominal",
      "title": "Phase / Iteration"
    },
    "tooltip": [
      {
        "field": "id",
        "type": "nominal",
        "title": "Issue"
      },
      {
        "field": "title",
        "type": "nominal",
        "title": "Title"
      },
      {
        "field": "type",
        "type": "nominal",
        "title": "Type"
      },
      {
        "field": "status",
        "type": "nominal",
        "title": "Status"
      },
      {
        "field": "phase",
        "type": "nominal",
        "title": "Phase"
      },
      {
        "field": "iteration",
        "type": "nominal",
        "title": "Iteration"
      },
      {
        "field": "estimate",
        "type": "nominal",
        "title": "Estimate"
      },
      {
        "field": "difficulty",
        "type": "quantitative",
        "title": "Difficulty"
      }
    ],
    "href": {
      "field": "url",
      "type": "nominal"
    }
  },
  "config": {
    "view": {
      "stroke": null
    },
    "axis": {
      "labelFontSize": 11,
      "titleFontSize": 12
    },
    "legend": {
      "labelFontSize": 11,
      "titleFontSize": 12
    }
  }
}
```
