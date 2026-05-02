#!/bin/bash
# GitHub Project fields bulk update script.
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

# --- Item IDs ---
ID_PH0=PVTI_lAHOAO_tRc4BWapmzgrn31U
ID_PH1=PVTI_lAHOAO_tRc4BWapmzgrn31c
ID_PH2=PVTI_lAHOAO_tRc4BWapmzgrn31g
ID_IT01=PVTI_lAHOAO_tRc4BWapmzgrn31w
ID_IT02=PVTI_lAHOAO_tRc4BWapmzgrn32I
ID_IT03=PVTI_lAHOAO_tRc4BWapmzgrn32c
ID_IT11=PVTI_lAHOAO_tRc4BWapmzgrn32k
ID_IT12=PVTI_lAHOAO_tRc4BWapmzgrn32w
ID_IT13=PVTI_lAHOAO_tRc4BWapmzgrn320
ID_TK11=PVTI_lAHOAO_tRc4BWapmzgrn4B8
ID_TK12=PVTI_lAHOAO_tRc4BWapmzgrn4CI
ID_TK13=PVTI_lAHOAO_tRc4BWapmzgrn4CQ
ID_TK14=PVTI_lAHOAO_tRc4BWapmzgrn4Cc
ID_TK15=PVTI_lAHOAO_tRc4BWapmzgrn4Cg
ID_TK16=PVTI_lAHOAO_tRc4BWapmzgrn4Cs
ID_TK17=PVTI_lAHOAO_tRc4BWapmzgrn4DA

set_ss() { gh project item-edit --project-id "$PROJ_ID" --id "$1" --field-id "$2" --single-select-option-id "$3"; }
set_tx() { gh project item-edit --project-id "$PROJ_ID" --id "$1" --field-id "$2" --text "$3"; }

echo "=== PH0 ==="
set_ss "$ID_PH0" "$F_TYPE" "$T_PHASE"
set_ss "$ID_PH0" "$F_PHASE" "$P_PH0"
set_ss "$ID_PH0" "$F_PRI" "$PRI_P0"
set_ss "$ID_PH0" "$F_EST" "$E_L"

echo "=== PH1 ==="
set_ss "$ID_PH1" "$F_TYPE" "$T_PHASE"
set_ss "$ID_PH1" "$F_PHASE" "$P_PH1"
set_ss "$ID_PH1" "$F_PRI" "$PRI_P1"
set_ss "$ID_PH1" "$F_EST" "$E_L"
set_tx "$ID_PH1" "$F_DEPS" "PH0"

echo "=== PH2 ==="
set_ss "$ID_PH2" "$F_TYPE" "$T_PHASE"
set_ss "$ID_PH2" "$F_PHASE" "$P_PH2"
set_ss "$ID_PH2" "$F_PRI" "$PRI_P1"
set_ss "$ID_PH2" "$F_EST" "$E_L"
set_tx "$ID_PH2" "$F_DEPS" "PH1"

echo "=== IT0-1 ==="
set_ss "$ID_IT01" "$F_TYPE" "$T_ITER"
set_ss "$ID_IT01" "$F_PHASE" "$P_PH0"
set_ss "$ID_IT01" "$F_ITER" "$I_1"
set_ss "$ID_IT01" "$F_PRI" "$PRI_P0"
set_ss "$ID_IT01" "$F_EST" "$E_L"
set_tx "$ID_IT01" "$F_PARENT" "PH0"

echo "=== IT0-2 ==="
set_ss "$ID_IT02" "$F_TYPE" "$T_ITER"
set_ss "$ID_IT02" "$F_PHASE" "$P_PH0"
set_ss "$ID_IT02" "$F_ITER" "$I_2"
set_ss "$ID_IT02" "$F_PRI" "$PRI_P0"
set_ss "$ID_IT02" "$F_EST" "$E_L"
set_tx "$ID_IT02" "$F_PARENT" "PH0"
set_tx "$ID_IT02" "$F_DEPS" "IT0-1"

echo "=== IT0-3 ==="
set_ss "$ID_IT03" "$F_TYPE" "$T_ITER"
set_ss "$ID_IT03" "$F_PHASE" "$P_PH0"
set_ss "$ID_IT03" "$F_ITER" "$I_3"
set_ss "$ID_IT03" "$F_PRI" "$PRI_P0"
set_ss "$ID_IT03" "$F_EST" "$E_L"
set_tx "$ID_IT03" "$F_PARENT" "PH0"
set_tx "$ID_IT03" "$F_DEPS" "IT0-2"

echo "=== IT1-1 ==="
set_ss "$ID_IT11" "$F_TYPE" "$T_ITER"
set_ss "$ID_IT11" "$F_PHASE" "$P_PH1"
set_ss "$ID_IT11" "$F_ITER" "$I_4"
set_ss "$ID_IT11" "$F_PRI" "$PRI_P1"
set_ss "$ID_IT11" "$F_EST" "$E_L"
set_tx "$ID_IT11" "$F_PARENT" "PH1"
set_tx "$ID_IT11" "$F_DEPS" "IT0-3"

