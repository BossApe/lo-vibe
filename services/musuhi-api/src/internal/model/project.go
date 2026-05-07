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

// ProjectInitResult は初期ディレクトリ作成結果を表す。
type ProjectInitResult struct {
	ID              uuid.UUID `json:"id"`
	DirectoryStatus string    `json:"directoryStatus"`
}
