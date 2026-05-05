#!/bin/bash
# GitHub Project fields bulk update script.
set -euo pipefail

PROJ_NUM=2
PROJ_ID=PVT_kwHOAO_tRc4BWapm
OWNER=BossApe

# --- Field IDs ---
F_TYPE=PVTSSF_lAHOAO_tRc4BWapmzhRvLmI
F_PHASE=PVTSSF_lAHOAO_tRc4BWapmzhRvLmU
F_ITER=PVTSSF_lAHOAO_tRc4BWapmzhRvLoA
F_SVC=PVTSSF_lAHOAO_tRc4BWapmzhRvLoE
F_PRI=PVTSSF_lAHOAO_tRc4BWapmzhRvLo8
F_EST=PVTSSF_lAHOAO_tRc4BWapmzhRvLqo
F_STATUS=PVTSSF_lAHOAO_tRc4BWapmzhRvJ2s
F_PARENT=PVTF_lAHOAO_tRc4BWapmzhRvLqs
F_DEPS=PVTF_lAHOAO_tRc4BWapmzhRvLrk

# --- Type option IDs ---
T_PHASE=94ff9c39
T_ITER=e58b0dc4
T_TICKET=b9a6e6c6

# --- Phase option IDs ---
P_PH0=1ca7cff1
P_PH1=7ead2372
P_PH2=f44b5b26

# --- Iteration option IDs ---
I_1=96080651
I_2=ab950554
I_3=b2e16053
I_4=a6313b8a
I_5=42d057ec
I_6=b1dae72f

# --- Service option IDs ---
S_API=a864fb8f
S_UI=7d002c35
S_INFRA=3dc39355
S_TEST=0ad5bb30
S_ALL=de7d7fc1

# --- Priority option IDs ---
PRI_P0=a4b0add0
PRI_P1=d7b2ed15
PRI_P2=4d8b18a1

# --- Estimate option IDs ---
E_XS=3566e673
E_S=10c2ad3f
E_M=82699b4f
E_L=616d014e
E_XL=998af411

# --- Status option IDs ---
ST_TODO=f75ad846
ST_IN_PROGRESS=47fc9ee4
ST_DONE=98236657

# --- Item IDs (issue# → task) ---
ID_PH0=PVTI_lAHOAO_tRc4BWapmzgrn31U    # #7  PH0
ID_PH1=PVTI_lAHOAO_tRc4BWapmzgrn31c    # #8  PH1
ID_PH2=PVTI_lAHOAO_tRc4BWapmzgrn31g    # #9  PH2
ID_IT01=PVTI_lAHOAO_tRc4BWapmzgrn31w   # #10 IT0-1
ID_IT02=PVTI_lAHOAO_tRc4BWapmzgrn32I   # #11 IT0-2
ID_IT03=PVTI_lAHOAO_tRc4BWapmzgrn32c   # #12 IT0-3
ID_IT04=PVTI_lAHOAO_tRc4BWapmzgrn32k   # #13 IT0-4
ID_IT11=PVTI_lAHOAO_tRc4BWapmzgrn32w   # #14 IT1-1
ID_IT12=PVTI_lAHOAO_tRc4BWapmzgrn320   # #15 IT1-2
ID_TK011=PVTI_lAHOAO_tRc4BWapmzgrn4B8  # #16 TK0-1-1
ID_TK012=PVTI_lAHOAO_tRc4BWapmzgrn4CI  # #17 TK0-1-2
ID_TK021=PVTI_lAHOAO_tRc4BWapmzgrn4CQ  # #18 TK0-2-1
ID_TK022=PVTI_lAHOAO_tRc4BWapmzgrn4Cc  # #19 TK0-2-2
ID_TK031=PVTI_lAHOAO_tRc4BWapmzgrn4Cg  # #20 TK0-3-1
ID_TK041=PVTI_lAHOAO_tRc4BWapmzgrn4Cs  # #21 TK0-4-1
ID_TK042=PVTI_lAHOAO_tRc4BWapmzgrn4DA  # #22 TK0-4-2
ID_IT13=PVTI_lAHOAO_tRc4BWapmzgr0eJ8   # #23 IT1-3
ID_IT14=PVTI_lAHOAO_tRc4BWapmzgr0eLA   # #24 IT1-4
ID_TK111=PVTI_lAHOAO_tRc4BWapmzgr0eLY  # #25 TK1-1-1
ID_TK112=PVTI_lAHOAO_tRc4BWapmzgr0eL8  # #26 TK1-1-2
ID_TK113=PVTI_lAHOAO_tRc4BWapmzgr0eMc  # #27 TK1-1-3
ID_TK114=PVTI_lAHOAO_tRc4BWapmzgr0eOM  # #28 TK1-1-4
ID_TK115=PVTI_lAHOAO_tRc4BWapmzgr0eO4  # #29 TK1-1-5
ID_TK116=PVTI_lAHOAO_tRc4BWapmzgr0ePk  # #30 TK1-1-6

