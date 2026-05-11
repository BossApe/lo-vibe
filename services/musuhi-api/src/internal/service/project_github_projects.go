package service

// GitHubProjectsClient は GitHub Projects v2 の操作を抽象化するインターフェース。

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"musuhi-api/internal/model"
)

// GitHubProjectsClient は GitHub Projects v2 の操作を抽象化する。
type GitHubProjectsClient interface {
	CreateProject(ctx context.Context, owner, title string) (*model.GitHubProjectsResult, error)
	AddPhase0Tasks(ctx context.Context, owner, projectID string) ([]*model.Phase0Task, error)
}

// graphqlRunner は GitHub GraphQL API リクエストを実行する。
// graphqlRunner は GitHub GraphQL API リクエストを実行するインターフェース。
	RunGraphQL(ctx context.Context, body []byte) ([]byte, error)
}

// defaultGraphQLRunner は graphqlRunner のデフォルト実装。

// RunGraphQL は gh CLI でGraphQL APIを実行します。
func (r *defaultGraphQLRunner) RunGraphQL(ctx context.Context, body []byte) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "gh", "api", "graphql", "--input", "-")
	cmd.Stdin = bytes.NewReader(body)
	return cmd.CombinedOutput()
}

// ghProjectsClient は GitHub Projects v2 操作用のクライアント実装。
type ghProjectsClient struct {
	runner graphqlRunner
}

// newDefaultGitHubProjectsClient は ghProjectsClient を生成します。
func newDefaultGitHubProjectsClient() GitHubProjectsClient {
	return &ghProjectsClient{runner: &defaultGraphQLRunner{}}
}

// --- GraphQL helpers ---

// buildGraphQLBody はGraphQLクエリと変数からリクエストボディを生成します。
func buildGraphQLBody(query string, variables map[string]any) ([]byte, error) {
	return json.Marshal(map[string]any{
		"query":     query,
		"variables": variables,
	})
}

// extractGraphQLErrors はGraphQLレスポンスからエラーを抽出します。
func extractGraphQLErrors(data []byte) error {
	var resp struct {
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil
	}
	if len(resp.Errors) > 0 {
		msgs := make([]string, 0, len(resp.Errors))
		for _, e := range resp.Errors {
			msgs = append(msgs, e.Message)
		}
		return fmt.Errorf("graphql errors: %s", strings.Join(msgs, "; "))
	}
	return nil
}

// runGraphQL はGraphQLクエリを実行し、エラーも判定します。
func (c *ghProjectsClient) runGraphQL(ctx context.Context, query string, variables map[string]any) ([]byte, error) {
	body, err := buildGraphQLBody(query, variables)
	if err != nil {
		return nil, fmt.Errorf("marshal graphql body: %w", err)
	}
	out, err := c.runner.RunGraphQL(ctx, body)
	if err != nil {
		return nil, fmt.Errorf("gh api graphql: %w: %s", err, strings.TrimSpace(string(out)))
	}
	if gqlErr := extractGraphQLErrors(out); gqlErr != nil {
		return nil, gqlErr
	}
	return out, nil
}

// --- Owner ID resolution ---

// resolveOwnerID はオーナー名からGitHubのユーザーまたは組織IDを解決します。
func (c *ghProjectsClient) resolveOwnerID(ctx context.Context, owner string) (string, error) {
	// Try user first
	out, err := c.runGraphQL(ctx,
		`query($login:String!){user(login:$login){id}}`,
		map[string]any{"login": owner},
	)
	if err == nil {
		var resp struct {
			Data struct {
				User *struct {
					ID string `json:"id"`
				} `json:"user"`
			} `json:"data"`
		}
		if jsonErr := json.Unmarshal(out, &resp); jsonErr == nil && resp.Data.User != nil && resp.Data.User.ID != "" {
			return resp.Data.User.ID, nil
		}
	}

	// Try org
	out, err = c.runGraphQL(ctx,
		`query($login:String!){organization(login:$login){id}}`,
		map[string]any{"login": owner},
	)
	if err != nil {
		return "", fmt.Errorf("resolveOwnerID(%q): %w", owner, err)
	}
	var resp struct {
		Data struct {
			Organization *struct {
				ID string `json:"id"`
			} `json:"organization"`
		} `json:"data"`
	}
	if err := json.Unmarshal(out, &resp); err != nil || resp.Data.Organization == nil || resp.Data.Organization.ID == "" {
		return "", fmt.Errorf("resolveOwnerID: could not resolve owner %q as user or organization", owner)
	}
	return resp.Data.Organization.ID, nil
}

