package service

import (
	"archive/zip"
	"backend/dto"
	"backend/global"
	"backend/pkg/utils"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/cloudwego/eino/schema"
	"github.com/ledongthuc/pdf"
	qcred "github.com/qiniu/go-sdk/v7/storagev2/credentials"
	"github.com/qiniu/go-sdk/v7/storagev2/downloader"
	httpclient "github.com/qiniu/go-sdk/v7/storagev2/http_client"
)

const (
	maxResumeDownloadBytes = 8 << 20
	maxResumeExtractRunes  = 16000
	maxResumeRefRunes      = 1500
	maxResumeQueryCount    = 3
	maxResumeRefCount      = 3
)

var (
	docxTagRE          = regexp.MustCompile(`<[^>]+>`)
	resumeSkillAliasRE = regexp.MustCompile(`(?i)(resume|cv).*(interview|question)|interview.*(resume|cv)`)
	englishTokenRE     = regexp.MustCompile(`[A-Za-z][A-Za-z0-9+#.\-/]{1,30}`)
	multiSpaceRE       = regexp.MustCompile(`\s+`)
)

type resumeFileContext struct {
	Name   string
	Type   string
	URL    string
	Text   string
	Notice string
}

type resumeSearchResult struct {
	Title string
	URL   string
}

type resumeReference struct {
	Title   string
	URL     string
	Content string
}

type resumeKnowledgeBundle struct {
	FileContexts []resumeFileContext
	ResumeText   string
	References   []resumeReference
	Notices      []string
}

func isResumeInterviewSkill(skillID string) bool {
	normalized := strings.ToLower(strings.TrimSpace(skillID))
	switch normalized {
	case "resume_interview", "resume_interview_agent", "resume_question", "cv_interview":
		return true
	}
	return resumeSkillAliasRE.MatchString(normalized)
}

func (s *AIService) buildResumeInterviewContext(ctx context.Context, req dto.AIChatReq) (string, error) {
	if !isResumeInterviewSkill(req.SkillID) {
		return "", nil
	}

	bundle, err := s.prepareResumeKnowledge(ctx, req)
	if err != nil {
		return "", err
	}
	return s.renderResumeKnowledgeBundle(bundle), nil
}

func (s *AIService) prepareResumeKnowledge(ctx context.Context, req dto.AIChatReq) (resumeKnowledgeBundle, error) {
	if len(req.Files) == 0 {
		return resumeKnowledgeBundle{}, errors.New("未上传简历文件，请先上传文件后再生成面试题")
	}

	emitAIToolEvent(ctx, "正在解析简历附件...")
	fileContexts, resumeText, notices := s.extractResumeFiles(ctx, req.Files)
	resumeText = strings.TrimSpace(resumeText)

	var references []resumeReference
	if resumeText != "" {
		emitAIToolEvent(ctx, "正在检索与简历相关的补充资料...")
		var refErr error
		references, refErr = s.searchResumeReferences(ctx, resumeText, strings.TrimSpace(req.UserInput))
		if refErr != nil && strings.TrimSpace(refErr.Error()) != "" {
			notices = append(notices, "补充网页检索失败："+refErr.Error())
		}
	} else {
		notices = append(notices, "简历正文未成功提取，已跳过补充网页检索，避免引入无关资料")
	}

	return resumeKnowledgeBundle{
		FileContexts: fileContexts,
		ResumeText:   resumeText,
		References:   references,
		Notices:      notices,
	}, nil
}

func validateResumeKnowledgeBundle(bundle resumeKnowledgeBundle) error {
	if len(bundle.FileContexts) == 0 {
		return errors.New("未上传简历文件，请先上传文件后再生成面试题")
	}
	if strings.TrimSpace(bundle.ResumeText) != "" {
		return nil
	}
	return errors.New("文件解析失败，请重新上传清晰且可读取的简历文件")
}