set_ss() { gh project item-edit --project-id "$PROJ_ID" --id "$1" --field-id "$2" --single-select-option-id "$3"; }
set_tx() { gh project item-edit --project-id "$PROJ_ID" --id "$1" --field-id "$2" --text "$3"; }

if ! command -v gh >/dev/null 2>&1; then
	echo "ERROR: gh コマンドが見つかりません。GitHub CLI をインストールしてください。" >&2
	exit 1
fi

if [ -z "$PROJ_ID" ] || [ -z "$OWNER" ] || [ "$PROJ_NUM" -le 0 ]; then
	echo "ERROR: PROJ_ID / OWNER / PROJ_NUM の設定が不正です。" >&2
	exit 1
fi

# ===== Phase =====

echo "=== PH0 ==="
set_ss "$ID_PH0" "$F_TYPE"   "$T_PHASE"
set_ss "$ID_PH0" "$F_PHASE"  "$P_PH0"
set_ss "$ID_PH0" "$F_PRI"    "$PRI_P0"
set_ss "$ID_PH0" "$F_EST"    "$E_L"
set_ss "$ID_PH0" "$F_STATUS" "$ST_DONE"

echo "=== PH1 ==="
set_ss "$ID_PH1" "$F_TYPE"   "$T_PHASE"
set_ss "$ID_PH1" "$F_PHASE"  "$P_PH1"
set_ss "$ID_PH1" "$F_PRI"    "$PRI_P1"
set_ss "$ID_PH1" "$F_EST"    "$E_L"
set_ss "$ID_PH1" "$F_STATUS" "$ST_TODO"
set_tx "$ID_PH1" "$F_DEPS"   "PH0"

echo "=== PH2 ==="
set_ss "$ID_PH2" "$F_TYPE"   "$T_PHASE"
set_ss "$ID_PH2" "$F_PHASE"  "$P_PH2"
set_ss "$ID_PH2" "$F_PRI"    "$PRI_P1"
set_ss "$ID_PH2" "$F_EST"    "$E_L"
set_ss "$ID_PH2" "$F_STATUS" "$ST_TODO"
set_tx "$ID_PH2" "$F_DEPS"   "PH1"

# ===== Phase 0 Iterations =====

echo "=== IT0-1 ==="
set_ss "$ID_IT01" "$F_TYPE"   "$T_ITER"
set_ss "$ID_IT01" "$F_PHASE"  "$P_PH0"
set_ss "$ID_IT01" "$F_ITER"   "$I_1"
set_ss "$ID_IT01" "$F_PRI"    "$PRI_P0"
set_ss "$ID_IT01" "$F_EST"    "$E_M"
set_ss "$ID_IT01" "$F_STATUS" "$ST_DONE"
set_tx "$ID_IT01" "$F_PARENT" "PH0"

echo "=== IT0-2 ==="
set_ss "$ID_IT02" "$F_TYPE"   "$T_ITER"
set_ss "$ID_IT02" "$F_PHASE"  "$P_PH0"
set_ss "$ID_IT02" "$F_ITER"   "$I_2"
set_ss "$ID_IT02" "$F_PRI"    "$PRI_P0"
set_ss "$ID_IT02" "$F_EST"    "$E_M"
set_ss "$ID_IT02" "$F_STATUS" "$ST_DONE"
set_tx "$ID_IT02" "$F_PARENT" "PH0"
set_tx "$ID_IT02" "$F_DEPS"   "IT0-1"

echo "=== IT0-3 ==="
set_ss "$ID_IT03" "$F_TYPE"   "$T_ITER"
set_ss "$ID_IT03" "$F_PHASE"  "$P_PH0"
set_ss "$ID_IT03" "$F_ITER"   "$I_3"
set_ss "$ID_IT03" "$F_PRI"    "$PRI_P0"
set_ss "$ID_IT03" "$F_EST"    "$E_M"
set_ss "$ID_IT03" "$F_STATUS" "$ST_DONE"
set_tx "$ID_IT03" "$F_PARENT" "PH0"
set_tx "$ID_IT03" "$F_DEPS"   "IT0-2"