// --- Field structs ---

type singleSelectOption struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type projectFieldSet struct {
	projectID       string
	typeFieldID     string
	typeOptions     map[string]string // name -> optionID
	phaseFieldID    string
	sprintFieldID   string
	priorityFieldID string
	priorityOptions map[string]string
	estimateFieldID string
}

// --- Create project ---

// CreateProject は指定オーナー配下にProjectsボードを作成します。
func (c *ghProjectsClient) CreateProject(ctx context.Context, owner, title string) (*model.GitHubProjectsResult, error) {
	ownerID, err := c.resolveOwnerID(ctx, owner)
	if err != nil {
		return nil, fmt.Errorf("CreateProject: %w", err)
	}

	out, err := c.runGraphQL(ctx,
		`mutation($ownerId:ID!,$title:String!){createProjectV2(input:{ownerId:$ownerId,title:$title}){projectV2{id,url,number}}}`,
		map[string]any{"ownerId": ownerID, "title": title},
	)
	if err != nil {
		return nil, fmt.Errorf("CreateProject: createProjectV2: %w", err)
	}

	var createResp struct {
		Data struct {
			CreateProjectV2 struct {
				ProjectV2 struct {
					ID  string `json:"id"`
					URL string `json:"url"`
				} `json:"projectV2"`
			} `json:"createProjectV2"`
		} `json:"data"`
	}
	if err := json.Unmarshal(out, &createResp); err != nil {
		return nil, fmt.Errorf("CreateProject: unmarshal createProjectV2: %w", err)
	}
	projectID := createResp.Data.CreateProjectV2.ProjectV2.ID
	projectURL := createResp.Data.CreateProjectV2.ProjectV2.URL
	if projectID == "" {
		return nil, fmt.Errorf("CreateProject: createProjectV2 returned empty project id")
	}

	if err := c.createCustomFields(ctx, projectID); err != nil {
		return nil, fmt.Errorf("CreateProject: %w", err)
	}

	return &model.GitHubProjectsResult{
		ProjectsURL: projectURL,
		ProjectsID:  projectID,
		Status:      "success",
	}, nil
}

// createCustomFields はProjectsボードにカスタムフィールドを追加します。
func (c *ghProjectsClient) createCustomFields(ctx context.Context, projectID string) error {
	// Type (single select: Phase, Sprint, Ticket)
	if _, err := c.createSingleSelectField(ctx, projectID, "Type",
		[]string{"Phase", "Sprint", "Ticket"}); err != nil {
		return fmt.Errorf("createCustomFields: Type: %w", err)
	}
	// Phase (text)
	if _, err := c.createTextField(ctx, projectID, "Phase"); err != nil {
		return fmt.Errorf("createCustomFields: Phase: %w", err)
	}
	// Sprint (text)
	if _, err := c.createTextField(ctx, projectID, "Sprint"); err != nil {
		return fmt.Errorf("createCustomFields: Sprint: %w", err)
	}
	// Priority (single select: High, Medium, Low)
	if _, err := c.createSingleSelectField(ctx, projectID, "Priority",
		[]string{"High", "Medium", "Low"}); err != nil {
		return fmt.Errorf("createCustomFields: Priority: %w", err)
	}
	// Estimate (number)
	if _, err := c.createNumberField(ctx, projectID, "Estimate"); err != nil {
		return fmt.Errorf("createCustomFields: Estimate: %w", err)
	}
	// Depends on (text)
	if _, err := c.createTextField(ctx, projectID, "Depends on"); err != nil {
		return fmt.Errorf("createCustomFields: Depends on: %w", err)
	}
	return nil
}

