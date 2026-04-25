package service

import (
	"backend/global"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	neturl "net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
	"github.com/cloudwego/eino/components/tool"
	toolutils "github.com/cloudwego/eino/components/tool/utils"
	"github.com/google/uuid"
	"github.com/signintech/gopdf"
)

const (
	maxWebFetchBytes       = 2 << 20
	maxWebPageContentRunes = 12000
	exportTicketTTL        = 5 * time.Minute
	pdfPageWidthMM         = 210.0
	pdfPageHeightMM        = 297.0
	pdfMarginLeftMM        = 15.0
	pdfMarginTopMM         = 18.0
	pdfMarginRightMM       = 15.0
	pdfMarginBottomMM      = 18.0
	pdfTitleLineHeightMM   = 9.0
	pdfMetaLineHeightMM    = 5.5
	pdfBodyLineHeightMM    = 6.5
	pdfParagraphSpacingMM  = 2.0
	pdfTitleFontSize       = 18.0
	pdfMetaFontSize        = 10.0
	pdfBodyFontSize        = 12.0
	pdfHeading1FontSize    = 16.0
	pdfHeading2FontSize    = 14.0
	pdfHeading3FontSize    = 13.0
	pdfCodeFontSize        = 10.5
	pdfQuoteFontSize       = 11.0
	pdfListIndentMM        = 6.0
	pdfListMarkerWidthMM   = 5.0
	pdfQuoteIndentMM       = 5.0
	pdfTableFontSize       = 10.5
	pdfTableLineHeightMM   = 5.5
	pdfTableCellPaddingMM  = 1.2
	pdfCodeBlockPaddingMM  = 1.8
)

var (
	fileNameCleaner     = regexp.MustCompile(`[^a-zA-Z0-9\p{Han}_-]+`)
	markdownHeadingRE   = regexp.MustCompile(`^(#{1,6})\s+(.*)$`)
	markdownUnorderedRE = regexp.MustCompile(`^\s*[-*+]\s+(.*)$`)
	markdownOrderedRE   = regexp.MustCompile(`^\s*(\d+)\.\s+(.*)$`)
	markdownLinkRE      = regexp.MustCompile(`\[(.*?)\]\((.*?)\)`)
	markdownImageRE     = regexp.MustCompile(`!\[(.*?)\]\((.*?)\)`)
	markdownEmphasisRE  = regexp.MustCompile("(\\*\\*|__|\\*|_|~~|`)")
	bareURLRE           = regexp.MustCompile(`https?://[^\s<>()]+`)
)

type aiToolEventSender func(content string)

type aiToolEventSenderCtxKey struct{}
type aiCurrentPageURLCtxKey struct{}

type fetchWebPageInput struct {
	URL string `json:"url"`
}

