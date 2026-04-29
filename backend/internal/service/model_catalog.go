package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type ModelCatalogConfig struct {
	Categories []ModelCatalogCategory `json:"categories"`
}

type ModelCatalogCategory struct {
	ID          string   `json:"id"`
	Label       string   `json:"label"`
	Description string   `json:"description,omitempty"`
	Models      []string `json:"models"`
}

var defaultModelCatalogConfig = ModelCatalogConfig{Categories: []ModelCatalogCategory{
	{
		ID:          "openai",
		Label:       "OpenAI 兼容 / 奥特曼加",
		Description: "本站当前主推的奥特曼加模型，适配 OpenAI 兼容接口。",
		Models: []string{
			"gpt-5.2",
			"gpt-5.3-codex",
			"gpt-5.3-codex-spark",
			"gpt-5.4",
			"gpt-5.4-mini",
			"gpt-5.5",
			"gpt-image-2",
		},
	},
}}

func DefaultModelCatalogConfig() ModelCatalogConfig {
	return defaultModelCatalogConfig
}

func NormalizeModelCatalogConfig(config ModelCatalogConfig) ModelCatalogConfig {
	categories := make([]ModelCatalogCategory, 0, len(config.Categories))
	for i, category := range config.Categories {
		models := make([]string, 0, len(category.Models))
		seen := make(map[string]struct{}, len(category.Models))
		for _, model := range category.Models {
			model = strings.TrimSpace(model)
			if model == "" {
				continue
			}
			if _, ok := seen[model]; ok {
				continue
			}
			seen[model] = struct{}{}
			models = append(models, model)
		}
		if len(models) == 0 {
			continue
		}
		id := strings.TrimSpace(category.ID)
		if id == "" {
			id = fmt.Sprintf("category-%d", i+1)
		}
		label := strings.TrimSpace(category.Label)
		if label == "" {
			label = id
		}
		categories = append(categories, ModelCatalogCategory{
			ID:          id,
			Label:       label,
			Description: strings.TrimSpace(category.Description),
			Models:      models,
		})
	}
	if len(categories) == 0 {
		return DefaultModelCatalogConfig()
	}
	return ModelCatalogConfig{Categories: categories}
}

func (s *SettingService) GetModelCatalog(ctx context.Context) (ModelCatalogConfig, error) {
	raw, err := s.settingRepo.GetValue(ctx, SettingKeyModelCatalog)
	if err != nil || strings.TrimSpace(raw) == "" {
		return DefaultModelCatalogConfig(), nil
	}
	var config ModelCatalogConfig
	if err := json.Unmarshal([]byte(raw), &config); err != nil {
		return DefaultModelCatalogConfig(), nil
	}
	return NormalizeModelCatalogConfig(config), nil
}

func (s *SettingService) UpdateModelCatalog(ctx context.Context, config ModelCatalogConfig) (ModelCatalogConfig, error) {
	normalized := NormalizeModelCatalogConfig(config)
	payload, err := json.Marshal(normalized)
	if err != nil {
		return ModelCatalogConfig{}, err
	}
	if err := s.settingRepo.Set(ctx, SettingKeyModelCatalog, string(payload)); err != nil {
		return ModelCatalogConfig{}, err
	}
	return normalized, nil
}
