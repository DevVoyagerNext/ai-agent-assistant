package service

import (
	"bytes"
	"context"
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
)

var fileNameCleaner = regexp.MustCompile(`[^a-zA-Z0-9\p{Han}_-]+`)

type aiToolEventSender func(content string)

type aiToolEventSenderCtxKey struct{}

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
	FileName    string `json:"fileName"`
	DownloadURL string `json:"downloadUrl"`
	SavedAt     string `json:"savedAt"`
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
		"用于把已经整理好的总结内容导出为 PDF，并返回可下载地址。适用于用户明确要求导出 PDF、生成 PDF 或保存总结结果的场景。输入应为最终总结内容，而不是原始网页全文。",
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

	fetchURL, err := normalizeFetchURL(input.URL)
	if err != nil {
		return fetchWebPageResult{}, err
	}

	parsedURL, err := neturl.Parse(fetchURL)
	if err != nil {
		return fetchWebPageResult{}, errors.New("网页地址格式不正确")
	}

	if err := validateSafeURL(ctx, parsedURL); err != nil {
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

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fetchURL, nil)
	if err != nil {
		return fetchWebPageResult{}, errors.New("创建网页请求失败")
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; AIAgentAssistant/1.0; +https://example.local)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,text/plain;q=0.8,*/*;q=0.7")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		return fetchWebPageResult{}, errors.New("抓取网页失败，请稍后重试")
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

	result := exportSummaryPDFResult{
		FileName:    fileName,
		DownloadURL: "/v1/ai/exports/" + neturl.PathEscape(fileName),
		SavedAt:     time.Now().Format(time.RFC3339),
	}
	emitAIToolEvent(ctx, "PDF 导出完成")
	return result, nil
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

	host := strings.TrimSpace(parsedURL.Hostname())
	if host == "" {
		return errors.New("网页地址缺少主机名")
	}
	if strings.EqualFold(host, "localhost") || strings.HasSuffix(strings.ToLower(host), ".local") {
		return errors.New("不允许抓取本地或内网地址")
	}

	if ip := net.ParseIP(host); ip != nil {
		if isPrivateIP(ip) {
			return errors.New("不允许抓取本地或内网地址")
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
		if isPrivateIP(addr.IP) {
			return errors.New("不允许抓取本地或内网地址")
		}
	}
	return nil
}

func isPrivateIP(ip net.IP) bool {
	if ip == nil {
		return true
	}
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified()
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
		if err := writePDFWrappedParagraph(&pdf, "ai_summary_font", pdfMetaFontSize, pdfMetaLineHeightMM, pdfMarginLeftMM, pdfMarginTopMM, pdfMarginBottomMM, contentWidth, &y, line); err != nil {
			return err
		}
	}

	y += 3
	if err := writePDFWrappedParagraph(&pdf, "ai_summary_font", pdfBodyFontSize, pdfBodyLineHeightMM, pdfMarginLeftMM, pdfMarginTopMM, pdfMarginBottomMM, contentWidth, &y, content); err != nil {
		return err
	}

	if err := pdf.WritePdf(filePath); err != nil {
		return errors.New("写入 PDF 文件失败")
	}
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

func emitAIToolEvent(ctx context.Context, content string) {
	if ctx == nil || strings.TrimSpace(content) == "" {
		return
	}
	sender, ok := ctx.Value(aiToolEventSenderCtxKey{}).(aiToolEventSender)
	if ok && sender != nil {
		sender(content)
	}
}