type fetchWebPageResult struct {
	URL       string `json:"url"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Truncated bool   `json:"truncated"`
	Notice    string `json:"notice,omitempty"`
}

type exportSummaryPDFInput struct {
	Title     string `json:"title"`
	Content   string `json:"content"`
	SourceURL string `json:"sourceUrl,omitempty"`
}

type exportSummaryPDFResult struct {
	FileName           string `json:"fileName"`
	OneTimeDownloadURL string `json:"oneTimeDownloadUrl"`
	SavedAt            string `json:"savedAt"`
}

type markdownBlock struct {
	Kind        string
	Level       int
	Lines       []string
	Text        string
	TableHeader []string
	TableRows   [][]string
}

type markdownInlineSegment struct {
	Text string
	Kind string
	URL  string
}

type pdfColor struct {
	R uint8
	G uint8
	B uint8
}

type pdfTextStyle struct {
	FontSize    float64
	TextColor   pdfColor
	FillColor   *pdfColor
	Underline   bool
	ExternalURL string
}

func (s *AIService) newAITools(userID uint) ([]tool.BaseTool, error) {
	fetchTool, err := toolutils.InferTool(
		"fetch_web_page",
		"用于访问网页链接并提取网页标题与正文内容。适用于阅读网页、总结网页、分析网页，或者当问题里包含需要实际读取的 URL 时。",
		s.fetchWebPageTool,
	)
	if err != nil {
		return nil, err
	}

	exportTool, err := toolutils.InferTool(
		"export_summary_pdf",
		"用于把已经整理好的总结内容导出为 PDF，并返回一次性下载地址。适用于用户明确要求导出 PDF、生成 PDF 或保存总结结果的场景。输入应为最终总结内容，而不是原始网页全文。",
		func(ctx context.Context, input exportSummaryPDFInput) (exportSummaryPDFResult, error) {
			return s.exportSummaryPDFTool(ctx, userID, input)
		},
	)
	if err != nil {
		return nil, err
	}

	return []tool.BaseTool{fetchTool, exportTool}, nil
}

func (s *AIService) fetchWebPageTool(ctx context.Context, input fetchWebPageInput) (fetchWebPageResult, error) {
	emitAIToolEvent(ctx, "正在抓取网页内容...")

	rawURL := strings.TrimSpace(input.URL)
	if rawURL == "" {
		rawURL = currentPageURLFromContext(ctx)
	}
	hasScheme := strings.Contains(rawURL, "://")

	fetchURL, err := normalizeFetchURL(rawURL)
	if err != nil {
		return fetchWebPageResult{}, err
	}

	client := &http.Client{
		Timeout: 20 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return errors.New("网页跳转次数过多")
			}
			return validateSafeURL(req.Context(), req.URL)
		},
	}

	doRequest := func(targetURL string) (*http.Response, *neturl.URL, error) {
		parsed, err := neturl.Parse(targetURL)
		if err != nil {
			return nil, nil, errors.New("网页地址格式不正确")
		}
		if err := validateSafeURL(ctx, parsed); err != nil {
			return nil, nil, err
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
		if err != nil {
			return nil, nil, errors.New("创建网页请求失败")
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; AIAgentAssistant/1.0; +https://example.local)")
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,text/plain;q=0.8,*/*;q=0.7")
		req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")

		resp, err := client.Do(req)
		if err != nil {
			return nil, nil, err
		}
		return resp, parsed, nil
	}

	resp, parsedURL, err := doRequest(fetchURL)
	if err != nil && !hasScheme && strings.HasPrefix(fetchURL, "https://") {
		emitAIToolEvent(ctx, "HTTPS 抓取失败，尝试使用 HTTP...")
		httpURL := "http://" + strings.TrimPrefix(fetchURL, "https://")
		resp, parsedURL, err = doRequest(httpURL)
		if err == nil {
			fetchURL = httpURL
		}
	}
	if err != nil {
		return fetchWebPageResult{}, errors.New("抓取网页失败，请检查链接是否可访问或补充 http/https 协议")
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fetchWebPageResult{}, fmt.Errorf("抓取网页失败，状态码为 %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxWebFetchBytes))
	if err != nil {
		return fetchWebPageResult{}, errors.New("读取网页内容失败")
	}

	contentType := strings.ToLower(resp.Header.Get("Content-Type"))
	title := ""
	content := ""

	switch {
	case strings.Contains(contentType, "text/plain"):
		content = cleanExtractedText(string(body))
	default:
		title, content, err = extractReadableHTML(body)
		if err != nil {
			return fetchWebPageResult{}, err
		}
	}

	if content == "" {
		return fetchWebPageResult{}, errors.New("网页正文为空，暂时无法总结")
	}

	content, truncated := truncateRunes(content, maxWebPageContentRunes)
	result := fetchWebPageResult{
		URL:       fetchURL,
		Title:     defaultIfEmpty(title, parsedURL.Hostname()),
		Content:   content,
		Truncated: truncated,
	}
	if truncated {
		result.Notice = "网页内容过长，已截取核心正文片段供总结使用"
	}

	emitAIToolEvent(ctx, "网页内容抓取完成")
	return result, nil
}

func (s *AIService) exportSummaryPDFTool(ctx context.Context, userID uint, input exportSummaryPDFInput) (exportSummaryPDFResult, error) {
	emitAIToolEvent(ctx, "正在导出 PDF...")

	title := strings.TrimSpace(input.Title)
	content := strings.TrimSpace(input.Content)
	if title == "" {
		title = "网页总结"
	}
	if content == "" {
		return exportSummaryPDFResult{}, errors.New("导出 PDF 失败，缺少总结内容")
	}

	fileName := buildPDFFileName(title)
	absPath, err := s.buildAIExportAbsolutePath(ctx, userID, fileName)
	if err != nil {
		return exportSummaryPDFResult{}, err
	}

	if err := renderSummaryPDF(absPath, title, input.SourceURL, content); err != nil {
		return exportSummaryPDFResult{}, err
	}

	oneTimeURL := ""
	ticket, ticketErr := s.createExportDownloadTicket(ctx, userID, fileName)
	if ticketErr == nil && ticket != "" {
		baseURL := strings.TrimSuffix(global.GVA_CONFIG.System.BaseURL, "/")
		if baseURL == "" {
			baseURL = "http://localhost:8080" // 兜底方案
		}
		oneTimeURL = baseURL + "/v1/ai/exports/tickets/" + neturl.PathEscape(ticket)
	}

	result := exportSummaryPDFResult{
		FileName:           fileName,
		OneTimeDownloadURL: oneTimeURL,
		SavedAt:            time.Now().Format(time.RFC3339),
	}
	emitAIToolEvent(ctx, "PDF 导出完成")
	return result, nil
}

type exportTicketPayload struct {
	UserID   uint   `json:"userId"`
	FileName string `json:"fileName"`
}

func (s *AIService) createExportDownloadTicket(ctx context.Context, userID uint, fileName string) (string, error) {
	_ = s
	ticket := strings.ReplaceAll(uuid.NewString(), "-", "")
	payload := exportTicketPayload{UserID: userID, FileName: fileName}
	data, err := json.Marshal(payload)
	if err != nil {
		return "", errors.New("生成下载凭证失败")
	}
	key := "ai:export_ticket:" + ticket
	if err := global.GVA_REDIS.Set(ctx, key, string(data), exportTicketTTL).Err(); err != nil {
		return "", errors.New("保存下载凭证失败")
	}
	return ticket, nil
}

// ConsumeExportDownloadTicket 消费一次性下载凭证并返回对应的 userID 与 fileName。
func (s *AIService) ConsumeExportDownloadTicket(ctx context.Context, ticket string) (uint, string, error) {
	_ = s
	safeTicket := strings.TrimSpace(ticket)
	if safeTicket == "" {
		return 0, "", errors.New("下载凭证不能为空")
	}
	key := "ai:export_ticket:" + safeTicket
	script := `local v = redis.call("GET", KEYS[1]); if not v then return "" end; redis.call("DEL", KEYS[1]); return v;`
	res, err := global.GVA_REDIS.Eval(ctx, script, []string{key}).Result()
	if err != nil {
		return 0, "", errors.New("读取下载凭证失败")
	}
	val, _ := res.(string)
	if strings.TrimSpace(val) == "" {
		return 0, "", errors.New("下载链接已失效")
	}
	var payload exportTicketPayload
	if err := json.Unmarshal([]byte(val), &payload); err != nil {
		return 0, "", errors.New("下载凭证格式错误")
	}
	if payload.UserID == 0 || strings.TrimSpace(payload.FileName) == "" {
		return 0, "", errors.New("下载凭证无效")
	}
	return payload.UserID, payload.FileName, nil
}

// GetExportFilePath 获取当前用户的 AI 导出文件绝对路径。
func (s *AIService) GetExportFilePath(ctx context.Context, userID uint, fileName string) (string, error) {
	_ = ctx
	safeName := filepath.Base(strings.TrimSpace(fileName))
	if safeName == "." || safeName == "" || safeName != fileName {
		return "", errors.New("导出文件名不合法")
	}
	if !strings.HasSuffix(strings.ToLower(safeName), ".pdf") {
		return "", errors.New("仅支持下载 PDF 文件")
	}

	absPath, err := s.buildAIExportAbsolutePath(context.Background(), userID, safeName)
	if err != nil {
		return "", err
	}

	info, statErr := os.Stat(absPath)
	if statErr != nil || info.IsDir() {
		return "", errors.New("导出文件不存在")
	}

	return absPath, nil
}

func (s *AIService) buildAIExportAbsolutePath(ctx context.Context, userID uint, fileName string) (string, error) {
	_ = ctx
	exportDir, err := filepath.Abs(filepath.Join("storage", "ai_exports", strconv.FormatUint(uint64(userID), 10)))
	if err != nil {
		return "", errors.New("导出目录初始化失败")
	}
	if err := os.MkdirAll(exportDir, 0o755); err != nil {
		return "", errors.New("创建导出目录失败")
	}
	return filepath.Join(exportDir, fileName), nil
}

func normalizeFetchURL(rawURL string) (string, error) {
	trimmed := strings.TrimSpace(rawURL)
	if trimmed == "" {
		return "", errors.New("网页地址不能为空")
	}
	if !strings.Contains(trimmed, "://") {
		trimmed = "https://" + trimmed
	}
	parsedURL, err := neturl.Parse(trimmed)
	if err != nil || parsedURL.Host == "" {
		return "", errors.New("网页地址格式不正确")
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return "", errors.New("仅支持抓取 http 或 https 网页")
	}
	return parsedURL.String(), nil
}

func validateSafeURL(ctx context.Context, parsedURL *neturl.URL) error {
	if parsedURL == nil {
		return errors.New("网页地址为空")
	}

	aiConfig := global.GVA_CONFIG.AI
	allowIntranet := aiConfig.AllowIntranetFetch
	allowLocalhost := aiConfig.AllowLocalhostFetch

	host := strings.TrimSpace(parsedURL.Hostname())
	if host == "" {
		return errors.New("网页地址缺少主机名")
	}
	if strings.EqualFold(host, "localhost") || strings.HasSuffix(strings.ToLower(host), ".local") {
		if allowLocalhost {
			return nil
		}
		return errors.New("不允许抓取本地地址（如需启用请在 settings.yaml 设置 ai.allow-localhost-fetch: true）")
	}

	if ip := net.ParseIP(host); ip != nil {
		if isLinkLocalIP(ip) {
			return errors.New("不允许抓取链路本地地址")
		}
		if ip.IsLoopback() || ip.IsUnspecified() {
			if allowLocalhost {
				return nil
			}
			return errors.New("不允许抓取本地地址（如需启用请在 settings.yaml 设置 ai.allow-localhost-fetch: true）")
		}
		if ip.IsPrivate() {
			if allowIntranet {
				return nil
			}
			return errors.New("不允许抓取内网地址（如需启用请在 settings.yaml 设置 ai.allow-intranet-fetch: true）")
		}
		return nil
	}

	addrs, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return errors.New("解析网页地址失败")
	}
	if len(addrs) == 0 {
		return errors.New("网页地址解析结果为空")
	}
	for _, addr := range addrs {
		ip := addr.IP
		if isLinkLocalIP(ip) {
			return errors.New("不允许抓取链路本地地址")
		}
		if ip.IsLoopback() || ip.IsUnspecified() {
			if !allowLocalhost {
				return errors.New("不允许抓取本地地址（如需启用请在 settings.yaml 设置 ai.allow-localhost-fetch: true）")
			}
			continue
		}
		if ip.IsPrivate() {
			if !allowIntranet {
				return errors.New("不允许抓取内网地址（如需启用请在 settings.yaml 设置 ai.allow-intranet-fetch: true）")
			}
			continue
		}
	}
	return nil
}

func isLinkLocalIP(ip net.IP) bool {
	if ip == nil {
		return true
	}
	return ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast()
}

func extractReadableHTML(body []byte) (string, string, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return "", "", errors.New("解析网页内容失败")
	}

	doc.Find("script,style,noscript,iframe,svg,canvas,nav,footer,header,form,button,aside").Each(func(i int, s *goquery.Selection) {
		s.Remove()
	})

	title := cleanExtractedText(doc.Find("title").First().Text())
	root := pickReadableRoot(doc)
	content := extractStructuredText(root)
	if utf8.RuneCountInString(content) < 200 {
		content = cleanExtractedText(root.Text())
	}
	if content == "" {
		content = cleanExtractedText(doc.Find("body").Text())
	}
	if content == "" {
		return "", "", errors.New("未能提取到有效网页正文")
	}

	return title, content, nil
}

func pickReadableRoot(doc *goquery.Document) *goquery.Selection {
	selectors := []string{
		"article",
		"main",
		"[role='main']",
		".article",
		".article-content",
		".post",
		".post-content",
		".entry-content",
		".content",
		"#content",
		".markdown-body",
		".rich_text",
		".reader-content",
	}

	for _, selector := range selectors {
		selection := doc.Find(selector).First()
		if selection.Length() > 0 && strings.TrimSpace(selection.Text()) != "" {
			return selection
		}
	}

	body := doc.Find("body").First()
	if body.Length() > 0 {
		return body
	}
	return doc.Selection
}

func extractStructuredText(root *goquery.Selection) string {
	var parts []string
	root.Find("h1,h2,h3,h4,h5,h6,p,li,pre,blockquote").Each(func(i int, s *goquery.Selection) {
		text := cleanExtractedText(s.Text())
		if text != "" {
			if len(parts) == 0 || parts[len(parts)-1] != text {
				parts = append(parts, text)
			}
		}
	})
	return strings.Join(parts, "\n")
}

func cleanExtractedText(text string) string {
	normalized := strings.ReplaceAll(text, "\u00a0", " ")
	lines := strings.Split(normalized, "\n")
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.Join(strings.Fields(strings.TrimSpace(line)), " ")
		if line != "" {
			result = append(result, line)
		}
	}
	return strings.Join(result, "\n")
}

func truncateRunes(text string, limit int) (string, bool) {
	runes := []rune(text)
	if len(runes) <= limit {
		return text, false
	}
	return strings.TrimSpace(string(runes[:limit])) + "\n\n[内容已截断]", true
}

func defaultIfEmpty(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func buildPDFFileName(title string) string {
	base := strings.TrimSpace(fileNameCleaner.ReplaceAllString(title, "_"))
	base = strings.Trim(base, "_")
	if base == "" {
		base = "summary"
	}
	if utf8.RuneCountInString(base) > 40 {
		base = string([]rune(base)[:40])
	}
	return fmt.Sprintf("%s_%s.pdf", base, uuid.NewString()[:8])
}

func renderSummaryPDF(filePath, title, sourceURL, content string) error {
	fontPath, err := findCJKFontPath()
	if err != nil {
		return err
	}

	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{
		PageSize: *gopdf.PageSizeA4,
		Unit:     gopdf.Unit_MM,
	})
	pdf.SetMargins(pdfMarginLeftMM, pdfMarginTopMM, pdfMarginRightMM, pdfMarginBottomMM)
	pdf.AddPage()

	if err := pdf.AddTTFFont("ai_summary_font", fontPath); err != nil {
		return errors.New("加载 PDF 字体失败")
	}

	y := pdfMarginTopMM
	contentWidth := pdfPageWidthMM - pdfMarginLeftMM - pdfMarginRightMM

	if err := writePDFWrappedParagraph(&pdf, "ai_summary_font", pdfTitleFontSize, pdfTitleLineHeightMM, pdfMarginLeftMM, pdfMarginTopMM, pdfMarginBottomMM, contentWidth, &y, title); err != nil {
		return err
	}
	y += 2

	metaLines := []string{
		"导出时间：" + time.Now().Format("2006-01-02 15:04:05"),
	}
	if strings.TrimSpace(sourceURL) != "" {
		metaLines = append(metaLines, "来源链接："+strings.TrimSpace(sourceURL))
	}
	for _, line := range metaLines {
		if err := writeMarkdownInlineParagraphToPDF(&pdf, "ai_summary_font", pdfMetaFontSize, pdfMetaLineHeightMM, pdfMarginLeftMM, pdfMarginLeftMM, pdfMarginTopMM, pdfMarginBottomMM, contentWidth, contentWidth, &y, line); err != nil {
			return err
		}
	}

	y += 3
	if err := writeMarkdownContentToPDF(&pdf, "ai_summary_font", pdfMarginLeftMM, pdfMarginTopMM, pdfMarginBottomMM, contentWidth, &y, content); err != nil {
		return err
	}

	if err := pdf.WritePdf(filePath); err != nil {
		return errors.New("写入 PDF 文件失败")
	}
	return nil
}

func writeMarkdownContentToPDF(pdf *gopdf.GoPdf, fontName string, left, top, bottom, width float64, y *float64, content string) error {
	blocks := parseMarkdownBlocks(content)
	if len(blocks) == 0 {
		return writeMarkdownInlineParagraphToPDF(pdf, fontName, pdfBodyFontSize, pdfBodyLineHeightMM, left, left, top, bottom, width, width, y, content)
	}

	for _, block := range blocks {
		switch block.Kind {
		case "heading":
			fontSize := pdfHeading3FontSize
			lineHeight := pdfBodyLineHeightMM
			if block.Level <= 1 {
				fontSize = pdfHeading1FontSize
				lineHeight = 8.0
			} else if block.Level == 2 {
				fontSize = pdfHeading2FontSize
				lineHeight = 7.0
			}
			*y += 1.5
			if err := writeMarkdownInlineParagraphToPDF(pdf, fontName, fontSize, lineHeight, left, left, top, bottom, width, width, y, block.Text); err != nil {
				return err
			}
			*y += 1
		case "unordered_list":
			for _, line := range block.Lines {
				if err := writeMarkdownListItemToPDF(pdf, fontName, pdfBodyFontSize, pdfBodyLineHeightMM, left, top, bottom, width, y, "-", line); err != nil {
					return err
				}
			}
		case "ordered_list":
			for idx, line := range block.Lines {
				prefix := fmt.Sprintf("%d.", idx+1)
				if err := writeMarkdownListItemToPDF(pdf, fontName, pdfBodyFontSize, pdfBodyLineHeightMM, left, top, bottom, width, y, prefix, line); err != nil {
					return err
				}
			}
		case "horizontal_rule":
			if err := drawMarkdownHorizontalRule(pdf, left, top, bottom, width, y); err != nil {
				return err
			}
		case "blockquote":
			for _, line := range block.Lines {
				if err := writeMarkdownInlineParagraphToPDF(pdf, fontName, pdfQuoteFontSize, pdfBodyLineHeightMM, left+pdfQuoteIndentMM, left+pdfQuoteIndentMM, top, bottom, width-pdfQuoteIndentMM, width-pdfQuoteIndentMM, y, "引用: "+line); err != nil {
					return err
				}
			}
		case "code":
			if err := writeMarkdownCodeBlockToPDF(pdf, fontName, left, top, bottom, width, y, block.Lines); err != nil {
				return err
			}
		case "table":
			if err := writeMarkdownTableToPDF(pdf, fontName, left, top, bottom, width, y, block.TableHeader, block.TableRows); err != nil {
				return err
			}
		default:
			if err := writeMarkdownInlineParagraphToPDF(pdf, fontName, pdfBodyFontSize, pdfBodyLineHeightMM, left, left, top, bottom, width, width, y, block.Text); err != nil {
				return err
			}
		}
	}

	return nil
}

func parseMarkdownBlocks(content string) []markdownBlock {
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	blocks := make([]markdownBlock, 0, len(lines))
	var paragraph []string
	var listLines []string
	var listKind string
	var quoteLines []string
	var codeLines []string
	inCodeBlock := false

	flushParagraph := func() {
		if len(paragraph) == 0 {
			return
		}
		blocks = append(blocks, markdownBlock{
			Kind: "paragraph",
			Text: strings.Join(paragraph, " "),
		})
		paragraph = nil
	}
	flushList := func() {
		if len(listLines) == 0 {
			return
		}
		blocks = append(blocks, markdownBlock{
			Kind:  listKind,
			Lines: append([]string(nil), listLines...),
		})
		listLines = nil
		listKind = ""
	}
	flushQuote := func() {
		if len(quoteLines) == 0 {
			return
		}
		blocks = append(blocks, markdownBlock{
			Kind:  "blockquote",
			Lines: append([]string(nil), quoteLines...),
		})
		quoteLines = nil
	}
	flushCode := func() {
		if len(codeLines) == 0 {
			return
		}
		blocks = append(blocks, markdownBlock{
			Kind:  "code",
			Lines: append([]string(nil), codeLines...),
		})
		codeLines = nil
	}

	for i := 0; i < len(lines); i++ {
		rawLine := lines[i]
		line := strings.TrimRight(rawLine, " \t")
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "```") {
			flushParagraph()
			flushList()
			flushQuote()
			if inCodeBlock {
				flushCode()
				inCodeBlock = false
			} else {
				inCodeBlock = true
				codeLines = nil
			}
			continue
		}

		if inCodeBlock {
			codeLines = append(codeLines, line)
			continue
		}

		if trimmed == "" {
			flushParagraph()
			flushList()
			flushQuote()
			continue
		}

		if matches := markdownHeadingRE.FindStringSubmatch(trimmed); len(matches) == 3 {
			flushParagraph()
			flushList()
			flushQuote()
			blocks = append(blocks, markdownBlock{
				Kind:  "heading",
				Level: len(matches[1]),
				Text:  matches[2],
			})
			continue
		}

		if isMarkdownHorizontalRule(trimmed) {
			flushParagraph()
			flushList()
			flushQuote()
			blocks = append(blocks, markdownBlock{Kind: "horizontal_rule"})
			continue
		}

		if isMarkdownTableRow(trimmed) && i+1 < len(lines) && isMarkdownTableSeparator(strings.TrimSpace(lines[i+1])) {
			flushParagraph()
			flushList()
			flushQuote()

			header := parseMarkdownTableRow(trimmed)
			var rows [][]string
			i += 2
			for ; i < len(lines); i++ {
				tableLine := strings.TrimSpace(lines[i])
				if !isMarkdownTableRow(tableLine) {
					i--
					break
				}
				rows = append(rows, parseMarkdownTableRow(tableLine))
			}

			if len(header) > 0 {
				blocks = append(blocks, markdownBlock{
					Kind:        "table",
					TableHeader: header,
					TableRows:   rows,
				})
				continue
			}
		}

		if matches := markdownUnorderedRE.FindStringSubmatch(line); len(matches) == 2 {
			flushParagraph()
			flushQuote()
			if listKind != "" && listKind != "unordered_list" {
				flushList()
			}
			listKind = "unordered_list"
			listLines = append(listLines, matches[1])
			continue
		}

		if matches := markdownOrderedRE.FindStringSubmatch(line); len(matches) == 3 {
			flushParagraph()
			flushQuote()
			if listKind != "" && listKind != "ordered_list" {
				flushList()
			}
			listKind = "ordered_list"
			listLines = append(listLines, matches[2])
			continue
		}

		if strings.HasPrefix(trimmed, ">") {
			flushParagraph()
			flushList()
			quoteLine := strings.TrimSpace(strings.TrimPrefix(trimmed, ">"))
			quoteLines = append(quoteLines, quoteLine)
			continue
		}

		flushList()
		flushQuote()
		paragraph = append(paragraph, trimmed)
	}

	if inCodeBlock {
		flushCode()
	}
	flushParagraph()
	flushList()
	flushQuote()

	return blocks
}

