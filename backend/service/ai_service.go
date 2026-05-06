package service

import (
	"backend/dto"
	"backend/global"
	"backend/model"
	"backend/pkg/utils"
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	einoOpenAI "github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
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

func (s *AIService) getAgentSystemPrompt(agentKey, fallback string) string {
	var agentConfig model.AIAgentConfig
	if err := global.GVA_DB.Where("agent_key = ? AND is_active = ?", strings.TrimSpace(agentKey), 1).First(&agentConfig).Error; err == nil {
		if strings.TrimSpace(agentConfig.SystemPrompt) != "" {
			return strings.TrimSpace(agentConfig.SystemPrompt)
		}
	}
	return strings.TrimSpace(fallback)
}

func (s *AIService) buildChatSystemPrompt(session model.Session, hasCurrentPageURL bool, skillID string) string {
	var agentConfig model.AIAgentConfig

	// 如果没有传 skillID，默认使用 main_agent
	agentKey := "main_agent"
	if strings.TrimSpace(skillID) != "" {
		agentKey = strings.TrimSpace(skillID)
	}

	err := global.GVA_DB.Where("agent_key = ?", agentKey).First(&agentConfig).Error
	if err != nil && agentKey != "main_agent" {
		err = global.GVA_DB.Where("agent_key = ?", "main_agent").First(&agentConfig).Error
	}

	systemPrompt := ""
	if err == nil && agentConfig.SystemPrompt != "" {
		systemPrompt = agentConfig.SystemPrompt
	} else {
		// 默认回退提示词（当数据库未配置或查询失败时使用）
		if isResumeInterviewSkill(skillID) {
			systemPrompt = `你是一名资深技术面试官兼人才评估专家，专门负责根据候选人的简历、项目经历与补充资料生成高质量面试题。

你的目标不是泛泛地罗列八股题，而是围绕候选人的真实经历出题，重点考察：
1. 技术基础是否扎实；
2. 项目经历是否真实且理解深入；
3. 遇到复杂场景时的分析、设计与取舍能力；
4. 是否具备从实践中总结问题与优化方案的能力。

请严格根据输入的简历内容、附件解析结果、网页补充资料和用户要求来组织输出。
如果某些附件无法解析，不要编造信息；你可以基于已有内容保守出题，并明确说明信息边界。`
		} else {
			systemPrompt = `你是一个全能人工智能学习助手，具备渊博的知识，可以回答用户关于编程、计算机科学以及各个学科的问题。你需要表现得专业、耐心且富有逻辑。

另外你还具备两个外部工具：
1. ` + "`fetch_web_page`" + `：可用于访问网页并提取标题与正文；
2. ` + "`export_summary_pdf`" + `：可用于将整理好的总结内容导出为 PDF 并返回下载地址。

请根据用户问题自行思考是否需要调用工具、调用哪个工具，以及如何基于工具结果继续观察、分析和回答。
如果一个请求需要多个工具，请先完成全部必要的工具调用，再输出面向用户的正式正文。不要在工具尚未执行完成前就先输出“我现在为你生成 PDF”这类正文说明。
如果工具返回了 PDF 下载地址，请在最终答复中明确告诉用户文件已生成，并以 Markdown 链接的形式提供下载地址，确保用户可以直接点击下载。`
		}
	}

	if hasCurrentPageURL {
		systemPrompt += "\n当前会话存在一个可供工具使用的当前页面 URL，但这不代表用户要你总结当前页面。只有当用户明确要求你基于当前页面、当前网页、上文页面、页面内容或用户选中文本来分析时，才调用 `fetch_web_page`；此时如果要读取当前页面，请将 `url` 置空，让系统自动使用当前页面 URL。若用户只是询问某个主题、概念或框架本身，例如“总结 eino 框架”，不要擅自把问题绑定到当前页面。"
	}

	if session.Summary != "" {
		systemPrompt += fmt.Sprintf("\n\n以下是之前的对话摘要，请作为背景参考：\n%s", session.Summary)
	}

	return systemPrompt
}

func (s *AIService) buildUserPrompt(req dto.AIChatReq) string {
	var builder strings.Builder

	// 1. 处理上传的文件
	if len(req.Files) > 0 {
		builder.WriteString("用户提供了以下文件/图片：\n")
		for _, file := range req.Files {
			// 将文件的详细信息拼接，供 AI 参考
			builder.WriteString(fmt.Sprintf("- 链接: %s (文件名: %s, 类型: %s, 大小: %d 字节)\n", file.FileURL, file.FileName, file.FileType, file.FileSize))
		}
		builder.WriteString("\n")
	}

	// 2. 处理页面文本或选中内容
	if strings.TrimSpace(req.SelectedText) != "" {
		builder.WriteString("用户选中的原文：\n")
		builder.WriteString(strings.TrimSpace(req.SelectedText))
		builder.WriteString("\n")
	}

	// 3. 处理用户输入（兼容新老字段）
	builder.WriteString("用户指令：\n")

	promptContent := strings.TrimSpace(req.UserInput)
	if promptContent == "" {
		promptContent = strings.TrimSpace(req.Prompt)
	}
	builder.WriteString(promptContent)

	return builder.String()
}

func (s *AIService) newChatAgent(ctx context.Context, userID uint) (*react.Agent, error) {
	chatModel, err := s.newChatModel(ctx)
	if err != nil {
		return nil, err
	}

	tools, err := s.newAITools(userID)
	if err != nil {
		return nil, errors.New("初始化 AI 工具失败")
	}

	agent, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: chatModel,
		ToolsConfig: compose.ToolsNodeConfig{
			Tools:               tools,
			ExecuteSequentially: true,
			ToolCallMiddlewares: []compose.ToolMiddleware{
				{
					Invokable: func(next compose.InvokableToolEndpoint) compose.InvokableToolEndpoint {
						return func(ctx context.Context, input *compose.ToolInput) (*compose.ToolOutput, error) {
							output, err := next(ctx, input)
							if err != nil {
								emitAIToolEvent(ctx, "工具执行失败："+err.Error())
								return &compose.ToolOutput{Result: err.Error()}, nil
							}
							return output, nil
						}
					},
				},
			},
		},
		MaxStep: 8,
	})
	if err != nil {
		return nil, errors.New("初始化 AI 智能体失败")
	}

	return agent, nil
}