func (s *AIService) renderResumeKnowledgeBundle(bundle resumeKnowledgeBundle) string {
	var builder strings.Builder
	builder.WriteString("【专项任务】你正在执行“根据候选人简历生成面试题”的任务。\n")
	builder.WriteString("请优先依据简历真实内容出题，再结合补充资料增强题目的深度与延展性。\n\n")

	if len(bundle.FileContexts) > 0 {
		builder.WriteString("【简历附件解析】\n")
		for idx, item := range bundle.FileContexts {
			builder.WriteString(fmt.Sprintf("%d. 文件名：%s\n", idx+1, defaultIfEmpty(item.Name, "未命名文件")))
			builder.WriteString(fmt.Sprintf("   文件类型：%s\n", defaultIfEmpty(item.Type, "未知")))
			if strings.TrimSpace(item.URL) != "" {
				builder.WriteString(fmt.Sprintf("   文件链接：%s\n", item.URL))
			}
			if strings.TrimSpace(item.Text) != "" {
				builder.WriteString("   提取内容：\n")
				builder.WriteString(indentMultiline(item.Text, "   "))
				builder.WriteString("\n")
			}
			if strings.TrimSpace(item.Notice) != "" {
				builder.WriteString(fmt.Sprintf("   说明：%s\n", item.Notice))
			}
		}
		builder.WriteString("\n")
	}

	if len(bundle.References) > 0 {
		builder.WriteString("【补充网页资料】\n")
		for idx, ref := range bundle.References {
			builder.WriteString(fmt.Sprintf("%d. 标题：%s\n", idx+1, defaultIfEmpty(ref.Title, "未命名资料")))
			builder.WriteString(fmt.Sprintf("   链接：%s\n", ref.URL))
			builder.WriteString("   要点：\n")
			builder.WriteString(indentMultiline(ref.Content, "   "))
			builder.WriteString("\n")
		}
		builder.WriteString("\n")
	}

	if len(bundle.Notices) > 0 {
		builder.WriteString("【处理说明】\n")
		for _, notice := range bundle.Notices {
			builder.WriteString("- " + notice + "\n")
		}
		builder.WriteString("\n")
	}

	builder.WriteString("【出题要求】\n")
	builder.WriteString("1. 先给出候选人画像总结，概括候选人的岗位方向、核心技术栈、项目亮点与潜在薄弱点。\n")
	builder.WriteString("2. 再生成面试题，至少覆盖：基础原理、项目深挖、场景设计、追问延展 4 个维度。\n")
	builder.WriteString("3. 每道题都要包含：题目、考察点、参考追问。\n")
	builder.WriteString("4. 题目要尽量绑定简历中的真实项目、技术和业务场景，避免空泛八股。\n")
	builder.WriteString("5. 如果附件无法完全解析，要基于已解析内容和用户输入保守出题，并明确说明局限。\n")
	builder.WriteString("6. 输出使用中文，结构清晰，适合直接给面试官使用。\n")

	return builder.String()
}

func (s *AIService) executeResumeInterviewAgents(ctx context.Context, req dto.AIChatReq, bundle resumeKnowledgeBundle) (string, error) {
	if !isResumeInterviewSkill(req.SkillID) {
		return "", errors.New("当前技能不是简历面试场景")
	}

	rawContext := s.renderResumeKnowledgeBundle(bundle)
	emitAIToolEvent(ctx, "简历资料整理完成")

	emitAIToolEvent(ctx, "简历解析 Agent 正在分析候选人画像...")
	profileResult, err := s.runResumeParserAgent(ctx, req, rawContext)
	if err != nil {
		return "", err
	}

	emitAIToolEvent(ctx, "资料研究 Agent 正在补充关键知识点...")
	researchResult, err := s.runResumeResearchAgent(ctx, req, rawContext, profileResult)
	if err != nil {
		return "", err
	}

	emitAIToolEvent(ctx, "出题 Agent 正在生成面试题...")
	questionResult, err := s.runResumeQuestionAgent(ctx, req, rawContext, profileResult, researchResult)
	if err != nil {
		return "", err
	}

	emitAIToolEvent(ctx, "审校 Agent 正在优化题目质量...")
	finalResult, err := s.runResumeReviewAgent(ctx, req, rawContext, profileResult, researchResult, questionResult)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(finalResult), nil
}

func (s *AIService) runResumeParserAgent(ctx context.Context, req dto.AIChatReq, rawContext string) (string, error) {
	systemPrompt := s.getAgentSystemPrompt("resume_parser_agent", `你是简历解析 Agent，专门负责从候选人的简历中提炼结构化候选人画像。
请输出：
1. 候选人目标岗位
2. 核心技术栈
3. 重点项目与职责
4. 可深挖亮点
5. 潜在薄弱点与风险点
要求：只基于输入信息总结，不编造。`)

	userPrompt := "请你解析下面的简历资料，并输出结构化候选人画像。\n\n" + rawContext + "\n\n用户额外要求：\n" + strings.TrimSpace(req.UserInput)
	return s.generateText(ctx, []*schema.Message{
		schema.SystemMessage(systemPrompt),
		schema.UserMessage(userPrompt),
	})
}