func isMarkdownHorizontalRule(line string) bool {
	trimmed := strings.TrimSpace(line)
	if len(trimmed) < 3 {
		return false
	}
	if strings.Trim(trimmed, "-") == "" {
		return true
	}
	if strings.Trim(trimmed, "_") == "" {
		return true
	}
	if strings.Trim(trimmed, "*") == "" {
		return true
	}
	return false
}

func isMarkdownTableRow(line string) bool {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return false
	}
	return strings.Count(trimmed, "|") >= 2
}

func isMarkdownTableSeparator(line string) bool {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return false
	}
	if !isMarkdownTableRow(trimmed) {
		return false
	}

	cells := parseMarkdownTableRow(trimmed)
	if len(cells) == 0 {
		return false
	}
	for _, cell := range cells {
		cell = strings.TrimSpace(cell)
		if cell == "" {
			return false
		}
		for _, r := range cell {
			if r != '-' && r != ':' {
				return false
			}
		}
	}
	return true
}

func parseMarkdownTableRow(line string) []string {
	trimmed := strings.TrimSpace(line)
	trimmed = strings.TrimPrefix(trimmed, "|")
	trimmed = strings.TrimSuffix(trimmed, "|")
	parts := strings.Split(trimmed, "|")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		result = append(result, cleanMarkdownInline(part))
	}
	return result
}

