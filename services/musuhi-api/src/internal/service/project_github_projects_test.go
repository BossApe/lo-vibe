package service

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type sequentialGraphQLRunner struct {
	responses [][]byte
	errs      []error
	calls     int
}

func (r *sequentialGraphQLRunner) RunGraphQL(_ context.Context, _ []byte) ([]byte, error) {
	if r.calls >= len(r.responses) {
		return nil, assert.AnError
	}
	out := r.responses[r.calls]
	var err error
	if r.calls < len(r.errs) {
		err = r.errs[r.calls]
	}
	r.calls++
	return out, err
}

func TestGHProjectsClient_CreateProject_GitHubProjectsを作成してカスタムフィールドを追加する_正常系(t *testing.T) {
	r := &sequentialGraphQLRunner{
		responses: [][]byte{
			[]byte(`{"data":{"user":{"id":"U_owner"}}}`),
			[]byte(`{"data":{"createProjectV2":{"projectV2":{"id":"PVT_001","url":"https://github.com/orgs/BossApe/projects/77","number":77}}}}`),
			[]byte(`{"data":{"createProjectV2Field":{"projectV2Field":{"id":"F_TYPE"}}}}`),
			[]byte(`{"data":{"createProjectV2Field":{"projectV2Field":{"id":"F_PHASE"}}}}`),
			[]byte(`{"data":{"createProjectV2Field":{"projectV2Field":{"id":"F_SPRINT"}}}}`),
			[]byte(`{"data":{"createProjectV2Field":{"projectV2Field":{"id":"F_PRIORITY"}}}}`),
			[]byte(`{"data":{"createProjectV2Field":{"projectV2Field":{"id":"F_EST"}}}}`),
			[]byte(`{"data":{"createProjectV2Field":{"projectV2Field":{"id":"F_DEPENDS"}}}}`),
		},
	}
	client := &ghProjectsClient{runner: r}

	got, err := client.CreateProject(context.Background(), "BossApe", "Musuhi Board")
	require.NoError(t, err)
	assert.Equal(t, "success", got.Status)
	assert.Equal(t, "PVT_001", got.ProjectsID)
	assert.Equal(t, "https://github.com/orgs/BossApe/projects/77", got.ProjectsURL)
	assert.Equal(t, 8, r.calls)
}

func TestGHProjectsClient_AddPhase0Tasks_Phase0タスクをProjectsへ登録する_正常系(t *testing.T) {
	backupDefs := phase0TaskDefs
	phase0TaskDefs = []phase0TaskDef{
		{TaskID: "TK0-1-1", Title: "TK0-1-1: 提案・要求仕様書自動生成", ItemType: "Ticket", Phase: "PH0", Sprint: "SP0-1", Priority: "High", Estimate: 1},
	}
	defer func() { phase0TaskDefs = backupDefs }()

	fieldsJSON := map[string]any{
		"data": map[string]any{
			"node": map[string]any{
				"fields": map[string]any{
					"nodes": []map[string]any{
						{"id": "F_TYPE", "name": "Type", "options": []map[string]any{{"id": "OPT_TYPE_TICKET", "name": "Ticket"}}},
						{"id": "F_PHASE", "name": "Phase"},
						{"id": "F_SPRINT", "name": "Sprint"},
						{"id": "F_PRIORITY", "name": "Priority", "options": []map[string]any{{"id": "OPT_PRIO_HIGH", "name": "High"}}},
						{"id": "F_EST", "name": "Estimate"},
					},
				},
			},
		},
	}
	fieldsBytes, _ := json.Marshal(fieldsJSON)

	r := &sequentialGraphQLRunner{
		responses: [][]byte{
			fieldsBytes,
			[]byte(`{"data":{"addProjectV2DraftIssue":{"projectItem":{"id":"PVTI_001"}}}}`),
			[]byte(`{"data":{"updateProjectV2ItemFieldValue":{"projectV2Item":{"id":"PVTI_001"}}}}`),
			[]byte(`{"data":{"updateProjectV2ItemFieldValue":{"projectV2Item":{"id":"PVTI_001"}}}}`),
			[]byte(`{"data":{"updateProjectV2ItemFieldValue":{"projectV2Item":{"id":"PVTI_001"}}}}`),
			[]byte(`{"data":{"updateProjectV2ItemFieldValue":{"projectV2Item":{"id":"PVTI_001"}}}}`),
			[]byte(`{"data":{"updateProjectV2ItemFieldValue":{"projectV2Item":{"id":"PVTI_001"}}}}`),
		},
	}
	client := &ghProjectsClient{runner: r}

	tasks, err := client.AddPhase0Tasks(context.Background(), "BossApe", "PVT_001")
	require.NoError(t, err)
	require.Len(t, tasks, 1)
	assert.Equal(t, "PVTI_001", tasks[0].ID)
	assert.Equal(t, "Ticket", tasks[0].Type)
	assert.Equal(t, 7, r.calls)
}