// createSingleSelectField は単一選択フィールドをProjectsボードに追加します。
func (c *ghProjectsClient) createSingleSelectField(ctx context.Context, projectID, name string, options []string) (string, error) {
	opts := make([]map[string]string, 0, len(options))
	for _, o := range options {
		opts = append(opts, map[string]string{"name": o, "color": "GRAY", "description": ""})
	}
	out, err := c.runGraphQL(ctx,
		`mutation($projectId:ID!,$name:String!,$options:[ProjectV2SingleSelectFieldOptionInput!]!){`+
			`createProjectV2Field(input:{projectId:$projectId,dataType:SINGLE_SELECT,name:$name,singleSelectOptions:$options}){`+
			`projectV2Field{...on ProjectV2SingleSelectField{id}}}}`,
		map[string]any{"projectId": projectID, "name": name, "options": opts},
	)
	if err != nil {
		return "", fmt.Errorf("createSingleSelectField(%q): %w", name, err)
	}
	var resp struct {
		Data struct {
			CreateProjectV2Field struct {
				ProjectV2Field struct {
					ID string `json:"id"`
				} `json:"projectV2Field"`
			} `json:"createProjectV2Field"`
		} `json:"data"`
	}
	if err := json.Unmarshal(out, &resp); err != nil {
		return "", fmt.Errorf("createSingleSelectField(%q): unmarshal: %w", name, err)
	}
	return resp.Data.CreateProjectV2Field.ProjectV2Field.ID, nil
}

// createTextField はテキストフィールドをProjectsボードに追加します。
func (c *ghProjectsClient) createTextField(ctx context.Context, projectID, name string) (string, error) {
	out, err := c.runGraphQL(ctx,
		`mutation($projectId:ID!,$name:String!){createProjectV2Field(input:{projectId:$projectId,dataType:TEXT,name:$name}){projectV2Field{...on ProjectV2Field{id}}}}`,
		map[string]any{"projectId": projectID, "name": name},
	)
	if err != nil {
		return "", fmt.Errorf("createTextField(%q): %w", name, err)
	}
	var resp struct {
		Data struct {
			CreateProjectV2Field struct {
				ProjectV2Field struct {
					ID string `json:"id"`
				} `json:"projectV2Field"`
			} `json:"createProjectV2Field"`
		} `json:"data"`
	}
	if err := json.Unmarshal(out, &resp); err != nil {
		return "", fmt.Errorf("createTextField(%q): unmarshal: %w", name, err)
	}
	return resp.Data.CreateProjectV2Field.ProjectV2Field.ID, nil
}

// createNumberField は数値フィールドをProjectsボードに追加します。
func (c *ghProjectsClient) createNumberField(ctx context.Context, projectID, name string) (string, error) {
	out, err := c.runGraphQL(ctx,
		`mutation($projectId:ID!,$name:String!){createProjectV2Field(input:{projectId:$projectId,dataType:NUMBER,name:$name}){projectV2Field{...on ProjectV2Field{id}}}}`,
		map[string]any{"projectId": projectID, "name": name},
	)
	if err != nil {
		return "", fmt.Errorf("createNumberField(%q): %w", name, err)
	}
	var resp struct {
		Data struct {
			CreateProjectV2Field struct {
				ProjectV2Field struct {
					ID string `json:"id"`
				} `json:"projectV2Field"`
			} `json:"createProjectV2Field"`
		} `json:"data"`
	}
	if err := json.Unmarshal(out, &resp); err != nil {
		return "", fmt.Errorf("createNumberField(%q): unmarshal: %w", name, err)
	}
	return resp.Data.CreateProjectV2Field.ProjectV2Field.ID, nil
}

// --- Phase0 task definitions ---

// phase0TaskDef はPhase0タスクの定義情報です。
type phase0TaskDef struct {
	TaskID   string
	Title    string
	ItemType string // Phase, Sprint, Ticket
	Phase    string
	Sprint   string
	Priority string
	Estimate float64
}