func cleanMarkdownInline(text string) string {
	return cleanMarkdownPlainText(text)
}

func cleanMarkdownPlainText(text string) string {
	cleaned := strings.TrimSpace(text)
	if cleaned == "" {
		return ""
	}

	cleaned = markdownImageRE.ReplaceAllString(cleaned, "$1 ($2)")
	cleaned = markdownLinkRE.ReplaceAllString(cleaned, "$1")
	cleaned = markdownEmphasisRE.ReplaceAllString(cleaned, "")
	cleaned = strings.ReplaceAll(cleaned, "&nbsp;", " ")
	cleaned = strings.ReplaceAll(cleaned, "&lt;", "<")
	cleaned = strings.ReplaceAll(cleaned, "&gt;", ">")
	cleaned = strings.ReplaceAll(cleaned, "&amp;", "&")
	replacer := strings.NewReplacer(
		`\*`, "*",
		`\_`, "_",
		`\[`, "[",
		`\]`, "]",
		`\(`, "(",
		`\)`, ")",
		`\-`, "-",
		`\#`, "#",
		`\+`, "+",
		`\!`, "!",
		`\~`, "~",
		"\\`", "`",
	)
	cleaned = replacer.Replace(cleaned)
	return strings.Join(strings.Fields(cleaned), " ")
}

