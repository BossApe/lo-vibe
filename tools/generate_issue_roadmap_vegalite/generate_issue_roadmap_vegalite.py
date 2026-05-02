#!/usr/bin/env python3
import argparse
import datetime as dt
import json
import re
import sys
from pathlib import Path

ESTIMATE_SCORE = {
    "XS": 1,
    "S": 2,
    "M": 3,
    "L": 5,
    "XL": 8,
}


def load_project(payload):
    data = payload.get("data", {})
    user_owner = data.get("user") or {}
    org_owner = data.get("organization") or {}
    return user_owner.get("projectV2") or org_owner.get("projectV2")


def field_map(field_nodes):
    mapped = {}
    for node in field_nodes:
        field = node.get("field") or {}
        field_name = field.get("name")
        if not field_name:
            continue
        if "name" in node and node.get("name") is not None:
            mapped[field_name] = node.get("name")
        elif "text" in node and node.get("text") is not None:
            mapped[field_name] = node.get("text")
        elif "title" in node and node.get("title") is not None:
            mapped[field_name] = node.get("title")
    return mapped


def normalize_row(item_node, repo_filter):
    content = item_node.get("content") or {}
    if not content or content.get("repository", {}).get("nameWithOwner") != repo_filter:
        return None

    fields = field_map(((item_node.get("fieldValues") or {}).get("nodes") or []))
    issue_no = content.get("number")
    issue_title = content.get("title")

    estimate_label = (fields.get("Estimate") or "").strip().upper()
    difficulty = ESTIMATE_SCORE.get(estimate_label, 1)

    row = {
        "id": f"#{issue_no}",
        "number": issue_no,
        "title": issue_title,
        "issue": f"#{issue_no} {issue_title}",
        "url": content.get("url"),
        "status": fields.get("Status", "Unknown"),
        "type": fields.get("Type", "Unknown"),
        "phase": fields.get("Phase", "Unknown"),
        "iteration": fields.get("Iteration", "Unknown"),
        "parent": fields.get("Parent", "No Parent"),
        "depends_on": fields.get("Depends on", ""),
        "estimate": estimate_label or "N/A",
        "difficulty": difficulty,
    }
    return row


def extract_ticket_order(title, fallback):
    m = re.search(r"TK-\d+-(\d+)", title)
    if m:
        return int(m.group(1))
    return fallback


def extract_ticket_code(title, fallback):
    m = re.search(r"(TK-\d+-\d+)", title)
    if m:
        return m.group(1)
    return fallback


def extract_parent_key(row):
    title = row.get("title", "")
    parent = (row.get("parent") or "").strip()

    if row.get("type") == "Iteration":
        m = re.search(r"(IT\d+-\d+)", title)
        if m:
            return m.group(1)
    if row.get("type") == "Phase":
        m = re.search(r"(PH\d+)", title)
        if m:
            return m.group(1)

    if parent:
        return parent

    return row.get("id", "Unknown")


def phase_iteration_sort_tuple(parent_key):
    m_it = re.match(r"IT(\d+)-(\d+)", parent_key)
    if m_it:
        return (int(m_it.group(1)), 1, int(m_it.group(2)), parent_key)

    m_ph = re.match(r"PH(\d+)", parent_key)
    if m_ph:
        return (int(m_ph.group(1)), 0, 0, parent_key)

    return (999, 2, 999, parent_key)