var phase0TaskDefs = []phase0TaskDef{
	{TaskID: "PH0", Title: "PH0: 提案・要求仕様・要件定義", ItemType: "Phase", Phase: "PH0", Sprint: "", Priority: "High", Estimate: 8},
	{TaskID: "SP0-1", Title: "SP0-1: 提案・要求仕様作成", ItemType: "Sprint", Phase: "PH0", Sprint: "SP0-1", Priority: "High", Estimate: 2},
	{TaskID: "TK0-1-1", Title: "TK0-1-1: 提案・要求仕様書自動生成", ItemType: "Ticket", Phase: "PH0", Sprint: "SP0-1", Priority: "High", Estimate: 1},
	{TaskID: "TK0-1-2", Title: "TK0-1-2: 提案・要求仕様書ユーザ承認", ItemType: "Ticket", Phase: "PH0", Sprint: "SP0-1", Priority: "High", Estimate: 1},
	{TaskID: "SP0-2", Title: "SP0-2: 要件定義作成", ItemType: "Sprint", Phase: "PH0", Sprint: "SP0-2", Priority: "High", Estimate: 2},
	{TaskID: "TK0-2-1", Title: "TK0-2-1: 要件定義書自動生成", ItemType: "Ticket", Phase: "PH0", Sprint: "SP0-2", Priority: "High", Estimate: 1},
	{TaskID: "TK0-2-2", Title: "TK0-2-2: 要件定義書ユーザ承認", ItemType: "Ticket", Phase: "PH0", Sprint: "SP0-2", Priority: "High", Estimate: 1},
	{TaskID: "SP0-3", Title: "SP0-3: 開発タスク生成", ItemType: "Sprint", Phase: "PH0", Sprint: "SP0-3", Priority: "High", Estimate: 1},
	{TaskID: "TK0-3-1", Title: "TK0-3-1: Phase/Sprint/Ticket分割・登録", ItemType: "Ticket", Phase: "PH0", Sprint: "SP0-3", Priority: "High", Estimate: 1},
	{TaskID: "SP0-4", Title: "SP0-4: 開発準備", ItemType: "Sprint", Phase: "PH0", Sprint: "SP0-4", Priority: "Medium", Estimate: 2},
	{TaskID: "TK0-4-1", Title: "TK0-4-1: 開発規約自動生成", ItemType: "Ticket", Phase: "PH0", Sprint: "SP0-4", Priority: "Medium", Estimate: 1},
	{TaskID: "TK0-4-2", Title: "TK0-4-2: tools準備", ItemType: "Ticket", Phase: "PH0", Sprint: "SP0-4", Priority: "Medium", Estimate: 1},
}

// --- Add Phase0 tasks ---

// AddPhase0Tasks はPhase0タスクをProjectsボードに一括登録します。
func (c *ghProjectsClient) AddPhase0Tasks(ctx context.Context, owner, projectID string) ([]*model.Phase0Task, error) {
	fields, err := c.getProjectFields(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("AddPhase0Tasks: getProjectFields: %w", err)
	}
	fields.projectID = projectID

	var tasks []*model.Phase0Task
	for _, def := range phase0TaskDefs {
		itemID, err := c.addDraftIssue(ctx, projectID, def.Title, def.TaskID)
		if err != nil {
			return nil, fmt.Errorf("AddPhase0Tasks: addDraftIssue(%q): %w", def.TaskID, err)
		}
		if err := c.setItemFields(ctx, fields, itemID, def); err != nil {
			return nil, fmt.Errorf("AddPhase0Tasks: setItemFields(%q): %w", def.TaskID, err)
		}
		tasks = append(tasks, &model.Phase0Task{ID: itemID, Title: def.Title, Type: def.ItemType})
	}
	return tasks, nil
}