func parseMarkdownInlineSegments(text string) []markdownInlineSegment {
	raw := strings.TrimSpace(text)
	if raw == "" {
		return nil
	}

	segments := make([]markdownInlineSegment, 0, 8)
	for len(raw) > 0 {
		switch {
		case strings.HasPrefix(raw, "!["):
			if alt, url, rest, ok := parseMarkdownImageToken(raw); ok {
				imageText := cleanMarkdownPlainText(alt)
				if imageText == "" {
					imageText = "图片"
				}
				if linkURL := strings.TrimSpace(url); linkURL != "" {
					segments = append(segments, markdownInlineSegment{Text: imageText, Kind: "link", URL: linkURL})
				} else {
					segments = append(segments, markdownInlineSegment{Text: imageText, Kind: "plain"})
				}
				raw = rest
				continue
			}
		case strings.HasPrefix(raw, "["):
			if label, url, rest, ok := parseMarkdownLinkToken(raw); ok {
				linkText := cleanMarkdownPlainText(label)
				if linkText == "" {
					linkText = strings.TrimSpace(url)
				}
				if linkText != "" {
					segments = append(segments, markdownInlineSegment{Text: linkText, Kind: "link", URL: strings.TrimSpace(url)})
				}
				raw = rest
				continue
			}
		case strings.HasPrefix(raw, "`"):
			if end := strings.Index(raw[1:], "`"); end >= 0 {
				codeText := strings.TrimSpace(raw[1 : end+1])
				if codeText != "" {
					segments = append(segments, markdownInlineSegment{Text: codeText, Kind: "inline_code"})
				}
				raw = raw[end+2:]
				continue
			}
		case strings.HasPrefix(raw, "$"):
			if end := strings.Index(raw[1:], "$"); end >= 0 {
				mathText := strings.TrimSpace(raw[:end+2])
				if mathText != "" {
					segments = append(segments, markdownInlineSegment{Text: mathText, Kind: "math"})
				}
				raw = raw[end+2:]
				continue
			}
		}

		next := len(raw)
		for _, marker := range []string{"![", "[", "`", "$"} {
			if idx := strings.Index(raw, marker); idx >= 0 && idx < next {
				next = idx
			}
		}

		plainText := cleanMarkdownPlainText(raw[:next])
		segments = append(segments, splitPlainTextWithBareLinks(plainText)...)
		raw = raw[next:]
	}

	return compactMarkdownInlineSegments(segments)
}