echo "=== IT0-4 ==="
set_ss "$ID_IT04" "$F_TYPE"   "$T_ITER"
set_ss "$ID_IT04" "$F_PHASE"  "$P_PH0"
set_ss "$ID_IT04" "$F_ITER"   "$I_4"
set_ss "$ID_IT04" "$F_PRI"    "$PRI_P0"
set_ss "$ID_IT04" "$F_EST"    "$E_M"
set_ss "$ID_IT04" "$F_STATUS" "$ST_DONE"
set_tx "$ID_IT04" "$F_PARENT" "PH0"
set_tx "$ID_IT04" "$F_DEPS"   "IT0-3"

# ===== Phase 1 Iterations =====

echo "=== IT1-1 ==="
set_ss "$ID_IT11" "$F_TYPE"   "$T_ITER"
set_ss "$ID_IT11" "$F_PHASE"  "$P_PH1"
set_ss "$ID_IT11" "$F_ITER"   "$I_1"
set_ss "$ID_IT11" "$F_PRI"    "$PRI_P1"
set_ss "$ID_IT11" "$F_EST"    "$E_L"
set_ss "$ID_IT11" "$F_STATUS" "$ST_TODO"
set_tx "$ID_IT11" "$F_PARENT" "PH1"
set_tx "$ID_IT11" "$F_DEPS"   "PH0"

echo "=== IT1-2 ==="
set_ss "$ID_IT12" "$F_TYPE"   "$T_ITER"
set_ss "$ID_IT12" "$F_PHASE"  "$P_PH1"
set_ss "$ID_IT12" "$F_ITER"   "$I_2"
set_ss "$ID_IT12" "$F_PRI"    "$PRI_P1"
set_ss "$ID_IT12" "$F_EST"    "$E_L"
set_ss "$ID_IT12" "$F_STATUS" "$ST_TODO"
set_tx "$ID_IT12" "$F_PARENT" "PH1"
set_tx "$ID_IT12" "$F_DEPS"   "IT1-1"

echo "=== IT1-3 ==="
set_ss "$ID_IT13" "$F_TYPE"   "$T_ITER"
set_ss "$ID_IT13" "$F_PHASE"  "$P_PH1"
set_ss "$ID_IT13" "$F_ITER"   "$I_3"
set_ss "$ID_IT13" "$F_PRI"    "$PRI_P1"
set_ss "$ID_IT13" "$F_EST"    "$E_L"
set_ss "$ID_IT13" "$F_STATUS" "$ST_TODO"
set_tx "$ID_IT13" "$F_PARENT" "PH1"
set_tx "$ID_IT13" "$F_DEPS"   "IT1-2"

echo "=== IT1-4 ==="
set_ss "$ID_IT14" "$F_TYPE"   "$T_ITER"
set_ss "$ID_IT14" "$F_PHASE"  "$P_PH1"
set_ss "$ID_IT14" "$F_ITER"   "$I_4"
set_ss "$ID_IT14" "$F_PRI"    "$PRI_P1"
set_ss "$ID_IT14" "$F_EST"    "$E_L"
set_ss "$ID_IT14" "$F_STATUS" "$ST_TODO"
set_tx "$ID_IT14" "$F_PARENT" "PH1"
set_tx "$ID_IT14" "$F_DEPS"   "IT1-3"

# ===== Phase 0 Tickets =====

echo "=== TK0-1-1 ==="
set_ss "$ID_TK011" "$F_TYPE"   "$T_TICKET"
set_ss "$ID_TK011" "$F_PHASE"  "$P_PH0"
set_ss "$ID_TK011" "$F_ITER"   "$I_1"
set_ss "$ID_TK011" "$F_SVC"    "$S_ALL"
set_ss "$ID_TK011" "$F_PRI"    "$PRI_P0"
set_ss "$ID_TK011" "$F_EST"    "$E_M"
set_ss "$ID_TK011" "$F_STATUS" "$ST_DONE"
set_tx "$ID_TK011" "$F_PARENT" "IT0-1"

