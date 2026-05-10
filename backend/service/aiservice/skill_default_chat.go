package aiservice

import (
	"backend/global"
	"backend/model"
	"errors"
	"io"
	"strings"

	"github.com/cloudwego/eino/schema"
	"go.uber.org/zap"
)

type DefaultChatSkillHandler struct{}

// CanHandle 始终返回 true，因为它是默认的聊天兜底处理器
func (h *DefaultChatSkillHandler) CanHandle(skillID string) bool {
	return true
}

// Handle 执行默认的单轮或多轮流式对话处理
func (h *DefaultChatSkillHandler) Handle(ctx *SkillContext, s *AIService) error {
	chatAgent, err := s.newChatAgent(ctx.AgentCtx, ctx.UserID)
	if err != nil {
		return err
	}

	systemPrompt := s.buildChatSystemPrompt(*ctx.Session, strings.TrimSpace(ctx.Req.CurrentPageURL) != "", ctx.Req.SkillID)

	// 组装消息列表
	// 获取最近4轮历史对话 (8条消息)
	var historyMsgs []model.Message
	if !ctx.IsNewSession {
		historyMsgs, _ = s.dao.GetRecentMessages(ctx.AgentCtx, ctx.Session.ID, 8)
	}

	userPrompt := s.buildUserPrompt(ctx.Req)
	messages := s.buildConversationMessages(systemPrompt, historyMsgs, userPrompt)

	go func() {
		defer close(ctx.MsgChan)
		streamResp, err := chatAgent.Stream(ctx.AgentCtx, messages)
		if err != nil {
			global.GVA_LOG.Error("Eino AI agent stream failed", zap.Error(err))
			fallbackReply := "工具链执行失败，请稍后重试"
			s.emitStreamChunks(ctx.MsgChan, "message", fallbackReply)
			s.dao.UpdateMessageContent(ctx.AgentCtx, ctx.AIMsg.ID, fallbackReply)
			s.finalizeChatSideEffects(ctx.DB, *ctx.Session, ctx.IsNewSession, ctx.PromptContent, ctx.UserMsg.ID)
			return
		}
		defer streamResp.Close()

		streamMsgs := make([]*schema.Message, 0, 32)
		for {
			chunk, err := streamResp.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				global.GVA_LOG.Error("Eino AI agent recv stream failed", zap.Error(err))
				fallbackReply := "AI 流式输出中断，请稍后重试"
				s.emitStreamChunks(ctx.MsgChan, "message", fallbackReply)
				s.dao.UpdateMessageContent(ctx.AgentCtx, ctx.AIMsg.ID, fallbackReply)
				s.finalizeChatSideEffects(ctx.DB, *ctx.Session, ctx.IsNewSession, ctx.PromptContent, ctx.UserMsg.ID)
				return
			}

			streamMsgs = append(streamMsgs, chunk)
			s.emitMessageChunk(ctx.MsgChan, chunk)
		}

		if len(streamMsgs) == 0 {
			fallbackReply := "AI 未生成最终答案，请稍后重试"
			s.emitStreamChunks(ctx.MsgChan, "message", fallbackReply)
			s.dao.UpdateMessageContent(ctx.AgentCtx, ctx.AIMsg.ID, fallbackReply)
			s.finalizeChatSideEffects(ctx.DB, *ctx.Session, ctx.IsNewSession, ctx.PromptContent, ctx.UserMsg.ID)
			return
		}

		finalResp, err := schema.ConcatMessages(streamMsgs)
		if err != nil {
			global.GVA_LOG.Error("Concat AI stream chunks failed", zap.Error(err))
			fallbackReply := "AI 流式结果合并失败，请稍后重试"
			s.emitStreamChunks(ctx.MsgChan, "message", fallbackReply)
			s.dao.UpdateMessageContent(ctx.AgentCtx, ctx.AIMsg.ID, fallbackReply)
			s.finalizeChatSideEffects(ctx.DB, *ctx.Session, ctx.IsNewSession, ctx.PromptContent, ctx.UserMsg.ID)
			return
		}

		messageText := s.extractMessageText(finalResp)
		if strings.TrimSpace(messageText) == "" {
			fallbackReply := "AI 未生成最终答案，请稍后重试"
			s.emitStreamChunks(ctx.MsgChan, "message", fallbackReply)
			s.dao.UpdateMessageContent(ctx.AgentCtx, ctx.AIMsg.ID, fallbackReply)
			s.finalizeChatSideEffects(ctx.DB, *ctx.Session, ctx.IsNewSession, ctx.PromptContent, ctx.UserMsg.ID)
			return
		}

		s.dao.UpdateMessageContent(ctx.AgentCtx, ctx.AIMsg.ID, messageText)
		s.finalizeChatSideEffects(ctx.DB, *ctx.Session, ctx.IsNewSession, ctx.PromptContent, ctx.UserMsg.ID)
	}()

	return nil
}