func compactMarkdownInlineSegments(segments []markdownInlineSegment) []markdownInlineSegment {
	compacted := make([]markdownInlineSegment, 0, len(segments))
	for _, segment := range segments {
		segment.Text = strings.TrimSpace(segment.Text)
		if segment.Text == "" {
			continue
		}
		if len(compacted) > 0 {
			last := &compacted[len(compacted)-1]
			if last.Kind == segment.Kind && last.URL == segment.URL {
				last.Text += " " + segment.Text
				continue
			}
		}
		compacted = append(compacted, segment)
	}
	return compacted
}

func splitPlainTextWithBareLinks(text string) []markdownInlineSegment {
	if strings.TrimSpace(text) == "" {
		return nil
	}

	matches := bareURLRE.FindAllStringIndex(text, -1)
	if len(matches) == 0 {
		return []markdownInlineSegment{{Text: text, Kind: "plain"}}
	}

	segments := make([]markdownInlineSegment, 0, len(matches)*2+1)
	last := 0
	for _, match := range matches {
		if match[0] > last {
			plain := strings.TrimSpace(text[last:match[0]])
			if plain != "" {
				segments = append(segments, markdownInlineSegment{Text: plain, Kind: "plain"})
			}
		}
		urlText := strings.TrimSpace(text[match[0]:match[1]])
		urlText = strings.TrimRightFunc(urlText, func(r rune) bool {
			return unicode.IsPunct(r) && r != '/' && r != ':' && r != '.' && r != '_' && r != '-' && r != '?' && r != '&' && r != '=' && r != '#'
		})
		if urlText != "" {
			segments = append(segments, markdownInlineSegment{Text: urlText, Kind: "link", URL: urlText})
		}
		last = match[1]
	}
	if last < len(text) {
		plain := strings.TrimSpace(text[last:])
		if plain != "" {
			segments = append(segments, markdownInlineSegment{Text: plain, Kind: "plain"})
		}
	}
	return segments
}

func parseMarkdownLinkToken(text string) (label, url, rest string, ok bool) {
	labelEnd := strings.Index(text, "]")
	if labelEnd <= 0 || labelEnd+1 >= len(text) || text[labelEnd+1] != '(' {
		return "", "", "", false
	}
	urlEnd := strings.Index(text[labelEnd+2:], ")")
	if urlEnd < 0 {
		return "", "", "", false
	}
	urlEnd += labelEnd + 2
	return text[1:labelEnd], text[labelEnd+2 : urlEnd], text[urlEnd+1:], true
}

func parseMarkdownImageToken(text string) (alt, url, rest string, ok bool) {
	if !strings.HasPrefix(text, "![") {
		return "", "", "", false
	}
	labelEnd := strings.Index(text[2:], "]")
	if labelEnd < 0 {
		return "", "", "", false
	}
	labelEnd += 2
	if labelEnd+1 >= len(text) || text[labelEnd+1] != '(' {
		return "", "", "", false
	}
	urlEnd := strings.Index(text[labelEnd+2:], ")")
	if urlEnd < 0 {
		return "", "", "", false
	}
	urlEnd += labelEnd + 2
	return text[2:labelEnd], text[labelEnd+2 : urlEnd], text[urlEnd+1:], true
}

func writeMarkdownTableToPDF(pdf *gopdf.GoPdf, fontName string, left, top, bottom, width float64, y *float64, header []string, rows [][]string) error {
	columnCount := len(header)
	if columnCount == 0 {
		return nil
	}
	for _, row := range rows {
		if len(row) > columnCount {
			columnCount = len(row)
		}
	}
	if columnCount == 0 {
		return nil
	}

	normalizedHeader := normalizeTableRow(header, columnCount)
	normalizedRows := make([][]string, 0, len(rows))
	for _, row := range rows {
		normalizedRows = append(normalizedRows, normalizeTableRow(row, columnCount))
	}

	columnWidth := width / float64(columnCount)
	tableTop := *y

	if err := drawMarkdownTableRow(pdf, fontName, left, top, bottom, y, columnWidth, normalizedHeader, true); err != nil {
		return err
	}
	for _, row := range normalizedRows {
		if err := drawMarkdownTableRow(pdf, fontName, left, top, bottom, y, columnWidth, row, false); err != nil {
			return err
		}
	}

	if *y == tableTop {
		*y += pdfTableLineHeightMM
	}
	*y += pdfParagraphSpacingMM
	return nil
}

func normalizeTableRow(row []string, columnCount int) []string {
	normalized := make([]string, columnCount)
	for i := 0; i < columnCount; i++ {
		if i < len(row) {
			normalized[i] = cleanMarkdownInline(row[i])
		}
	}
	return normalized
}

func drawMarkdownTableRow(pdf *gopdf.GoPdf, fontName string, left, top, bottom float64, y *float64, columnWidth float64, row []string, isHeader bool) error {
	fontSize := pdfTableFontSize
	lineHeight := pdfTableLineHeightMM
	if isHeader {
		fontSize = pdfBodyFontSize
		lineHeight = pdfBodyLineHeightMM
	}

	if err := pdf.SetFont(fontName, "", fontSize); err != nil {
		return errors.New("设置 PDF 字体失败")
	}

	cellLines := make([][]string, len(row))
	maxLineCount := 1
	textWidth := columnWidth - pdfTableCellPaddingMM*2
	if textWidth <= 1 {
		textWidth = columnWidth
	}

	for i, cell := range row {
		lines, err := pdf.SplitTextWithWordWrap(defaultIfEmpty(cell, " "), textWidth)
		if err != nil {
			return errors.New("计算 PDF 表格换行失败")
		}
		if len(lines) == 0 {
			lines = []string{" "}
		}
		cellLines[i] = lines
		if len(lines) > maxLineCount {
			maxLineCount = len(lines)
		}
	}

	rowHeight := float64(maxLineCount)*lineHeight + pdfTableCellPaddingMM*2
	if *y+rowHeight > pdfPageHeightMM-bottom {
		pdf.AddPage()
		*y = top
		if err := pdf.SetFont(fontName, "", fontSize); err != nil {
			return errors.New("设置 PDF 字体失败")
		}
	}

	pdf.SetLineWidth(0.1)
	for col := 0; col < len(row); col++ {
		x := left + float64(col)*columnWidth
		pdf.Line(x, *y, x+columnWidth, *y)
		pdf.Line(x, *y+rowHeight, x+columnWidth, *y+rowHeight)
		pdf.Line(x, *y, x, *y+rowHeight)
		pdf.Line(x+columnWidth, *y, x+columnWidth, *y+rowHeight)

		textY := *y + pdfTableCellPaddingMM
		for _, line := range cellLines[col] {
			pdf.SetXY(x+pdfTableCellPaddingMM, textY)
			if err := pdf.Cell(&gopdf.Rect{W: textWidth, H: lineHeight}, line); err != nil {
				return errors.New("写入 PDF 表格内容失败")
			}
			textY += lineHeight
		}
	}

	*y += rowHeight
	return nil
}