echo "=== TK0-1-2 ==="
set_ss "$ID_TK012" "$F_TYPE"   "$T_TICKET"
set_ss "$ID_TK012" "$F_PHASE"  "$P_PH0"
set_ss "$ID_TK012" "$F_ITER"   "$I_1"
set_ss "$ID_TK012" "$F_SVC"    "$S_ALL"
set_ss "$ID_TK012" "$F_PRI"    "$PRI_P0"
set_ss "$ID_TK012" "$F_EST"    "$E_S"
set_ss "$ID_TK012" "$F_STATUS" "$ST_DONE"
set_tx "$ID_TK012" "$F_PARENT" "IT0-1"
set_tx "$ID_TK012" "$F_DEPS"   "TK0-1-1"

echo "=== TK0-2-1 ==="
set_ss "$ID_TK021" "$F_TYPE"   "$T_TICKET"
set_ss "$ID_TK021" "$F_PHASE"  "$P_PH0"
set_ss "$ID_TK021" "$F_ITER"   "$I_2"
set_ss "$ID_TK021" "$F_SVC"    "$S_ALL"
set_ss "$ID_TK021" "$F_PRI"    "$PRI_P0"
set_ss "$ID_TK021" "$F_EST"    "$E_M"
set_ss "$ID_TK021" "$F_STATUS" "$ST_DONE"
set_tx "$ID_TK021" "$F_PARENT" "IT0-2"
set_tx "$ID_TK021" "$F_DEPS"   "TK0-1-2"

echo "=== TK0-2-2 ==="
set_ss "$ID_TK022" "$F_TYPE"   "$T_TICKET"
set_ss "$ID_TK022" "$F_PHASE"  "$P_PH0"
set_ss "$ID_TK022" "$F_ITER"   "$I_2"
set_ss "$ID_TK022" "$F_SVC"    "$S_ALL"
set_ss "$ID_TK022" "$F_PRI"    "$PRI_P0"
set_ss "$ID_TK022" "$F_EST"    "$E_S"
set_ss "$ID_TK022" "$F_STATUS" "$ST_DONE"
set_tx "$ID_TK022" "$F_PARENT" "IT0-2"
set_tx "$ID_TK022" "$F_DEPS"   "TK0-2-1"

echo "=== TK0-3-1 ==="
set_ss "$ID_TK031" "$F_TYPE"   "$T_TICKET"
set_ss "$ID_TK031" "$F_PHASE"  "$P_PH0"
set_ss "$ID_TK031" "$F_ITER"   "$I_3"
set_ss "$ID_TK031" "$F_SVC"    "$S_ALL"
set_ss "$ID_TK031" "$F_PRI"    "$PRI_P0"
set_ss "$ID_TK031" "$F_EST"    "$E_M"
set_ss "$ID_TK031" "$F_STATUS" "$ST_DONE"
set_tx "$ID_TK031" "$F_PARENT" "IT0-3"
set_tx "$ID_TK031" "$F_DEPS"   "TK0-2-2"

echo "=== TK0-4-1 ==="
set_ss "$ID_TK041" "$F_TYPE"   "$T_TICKET"
set_ss "$ID_TK041" "$F_PHASE"  "$P_PH0"
set_ss "$ID_TK041" "$F_ITER"   "$I_4"
set_ss "$ID_TK041" "$F_SVC"    "$S_ALL"
set_ss "$ID_TK041" "$F_PRI"    "$PRI_P0"
set_ss "$ID_TK041" "$F_EST"    "$E_M"
set_ss "$ID_TK041" "$F_STATUS" "$ST_DONE"
set_tx "$ID_TK041" "$F_PARENT" "IT0-4"
set_tx "$ID_TK041" "$F_DEPS"   "TK0-3-1"

echo "=== TK0-4-2 ==="
set_ss "$ID_TK042" "$F_TYPE"   "$T_TICKET"
set_ss "$ID_TK042" "$F_PHASE"  "$P_PH0"
set_ss "$ID_TK042" "$F_ITER"   "$I_4"
set_ss "$ID_TK042" "$F_SVC"    "$S_ALL"
set_ss "$ID_TK042" "$F_PRI"    "$PRI_P0"
set_ss "$ID_TK042" "$F_EST"    "$E_S"
set_ss "$ID_TK042" "$F_STATUS" "$ST_DONE"
set_tx "$ID_TK042" "$F_PARENT" "IT0-4"
set_tx "$ID_TK042" "$F_DEPS"   "TK0-4-1"

# ===== Phase 1 Iteration 1 Tickets =====

