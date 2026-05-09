package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"musuhi-api/internal/model"
)

type openAIProjectNameSuggester struct {
	endpoint       string
	apiKey         string
	config         llmModelConfig
	client         *http.Client
	mu             sync.RWMutex
	currentProfile string
}

type llmModelConfig struct {
	FastModel     string
	BalancedModel string
	QualityModel  string
	Profile       string
}

type chatCompletionsRequest struct {
	Model       string                   `json:"model"`
	Messages    []chatCompletionsMessage `json:"messages"`
	Temperature float64                  `json:"temperature,omitempty"`
}

type chatCompletionsMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionsResponse struct {
	Choices []struct {
		Message chatCompletionsMessage `json:"message"`
	} `json:"choices"`
}

type aiCandidateEnvelope struct {
	Items []struct {
		Name   string `json:"name"`
		Reason string `json:"reason"`
	} `json:"items"`
}

func newEnvProjectNameSuggester() ProjectNameSuggester {
	endpoint := strings.TrimSpace(os.Getenv("MUSUHI_LLM_ENDPOINT"))
	apiKey := strings.TrimSpace(os.Getenv("MUSUHI_LLM_API_KEY"))
	config := loadLLMModelConfigFromEnv()
	if endpoint == "" {
		return nil
	}

	return &openAIProjectNameSuggester{
		endpoint:       endpoint,
		apiKey:         apiKey,
		config:         config,
		currentProfile: config.Profile,
		client: &http.Client{
			Timeout: 8 * time.Second,
		},
	}
}

func (s *openAIProjectNameSuggester) SuggestGodNames(ctx context.Context, overviewContent string) ([]model.ProjectNameCandidate, error) {
	if strings.TrimSpace(overviewContent) == "" {
		return nil, nil
	}

	requestBody := chatCompletionsRequest{
		Model: resolveModelByProfile(s.config, s.getCurrentProfile(), overviewContent),
		Messages: []chatCompletionsMessage{
			{
				Role:    "system",
				Content: "あなたは日本神話に詳しい命名アシスタントです。入力概要に意味的に近い神様名を3件提案してください。出力はJSONのみ。形式: {\"items\":[{\"name\":\"lowercase-name\",\"reason\":\"理由\"}]}。nameは英数字・ハイフン・アンダースコアのみ。",
			},
			{
				Role:    "user",
				Content: overviewContent,
			},
		},
		Temperature: 0.2,
	}

	payload, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, buildLLMEndpoint(s.endpoint), bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if strings.TrimSpace(s.apiKey) != "" {
		req.Header.Set("Authorization", "Bearer "+s.apiKey)
	}

	res, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("llm request failed: status=%d", res.StatusCode)
	}

	var response chatCompletionsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}
	if len(response.Choices) == 0 {
		return nil, nil
	}

	aiItems, err := parseAICandidates(response.Choices[0].Message.Content)
	if err != nil {
		return nil, err
	}

	out := make([]model.ProjectNameCandidate, 0, len(aiItems))
	for _, item := range aiItems {
		out = append(out, model.ProjectNameCandidate{
			Name:        item.Name,
			Reason:      item.Reason,
			AISuggested: true,
		})
	}
	return out, nil
}

func (s *openAIProjectNameSuggester) GetProfile() model.NameSuggestionProfile {
	return model.NameSuggestionProfile{
		Profile:           s.getCurrentProfile(),
		AvailableProfiles: []string{"fast", "balanced", "quality"},
		Enabled:           true,
	}
}

func (s *openAIProjectNameSuggester) SetProfile(profile string) error {
	p := strings.ToLower(strings.TrimSpace(profile))
	if p != "fast" && p != "balanced" && p != "quality" {
		return fmt.Errorf("unsupported profile")
	}
	s.mu.Lock()
	s.currentProfile = p
	s.mu.Unlock()
	return nil
}

func (s *openAIProjectNameSuggester) getCurrentProfile() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.currentProfile == "" {
		return "balanced"
	}
	return s.currentProfile
}

func loadLLMModelConfigFromEnv() llmModelConfig {
	legacy := strings.TrimSpace(os.Getenv("MUSUHI_LLM_MODEL"))
	if legacy == "" {
		legacy = "gpt-4o-mini"
	}

	fast := strings.TrimSpace(os.Getenv("MUSUHI_LLM_MODEL_FAST"))
	if fast == "" {
		fast = legacy
	}

	balanced := strings.TrimSpace(os.Getenv("MUSUHI_LLM_MODEL_BALANCED"))
	if balanced == "" {
		balanced = fast
	}

	quality := strings.TrimSpace(os.Getenv("MUSUHI_LLM_MODEL_QUALITY"))
	if quality == "" {
		quality = balanced
	}

	profile := strings.ToLower(strings.TrimSpace(os.Getenv("MUSUHI_LLM_PROFILE")))
	if profile == "" {
		profile = "balanced"
	}

	return llmModelConfig{
		FastModel:     fast,
		BalancedModel: balanced,
		QualityModel:  quality,
		Profile:       profile,
	}
}

func resolveModelByProfile(config llmModelConfig, profile string, overviewContent string) string {
	if profile == "auto" {
		profile = inferProfileFromOverview(overviewContent)
	}

	switch profile {
	case "fast":
		return config.FastModel
	case "quality":
		return config.QualityModel
	case "balanced":
		fallthrough
	default:
		return config.BalancedModel
	}
}

func inferProfileFromOverview(content string) string {
	lower := strings.ToLower(content)
	if containsAny(lower, []string{"最終", "レビュー", "監査", "品質", "回帰", "リスク", "厳密"}) {
		return "quality"
	}
	if containsAny(lower, []string{"設計", "分析", "推論", "複雑", "最適化", "architecture", "design"}) {
		return "balanced"
	}
	return "fast"
}

func buildLLMEndpoint(base string) string {
	trimmed := strings.TrimRight(strings.TrimSpace(base), "/")
	if strings.HasSuffix(trimmed, "/v1/chat/completions") {
		return trimmed
	}
	return trimmed + "/v1/chat/completions"
}

func parseAICandidates(content string) ([]struct {
	Name   string
	Reason string
}, error) {
	raw := strings.TrimSpace(content)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	var envelope aiCandidateEnvelope
	if err := json.Unmarshal([]byte(raw), &envelope); err != nil {
		return nil, err
	}

	items := make([]struct {
		Name   string
		Reason string
	}, 0, len(envelope.Items))
	for _, item := range envelope.Items {
		items = append(items, struct {
			Name   string
			Reason string
		}{
			Name:   item.Name,
			Reason: item.Reason,
		})
	}
	return items, nil
}

func validateAIProjectNameCandidates(values []model.ProjectNameCandidate) []model.ProjectNameCandidate {
	out := make([]model.ProjectNameCandidate, 0, len(values))
	for _, value := range values {
		name := strings.ToLower(strings.TrimSpace(value.Name))
		reason := strings.TrimSpace(value.Reason)
		if name == "" || reason == "" {
			continue
		}
		if !projectNamePattern.MatchString(name) {
			continue
		}
		if len([]rune(reason)) < 8 {
			continue
		}
		out = append(out, model.ProjectNameCandidate{
			Name:        name,
			Reason:      reason,
			AISuggested: true,
		})
		if len(out) >= 3 {
			break
		}
	}
	return uniqueProjectNameCandidates(out)
}
