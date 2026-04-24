package service

import (
	"backend/dto"
	"backend/global"
	"backend/model"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	einoOpenAI "github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AIService struct{}

func (s *AIService) newChatModel(ctx context.Context) (*einoOpenAI.ChatModel, error) {
	aiConfig := global.GVA_CONFIG.AI
	if aiConfig.APIKey == "" || aiConfig.BaseURL == "" {
		global.GVA_LOG.Error("AI config is missing")
		return nil, errors.New("AI 服务未配置")
	}

	chatModel, err := einoOpenAI.NewChatModel(ctx, &einoOpenAI.ChatModelConfig{
		APIKey:  aiConfig.APIKey,
		BaseURL: aiConfig.BaseURL,
		Model:   aiConfig.Model,
	})
	if err != nil {
		global.GVA_LOG.Error("Failed to initialize Eino chat model", zap.Error(err))
		return nil, errors.New("初始化 AI 模型失败")
	}

	return chatModel, nil
}

func (s *AIService) buildConversationMessages(systemPrompt string, historyMsgs []model.Message, prompt string) []*schema.Message {
	messages := []*schema.Message{
		schema.SystemMessage(systemPrompt),
	}

	for _, historyMsg := range historyMsgs {
		role := schema.RoleType(historyMsg.Role)
		if role == "" {
			continue
		}
		messages = append(messages, &schema.Message{
			Role:    role,
			Content: historyMsg.Content,
		})
	}

	messages = append(messages, schema.UserMessage(prompt))
	return messages
}

func (s *AIService) extractMessageText(msg *schema.Message) string {
	if msg == nil {
		return ""
	}
	if msg.Content != "" {
		return msg.Content
	}

	var builder strings.Builder
	for _, part := range msg.AssistantGenMultiContent {
		if part.Type == schema.ChatMessagePartTypeText && part.Text != "" {
			builder.WriteString(part.Text)
		}
	}
	return builder.String()
}

func (s *AIService) extractReasoningText(msg *schema.Message) string {
	if msg == nil {
		return ""
	}
	if msg.ReasoningContent != "" {
		return msg.ReasoningContent
	}

	var builder strings.Builder
	for _, part := range msg.AssistantGenMultiContent {
		if part.Type == schema.ChatMessagePartTypeReasoning && part.Reasoning != nil {
			builder.WriteString(part.Reasoning.Text)
		}
	}
	return builder.String()
}

func (s *AIService) generateText(ctx context.Context, messages []*schema.Message) (string, error) {
	chatModel, err := s.newChatModel(ctx)
	if err != nil {
		return "", err
	}

	resp, err := chatModel.Generate(ctx, messages)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(s.extractMessageText(resp)), nil
}

// UpdateSessionTitle 修改会话标题
func (s *AIService) UpdateSessionTitle(ctx context.Context, userId uint, sessionId int64, title string) error {
	db := global.GVA_DB.WithContext(ctx)
	res := db.Model(&model.Session{}).
		Where("id = ? AND user_id = ? AND is_deleted = false", sessionId, userId).
		Update("title", title)

	if res.Error != nil {
		return errors.New("更新标题失败")
	}
	if res.RowsAffected == 0 {
		return errors.New("会话不存在或无权修改")
	}
	return nil
}

// GetUserSessions 获取用户的历史会话列表（游标分页）
func (s *AIService) GetUserSessions(ctx context.Context, userId uint, lastId int64) (dto.SessionListRes, error) {
	db := global.GVA_DB.WithContext(ctx)
	var sessions []model.Session
	limit := 10

	query := db.Model(&model.Session{}).Where("user_id = ? AND is_deleted = false", userId)

	// 如果传了 lastId，则需要找到 lastId 对应的 updated_at，并获取比这个时间更早的会话
	if lastId > 0 {
		var lastSession model.Session
		if err := db.Where("id = ? AND user_id = ?", lastId, userId).First(&lastSession).Error; err == nil {
			// 根据 updated_at 倒序排列。如果 updated_at 相同，则根据 id 倒序保证稳定性
			query = query.Where("(updated_at < ?) OR (updated_at = ? AND id < ?)", lastSession.UpdatedAt, lastSession.UpdatedAt, lastId)
		}
	}

	// 多查一条用于判断是否有下一页
	if err := query.Order("updated_at desc, id desc").Limit(limit + 1).Find(&sessions).Error; err != nil {
		return dto.SessionListRes{}, err
	}

	hasMore := false
	if len(sessions) > limit {
		hasMore = true
		sessions = sessions[:limit] // 截断最后一条
	}

	var list []dto.SessionItemRes
	for _, session := range sessions {
		list = append(list, dto.SessionItemRes{
			ID:        session.ID,
			Title:     session.Title,
			ModelID:   session.ModelID,
			UpdatedAt: session.UpdatedAt.Format(time.RFC3339),
			CreatedAt: session.CreatedAt.Format(time.RFC3339),
		})
	}

	return dto.SessionListRes{
		List:    list,
		HasMore: hasMore,
	}, nil
}

// GetSessionMessages 获取指定会话的消息列表（游标分页）
func (s *AIService) GetSessionMessages(ctx context.Context, userId uint, sessionId int64, lastId int64) (dto.MessageListRes, error) {
	db := global.GVA_DB.WithContext(ctx)

	// 校验会话所属权
	var session model.Session
	if err := db.Where("id = ? AND user_id = ? AND is_deleted = false", sessionId, userId).First(&session).Error; err != nil {
		return dto.MessageListRes{}, errors.New("会话不存在或无权访问")
	}

	var messages []model.Message
	limit := 10

	query := db.Where("session_id = ? AND status = 'active'", sessionId)
	if lastId > 0 {
		query = query.Where("id < ?", lastId)
	}

	// 取最新的 10 条（按 ID 降序），多查一条用于判断 hasMore
	if err := query.Order("id desc").Limit(limit + 1).Find(&messages).Error; err != nil {
		return dto.MessageListRes{}, err
	}

	hasMore := false
	if len(messages) > limit {
		hasMore = true
		messages = messages[:limit]
	}

	// 此时 messages 是按照 id 降序排列的，为了让前端展示时顺序正常，需要反转为正序
	for i := len(messages)/2 - 1; i >= 0; i-- {
		opp := len(messages) - 1 - i
		messages[i], messages[opp] = messages[opp], messages[i]
	}

	var list []dto.MessageItemRes
	for _, msg := range messages {
		list = append(list, dto.MessageItemRes{
			ID:        msg.ID,
			SessionID: msg.SessionID,
			ParentID:  msg.ParentID,
			Role:      msg.Role,
			Content:   msg.Content,
			Status:    msg.Status,
			CreatedAt: msg.CreatedAt.Format(time.RFC3339),
		})
	}

	return dto.MessageListRes{
		List:    list,
		HasMore: hasMore,
	}, nil
}

// Chat 处理与 AI 模型的单次对话（流式）
func (s *AIService) Chat(ctx context.Context, userId uint, req dto.AIChatReq) (<-chan dto.ChatStreamChunk, int64, int64, error) {
	chatModel, err := s.newChatModel(ctx)
	if err != nil {
		return nil, 0, 0, err
	}

	var session model.Session
	db := global.GVA_DB.WithContext(ctx)
	isNewSession := false

	if req.SessionID == 0 {
		// 1. 无 sessionId 时创建新会话
		isNewSession = true
		session = model.Session{
			UserID:  int64(userId),
			Title:   "新对话",
			ModelID: global.GVA_CONFIG.AI.Model,
		}
		if err := db.Create(&session).Error; err != nil {
			global.GVA_LOG.Error("Failed to create AI session", zap.Error(err))
			return nil, 0, 0, errors.New("创建会话失败")
		}
	} else {
		// 2. 有 sessionId 时查找会话
		if err := db.Where("id = ? AND user_id = ?", req.SessionID, userId).First(&session).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, 0, 0, errors.New("会话不存在或无权访问")
			}
			return nil, 0, 0, errors.New("查询会话失败")
		}
	}

	// 教师系统提示词
	systemPrompt := `你现在是一位顶尖的全能型高级教师。你不仅学识渊博，更精通教学之道。在回答学生的问题时，请务必做到以下几点：
1. 深入浅出：知识讲解既要有专业的深度与丰富的细节，又要通俗易懂；
2. 逻辑严密：条理极其清晰，结构分明，善于使用序号、分类或对比来组织内容；
3. 启发思考：不仅给出答案，还要引导学生思考背后的原理，培养其举一反三的能力。
请始终保持专业、耐心且富有启发性的教育者语气进行解答。`

	// 如果存在会话摘要，则添加到系统提示词中
	if session.Summary != "" {
		systemPrompt += fmt.Sprintf("\n\n以下是之前的对话摘要，请作为背景参考：\n%s", session.Summary)
	}

	// 组装消息列表
	// 获取最近4轮历史对话 (8条消息)
	var historyMsgs []model.Message
	if !isNewSession {
		// 取最近 8 条（4轮），按时间降序查出后，再反转顺序加入
		db.Where("session_id = ? AND status = 'active'", session.ID).
			Order("created_at desc").
			Limit(8).
			Find(&historyMsgs)

		// 反转，使其按时间正序
		for i := len(historyMsgs)/2 - 1; i >= 0; i-- {
			opp := len(historyMsgs) - 1 - i
			historyMsgs[i], historyMsgs[opp] = historyMsgs[opp], historyMsgs[i]
		}
	}

	messages := s.buildConversationMessages(systemPrompt, historyMsgs, req.Prompt)

	stream, err := chatModel.Stream(ctx, messages)
	if err != nil {
		global.GVA_LOG.Error("Eino AI stream call failed", zap.Error(err))
		if isNewSession {
			db.Unscoped().Where("id = ?", session.ID).Delete(&model.Session{})
		}
		return nil, 0, 0, errors.New("AI 服务调用失败")
	}

	// 阻塞读取第一个数据块，确保 AI 真的有响应，避免空占位存入数据库
	firstResponse, err := stream.Recv()
	if err != nil {
		stream.Close()
		global.GVA_LOG.Error("Eino AI stream receive failed on first chunk", zap.Error(err))
		if isNewSession {
			db.Unscoped().Where("id = ?", session.ID).Delete(&model.Session{})
		}
		return nil, 0, 0, errors.New("AI 服务响应失败")
	}

	// 确认收到第一块数据（成功连接大模型）后，再将用户消息和 AI 占位消息存入数据库
	userMsg := model.Message{
		SessionID: session.ID,
		ParentID:  req.ParentID,
		Role:      string(schema.User),
		Content:   req.Prompt,
	}
	db.Create(&userMsg)

	// 保存 AI 回复消息占位符
	aiMsg := model.Message{
		SessionID: session.ID,
		ParentID:  &userMsg.ID,
		Role:      string(schema.Assistant),
		Content:   "", // 留空，流式输出完成后再更新
	}
	db.Create(&aiMsg)

	msgChan := make(chan dto.ChatStreamChunk)

	go func() {
		defer stream.Close()
		defer close(msgChan)

		var fullReply string

		// 先处理已经收到的第一个 chunk
		if reasoning := s.extractReasoningText(firstResponse); reasoning != "" {
			msgChan <- dto.ChatStreamChunk{Type: "reasoning", Content: reasoning}
		}
		if content := s.extractMessageText(firstResponse); content != "" {
			fullReply += content
			msgChan <- dto.ChatStreamChunk{Type: "message", Content: content}
		}

		for {
			response, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				// 即使出错（比如 context canceled），也保存已经生成的部分内容
				global.GVA_LOG.Warn("AI stream interrupted", zap.Error(err))
				break
			}

			// 处理思考过程 (Reasoning Content)
			if reasoning := s.extractReasoningText(response); reasoning != "" {
				msgChan <- dto.ChatStreamChunk{Type: "reasoning", Content: reasoning}
			}

			// 处理正式回复 (Content)
			if content := s.extractMessageText(response); content != "" {
				fullReply += content
				msgChan <- dto.ChatStreamChunk{Type: "message", Content: content}
			}
		}

		// 流式结束或异常中断后，更新完整内容到数据库
		if fullReply != "" {
			global.GVA_DB.Model(&model.Message{}).Where("id = ?", aiMsg.ID).Update("content", fullReply)
		} else {
			// 如果 AI 没有任何回复（发送失败或异常中断），清理掉刚才占位的数据库记录
			global.GVA_DB.Unscoped().Where("id IN ?", []int64{userMsg.ID, aiMsg.ID}).Delete(&model.Message{})
			// 如果是新建的会话，且第一条消息就失败了，同时清理空会话
			if isNewSession {
				global.GVA_DB.Unscoped().Where("id = ?", session.ID).Delete(&model.Session{})
			}
			return // 直接返回，不再执行后续的生成标题或摘要逻辑
		}

		// 异步任务处理：新会话生成标题，老会话检查是否需要生成摘要
		if isNewSession {
			go func(sId int64, prompt string) {
				titleCtx := context.Background()
				titlePrompt := fmt.Sprintf("请根据以下用户的提问，生成一个简短的对话标题（不超过15个字），不要包含任何标点符号：\n\n%s", prompt)
				title, err := s.generateText(titleCtx, []*schema.Message{
					schema.UserMessage(titlePrompt),
				})
				if err == nil && title != "" {
					global.GVA_DB.Model(&model.Session{}).Where("id = ?", sId).Update("title", title)
				}
			}(session.ID, req.Prompt)
		} else {
			// 检查是否需要生成摘要 (经过每4轮对话)
			var count int64
			query := db.Model(&model.Message{}).Where("session_id = ? AND role = ?", session.ID, string(schema.User))
			if session.LastSummaryMessageID > 0 {
				query = query.Where("id > ?", session.LastSummaryMessageID)
			}
			query.Count(&count)

			if count >= 4 {
				// 触发摘要总结
				go func(sessionModel model.Session, latestUserMsgId int64) {
					summaryCtx := context.Background()
					// 获取自上次总结以来的新对话
					var msgsToSummarize []model.Message
					q := global.GVA_DB.Where("session_id = ? AND status = 'active'", sessionModel.ID)
					if sessionModel.LastSummaryMessageID > 0 {
						q = q.Where("id > ?", sessionModel.LastSummaryMessageID)
					}
					q.Order("created_at asc").Find(&msgsToSummarize)

					contentToSummarize := ""
					if sessionModel.Summary != "" {
						contentToSummarize += fmt.Sprintf("【之前的对话背景摘要】\n%s\n\n", sessionModel.Summary)
					}
					contentToSummarize += "【最新的对话记录】\n"
					for _, m := range msgsToSummarize {
						roleName := "用户"
						if m.Role == string(schema.Assistant) {
							roleName = "AI"
						}
						contentToSummarize += fmt.Sprintf("[%s]: %s\n", roleName, m.Content)
					}
					contentToSummarize += "\n请结合【之前的对话背景摘要】和【最新的对话记录】，重新生成一份全局的简明摘要。提取核心背景、用户意图和关键结论，以便作为后续对话的上下文。注意：总结字数请严格控制在500字以内！"

					newSummary, err := s.generateText(summaryCtx, []*schema.Message{
						schema.UserMessage(contentToSummarize),
					})
					if err == nil && newSummary != "" {
						global.GVA_DB.Model(&model.Session{}).Where("id = ?", sessionModel.ID).Updates(map[string]interface{}{
							"summary":                 newSummary,
							"last_summary_message_id": latestUserMsgId,
						})
					}
				}(session, userMsg.ID)
			}
		}
	}()

	return msgChan, session.ID, aiMsg.ID, nil
}