def build_sequential_rows(rows):
    tickets = [r for r in rows if r.get("type") == "Ticket"]

    parent_rows = [r for r in rows if r.get("type") in ("Phase", "Iteration")]
    if not tickets and not parent_rows:
        return []

    ticket_groups = {}
    for r in tickets:
        key = (r.get("parent") or "No Parent").strip() or "No Parent"
        ticket_groups.setdefault(key, []).append(r)

    parent_map = {}
    for r in parent_rows:
        key = extract_parent_key(r)
        parent_map[key] = r

    for key in ticket_groups.keys():
        if key not in parent_map:
            parent_map[key] = {
                "id": key,
                "title": f"{key} parent lane",
                "type": "Iteration",
                "phase": "Unknown",
                "iteration": "Unknown",
                "difficulty": 0,
            }

    sequential_rows = []
    row_order = 0
    sorted_parent_keys = sorted(parent_map.keys(), key=phase_iteration_sort_tuple)
    for parent_key in sorted_parent_keys:
        parent_row = parent_map[parent_key]
        phase = parent_row.get("phase", "Unknown")
        iteration = parent_row.get("iteration", "Unknown")
        items = ticket_groups.get(parent_key, [])
        items.sort(key=lambda r: (extract_ticket_order(r.get("title", ""), r.get("number", 0)), r.get("number", 0)))

        if items:
            total = sum(int(r.get("difficulty", 1)) for r in items)
        else:
            total = int(parent_row.get("difficulty", 0))

        row_order += 1
        sequential_rows.append(
            {
                "row_type": "parent",
                "row_order": row_order,
                "row_label": f"{parent_key} ({phase}/I{iteration})",
                "id": parent_key,
                "title": parent_row.get("title", f"{parent_key} parent lane"),
                "parent": parent_key,
                "depends_on": "",
                "estimate": "",
                "difficulty": total,
                "start": 0,
                "end": total,
                "mid": total / 2 if total else 0,
                "execution_order": 0,
                "url": parent_row.get("url", ""),
            }
        )

        cursor = 0
        for idx, r in enumerate(items, start=1):
            start = cursor
            end = cursor + int(r.get("difficulty", 1))
            cursor = end
            row_order += 1
            ticket_code = extract_ticket_code(r.get("title", ""), r.get("id", ""))
            sequential_rows.append(
                {
                    **r,
                    "row_type": "ticket",
                    "row_order": row_order,
                    "row_label": f"  - {ticket_code}",
                    "execution_order": idx,
                    "start": start,
                    "end": end,
                    "mid": (start + end) / 2,
                }
            )

    sequential_rows.sort(key=lambda r: r["row_order"])
    return sequential_rows


def make_spec(rows, chart_title, chart_subtitle):
    return {
        "$schema": "https://vega.github.io/schema/vega-lite/v5.json",
        "title": {"text": chart_title, "subtitle": chart_subtitle},
        "width": 1000,
        "height": {"step": 28},
        "data": {"values": rows},
        "layer": [
            {
                "transform": [{"filter": "datum.row_type === 'parent'"}],
                "mark": {"type": "bar", "cornerRadius": 5, "opacity": 0.25, "color": "#64748b"},
                "encoding": {
                    "y": {
                        "field": "row_label",
                        "type": "ordinal",
                        "title": "親子階層",
                        "sort": {"field": "row_order", "order": "ascending"},
                    },
                    "x": {
                        "field": "start",
                        "type": "quantitative",
                        "title": "実行進捗（難易度ポイント）",
                        "scale": {"zero": True},
                    },
                    "x2": {"field": "end"},
                    "tooltip": [
                        {"field": "id", "type": "nominal", "title": "チケット"},
                        {"field": "title", "type": "nominal", "title": "タイトル"},
                        {"field": "parent", "type": "nominal", "title": "親"},
                        {"field": "difficulty", "type": "quantitative", "title": "総合難易度"},
                    ],
                },
            },
            {
                "transform": [{"filter": "datum.row_type === 'ticket'"}],
                "mark": {"type": "bar", "cornerRadius": 5},
                "encoding": {
                    "y": {
                        "field": "row_label",
                        "type": "ordinal",
                        "title": "親子階層",
                        "sort": {"field": "row_order", "order": "ascending"},
                    },
                    "x": {
                        "field": "start",
                        "type": "quantitative",
                        "title": "実行進捗（難易度ポイント）",
                        "scale": {"zero": True},
                    },
                    "x2": {"field": "end"},
                    "color": {
                        "field": "execution_order",
                        "type": "ordinal",
                        "title": "実行順",
                    },
                    "tooltip": [
                        {"field": "id", "type": "nominal", "title": "チケット"},
                        {"field": "title", "type": "nominal", "title": "タイトル"},
                        {"field": "parent", "type": "nominal", "title": "親"},
                        {"field": "execution_order", "type": "ordinal", "title": "実行順"},
                        {"field": "depends_on", "type": "nominal", "title": "依存"},
                        {"field": "estimate", "type": "nominal", "title": "見積"},
                        {"field": "difficulty", "type": "quantitative", "title": "難易度"},
                        {"field": "start", "type": "quantitative", "title": "開始"},
                        {"field": "end", "type": "quantitative", "title": "終了"},
                    ],
                    "href": {"field": "url", "type": "nominal"},
                },
            },
            {
                "transform": [{"filter": "datum.row_type === 'ticket'"}],
                "mark": {"type": "text", "baseline": "middle", "fontSize": 10, "color": "#111827"},
                "encoding": {
                    "y": {
                        "field": "row_label",
                        "type": "ordinal",
                        "sort": {"field": "row_order", "order": "ascending"},
                    },
                    "x": {"field": "mid", "type": "quantitative"},
                    "text": {"field": "id", "type": "nominal"},
                },
            },
        ],
        "config": {
            "view": {"stroke": None},
            "axis": {"labelFontSize": 11, "titleFontSize": 12},
            "legend": {"labelFontSize": 11, "titleFontSize": 12},
        },
    }