func writePDFWrappedParagraph(pdf *gopdf.GoPdf, fontName string, fontSize, lineHeight, left, top, bottom, width float64, y *float64, text string) error {
	if err := pdf.SetFont(fontName, "", fontSize); err != nil {
		return errors.New("设置 PDF 字体失败")
	}

	paragraphs := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	for _, paragraph := range paragraphs {
		trimmed := strings.TrimSpace(paragraph)
		if trimmed == "" {
			*y += lineHeight / 2
			continue
		}

		lines, err := pdf.SplitTextWithWordWrap(trimmed, width)
		if err != nil {
			return errors.New("计算 PDF 换行失败")
		}

		for _, line := range lines {
			if *y+lineHeight > pdfPageHeightMM-bottom {
				pdf.AddPage()
				*y = top
				if err := pdf.SetFont(fontName, "", fontSize); err != nil {
					return errors.New("设置 PDF 字体失败")
				}
			}

			pdf.SetXY(left, *y)
			if err := pdf.Cell(&gopdf.Rect{W: width, H: lineHeight}, line); err != nil {
				return errors.New("写入 PDF 内容失败")
			}
			*y += lineHeight
		}

		*y += pdfParagraphSpacingMM
	}

	return nil
}

func writeMarkdownListItemToPDF(pdf *gopdf.GoPdf, fontName string, fontSize, lineHeight, left, top, bottom, width float64, y *float64, marker, text string) error {
	if err := pdf.SetFont(fontName, "", fontSize); err != nil {
		return errors.New("设置 PDF 字体失败")
	}

	contentLeft := left + pdfListMarkerWidthMM
	contentWidth := width - pdfListMarkerWidthMM
	if contentWidth <= 1 {
		contentWidth = width
		contentLeft = left
	}
	if err := ensurePDFLineSpace(pdf, top, bottom, y, lineHeight); err != nil {
		return err
	}
	if marker != "" {
		pdf.SetXY(left, *y)
		if err := pdf.Cell(&gopdf.Rect{W: pdfListMarkerWidthMM, H: lineHeight}, marker); err != nil {
			return errors.New("写入 PDF 列表标记失败")
		}
	}
	return writeMarkdownInlineParagraphToPDF(pdf, fontName, fontSize, lineHeight, contentLeft, contentLeft, top, bottom, contentWidth, contentWidth, y, text)
}

func writeMarkdownInlineParagraphToPDF(pdf *gopdf.GoPdf, fontName string, fontSize, lineHeight, firstLineLeft, otherLineLeft, top, bottom, firstLineWidth, otherLineWidth float64, y *float64, text string) error {
	segments := parseMarkdownInlineSegments(text)
	if len(segments) == 0 {
		segments = []markdownInlineSegment{{Text: cleanMarkdownPlainText(text), Kind: "plain"}}
	}

	lineIndex := 0
	currentX := firstLineLeft
	lineLeft := firstLineLeft
	lineWidth := firstLineWidth

	startNewLine := func() error {
		*y += lineHeight
		lineIndex++
		lineLeft = otherLineLeft
		lineWidth = otherLineWidth
		currentX = lineLeft
		return ensurePDFLineSpace(pdf, top, bottom, y, lineHeight)
	}

	if err := ensurePDFLineSpace(pdf, top, bottom, y, lineHeight); err != nil {
		return err
	}

	for _, segment := range segments {
		style := buildPDFTextStyle(segment.Kind, fontSize, segment.URL)
		remaining := segment.Text
		for strings.TrimSpace(remaining) != "" {
			if currentX-lineLeft >= lineWidth {
				if err := startNewLine(); err != nil {
					return err
				}
			}

			if currentX == lineLeft {
				remaining = strings.TrimLeftFunc(remaining, unicode.IsSpace)
			}
			if remaining == "" {
				break
			}

			availableWidth := lineWidth - (currentX - lineLeft)
			chunk, rest, err := extractWrappedTextChunk(pdf, fontName, style.FontSize, remaining, availableWidth)
			if err != nil {
				return err
			}
			if chunk == "" {
				if err := startNewLine(); err != nil {
					return err
				}
				continue
			}

			chunkWidth, err := writeStyledTextChunkToPDF(pdf, fontName, lineHeight, currentX, *y, chunk, style)
			if err != nil {
				return err
			}
			currentX += chunkWidth
			remaining = rest
			if strings.TrimSpace(remaining) != "" {
				if err := startNewLine(); err != nil {
					return err
				}
			}
		}
	}

	if lineIndex == 0 && currentX == lineLeft {
		*y += lineHeight
	}
	*y += pdfParagraphSpacingMM
	pdf.SetTextColor(0, 0, 0)
	return nil
}

func buildPDFTextStyle(kind string, fontSize float64, externalURL string) pdfTextStyle {
	style := pdfTextStyle{
		FontSize:  fontSize,
		TextColor: pdfColor{R: 0, G: 0, B: 0},
	}

	switch kind {
	case "link":
		style.TextColor = pdfColor{R: 26, G: 115, B: 232}
		style.Underline = true
		style.ExternalURL = externalURL
	case "inline_code":
		style.TextColor = pdfColor{R: 180, G: 35, B: 24}
		style.FillColor = &pdfColor{R: 245, G: 245, B: 245}
	case "math":
		style.TextColor = pdfColor{R: 105, G: 52, B: 171}
		style.FillColor = &pdfColor{R: 247, G: 243, B: 255}
	default:
	}

	return style
}

func ensurePDFLineSpace(pdf *gopdf.GoPdf, top, bottom float64, y *float64, lineHeight float64) error {
	if *y+lineHeight <= pdfPageHeightMM-bottom {
		return nil
	}
	pdf.AddPage()
	*y = top
	return nil
}