// getProjectFields はProjectsボードのカスタムフィールド情報を取得します。
func (c *ghProjectsClient) getProjectFields(ctx context.Context, projectID string) (*projectFieldSet, error) {
	out, err := c.runGraphQL(ctx,
		`query($id:ID!){node(id:$id){...on ProjectV2{fields(first:20){nodes{`+
			`...on ProjectV2Field{id,name}`+
			`...on ProjectV2SingleSelectField{id,name,options{id,name}}`+
			`}}}}}`,
		map[string]any{"id": projectID},
	)
	if err != nil {
		return nil, fmt.Errorf("getProjectFields: %w", err)
	}

	var resp struct {
		Data struct {
			Node struct {
				Fields struct {
					Nodes []struct {
						ID      string               `json:"id"`
						Name    string               `json:"name"`
						Options []singleSelectOption `json:"options"`
					} `json:"nodes"`
				} `json:"fields"`
			} `json:"node"`
		} `json:"data"`
	}
	if err := json.Unmarshal(out, &resp); err != nil {
		return nil, fmt.Errorf("getProjectFields: unmarshal: %w", err)
	}

	fs := &projectFieldSet{
		typeOptions:     make(map[string]string),
		priorityOptions: make(map[string]string),
	}
	for _, node := range resp.Data.Node.Fields.Nodes {
		switch node.Name {
		case "Type":
			fs.typeFieldID = node.ID
			for _, opt := range node.Options {
				fs.typeOptions[opt.Name] = opt.ID
			}
		case "Phase":
			fs.phaseFieldID = node.ID
		case "Sprint":
			fs.sprintFieldID = node.ID
		case "Priority":
			fs.priorityFieldID = node.ID
			for _, opt := range node.Options {
				fs.priorityOptions[opt.Name] = opt.ID
			}
		case "Estimate":
			fs.estimateFieldID = node.ID
		}
	}
	return fs, nil
}

// addDraftIssue はドラフトIssueをProjectsボードに追加します。
func (c *ghProjectsClient) addDraftIssue(ctx context.Context, projectID, title, body string) (string, error) {
	out, err := c.runGraphQL(ctx,
		`mutation($projectId:ID!,$title:String!,$body:String!){addProjectV2DraftIssue(input:{projectId:$projectId,title:$title,body:$body}){projectItem{id}}}`,
		map[string]any{"projectId": projectID, "title": title, "body": body},
	)
	if err != nil {
		return "", fmt.Errorf("addDraftIssue: %w", err)
	}
	var resp struct {
		Data struct {
			AddProjectV2DraftIssue struct {
				ProjectItem struct {
					ID string `json:"id"`
				} `json:"projectItem"`
			} `json:"addProjectV2DraftIssue"`
		} `json:"data"`
	}
	if err := json.Unmarshal(out, &resp); err != nil {
		return "", fmt.Errorf("addDraftIssue: unmarshal: %w", err)
	}
	if resp.Data.AddProjectV2DraftIssue.ProjectItem.ID == "" {
		return "", fmt.Errorf("addDraftIssue: returned empty item id")
	}
	return resp.Data.AddProjectV2DraftIssue.ProjectItem.ID, nil
}

// setItemFields はタスクに各種フィールド値を設定します。
func (c *ghProjectsClient) setItemFields(ctx context.Context, fs *projectFieldSet, itemID string, def phase0TaskDef) error {
	// Type (single select)
	if fs.typeFieldID != "" {
		if optID, ok := fs.typeOptions[def.ItemType]; ok {
			if err := c.setSingleSelectFieldValue(ctx, fs.projectID, itemID, fs.typeFieldID, optID); err != nil {
				return fmt.Errorf("setItemFields: Type: %w", err)
			}
		}
	}
	// Phase (text)
	if fs.phaseFieldID != "" && def.Phase != "" {
		if err := c.setTextFieldValue(ctx, fs.projectID, itemID, fs.phaseFieldID, def.Phase); err != nil {
			return fmt.Errorf("setItemFields: Phase: %w", err)
		}
	}
	// Sprint (text)
	if fs.sprintFieldID != "" && def.Sprint != "" {
		if err := c.setTextFieldValue(ctx, fs.projectID, itemID, fs.sprintFieldID, def.Sprint); err != nil {
			return fmt.Errorf("setItemFields: Sprint: %w", err)
		}
	}
	// Priority (single select)
	if fs.priorityFieldID != "" {
		if optID, ok := fs.priorityOptions[def.Priority]; ok {
			if err := c.setSingleSelectFieldValue(ctx, fs.projectID, itemID, fs.priorityFieldID, optID); err != nil {
				return fmt.Errorf("setItemFields: Priority: %w", err)
			}
		}
	}
	// Estimate (number)
	if fs.estimateFieldID != "" {
		if err := c.setNumberFieldValue(ctx, fs.projectID, itemID, fs.estimateFieldID, def.Estimate); err != nil {
			return fmt.Errorf("setItemFields: Estimate: %w", err)
		}
	}
	return nil
}