echo "=== TK1-1-1 ==="
set_ss "$ID_TK111" "$F_TYPE"   "$T_TICKET"
set_ss "$ID_TK111" "$F_PHASE"  "$P_PH1"
set_ss "$ID_TK111" "$F_ITER"   "$I_1"
set_ss "$ID_TK111" "$F_SVC"    "$S_ALL"
set_ss "$ID_TK111" "$F_PRI"    "$PRI_P1"
set_ss "$ID_TK111" "$F_EST"    "$E_M"
set_ss "$ID_TK111" "$F_STATUS" "$ST_TODO"
set_tx "$ID_TK111" "$F_PARENT" "IT1-1"

echo "=== TK1-1-2 ==="
set_ss "$ID_TK112" "$F_TYPE"   "$T_TICKET"
set_ss "$ID_TK112" "$F_PHASE"  "$P_PH1"
set_ss "$ID_TK112" "$F_ITER"   "$I_1"
set_ss "$ID_TK112" "$F_SVC"    "$S_ALL"
set_ss "$ID_TK112" "$F_PRI"    "$PRI_P1"
set_ss "$ID_TK112" "$F_EST"    "$E_M"
set_ss "$ID_TK112" "$F_STATUS" "$ST_TODO"
set_tx "$ID_TK112" "$F_PARENT" "IT1-1"
set_tx "$ID_TK112" "$F_DEPS"   "TK1-1-1"

echo "=== TK1-1-3 ==="
set_ss "$ID_TK113" "$F_TYPE"   "$T_TICKET"
set_ss "$ID_TK113" "$F_PHASE"  "$P_PH1"
set_ss "$ID_TK113" "$F_ITER"   "$I_1"
set_ss "$ID_TK113" "$F_SVC"    "$S_ALL"
set_ss "$ID_TK113" "$F_PRI"    "$PRI_P1"
set_ss "$ID_TK113" "$F_EST"    "$E_M"
set_ss "$ID_TK113" "$F_STATUS" "$ST_TODO"
set_tx "$ID_TK113" "$F_PARENT" "IT1-1"
set_tx "$ID_TK113" "$F_DEPS"   "TK1-1-2"

echo "=== TK1-1-4 ==="
set_ss "$ID_TK114" "$F_TYPE"   "$T_TICKET"
set_ss "$ID_TK114" "$F_PHASE"  "$P_PH1"
set_ss "$ID_TK114" "$F_ITER"   "$I_1"
set_ss "$ID_TK114" "$F_SVC"    "$S_ALL"
set_ss "$ID_TK114" "$F_PRI"    "$PRI_P1"
set_ss "$ID_TK114" "$F_EST"    "$E_M"
set_ss "$ID_TK114" "$F_STATUS" "$ST_TODO"
set_tx "$ID_TK114" "$F_PARENT" "IT1-1"
set_tx "$ID_TK114" "$F_DEPS"   "TK1-1-3"

echo "=== TK1-1-5 ==="
set_ss "$ID_TK115" "$F_TYPE"   "$T_TICKET"
set_ss "$ID_TK115" "$F_PHASE"  "$P_PH1"
set_ss "$ID_TK115" "$F_ITER"   "$I_1"
set_ss "$ID_TK115" "$F_SVC"    "$S_ALL"
set_ss "$ID_TK115" "$F_PRI"    "$PRI_P1"
set_ss "$ID_TK115" "$F_EST"    "$E_S"
set_ss "$ID_TK115" "$F_STATUS" "$ST_TODO"
set_tx "$ID_TK115" "$F_PARENT" "IT1-1"
set_tx "$ID_TK115" "$F_DEPS"   "TK1-1-4"

echo "=== TK1-1-6 ==="
set_ss "$ID_TK116" "$F_TYPE"   "$T_TICKET"
set_ss "$ID_TK116" "$F_PHASE"  "$P_PH1"
set_ss "$ID_TK116" "$F_ITER"   "$I_1"
set_ss "$ID_TK116" "$F_SVC"    "$S_ALL"
set_ss "$ID_TK116" "$F_PRI"    "$PRI_P1"
set_ss "$ID_TK116" "$F_EST"    "$E_S"
set_ss "$ID_TK116" "$F_STATUS" "$ST_TODO"
set_tx "$ID_TK116" "$F_PARENT" "IT1-1"
set_tx "$ID_TK116" "$F_DEPS"   "TK1-1-5"

echo "=== 完了 ==="