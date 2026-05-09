package model

import "github.com/google/uuid"

// ProjectExtraction は機能抽出結果を表す。
type ProjectExtraction struct {
	Features   []string `json:"features"`
	Components []string `json:"components"`
}

// ProjectNameSuggestion はプロジェクト名候補を表す。
type ProjectNameSuggestion struct {
	Candidates []string               `json:"candidates"`
	Items      []ProjectNameCandidate `json:"items"`
}

// ProjectNameCandidate はプロジェクト名候補の詳細を表す。
type ProjectNameCandidate struct {
	Name        string `json:"name"`
	Reason      string `json:"reason,omitempty"`
	AISuggested bool   `json:"aiSuggested"`
}

// NameSuggestionProfile は名前候補生成で使うLLM運用モードを表す。
type NameSuggestionProfile struct {
	Profile           string   `json:"profile"`
	AvailableProfiles []string `json:"availableProfiles"`
	Enabled           bool     `json:"enabled"`
}

// ProjectInitResult は初期ディレクトリ作成結果を表す。
type ProjectInitResult struct {
	ID              uuid.UUID `json:"id"`
	DirectoryStatus string    `json:"directoryStatus"`
}

// ProjectWithExternalResult は GitHub リポジトリ作成・initial push の結果を表す。
type ProjectWithExternalResult struct {
	RepositoryURL     string `json:"repositoryUrl"`
	ExternalProjectID string `json:"externalProjectId"`
	PushStatus        string `json:"pushStatus"`
}

// GitHubProjectsResult は GitHub Projects v2 ボード作成結果を表す。
type GitHubProjectsResult struct {
	ProjectsURL string `json:"projectsUrl"`
	ProjectsID  string `json:"projectsId"`
	Status      string `json:"status"`
}

// Phase0Task は Phase0 の1タスクアイテムを表す。
type Phase0Task struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Type  string `json:"type"`
}

// Phase0TasksResult は Phase0 タスク登録結果を表す。
type Phase0TasksResult struct {
	Tasks  []*Phase0Task `json:"tasks"`
	Status string        `json:"status"`
}