func extractWrappedTextChunk(pdf *gopdf.GoPdf, fontName string, fontSize float64, text string, maxWidth float64) (string, string, error) {
	if strings.TrimSpace(text) == "" {
		return "", "", nil
	}
	if maxWidth <= 1 {
		return "", text, nil
	}
	if err := pdf.SetFont(fontName, "", fontSize); err != nil {
		return "", "", errors.New("设置 PDF 字体失败")
	}

	runes := []rune(text)
	lastBreak := -1
	for idx := range runes {
		candidate := string(runes[:idx+1])
		width, err := pdf.MeasureTextWidth(candidate)
		if err != nil {
			return "", "", errors.New("测量 PDF 文本宽度失败")
		}
		if width <= maxWidth {
			if unicode.IsSpace(runes[idx]) {
				lastBreak = idx + 1
			}
			continue
		}

		breakPos := idx
		if lastBreak > 0 {
			breakPos = lastBreak
		}
		if breakPos <= 0 {
			breakPos = 1
		}

		fit := strings.TrimRightFunc(string(runes[:breakPos]), unicode.IsSpace)
		rest := strings.TrimLeftFunc(string(runes[breakPos:]), unicode.IsSpace)
		if fit == "" && len(runes) > 0 {
			fit = string(runes[:1])
			rest = strings.TrimLeftFunc(string(runes[1:]), unicode.IsSpace)
		}
		return fit, rest, nil
	}

	return string(runes), "", nil
}

func writeStyledTextChunkToPDF(pdf *gopdf.GoPdf, fontName string, lineHeight, x, y float64, text string, style pdfTextStyle) (float64, error) {
	if err := pdf.SetFont(fontName, "", style.FontSize); err != nil {
		return 0, errors.New("设置 PDF 字体失败")
	}

	textWidth, err := pdf.MeasureTextWidth(text)
	if err != nil {
		return 0, errors.New("测量 PDF 文本宽度失败")
	}

	if style.FillColor != nil {
		pdf.SetFillColor(style.FillColor.R, style.FillColor.G, style.FillColor.B)
		pdf.RectFromUpperLeftWithStyle(x-0.2, y+0.4, textWidth+0.8, lineHeight-0.8, "F")
	}

	pdf.SetTextColor(style.TextColor.R, style.TextColor.G, style.TextColor.B)
	pdf.SetXY(x, y)
	if err := pdf.Cell(&gopdf.Rect{W: textWidth + 0.2, H: lineHeight}, text); err != nil {
		return 0, errors.New("写入 PDF 内容失败")
	}

	if style.Underline {
		pdf.SetStrokeColor(style.TextColor.R, style.TextColor.G, style.TextColor.B)
		underlineY := y + lineHeight - 0.8
		pdf.SetLineWidth(0.18)
		pdf.Line(x, underlineY, x+textWidth, underlineY)
	}
	if strings.TrimSpace(style.ExternalURL) != "" {
		pdf.AddExternalLink(strings.TrimSpace(style.ExternalURL), x, y, textWidth, lineHeight)
	}
	return textWidth, nil
}

func writeMarkdownCodeBlockToPDF(pdf *gopdf.GoPdf, fontName string, left, top, bottom, width float64, y *float64, lines []string) error {
	codeLeft := left + pdfListIndentMM
	codeWidth := width - pdfListIndentMM
	if codeWidth <= 1 {
		codeLeft = left
		codeWidth = width
	}

	contentWidth := codeWidth - pdfCodeBlockPaddingMM*2
	if contentWidth <= 1 {
		contentWidth = codeWidth
	}

	if err := pdf.SetFont(fontName, "", pdfCodeFontSize); err != nil {
		return errors.New("设置 PDF 字体失败")
	}

	for idx, rawLine := range lines {
		lineText := strings.ReplaceAll(rawLine, "\t", "    ")
		if strings.TrimSpace(lineText) == "" {
			lineText = " "
		}

		wrappedLines, err := pdf.SplitTextWithWordWrap(lineText, contentWidth)
		if err != nil {
			return errors.New("计算 PDF 代码块换行失败")
		}
		if len(wrappedLines) == 0 {
			wrappedLines = []string{" "}
		}

		for _, wrappedLine := range wrappedLines {
			if err := ensurePDFLineSpace(pdf, top, bottom, y, pdfBodyLineHeightMM); err != nil {
				return err
			}
			pdf.SetFillColor(246, 248, 250)
			pdf.RectFromUpperLeftWithStyle(codeLeft, *y, codeWidth, pdfBodyLineHeightMM, "F")
			pdf.SetTextColor(31, 35, 40)
			pdf.SetXY(codeLeft+pdfCodeBlockPaddingMM, *y)
			if err := pdf.Cell(&gopdf.Rect{W: contentWidth, H: pdfBodyLineHeightMM}, wrappedLine); err != nil {
				return errors.New("写入 PDF 代码块内容失败")
			}
			*y += pdfBodyLineHeightMM
		}

		if idx == len(lines)-1 {
			break
		}
	}

	pdf.SetTextColor(0, 0, 0)
	*y += pdfParagraphSpacingMM
	return nil
}

func drawMarkdownHorizontalRule(pdf *gopdf.GoPdf, left, top, bottom, width float64, y *float64) error {
	ruleSpacing := pdfParagraphSpacingMM + 1
	ruleY := *y + ruleSpacing
	if ruleY > pdfPageHeightMM-bottom {
		pdf.AddPage()
		*y = top
		ruleY = *y + ruleSpacing
	}

	pdf.SetLineWidth(0.2)
	pdf.Line(left, ruleY, left+width, ruleY)
	*y = ruleY + ruleSpacing
	return nil
}

func findCJKFontPath() (string, error) {
	candidates := []string{
		`C:\Windows\Fonts\simhei.ttf`,
		`C:\Windows\Fonts\Deng.ttf`,
		`C:\Windows\Fonts\NotoSansSC-VF.ttf`,
		`C:\Windows\Fonts\Source Han Serif SC Heavy (TrueType).ttf`,
		`C:\Windows\Fonts\STSONG.TTF`,
	}

	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate, nil
		}
	}

	return "", errors.New("未找到可用的中文字体，无法导出 PDF")
}

func withAIToolEventSender(ctx context.Context, sender aiToolEventSender) context.Context {
	return context.WithValue(ctx, aiToolEventSenderCtxKey{}, sender)
}

func withAICurrentPageURL(ctx context.Context, currentPageURL string) context.Context {
	trimmed := strings.TrimSpace(currentPageURL)
	if ctx == nil || trimmed == "" {
		return ctx
	}
	return context.WithValue(ctx, aiCurrentPageURLCtxKey{}, trimmed)
}

func currentPageURLFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	currentPageURL, ok := ctx.Value(aiCurrentPageURLCtxKey{}).(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(currentPageURL)
}

func emitAIToolEvent(ctx context.Context, content string) {
	if ctx == nil || strings.TrimSpace(content) == "" {
		return
	}
	sender, ok := ctx.Value(aiToolEventSenderCtxKey{}).(aiToolEventSender)
	if ok && sender != nil {
		sender(content)
	}
}
