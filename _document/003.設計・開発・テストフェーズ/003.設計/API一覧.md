# Musuhi API一覧

[上位](README.md)

<details>
<summary>目次（クリックで展開）</summary>

- [1. 目的](#1-目的)
- [2. API一覧（全体）](#2-api一覧全体)
- [3. 機能別 API 仕様書リンク](#3-機能別-api-仕様書リンク)
- [4. エラーコード共通表](#4-エラーコード共通表)
- [5. 更新ルール](#5-更新ルール)

</details>

---

## 1. 目的

本書は Musuhi 全体で利用する API の ID・エンドポイント・対応機能・実装状況を一元管理する。
また、各機能の API 仕様書への参照点を統一し、設計と実装の追跡性を担保する。

---

## 2. API一覧（全体）

| API ID | メソッド | パス | 機能（FR） | 実装状況 | API仕様書リンク |
| --- | --- | --- | --- | --- | --- |
| API-FR001-POST-001 | POST | `/api/v1/system-overviews` | FR-001 システム概要入力・保存 | 実装済み | [FR-001 API仕様書 3.1](001.FR-001-システム概要入力/003.API仕様書.md#31-post-apiv1system-overviews--システム概要を保存) |
| API-FR001-GET-001 | GET | `/api/v1/system-overviews/{id}` | FR-001 システム概要入力・保存 | 実装済み | [FR-001 API仕様書 3.2](001.FR-001-システム概要入力/003.API仕様書.md#32-get-apiv1system-overviewsid--システム概要を取得) |
| API-FR002-POST-001 | POST | `/api/v1/projects/extract-features` | FR-002 機能抽出 | 実装済み | [FR-002 API仕様書 1](002.FR-002-機能抽出・ディレクトリ作成/002.API仕様書.md#1-post-apiv1projectsextract-features) |
| API-FR002-POST-002 | POST | `/api/v1/projects/suggest-name` | FR-002 プロジェクト名候補生成 | 実装済み | [FR-002 API仕様書 2](002.FR-002-機能抽出・ディレクトリ作成/002.API仕様書.md#2-post-apiv1projectssuggest-name) |
| API-FR002-POST-003 | POST | `/api/v1/projects/init-directory` | FR-002 初期ディレクトリ作成 | 実装済み | [FR-002 API仕様書 3](002.FR-002-機能抽出・ディレクトリ作成/002.API仕様書.md#3-post-apiv1projectsinit-directory) |
| API-FR003-POST-001 | POST | `/api/v1/projects/with-external` | FR-003 GitHubリポジトリ作成・push | 未実装 | [機能業務フロー図（FR-003）](001.FR-001-システム概要入力/001.機能業務フロー図.md) |
| API-FR004-POST-001 | POST | `/api/v1/projects/github-projects` | FR-004 GitHub Projects作成 | 未実装 | [機能業務フロー図（FR-004）](001.FR-001-システム概要入力/001.機能業務フロー図.md) |

---

## 3. 機能別 API 仕様書リンク

| 機能（FR） | API仕様書 | 補足 |
| --- | --- | --- |
| FR-001 | [FR-001 API仕様書](001.FR-001-システム概要入力/003.API仕様書.md) | 詳細定義あり（実装済み） |
| FR-002 | [FR-002 API仕様書](002.FR-002-機能抽出・ディレクトリ作成/002.API仕様書.md) | TK1-1-2 実装済み |
| FR-003 | [機能業務フロー図](001.FR-001-システム概要入力/001.機能業務フロー図.md) | API仕様書は未作成 |
| FR-004 | [機能業務フロー図](001.FR-001-システム概要入力/001.機能業務フロー図.md) | API仕様書は未作成 |

---

## 4. エラーコード共通表

FR-001 API仕様書で定義済みの共通エラーコード。

| エラーコード | HTTP | 意味 |
| --- | --- | --- |
| `BAD_REQUEST` | 400 | リクエスト形式不正 |
| `VALIDATION_ERROR` | 422 | バリデーション違反 |
| `NOT_FOUND` | 404 | リソース未存在 |
| `INTERNAL_ERROR` | 500 | サーバー内部エラー |

参照: [FR-001 API仕様書](001.FR-001-システム概要入力/003.API仕様書.md)

---

## 5. 更新ルール

- API 追加時は `API ID` を採番し本一覧へ追加する。
- 実装着手時に「実装状況」を更新し、API 仕様書作成後にリンク先を差し替える。
- パス変更時は API 一覧・機能業務フロー図・実装コードのルーティングを同一コミットで更新する。