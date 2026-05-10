package aiservice

import (
	"backend/global"
	"strings"

	"go.uber.org/zap"
)

type ResumeInterviewSkillHandler struct{}

// CanHandle 根据 skillID 判定是否由“简历面试”处理器进行接管
func (h *ResumeInterviewSkillHandler) CanHandle(skillID string) bool {
	return isResumeInterviewSkill(skillID)
}

// Handle 执行具体的简历面试相关 Agent 的处理逻辑
func (h *ResumeInterviewSkillHandler) Handle(ctx *SkillContext, s *AIService) error {
	go func() {
		defer close(ctx.MsgChan)

		bundle, err := s.prepareResumeKnowledge(ctx.AgentCtx, ctx.Req)
		if err != nil {
			fallbackReply := buildResumeFallbackReply(err)
			s.emitStreamChunks(ctx.MsgChan, "message", fallbackReply)
			s.dao.UpdateMessageContent(ctx.AgentCtx, ctx.AIMsg.ID, fallbackReply)
			s.finalizeChatSideEffects(ctx.DB, *ctx.Session, ctx.IsNewSession, ctx.PromptContent, ctx.UserMsg.ID)
			return
		}
		if err := validateResumeKnowledgeBundle(bundle); err != nil {
			fallbackReply := buildResumeFallbackReply(err)
			s.emitStreamChunks(ctx.MsgChan, "message", fallbackReply)
			s.dao.UpdateMessageContent(ctx.AgentCtx, ctx.AIMsg.ID, fallbackReply)
			s.finalizeChatSideEffects(ctx.DB, *ctx.Session, ctx.IsNewSession, ctx.PromptContent, ctx.UserMsg.ID)
			return
		}

		finalAnswer, err := s.executeResumeInterviewAgents(ctx.AgentCtx, ctx.Req, bundle)
		if err != nil {
			global.GVA_LOG.Error("resume interview multi-agent failed", zap.Error(err))
			fallbackReply := "抱歉，简历面试 Agent 执行过程中出现错误，未能成功生成面试题，请稍后重试。"
			s.emitStreamChunks(ctx.MsgChan, "message", fallbackReply)
			s.dao.UpdateMessageContent(ctx.AgentCtx, ctx.AIMsg.ID, fallbackReply)
			s.finalizeChatSideEffects(ctx.DB, *ctx.Session, ctx.IsNewSession, ctx.PromptContent, ctx.UserMsg.ID)
			return
		}

		if strings.TrimSpace(finalAnswer) == "" {
			fallbackReply := "抱歉，简历面试 Agent 未能生成有效的面试题内容，请稍后重试或尝试补充更多要求。"
			s.emitStreamChunks(ctx.MsgChan, "message", fallbackReply)
			s.dao.UpdateMessageContent(ctx.AgentCtx, ctx.AIMsg.ID, fallbackReply)
			s.finalizeChatSideEffects(ctx.DB, *ctx.Session, ctx.IsNewSession, ctx.PromptContent, ctx.UserMsg.ID)
			return
		}

		s.emitStreamChunks(ctx.MsgChan, "message", finalAnswer)
		s.dao.UpdateMessageContent(ctx.AgentCtx, ctx.AIMsg.ID, finalAnswer)
		s.finalizeChatSideEffects(ctx.DB, *ctx.Session, ctx.IsNewSession, ctx.PromptContent, ctx.UserMsg.ID)
	}()
	return nil
}

func init() {
	RegisterSkillHandler(&ResumeInterviewSkillHandler{})
}
