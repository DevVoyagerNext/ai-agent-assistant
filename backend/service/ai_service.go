package service

import (
	"backend/dto"
	"backend/global"
	"backend/model"
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AIService struct{}

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
func (s *AIService) Chat(ctx context.Context, userId uint, req dto.AIChatReq, fileContents []string) (<-chan dto.ChatStreamChunk, int64, int64, error) {
	aiConfig := global.GVA_CONFIG.AI
	if aiConfig.APIKey == "" || aiConfig.BaseURL == "" {
		global.GVA_LOG.Error("AI config is missing")
		return nil, 0, 0, errors.New("AI 服务未配置")
	}

	config := openai.DefaultConfig(aiConfig.APIKey)
	config.BaseURL = aiConfig.BaseURL

	client := openai.NewClientWithConfig(config)

	var session model.Session
	db := global.GVA_DB.WithContext(ctx)
	isNewSession := false

	if req.SessionID == 0 {
		// 1. 无 sessionId 时创建新会话
		isNewSession = true
		session = model.Session{
			UserID:  int64(userId),
			Title:   "新对话",
			ModelID: aiConfig.Model,
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
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemPrompt,
		},
	}

	// 获取最近4轮历史对话 (8条消息)
	if !isNewSession {
		var historyMsgs []model.Message
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

		for _, m := range historyMsgs {
			messages = append(messages, openai.ChatCompletionMessage{
				Role:    m.Role,
				Content: m.Content,
			})
		}
	}

	// 处理最终的用户 Prompt（拼接文件内容）
	finalPrompt := req.Prompt
	if len(fileContents) > 0 {
		finalPrompt += "\n\n【用户上传的文件内容如下】\n"
		for _, fc := range fileContents {
			finalPrompt += fc + "\n"
		}
	}

	// 加入当前用户的 prompt
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: finalPrompt,
	})

	// 保存用户消息
	userMsg := model.Message{
		SessionID: session.ID,
		ParentID:  req.ParentID,
		Role:      openai.ChatMessageRoleUser,
		Content:   finalPrompt,
	}
	db.Create(&userMsg)

	// 保存 AI 回复消息占位符
	aiMsg := model.Message{
		SessionID: session.ID,
		ParentID:  &userMsg.ID,
		Role:      openai.ChatMessageRoleAssistant,
		Content:   "", // 留空，流式输出完成后再更新
	}
	db.Create(&aiMsg)

	// 调用 AI (流式)
	reqStream := openai.ChatCompletionRequest{
		Model:    aiConfig.Model,
		Messages: messages,
		Stream:   true,
	}

	stream, err := client.CreateChatCompletionStream(ctx, reqStream)
	if err != nil {
		global.GVA_LOG.Error("AI API stream call failed", zap.Error(err))
		return nil, session.ID, aiMsg.ID, errors.New("AI 服务调用失败")
	}

	msgChan := make(chan dto.ChatStreamChunk)

	go func() {
		defer stream.Close()
		defer close(msgChan)

		var fullReply string

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

			if len(response.Choices) > 0 {
				delta := response.Choices[0].Delta

				// 处理思考过程 (Reasoning Content)
				if delta.ReasoningContent != "" {
					msgChan <- dto.ChatStreamChunk{Type: "reasoning", Content: delta.ReasoningContent}
				}

				// 处理正式回复 (Content)
				if delta.Content != "" {
					fullReply += delta.Content
					msgChan <- dto.ChatStreamChunk{Type: "message", Content: delta.Content}
				}
			}
		}

		// 流式结束或异常中断后，更新完整内容到数据库
		if fullReply != "" {
			global.GVA_DB.Model(&model.Message{}).Where("id = ?", aiMsg.ID).Update("content", fullReply)
		}

		// 异步任务处理：新会话生成标题，老会话检查是否需要生成摘要
		if isNewSession {
			go func(sId int64, prompt string) {
				titleCtx := context.Background()
				titlePrompt := fmt.Sprintf("请根据以下用户的提问，生成一个简短的对话标题（不超过15个字），不要包含任何标点符号：\n\n%s", prompt)
				titleResp, err := client.CreateChatCompletion(
					titleCtx,
					openai.ChatCompletionRequest{
						Model: aiConfig.Model,
						Messages: []openai.ChatCompletionMessage{
							{Role: openai.ChatMessageRoleUser, Content: titlePrompt},
						},
					},
				)
				if err == nil && len(titleResp.Choices) > 0 {
					title := titleResp.Choices[0].Message.Content
					global.GVA_DB.Model(&model.Session{}).Where("id = ?", sId).Update("title", title)
				}
			}(session.ID, req.Prompt)
		} else {
			// 检查是否需要生成摘要 (经过每4轮对话)
			var count int64
			query := db.Model(&model.Message{}).Where("session_id = ? AND role = ?", session.ID, openai.ChatMessageRoleUser)
			if session.LastSummaryMessageID > 0 {
				query = query.Where("id > ?", session.LastSummaryMessageID)
			}
			query.Count(&count)

			if count >= 4 {
				// 触发摘要总结
				go func(s model.Session, c *openai.Client, latestUserMsgId int64) {
					summaryCtx := context.Background()
					// 获取自上次总结以来的新对话
					var msgsToSummarize []model.Message
					q := global.GVA_DB.Where("session_id = ? AND status = 'active'", s.ID)
					if s.LastSummaryMessageID > 0 {
						q = q.Where("id > ?", s.LastSummaryMessageID)
					}
					q.Order("created_at asc").Find(&msgsToSummarize)

					contentToSummarize := ""
					if s.Summary != "" {
						contentToSummarize += fmt.Sprintf("【之前的对话背景摘要】\n%s\n\n", s.Summary)
					}
					contentToSummarize += "【最新的对话记录】\n"
					for _, m := range msgsToSummarize {
						roleName := "用户"
						if m.Role == openai.ChatMessageRoleAssistant {
							roleName = "AI"
						}
						contentToSummarize += fmt.Sprintf("[%s]: %s\n", roleName, m.Content)
					}
					contentToSummarize += "\n请结合【之前的对话背景摘要】和【最新的对话记录】，重新生成一份全局的简明摘要。提取核心背景、用户意图和关键结论，以便作为后续对话的上下文。注意：总结字数请严格控制在500字以内！"

					summaryResp, err := c.CreateChatCompletion(
						summaryCtx,
						openai.ChatCompletionRequest{
							Model: aiConfig.Model,
							Messages: []openai.ChatCompletionMessage{
								{Role: openai.ChatMessageRoleUser, Content: contentToSummarize},
							},
						},
					)
					if err == nil && len(summaryResp.Choices) > 0 {
						newSummary := summaryResp.Choices[0].Message.Content
						global.GVA_DB.Model(&model.Session{}).Where("id = ?", s.ID).Updates(map[string]interface{}{
							"summary":                 newSummary,
							"last_summary_message_id": latestUserMsgId,
						})
					}
				}(session, client, userMsg.ID)
			}
		}
	}()

	return msgChan, session.ID, aiMsg.ID, nil
}