func (s *AIService) runResumeResearchAgent(ctx context.Context, req dto.AIChatReq, rawContext, profile string) (string, error) {
	systemPrompt := s.getAgentSystemPrompt("resume_research_agent", `你是资料研究 Agent，负责围绕候选人简历中的技术栈、项目和业务场景，提炼出面试时值得追问的知识点。
请输出：
1. 候选人涉及技术的关键原理
2. 这些技术在项目中的典型追问方向
3. 结合补充网页资料得到的延展点
4. 面试官需要重点验证的真实性问题
要求：输出聚焦、结构化、面向面试。`)

	userPrompt := "以下是简历原始资料：\n" + rawContext + "\n\n以下是简历解析 Agent 的结果：\n" + profile + "\n\n请继续输出研究结论。"
	return s.generateText(ctx, []*schema.Message{
		schema.SystemMessage(systemPrompt),
		schema.UserMessage(userPrompt),
	})
}

func (s *AIService) runResumeQuestionAgent(ctx context.Context, req dto.AIChatReq, rawContext, profile, research string) (string, error) {
	systemPrompt := s.getAgentSystemPrompt("resume_question_agent", `你是面试题生成 Agent，负责根据候选人画像和研究资料生成高质量面试题。
请至少覆盖：
1. 基础原理题
2. 项目深挖题
3. 场景设计题
4. 延展追问题
每道题都要包含：题目、考察点、参考追问。`)

	userPrompt := "以下是简历资料：\n" + rawContext + "\n\n以下是候选人画像：\n" + profile + "\n\n以下是研究资料：\n" + research + "\n\n请基于这些信息生成完整的面试题初稿。"
	return s.generateText(ctx, []*schema.Message{
		schema.SystemMessage(systemPrompt),
		schema.UserMessage(userPrompt),
	})
}

func (s *AIService) runResumeReviewAgent(ctx context.Context, req dto.AIChatReq, rawContext, profile, research, questions string) (string, error) {
	systemPrompt := s.getAgentSystemPrompt("resume_review_agent", `你是面试题审校 Agent，负责检查并优化题目质量。
请完成：
1. 去掉重复或空泛题目
2. 让题目更贴合简历真实经历
3. 调整难度分层
4. 保证输出可直接给面试官使用
最终输出格式：
1. 候选人画像总结
2. 面试题清单（分模块）
3. 面试建议`)

	userPrompt := "以下是简历资料：\n" + rawContext + "\n\n候选人画像：\n" + profile + "\n\n研究资料：\n" + research + "\n\n面试题初稿：\n" + questions + "\n\n请产出最终可交付版本。"
	return s.generateText(ctx, []*schema.Message{
		schema.SystemMessage(systemPrompt),
		schema.UserMessage(userPrompt),
	})
}

func (s *AIService) extractResumeFiles(ctx context.Context, files []dto.AIChatFile) ([]resumeFileContext, string, []string) {
	if len(files) == 0 {
		return nil, "", []string{"未提供简历附件，将仅依据用户文本描述生成题目"}
	}

	contexts := make([]resumeFileContext, 0, len(files))
	notices := make([]string, 0, len(files))
	var corpus strings.Builder

	for _, item := range files {
		ctxItem := resumeFileContext{
			Name: strings.TrimSpace(item.FileName),
			Type: strings.TrimSpace(item.FileType),
			URL:  strings.TrimSpace(item.FileURL),
		}
		if ctxItem.URL == "" {
			ctxItem.Notice = "文件链接为空，已跳过解析"
			contexts = append(contexts, ctxItem)
			notices = append(notices, defaultIfEmpty(ctxItem.Name, "未知文件")+"：文件链接为空")
			continue
		}

		data, contentType, err := s.downloadResumeFile(ctx, ctxItem.URL)
		if err != nil {
			ctxItem.Notice = "下载失败：" + err.Error()
			contexts = append(contexts, ctxItem)
			notices = append(notices, defaultIfEmpty(ctxItem.Name, ctxItem.URL)+" 下载失败")
			continue
		}

		detectedType := detectResumeFileKind(item, contentType, data)
		ctxItem.Type = detectedType

		text, notice := s.extractResumeTextByType(data, detectedType)
		if strings.TrimSpace(text) != "" {
			text, _ = truncateRunes(cleanExtractedText(text), 5000)
			ctxItem.Text = text
			corpus.WriteString(text)
			corpus.WriteString("\n")
		}
		if strings.TrimSpace(notice) != "" {
			ctxItem.Notice = notice
			notices = append(notices, defaultIfEmpty(ctxItem.Name, ctxItem.URL)+"："+notice)
		}

		contexts = append(contexts, ctxItem)
	}

	combined, _ := truncateRunes(corpus.String(), maxResumeExtractRunes)
	return contexts, combined, notices
}

