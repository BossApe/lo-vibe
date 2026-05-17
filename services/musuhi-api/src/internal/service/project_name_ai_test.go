package service

import "testing"

func TestResolveModelByProfile_固定プロファイルで対応モデルを返す_正常系(t *testing.T) {
	config := llmModelConfig{
		FastModel:     "Claude Haiku 4.5",
		BalancedModel: "Claude Sonnet 4.6",
		QualityModel:  "GPT-5.4",
	}
	if got := resolveModelByProfile(config, "fast", "通常の候補生成"); got != "Claude Haiku 4.5" {
		t.Fatalf("expected fast model, got=%s", got)
	}

	if got := resolveModelByProfile(config, "balanced", "設計検討"); got != "Claude Sonnet 4.6" {
		t.Fatalf("expected balanced model, got=%s", got)
	}

	if got := resolveModelByProfile(config, "quality", "最終レビュー"); got != "GPT-5.4" {
		t.Fatalf("expected quality model, got=%s", got)
	}
}

func TestResolveModelByProfile_autoは概要文から推定する_正常系(t *testing.T) {
	config := llmModelConfig{
		FastModel:     "Claude Haiku 4.5",
		BalancedModel: "Claude Sonnet 4.6",
		QualityModel:  "GPT-5.4",
	}

	if got := resolveModelByProfile(config, "auto", "最終レビューで回帰リスクを監査したい"); got != "GPT-5.4" {
		t.Fatalf("expected quality model in auto, got=%s", got)
	}

	if got := resolveModelByProfile(config, "auto", "複雑な設計分析をしたい"); got != "Claude Sonnet 4.6" {
		t.Fatalf("expected balanced model in auto, got=%s", got)
	}

	if got := resolveModelByProfile(config, "auto", "簡単な候補名を出したい"); got != "Claude Haiku 4.5" {
		t.Fatalf("expected fast model in auto, got=%s", got)
	}
}