echo "=== IT1-2 ==="
set_ss "$ID_IT12" "$F_TYPE" "$T_ITER"
set_ss "$ID_IT12" "$F_PHASE" "$P_PH1"
set_ss "$ID_IT12" "$F_ITER" "$I_5"
set_ss "$ID_IT12" "$F_PRI" "$PRI_P1"
set_ss "$ID_IT12" "$F_EST" "$E_L"
set_tx "$ID_IT12" "$F_PARENT" "PH1"
set_tx "$ID_IT12" "$F_DEPS" "IT1-1"

echo "=== IT1-3 ==="
set_ss "$ID_IT13" "$F_TYPE" "$T_ITER"
set_ss "$ID_IT13" "$F_PHASE" "$P_PH1"
set_ss "$ID_IT13" "$F_ITER" "$I_6"
set_ss "$ID_IT13" "$F_PRI" "$PRI_P1"
set_ss "$ID_IT13" "$F_EST" "$E_L"
set_tx "$ID_IT13" "$F_PARENT" "PH1"
set_tx "$ID_IT13" "$F_DEPS" "IT1-2"

echo "=== TK-1-1 ==="
set_ss "$ID_TK11" "$F_TYPE" "$T_TICKET"
set_ss "$ID_TK11" "$F_PHASE" "$P_PH0"
set_ss "$ID_TK11" "$F_ITER" "$I_1"
set_ss "$ID_TK11" "$F_SVC" "$S_API"
set_ss "$ID_TK11" "$F_PRI" "$PRI_P0"
set_ss "$ID_TK11" "$F_EST" "$E_L"
set_tx "$ID_TK11" "$F_PARENT" "IT0-1"

echo "=== TK-1-2 ==="
set_ss "$ID_TK12" "$F_TYPE" "$T_TICKET"
set_ss "$ID_TK12" "$F_PHASE" "$P_PH0"
set_ss "$ID_TK12" "$F_ITER" "$I_1"
set_ss "$ID_TK12" "$F_SVC" "$S_API"
set_ss "$ID_TK12" "$F_PRI" "$PRI_P0"
set_ss "$ID_TK12" "$F_EST" "$E_M"
set_tx "$ID_TK12" "$F_PARENT" "IT0-1"
set_tx "$ID_TK12" "$F_DEPS" "TK-1-1"

echo "=== TK-1-3 ==="
set_ss "$ID_TK13" "$F_TYPE" "$T_TICKET"
set_ss "$ID_TK13" "$F_PHASE" "$P_PH0"
set_ss "$ID_TK13" "$F_ITER" "$I_1"
set_ss "$ID_TK13" "$F_SVC" "$S_API"
set_ss "$ID_TK13" "$F_PRI" "$PRI_P0"
set_ss "$ID_TK13" "$F_EST" "$E_M"
set_tx "$ID_TK13" "$F_PARENT" "IT0-1"
set_tx "$ID_TK13" "$F_DEPS" "TK-1-2"

echo "=== TK-1-4 ==="
set_ss "$ID_TK14" "$F_TYPE" "$T_TICKET"
set_ss "$ID_TK14" "$F_PHASE" "$P_PH0"
set_ss "$ID_TK14" "$F_ITER" "$I_1"
set_ss "$ID_TK14" "$F_SVC" "$S_INFRA"
set_ss "$ID_TK14" "$F_PRI" "$PRI_P0"
set_ss "$ID_TK14" "$F_EST" "$E_M"
set_tx "$ID_TK14" "$F_PARENT" "IT0-1"

echo "=== TK-1-5 ==="
set_ss "$ID_TK15" "$F_TYPE" "$T_TICKET"
set_ss "$ID_TK15" "$F_PHASE" "$P_PH0"
set_ss "$ID_TK15" "$F_ITER" "$I_1"
set_ss "$ID_TK15" "$F_SVC" "$S_INFRA"
set_ss "$ID_TK15" "$F_PRI" "$PRI_P0"
set_ss "$ID_TK15" "$F_EST" "$E_M"
set_tx "$ID_TK15" "$F_PARENT" "IT0-1"

echo "=== TK-1-6 ==="
set_ss "$ID_TK16" "$F_TYPE" "$T_TICKET"
set_ss "$ID_TK16" "$F_PHASE" "$P_PH0"
set_ss "$ID_TK16" "$F_ITER" "$I_1"
set_ss "$ID_TK16" "$F_SVC" "$S_TEST"
set_ss "$ID_TK16" "$F_PRI" "$PRI_P0"
set_ss "$ID_TK16" "$F_EST" "$E_M"
set_tx "$ID_TK16" "$F_PARENT" "IT0-1"
set_tx "$ID_TK16" "$F_DEPS" "TK-1-2, TK-1-3"

echo "=== TK-1-7 ==="
set_ss "$ID_TK17" "$F_TYPE" "$T_TICKET"
set_ss "$ID_TK17" "$F_PHASE" "$P_PH0"
set_ss "$ID_TK17" "$F_ITER" "$I_1"
set_ss "$ID_TK17" "$F_SVC" "$S_UI"
set_ss "$ID_TK17" "$F_PRI" "$PRI_P0"
set_ss "$ID_TK17" "$F_EST" "$E_L"
set_tx "$ID_TK17" "$F_PARENT" "IT0-1"
set_tx "$ID_TK17" "$F_DEPS" "TK-1-4, TK-1-5"

echo "=== 完了 ==="