func (s *AIService) downloadResumeFile(ctx context.Context, fileURL string) ([]byte, string, error) {
	cleanURL := utils.CleanQiniuFileURL(fileURL)
	if data, contentType, err := s.downloadResumeFileFromQiniu(ctx, cleanURL); err == nil {
		return data, contentType, nil
	}
	targetURL, err := normalizeFetchURL(cleanURL)
	if err != nil {
		return nil, "", err
	}

	parsed, err := neturl.Parse(targetURL)
	if err != nil {
		return nil, "", errors.New("文件链接格式不正确")
	}
	if err := validateSafeURL(ctx, parsed); err != nil {
		return nil, "", err
	}

	client := &http.Client{Timeout: 25 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, "", errors.New("创建文件请求失败")
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; AIAgentAssistant/1.0; +https://example.local)")
	req.Header.Set("Accept", "*/*")

	resp, err := client.Do(req)
	if err != nil {
		return nil, "", errors.New("下载文件失败")
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, "", fmt.Errorf("下载文件失败，状态码为 %d", resp.StatusCode)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxResumeDownloadBytes))
	if err != nil {
		return nil, "", errors.New("读取文件内容失败")
	}
	return data, strings.ToLower(strings.TrimSpace(resp.Header.Get("Content-Type"))), nil
}

func (s *AIService) downloadResumeFileFromQiniu(ctx context.Context, fileURL string) ([]byte, string, error) {
	_ = s
	q := global.GVA_CONFIG.Qiniu
	if strings.TrimSpace(q.AccessKey) == "" || strings.TrimSpace(q.SecretKey) == "" || strings.TrimSpace(q.Bucket) == "" {
		return nil, "", errors.New("七牛云配置不完整")
	}

	key := utils.ExtractQiniuKey(fileURL)
	if strings.TrimSpace(key) == "" {
		return nil, "", errors.New("无法识别七牛云文件 key")
	}

	looksLikeQiniu := strings.HasPrefix(key, "Agent/") ||
		(strings.TrimSpace(q.Domain) != "" && strings.Contains(fileURL, strings.TrimSpace(q.Domain))) ||
		!strings.Contains(fileURL, "://")
	if !looksLikeQiniu {
		return nil, "", errors.New("不是七牛云文件地址")
	}

	cred := qcred.NewCredentials(q.AccessKey, q.SecretKey)
	downloadManager := downloader.NewDownloadManager(&downloader.DownloadManagerOptions{
		Options: httpclient.Options{
			Credentials:         cred,
			UseInsecureProtocol: !q.UseHTTPS,
		},
	})

	var buf bytes.Buffer
	if _, err := downloadManager.DownloadToWriter(ctx, key, &buf, &downloader.ObjectOptions{
		GenerateOptions: downloader.GenerateOptions{
			BucketName:          q.Bucket,
			UseInsecureProtocol: !q.UseHTTPS,
		},
	}); err != nil {
		return nil, "", err
	}

	data := buf.Bytes()
	return data, strings.ToLower(http.DetectContentType(data)), nil
}

