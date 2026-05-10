package aiservice

import (
	"backend/dto"
	"backend/model"
	"context"

	"gorm.io/gorm"
)

// SkillContext 封装了技能处理器（Skill Handler）执行所需的所有上下文信息
type SkillContext struct {
	AgentCtx      context.Context          // Agent 运行时的 Context，包含追踪或取消信号
	Req           dto.AIChatReq            // 用户发起的 AI 聊天请求参数
	Session       *model.Session           // 当前的会话信息
	UserMsg       *model.Message           // 当前用户发送的消息记录
	AIMsg         *model.Message           // 预先创建的 AI 回复消息占位符（后续会被更新）
	IsNewSession  bool                     // 标识当前请求是否创建了新会话
	PromptContent string                   // 提取出的用户实际 Prompt 文本内容
	DB            *gorm.DB                 // 数据库事务或连接对象
	MsgChan       chan dto.ChatStreamChunk // 用于向前端流式推送响应数据块的通道
	UserID        uint                     // 当前请求用户的 ID
}

// SkillHandler 定义了不同 AI Agent 或技能的通用接口
type SkillHandler interface {
	// CanHandle 判断当前处理器是否能处理指定的 skillID
	CanHandle(skillID string) bool
	// Handle 执行具体的技能处理逻辑
	Handle(skillCtx *SkillContext, s *AIService) error
}

// defaultSkillRegistry 保存所有已注册的技能处理器列表
var defaultSkillRegistry []SkillHandler

// RegisterSkillHandler 将一个实现了 SkillHandler 接口的技能处理器注册到全局注册表中
func RegisterSkillHandler(handler SkillHandler) {
	defaultSkillRegistry = append(defaultSkillRegistry, handler)
}

// getSkillHandler 根据传入的 skillID 遍历注册表，返回第一个能够处理该请求的技能处理器；若无匹配则返回 nil
func getSkillHandler(skillID string) SkillHandler {
	for _, handler := range defaultSkillRegistry {
		if handler.CanHandle(skillID) {
			return handler
		}
	}
	return nil
}