func (s *AIService) emitStreamChunks(msgChan chan<- dto.ChatStreamChunk, chunkType, text string) {
	if strings.TrimSpace(text) == "" {
		return
	}

	for _, chunk := range splitTextForStreaming(text, 20) {
		msgChan <- dto.ChatStreamChunk{Type: chunkType, Content: chunk}
		time.Sleep(12 * time.Millisecond)
	}
}

// emitMessageChunk 将模型实时返回的片段直接推送给前端。
func (s *AIService) emitMessageChunk(msgChan chan<- dto.ChatStreamChunk, msg *schema.Message) {
	if msg == nil {
		return
	}

	if reasoning := s.extractReasoningText(msg); reasoning != "" {
		msgChan <- dto.ChatStreamChunk{Type: "reasoning", Content: reasoning}
	}
	if content := s.extractMessageText(msg); content != "" {
		msgChan <- dto.ChatStreamChunk{Type: "message", Content: content}
	}
}

func splitTextForStreaming(text string, chunkSize int) []string {
	if chunkSize <= 0 {
		chunkSize = 20
	}

	runes := []rune(text)
	if len(runes) == 0 {
		return nil
	}

	chunks := make([]string, 0, (len(runes)+chunkSize-1)/chunkSize)
	for start := 0; start < len(runes); start += chunkSize {
		end := start + chunkSize
		if end > len(runes) {
			end = len(runes)
		}
		chunks = append(chunks, string(runes[start:end]))
	}
	return chunks
}