def write_html(path: Path, owner: str, repo: str, project_number: int, spec: dict, count: int):
    generated_at = dt.datetime.now(dt.timezone.utc).astimezone().isoformat(timespec="seconds")
    spec_json = json.dumps(spec, ensure_ascii=False)
    html = f"""<!doctype html>
<html lang=\"ja\">
<head>
    <meta charset=\"utf-8\" />
    <meta name=\"viewport\" content=\"width=device-width, initial-scale=1\" />
    <title>Issue難易度ロードマップ</title>
    <style>
        body {{
            font-family: -apple-system, BlinkMacSystemFont, \"Segoe UI\", sans-serif;
            margin: 24px;
            color: #1f2937;
            background: #f8fafc;
        }}
        .meta {{
            margin-bottom: 16px;
            font-size: 14px;
            color: #374151;
        }}
        #vis {{
            max-width: 1200px;
            background: #ffffff;
            border: 1px solid #e5e7eb;
            border-radius: 12px;
            padding: 12px;
        }}
    </style>
    <script src=\"https://cdn.jsdelivr.net/npm/vega@5\"></script>
    <script src=\"https://cdn.jsdelivr.net/npm/vega-lite@5\"></script>
    <script src=\"https://cdn.jsdelivr.net/npm/vega-embed@6\"></script>
</head>
<body>
    <h1>Issue難易度ロードマップ</h1>
    <div class=\"meta\">生成日時: {generated_at}</div>
    <div class=\"meta\">ソース: GitHub Project #{project_number} ({owner}/{repo}) / アイテム数: {count}</div>
    <div class="meta">難易度マッピング: XS=1, S=2, M=3, L=5, XL=8</div>
    <div id=\"vis\"></div>
    <script>
        const spec = {spec_json};
        vegaEmbed('#vis', spec, {{ actions: true }}).catch(console.error);
    </script>
</body>
</html>
"""
    path.write_text(html, encoding="utf-8")


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--input", required=True)
    parser.add_argument("--owner", required=True)
    parser.add_argument("--repo", required=True)
    parser.add_argument("--project-number", required=True, type=int)
    parser.add_argument("--output", required=True)
    args = parser.parse_args()

    payload = json.loads(Path(args.input).read_text(encoding="utf-8"))
    project = load_project(payload)
    if not project:
        print("ERROR: projectV2 not found for the specified owner/number.", file=sys.stderr)
        sys.exit(2)

    repo_filter = f"{args.owner}/{args.repo}"
    rows = []
    for item in ((project.get("items") or {}).get("nodes") or []):
        row = normalize_row(item, repo_filter)
        if row:
            rows.append(row)

    if not rows:
        print("ERROR: no issue rows found for the specified repository in this project.", file=sys.stderr)
        sys.exit(3)

    sequential_rows = build_sequential_rows(rows)
    if not sequential_rows:
        print("ERROR: no ticket rows found. Check Type/Estimate/Parent fields.", file=sys.stderr)
        sys.exit(4)

    title = f"{args.repo} チケット実行ロードマップ"
    subtitle = "Y軸はGitHub Projectsの親子ビューに模した階層構造、順序付けされたチケット帯を表示"
    spec = make_spec(sequential_rows, title, subtitle)

    output_path = Path(args.output)
    output_path.parent.mkdir(parents=True, exist_ok=True)
    write_html(output_path, args.owner, args.repo, args.project_number, spec, len(sequential_rows))


if __name__ == "__main__":
    main()
