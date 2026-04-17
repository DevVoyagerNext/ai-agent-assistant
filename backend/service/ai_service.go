package service

import (
	"backend/global"
	"context"
	"errors"

	"github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
)

type AIService struct{}

// Chat 处理与 AI 模型的单次对话
func (s *AIService) Chat(ctx context.Context, prompt string) (string, error) {
	aiConfig := global.GVA_CONFIG.AI
	if aiConfig.APIKey == "" || aiConfig.BaseURL == "" {
		global.GVA_LOG.Error("AI config is missing")
		return "", errors.New("AI 服务未配置")
	}

	config := openai.DefaultConfig(aiConfig.APIKey)
	config.BaseURL = aiConfig.BaseURL

	client := openai.NewClientWithConfig(config)

	resp, err := client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: aiConfig.Model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)

	if err != nil {
		global.GVA_LOG.Error("AI API call failed", zap.Error(err))
		return "", errors.New("AI 服务调用失败")
	}

	if len(resp.Choices) > 0 {
		return resp.Choices[0].Message.Content, nil
	}

	return "", errors.New("AI 返回结果为空")
}