func detectResumeFileKind(file dto.AIChatFile, contentType string, data []byte) string {
	name := strings.ToLower(strings.TrimSpace(file.FileName))
	urlText := strings.ToLower(strings.TrimSpace(file.FileURL))
	typeText := strings.ToLower(strings.TrimSpace(file.FileType))
	merged := name + " " + urlText + " " + typeText + " " + contentType + " " + http.DetectContentType(data)

	switch {
	case strings.Contains(merged, ".pdf") || strings.Contains(merged, "application/pdf"):
		return "pdf"
	case strings.Contains(merged, ".docx") || strings.Contains(merged, "officedocument.wordprocessingml.document"):
		return "docx"
	case strings.Contains(merged, ".doc") || strings.Contains(merged, "application/msword"):
		return "doc"
	case strings.Contains(merged, "image/") ||
		strings.Contains(merged, ".png") ||
		strings.Contains(merged, ".jpg") ||
		strings.Contains(merged, ".jpeg") ||
		strings.Contains(merged, ".webp") ||
		strings.Contains(merged, ".gif"):
		return "image"
	case strings.Contains(merged, "text/") ||
		strings.Contains(merged, ".txt") ||
		strings.Contains(merged, ".md") ||
		strings.Contains(merged, ".markdown") ||
		strings.Contains(merged, ".csv") ||
		strings.Contains(merged, ".json"):
		return "text"
	default:
		return "unknown"
	}
}

func (s *AIService) extractResumeTextByType(data []byte, fileType string) (string, string) {
	switch fileType {
	case "pdf":
		text, err := extractPDFText(data)
		if err != nil {
			return "", "PDF 文本提取失败，已保留文件元信息"
		}
		return text, ""
	case "docx":
		text, err := extractDOCXText(data)
		if err != nil {
			return "", "DOCX 文本提取失败，已保留文件元信息"
		}
		return text, ""
	case "text":
		return string(data), ""
	case "image":
		return "", "图片简历暂未做本地 OCR，已将文件链接与元信息提供给模型参考"
	case "doc":
		return "", "暂不支持解析旧版 DOC 二进制文档，建议转为 PDF、DOCX 或 TXT"
	default:
		return "", "暂不支持解析该文件类型，已保留文件链接与元信息"
	}
}

