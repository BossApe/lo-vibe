#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

OWNER="${OWNER:-BossApe}"
REPO="${REPO:-Musuhi}"
PROJECT_NUMBER="${PROJECT_NUMBER:-2}"
OUTPUT="${OUTPUT:-$SCRIPT_DIR/output/issue_difficulty_roadmap.html}"

usage() {
  cat <<'EOF'
Usage:
  ./generate_issue_roadmap_vegalite.sh [options]

Options:
  -o, --owner <owner>              GitHub owner (user or org). Default: BossApe
  -r, --repo <repo>                Repository name. Default: Musuhi
  -p, --project-number <number>    GitHub Project (v2) number. Default: 2
  -O, --output <path>              Output html path.
                                   Default: tools/generate_issue_roadmap_vegalite/output/issue_difficulty_roadmap.html
  -h, --help                       Show this help.

Environment variables:
  OWNER, REPO, PROJECT_NUMBER, OUTPUT

Requirements:
  - gh CLI authenticated (gh auth status)
  - python3
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    -o|--owner)
      OWNER="$2"
      shift 2
      ;;
    -r|--repo)
      REPO="$2"
      shift 2
      ;;
    -p|--project-number)
      PROJECT_NUMBER="$2"
      shift 2
      ;;
    -O|--output)
      OUTPUT="$2"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown option: $1" >&2
      usage
      exit 1
      ;;
  esac
done

mkdir -p "$(dirname "$OUTPUT")"

TMP_JSON="$(mktemp)"
trap 'rm -f "$TMP_JSON"' EXIT

GH_QUERY_USER='query($owner: String!, $number: Int!) {
  user(login: $owner) {
    projectV2(number: $number) {
      title
      items(first: 100) {
        nodes {
          content {
            ... on Issue {
              number
              title
              url
              state
              repository {
                nameWithOwner
              }
            }
          }
          fieldValues(first: 30) {
            nodes {
              ... on ProjectV2ItemFieldSingleSelectValue {
                name
                field {
                  ... on ProjectV2SingleSelectField {
                    name
                  }
                }
              }
              ... on ProjectV2ItemFieldTextValue {
                text
                field {
                  ... on ProjectV2FieldCommon {
                    name
                  }
                }
              }
              ... on ProjectV2ItemFieldIterationValue {
                title
                field {
                  ... on ProjectV2IterationField {
                    name
                  }
                }
              }
            }
          }
        }
      }
    }
  }
}'

GH_QUERY_ORG='query($owner: String!, $number: Int!) {
  organization(login: $owner) {
    projectV2(number: $number) {
      title
      items(first: 100) {
        nodes {
          content {
            ... on Issue {
              number
              title
              url
              state
              repository {
                nameWithOwner
              }
            }
          }
          fieldValues(first: 30) {
            nodes {
              ... on ProjectV2ItemFieldSingleSelectValue {
                name
                field {
                  ... on ProjectV2SingleSelectField {
                    name
                  }
                }
              }
              ... on ProjectV2ItemFieldTextValue {
                text
                field {
                  ... on ProjectV2FieldCommon {
                    name
                  }
                }
              }
              ... on ProjectV2ItemFieldIterationValue {
                title
                field {
                  ... on ProjectV2IterationField {
                    name
                  }
                }
              }
            }
          }
        }
      }
    }
  }
}'

if ! gh api graphql \
  -F owner="$OWNER" \
  -F number="$PROJECT_NUMBER" \
  -f query="$GH_QUERY_USER" > "$TMP_JSON" 2>/dev/null; then
  gh api graphql \
    -F owner="$OWNER" \
    -F number="$PROJECT_NUMBER" \
    -f query="$GH_QUERY_ORG" > "$TMP_JSON"
fi

python3 "$SCRIPT_DIR/generate_issue_roadmap_vegalite.py" \
  --input "$TMP_JSON" \
  --owner "$OWNER" \
  --repo "$REPO" \
  --project-number "$PROJECT_NUMBER" \
  --output "$OUTPUT"

echo "Generated: $OUTPUT"
