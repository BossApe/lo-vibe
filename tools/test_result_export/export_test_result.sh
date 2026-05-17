#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Usage:
  export_test_result.sh [--template] <type> <target-id> [command ...]

Types:
  unit internal external acceptance api e2e security performance regression

Examples:
  export_test_result.sh unit FR-001 go test ./internal/handler ./internal/service -run SystemOverview -v
  export_test_result.sh internal FR-001 go test ./test/integration -run 'SystemOverview|ヘルスチェック' -v
  export_test_result.sh --template external FR-001
  export_test_result.sh --template acceptance FR-001
EOF
}

script_dir="$(cd "$(dirname "$0")" && pwd)"
repo_root="$(cd "$script_dir/../.." && pwd)"
output_root="${OUTPUT_ROOT:-$repo_root/_document/003.設計・開発・テストフェーズ/004.テスト結果}"

template_mode=false
if [[ "${1:-}" == "--template" ]]; then
  template_mode=true
  shift
fi

if [[ $# -lt 2 ]]; then
  usage >&2
  exit 1
fi

test_type="$1"
target_id="$2"
shift 2

dir_name_for_type() {
  case "$1" in
    unit) echo "001.単体テスト" ;;
    internal) echo "002.内部結合テスト" ;;
    external) echo "003.外部結合テスト" ;;
    acceptance) echo "004.受け入れテスト" ;;
    api) echo "005.APIテスト" ;;
    e2e) echo "006.E2Eテスト" ;;
    security) echo "007.セキュリティテスト" ;;
    performance) echo "008.性能テスト" ;;
    regression) echo "009.回帰テスト" ;;
    *) return 1 ;;
  esac
}

label_for_type() {
  case "$1" in
    unit) echo "単体テスト" ;;
    internal) echo "内部結合テスト" ;;
    external) echo "外部結合テスト" ;;
    acceptance) echo "受け入れテスト" ;;
    api) echo "APIテスト" ;;
    e2e) echo "E2Eテスト" ;;
    security) echo "セキュリティテスト" ;;
    performance) echo "性能テスト" ;;
    regression) echo "回帰テスト" ;;
    *) return 1 ;;
  esac
}

validate_target_id() {
  local type="$1"
  local id="$2"
  case "$type" in
    unit|internal|external|api|acceptance)
      [[ "$id" =~ ^FR-[0-9]{3}$ ]]
      ;;
    e2e|security)
      [[ "$id" =~ ^SPR-[0-9]+-[0-9]+$ ]]
      ;;
    performance)
      [[ "$id" =~ ^PH-[0-9]+$ ]]
      ;;
    regression)
      [[ "$id" =~ ^MRG-[0-9]+$ ]]
      ;;
    *)
      return 1
      ;;
  esac
}

if ! dir_name="$(dir_name_for_type "$test_type")"; then
  echo "Unknown type: $test_type" >&2
  usage >&2
  exit 1
fi

if ! test_label="$(label_for_type "$test_type")"; then
  echo "Unknown type: $test_type" >&2
  exit 1
fi

if ! validate_target_id "$test_type" "$target_id"; then
  echo "Invalid target-id '$target_id' for type '$test_type'" >&2
  exit 1
fi

output_dir="$output_root/$dir_name"
output_file="$output_dir/${target_id}${test_label}結果.txt"
mkdir -p "$output_dir"

if $template_mode; then
  cat > "$output_file" <<EOF
[テンプレート作成日時] $(date '+%Y-%m-%d %H:%M:%S')
[対象ID] $target_id
[テスト種別] $test_label
[状態] 未実行
[実行コマンド] 未設定

[記録欄]
- 実行環境:
- 前提条件:
- 結果:
- 備考:
EOF
  echo "$output_file"
  exit 0
fi

if [[ $# -eq 0 ]]; then
  echo "command is required unless --template is used" >&2
  exit 1
fi

command_string="$*"
{
  printf '[実行日時] %s\n' "$(date '+%Y-%m-%d %H:%M:%S')"
  printf '[対象ID] %s\n' "$target_id"
  printf '[テスト種別] %s\n' "$test_label"
  printf '[コマンド] %s\n\n' "$command_string"
} > "$output_file"

status=0
if ! "$@" >> "$output_file" 2>&1; then
  status=$?
fi

printf '\n[終了コード] %s\n' "$status" >> "$output_file"
echo "$output_file"
exit "$status"