func extractPDFText(data []byte) (string, error) {
	tmpFile, err := os.CreateTemp("", "resume-*.pdf")
	if err != nil {
		return "", err
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		return "", err
	}
	if err := tmpFile.Close(); err != nil {
		return "", err
	}

	if text, err := extractPDFTextWithPyMuPDF(tmpPath); err == nil && strings.TrimSpace(text) != "" {
		return text, nil
	}

	f, reader, err := pdf.Open(tmpPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	plainText, err := reader.GetPlainText()
	if err != nil {
		return "", err
	}

	buf, err := io.ReadAll(plainText)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

func extractPDFTextWithPyMuPDF(pdfPath string) (string, error) {
	pythonPath, err := exec.LookPath("python")
	if err != nil {
		return "", err
	}

	scriptFile, err := os.CreateTemp("", "resume-pdf-*.py")
	if err != nil {
		return "", err
	}
	scriptPath := scriptFile.Name()
	defer os.Remove(scriptPath)

	script := strings.Join([]string{
		"import fitz",
		"import sys",
		"",
		"path = sys.argv[1]",
		"doc = fitz.open(path)",
		`text = "\n".join(page.get_text("text") for page in doc)`,
		"sys.stdout.buffer.write(text.encode('utf-8', errors='ignore'))",
	}, "\n")

	if _, err := scriptFile.WriteString(script); err != nil {
		scriptFile.Close()
		return "", err
	}
	if err := scriptFile.Close(); err != nil {
		return "", err
	}

	cmd := exec.Command(pythonPath, scriptPath, pdfPath)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func extractDOCXText(data []byte) (string, error) {
	readerAt := bytes.NewReader(data)
	zr, err := zip.NewReader(readerAt, int64(len(data)))
	if err != nil {
		return "", err
	}

	for _, file := range zr.File {
		if file.Name != "word/document.xml" {
			continue
		}
		rc, err := file.Open()
		if err != nil {
			return "", err
		}
		defer rc.Close()

		xmlData, err := io.ReadAll(rc)
		if err != nil {
			return "", err
		}
		text := string(xmlData)
		text = strings.ReplaceAll(text, "</w:p>", "\n")
		text = strings.ReplaceAll(text, "</w:tr>", "\n")
		text = docxTagRE.ReplaceAllString(text, " ")
		return cleanExtractedText(text), nil
	}

	return "", errors.New("未找到 DOCX 正文")
}

func (s *AIService) searchResumeReferences(ctx context.Context, resumeText, userInput string) ([]resumeReference, error) {
	if strings.TrimSpace(resumeText) == "" {
		return nil, nil
	}
	queries := buildResumeSearchQueries(resumeText, userInput)
	if len(queries) == 0 {
		return nil, nil
	}

	results := make([]resumeReference, 0, maxResumeRefCount)
	seen := make(map[string]struct{})

	for _, query := range queries {
		searchResults, err := s.searchBing(ctx, query)
		if err != nil {
			continue
		}
		for _, item := range searchResults {
			if _, ok := seen[item.URL]; ok {
				continue
			}
			seen[item.URL] = struct{}{}

			ref, err := s.fetchResumeReference(ctx, item)
			if err != nil {
				continue
			}
			results = append(results, ref)
			if len(results) >= maxResumeRefCount {
				return results, nil
			}
		}
	}

	return results, nil
}

func buildResumeSearchQueries(resumeText, userInput string) []string {
	queries := make([]string, 0, maxResumeQueryCount)
	seen := make(map[string]struct{})

	techTerms := extractResumeTechTerms(resumeText, userInput)
	projectLines := extractResumeProjectLines(resumeText)

	addQuery := func(query string) {
		query = normalizeResumeQuery(query)
		if query == "" {
			return
		}
		if _, ok := seen[query]; ok {
			return
		}
		seen[query] = struct{}{}
		queries = append(queries, query)
	}

	if len(techTerms) >= 3 {
		addQuery(strings.Join(techTerms[:minInt(len(techTerms), 4)], " ") + " 原理 项目 实战")
	}

	for _, line := range projectLines {
		addQuery(line + " 技术方案 架构 实战")
		if len(queries) >= maxResumeQueryCount {
			return queries
		}
	}

	if len(techTerms) > 0 {
		for _, term := range techTerms {
			addQuery(term + " 原理 项目 面试 实战")
			if len(queries) >= maxResumeQueryCount {
				return queries
			}
		}
	}

	if len(queries) == 0 && strings.TrimSpace(userInput) != "" {
		query := strings.TrimSpace(userInput)
		if len([]rune(query)) > 30 {
			query = string([]rune(query)[:30])
		}
		addQuery(query + " 面试 原理 实战")
	}

	return queries
}

func extractResumeProjectLines(resumeText string) []string {
	lines := strings.Split(resumeText, "\n")
	result := make([]string, 0, 4)
	seen := make(map[string]struct{})

	for _, rawLine := range lines {
		line := normalizeResumeLine(rawLine)
		if line == "" {
			continue
		}
		if !isLikelyProjectLine(line) {
			continue
		}
		if _, ok := seen[line]; ok {
			continue
		}
		seen[line] = struct{}{}
		result = append(result, line)
		if len(result) >= maxResumeQueryCount {
			return result
		}
	}

	return result
}

func extractResumeTechTerms(resumeText, userInput string) []string {
	base := strings.ToLower(resumeText + "\n" + userInput)
	priorityTerms := []string{
		"golang", "go", "java", "python", "mysql", "postgresql", "redis", "mongodb", "elasticsearch",
		"kafka", "rabbitmq", "rocketmq", "docker", "kubernetes", "linux", "grpc", "http", "tcp",
		"spring", "springboot", "gin", "gorm", "react", "vue", "typescript", "javascript", "nginx",
		"微服务", "分布式", "高并发", "消息队列", "缓存", "数据库", "搜索", "推荐", "机器学习", "rag", "llm",
	}

	result := make([]string, 0, 8)
	seen := make(map[string]struct{})
	for _, term := range priorityTerms {
		if !strings.Contains(base, strings.ToLower(term)) {
			continue
		}
		if _, ok := seen[term]; ok {
			continue
		}
		seen[term] = struct{}{}
		result = append(result, term)
		if len(result) >= 6 {
			return result
		}
	}

	for _, token := range englishTokenRE.FindAllString(resumeText, -1) {
		normalized := strings.Trim(token, ".,;:()[]{}<>\"'")
		lower := strings.ToLower(normalized)
		if len(lower) < 2 || len(lower) > 30 {
			continue
		}
		if isCommonNoiseToken(lower) {
			continue
		}
		if _, ok := seen[lower]; ok {
			continue
		}
		seen[lower] = struct{}{}
		result = append(result, normalized)
		if len(result) >= 6 {
			return result
		}
	}

	return result
}

func normalizeResumeLine(line string) string {
	line = strings.TrimSpace(line)
	line = multiSpaceRE.ReplaceAllString(line, " ")
	line = strings.Trim(line, "-*#| \t")
	if len([]rune(line)) > 48 {
		line = string([]rune(line)[:48])
	}
	return strings.TrimSpace(line)
}

func normalizeResumeQuery(query string) string {
	query = normalizeResumeLine(query)
	query = strings.ReplaceAll(query, "：", " ")
	query = strings.ReplaceAll(query, ":", " ")
	query = multiSpaceRE.ReplaceAllString(query, " ")
	return strings.TrimSpace(query)
}

func isLikelyProjectLine(line string) bool {
	if len([]rune(line)) < 6 {
		return false
	}
	lower := strings.ToLower(line)
	if strings.Contains(lower, "@") || strings.Contains(lower, "电话") || strings.Contains(lower, "邮箱") {
		return false
	}
	hints := []string{
		"项目", "系统", "平台", "负责", "实现", "设计", "开发", "优化", "架构", "技术栈",
		"service", "system", "platform", "project", "design", "develop", "architecture",
	}
	for _, hint := range hints {
		if strings.Contains(lower, strings.ToLower(hint)) {
			return true
		}
	}
	return false
}

func isCommonNoiseToken(token string) bool {
	noise := map[string]struct{}{
		"http": {}, "https": {}, "com": {}, "www": {}, "email": {}, "phone": {}, "github": {},
		"resume": {}, "project": {}, "developer": {}, "engineer": {}, "china": {}, "work": {},
	}
	_, ok := noise[token]
	return ok
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (s *AIService) searchBing(ctx context.Context, query string) ([]resumeSearchResult, error) {
	searchURL := "https://www.bing.com/search?q=" + neturl.QueryEscape(query)
	parsedURL, err := neturl.Parse(searchURL)
	if err != nil {
		return nil, errors.New("搜索地址构造失败")
	}
	if err := validateSafeURL(ctx, parsedURL); err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 20 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; AIAgentAssistant/1.0; +https://example.local)")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("搜索失败，状态码为 %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxWebFetchBytes))
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	results := make([]resumeSearchResult, 0, 5)
	doc.Find("li.b_algo h2 a").EachWithBreak(func(i int, sel *goquery.Selection) bool {
		href, ok := sel.Attr("href")
		if !ok {
			return true
		}
		title := cleanExtractedText(sel.Text())
		href = strings.TrimSpace(href)
		if title == "" || href == "" {
			return true
		}
		results = append(results, resumeSearchResult{Title: title, URL: href})
		return len(results) < 5
	})

	return results, nil
}

func (s *AIService) fetchResumeReference(ctx context.Context, item resumeSearchResult) (resumeReference, error) {
	targetURL, err := normalizeFetchURL(item.URL)
	if err != nil {
		return resumeReference{}, err
	}

	parsed, err := neturl.Parse(targetURL)
	if err != nil {
		return resumeReference{}, err
	}
	if err := validateSafeURL(ctx, parsed); err != nil {
		return resumeReference{}, err
	}

	client := &http.Client{Timeout: 20 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return resumeReference{}, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; AIAgentAssistant/1.0; +https://example.local)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,text/plain;q=0.8,*/*;q=0.7")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		return resumeReference{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return resumeReference{}, fmt.Errorf("网页抓取失败，状态码为 %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxWebFetchBytes))
	if err != nil {
		return resumeReference{}, err
	}

	contentType := strings.ToLower(resp.Header.Get("Content-Type"))
	title := item.Title
	content := ""
	switch {
	case strings.Contains(contentType, "text/plain"):
		content = cleanExtractedText(string(body))
	default:
		pageTitle, pageContent, err := extractReadableHTML(body)
		if err != nil {
			return resumeReference{}, err
		}
		if strings.TrimSpace(pageTitle) != "" {
			title = pageTitle
		}
		content = pageContent
	}

	content, _ = truncateRunes(content, maxResumeRefRunes)
	if strings.TrimSpace(content) == "" {
		return resumeReference{}, errors.New("网页内容为空")
	}

	return resumeReference{
		Title:   title,
		URL:     targetURL,
		Content: content,
	}, nil
}

func indentMultiline(text, prefix string) string {
	lines := strings.Split(strings.TrimSpace(text), "\n")
	for i, line := range lines {
		lines[i] = prefix + strings.TrimRight(line, " ")
	}
	return strings.Join(lines, "\n")
}