func (c *ghProjectsClient) setTextFieldValue(ctx context.Context, projectID, itemID, fieldID, value string) error {
	_, err := c.runGraphQL(ctx,
		`mutation($projectId:ID!,$itemId:ID!,$fieldId:ID!,$value:String!){`+
			`updateProjectV2ItemFieldValue(input:{projectId:$projectId,itemId:$itemId,fieldId:$fieldId,value:{text:$value}}){projectV2Item{id}}}`,
		map[string]any{"projectId": projectID, "itemId": itemID, "fieldId": fieldID, "value": value},
	)
	if err != nil {
		return fmt.Errorf("setTextFieldValue: %w", err)
	}
	return nil
}

func (c *ghProjectsClient) setSingleSelectFieldValue(ctx context.Context, projectID, itemID, fieldID, optionID string) error {
	_, err := c.runGraphQL(ctx,
		`mutation($projectId:ID!,$itemId:ID!,$fieldId:ID!,$value:String!){`+
			`updateProjectV2ItemFieldValue(input:{projectId:$projectId,itemId:$itemId,fieldId:$fieldId,value:{singleSelectOptionId:$value}}){projectV2Item{id}}}`,
		map[string]any{"projectId": projectID, "itemId": itemID, "fieldId": fieldID, "value": optionID},
	)
	if err != nil {
		return fmt.Errorf("setSingleSelectFieldValue: %w", err)
	}
	return nil
}

func (c *ghProjectsClient) setNumberFieldValue(ctx context.Context, projectID, itemID, fieldID string, value float64) error {
	_, err := c.runGraphQL(ctx,
		`mutation($projectId:ID!,$itemId:ID!,$fieldId:ID!,$value:Float!){`+
			`updateProjectV2ItemFieldValue(input:{projectId:$projectId,itemId:$itemId,fieldId:$fieldId,value:{number:$value}}){projectV2Item{id}}}`,
		map[string]any{"projectId": projectID, "itemId": itemID, "fieldId": fieldID, "value": value},
	)
	if err != nil {
		return fmt.Errorf("setNumberFieldValue: %w", err)
	}
	return nil
}

// --- Service methods ---

func (s *projectService) CreateGitHubProjects(ctx context.Context, id, owner, title string) (*model.GitHubProjectsResult, error) {
	owner = strings.TrimSpace(owner)
	title = strings.TrimSpace(title)

	if owner == "" {
		return nil, fmt.Errorf("%w: owner is required", ErrValidation)
	}
	if title == "" {
		return nil, fmt.Errorf("%w: title is required", ErrValidation)
	}

	result, err := s.githubProjectsClient.CreateProject(ctx, owner, title)
	if err != nil {
		return nil, fmt.Errorf("projectService.CreateGitHubProjects: %w", err)
	}
	return result, nil
}

func (s *projectService) CreatePhase0Tasks(ctx context.Context, id, owner, projectsID string) (*model.Phase0TasksResult, error) {
	owner = strings.TrimSpace(owner)
	projectsID = strings.TrimSpace(projectsID)

	if owner == "" {
		return nil, fmt.Errorf("%w: owner is required", ErrValidation)
	}
	if projectsID == "" {
		return nil, fmt.Errorf("%w: projectsId is required", ErrValidation)
	}

	tasks, err := s.githubProjectsClient.AddPhase0Tasks(ctx, owner, projectsID)
	if err != nil {
		return nil, fmt.Errorf("projectService.CreatePhase0Tasks: %w", err)
	}
	return &model.Phase0TasksResult{Tasks: tasks, Status: "success"}, nil
}