func (s *AIService) finalizeChatSideEffects(db *gorm.DB, session model.Session, isNewSession bool, prompt string, latestUserMsgID int64) {
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
		}(session.ID, prompt)
		return
	}

	var count int64
	query := db.Model(&model.Message{}).Where("session_id = ? AND role = ?", session.ID, string(schema.User))
	if session.LastSummaryMessageID > 0 {
		query = query.Where("id > ?", session.LastSummaryMessageID)
	}
	query.Count(&count)

	if count < 4 {
		return
	}

	go func(sessionModel model.Session, latestUserMsgID int64) {
		summaryCtx := context.Background()
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
				"last_summary_message_id": latestUserMsgID,
			})
		}
	}(session, latestUserMsgID)
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
	msgChan := make(chan dto.ChatStreamChunk, 64)
	agentCtx := withAIToolEventSender(ctx, func(content string) {
		msgChan <- dto.ChatStreamChunk{Type: "tool", Content: content}
	})
	agentCtx = withAICurrentPageURL(agentCtx, req.CurrentPageURL)

	var session model.Session
	db := global.GVA_DB.WithContext(ctx)
	isNewSession := false

	// 兼容处理 SessionID
	var reqSessionID int64
	if req.SessionID != "" {
		parsed, _ := strconv.ParseInt(req.SessionID, 10, 64)
		reqSessionID = parsed
	}

	if reqSessionID == 0 {
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
		if err := db.Where("id = ? AND user_id = ?", reqSessionID, userId).First(&session).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, 0, 0, errors.New("会话不存在或无权访问")
			}
			return nil, 0, 0, errors.New("查询会话失败")
		}
	}

	// 记录数据库的用户输入（如果兼容老参数，则合并判断）
	promptContent := strings.TrimSpace(req.UserInput)
	if promptContent == "" {
		promptContent = strings.TrimSpace(req.Prompt)
	}

	userMsg := model.Message{
		SessionID: session.ID,
		ParentID:  req.ParentID,
		Role:      string(schema.User),
		Content:   promptContent,
	}
	db.Create(&userMsg)

	// 保存附件信息
	if len(req.Files) > 0 {
		for _, f := range req.Files {
			fileKey := utils.ExtractQiniuKey(f.FileURL)

			// 1. 存入 files 表 (如果尚未存在)
			var count int64
			db.Model(&model.File{}).Where("file_path = ?", fileKey).Count(&count)
			if count == 0 {
				fileRecord := model.File{
					FileName: f.FileName,
					FilePath: fileKey,
					FileType: f.FileType,
					FileSize: int(f.FileSize),
					UserID:   int(userId),
				}
				db.Create(&fileRecord)
			}

			// 2. 存入 message_attachments 表
			attachment := model.MessageAttachment{
				SessionID:  session.ID,
				MessageID:  userMsg.ID,
				UserID:     int64(userId),
				FileKey:    fileKey,
				FileName:   f.FileName,
				FileType:   f.FileType,
				FileSize:   f.FileSize,
				SenderRole: string(schema.User),
			}
			db.Create(&attachment)
		}
	}

	// 保存 AI 回复消息占位符
	aiMsg := model.Message{
		SessionID: session.ID,
		ParentID:  &userMsg.ID,
		Role:      string(schema.Assistant),
		Content:   "", // 留空，流式输出完成后再更新
	}
	db.Create(&aiMsg)

	if isResumeInterviewSkill(req.SkillID) {
		go func() {
			defer close(msgChan)

			finalAnswer, err := s.executeResumeInterviewAgents(agentCtx, req)
			if err != nil {
				global.GVA_LOG.Error("resume interview multi-agent failed", zap.Error(err))
				msgChan <- dto.ChatStreamChunk{Type: "tool", Content: "简历面试 Agent 执行失败，请稍后重试"}
				global.GVA_DB.Unscoped().Where("id IN ?", []int64{userMsg.ID, aiMsg.ID}).Delete(&model.Message{})
				if isNewSession {
					global.GVA_DB.Unscoped().Where("id = ?", session.ID).Delete(&model.Session{})
				}
				return
			}

			if strings.TrimSpace(finalAnswer) == "" {
				msgChan <- dto.ChatStreamChunk{Type: "tool", Content: "简历面试 Agent 未生成结果，请稍后重试"}
				global.GVA_DB.Unscoped().Where("id IN ?", []int64{userMsg.ID, aiMsg.ID}).Delete(&model.Message{})
				if isNewSession {
					global.GVA_DB.Unscoped().Where("id = ?", session.ID).Delete(&model.Session{})
				}
				return
			}

			s.emitStreamChunks(msgChan, "message", finalAnswer)
			global.GVA_DB.Model(&model.Message{}).Where("id = ?", aiMsg.ID).Update("content", finalAnswer)
			s.finalizeChatSideEffects(db, session, isNewSession, promptContent, userMsg.ID)
		}()
		return msgChan, session.ID, aiMsg.ID, nil
	}

	chatAgent, err := s.newChatAgent(ctx, userId)
	if err != nil {
		return nil, 0, 0, err
	}

	systemPrompt := s.buildChatSystemPrompt(session, strings.TrimSpace(req.CurrentPageURL) != "", req.SkillID)

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

	userPrompt := s.buildUserPrompt(req)
	messages := s.buildConversationMessages(systemPrompt, historyMsgs, userPrompt)

	go func() {
		defer close(msgChan)
		streamResp, err := chatAgent.Stream(agentCtx, messages)
		if err != nil {
			global.GVA_LOG.Error("Eino AI agent stream failed", zap.Error(err))
			msgChan <- dto.ChatStreamChunk{Type: "tool", Content: "工具链执行失败，请稍后重试"}
			global.GVA_DB.Unscoped().Where("id IN ?", []int64{userMsg.ID, aiMsg.ID}).Delete(&model.Message{})
			if isNewSession {
				global.GVA_DB.Unscoped().Where("id = ?", session.ID).Delete(&model.Session{})
			}
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
				msgChan <- dto.ChatStreamChunk{Type: "tool", Content: "AI 流式输出中断，请稍后重试"}
				global.GVA_DB.Unscoped().Where("id IN ?", []int64{userMsg.ID, aiMsg.ID}).Delete(&model.Message{})
				if isNewSession {
					global.GVA_DB.Unscoped().Where("id = ?", session.ID).Delete(&model.Session{})
				}
				return
			}

			streamMsgs = append(streamMsgs, chunk)
			s.emitMessageChunk(msgChan, chunk)
		}

		if len(streamMsgs) == 0 {
			msgChan <- dto.ChatStreamChunk{Type: "tool", Content: "AI 未生成最终答案，请稍后重试"}
			global.GVA_DB.Unscoped().Where("id IN ?", []int64{userMsg.ID, aiMsg.ID}).Delete(&model.Message{})
			if isNewSession {
				global.GVA_DB.Unscoped().Where("id = ?", session.ID).Delete(&model.Session{})
			}
			return
		}

		finalResp, err := schema.ConcatMessages(streamMsgs)
		if err != nil {
			global.GVA_LOG.Error("Concat AI stream chunks failed", zap.Error(err))
			msgChan <- dto.ChatStreamChunk{Type: "tool", Content: "AI 流式结果合并失败，请稍后重试"}
			global.GVA_DB.Unscoped().Where("id IN ?", []int64{userMsg.ID, aiMsg.ID}).Delete(&model.Message{})
			if isNewSession {
				global.GVA_DB.Unscoped().Where("id = ?", session.ID).Delete(&model.Session{})
			}
			return
		}

		messageText := s.extractMessageText(finalResp)
		if strings.TrimSpace(messageText) == "" {
			msgChan <- dto.ChatStreamChunk{Type: "tool", Content: "AI 未生成最终答案，请稍后重试"}
			global.GVA_DB.Unscoped().Where("id IN ?", []int64{userMsg.ID, aiMsg.ID}).Delete(&model.Message{})
			if isNewSession {
				global.GVA_DB.Unscoped().Where("id = ?", session.ID).Delete(&model.Session{})
			}
			return
		}

		global.GVA_DB.Model(&model.Message{}).Where("id = ?", aiMsg.ID).Update("content", messageText)
		s.finalizeChatSideEffects(db, session, isNewSession, promptContent, userMsg.ID)
	}()

	return msgChan, session.ID, aiMsg.ID, nil
}
