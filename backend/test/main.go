package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

// 定义节点结构
type Node struct {
	ID       int64
	ParentID int64
	Name     string
	Level    int
	Content  string
}

func main() {
	// 1. 数据库配置
	dsn := "root:123456@tcp(127.0.0.1:3306)/ai_study_assistant?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}
	defer db.Close()

	// 2. 初始化学科 (Subjects)
	subjectName := "数据结构"
	subjectSlug := "data_structure"
	res, err := db.Exec("INSERT IGNORE INTO subjects (slug, name, description) VALUES (?, ?, ?)",
		subjectSlug, subjectName, "探索数据的组织、管理和存储格式")
	if err != nil {
		log.Fatal("初始化学科失败:", err)
	}
	subjectID, _ := res.LastInsertId()
	if subjectID == 0 {
		// 如果已存在，查出 ID
		db.QueryRow("SELECT id FROM subjects WHERE slug = ?", subjectSlug).Scan(&subjectID)
	}
	fmt.Printf("已锁定学科: %s (ID: %d)\n", subjectName, subjectID)

	// 3. 解析 Markdown 文件
	filePath := "./数据结构-树.md"
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal("无法打开教材文件:", err)
	}
	defer file.Close()

	// 正则匹配层级
	reHeader := regexp.MustCompile(`^(#+)\s+(.*)$`)

	var lastLevelNodes [5]int64 // 记录每一层最后的 ID，[1]篇, [2]章, [3]节, [4]点
	lastLevelNodes[0] = 0       // 根节点的 ParentID 为 0

	// 特殊处理：文件名作为 Level 1 (篇)
	rootID := insertNode(db, subjectID, 0, "树结构全解析", 1)
	lastLevelNodes[1] = rootID

	scanner := bufio.NewScanner(file)
	var currentKPID int64
	var contentBuffer strings.Builder

	fmt.Println("开始解析并同步至数据库...")

	for scanner.Scan() {
		line := scanner.Text()
		match := reHeader.FindStringSubmatch(line)

		if match != nil {
			// 如果进入新标题，先保存上一个知识点的内容
			if currentKPID > 0 && contentBuffer.Len() > 0 {
				saveContent(db, currentKPID, contentBuffer.String())
				contentBuffer.Reset()
			}

			levelSign := match[1]
			title := match[2]
			currentLevel := len(levelSign) + 1 // Markdown ## 是二级，对应我们结构的 Level 2 (章)

			if currentLevel >= 2 && currentLevel <= 4 {
				parentID := lastLevelNodes[currentLevel-1]
				nodeID := insertNode(db, subjectID, parentID, title, currentLevel)
				lastLevelNodes[currentLevel] = nodeID

				if currentLevel == 4 {
					currentKPID = nodeID
				} else {
					currentKPID = 0 // 非知识点级别不挂载内容
				}
			}
		} else {
			// 收集内容
			if currentKPID > 0 {
				contentBuffer.WriteString(line + "\n")
			}
		}
	}

	// 保存最后一个知识点内容
	if currentKPID > 0 && contentBuffer.Len() > 0 {
		saveContent(db, currentKPID, contentBuffer.String())
	}

	fmt.Println("🎉 教材同步完成！")
}

// 插入节点并返回 ID
func insertNode(db *sql.DB, subjectID int64, parentID int64, name string, level int) int64 {
	query := "INSERT INTO knowledge_nodes (subject_id, parent_id, name, level) VALUES (?, ?, ?, ?)"
	res, err := db.Exec(query, subjectID, parentID, name, level)
	if err != nil {
		log.Printf("插入节点失败 [%s]: %v", name, err)
		return 0
	}
	id, _ := res.LastInsertId()
	return id
}

// 保存知识点内容
func saveContent(db *sql.DB, nodeID int64, content string) {
	query := "INSERT INTO knowledge_contents (node_id, content, source) VALUES (?, ?, ?)"
	_, err := db.Exec(query, nodeID, content, "Auto-Importer")
	if err != nil {
		log.Printf("保存内容失败 (NodeID: %d): %v", nodeID, err)
	}
}

//func main() {
//	client := arkruntime.NewClientWithApiKey("dd1b9139-7673-4021-a483-9c1ab1fed5e9")
//	ctx := context.Background()
//
//	fmt.Println("----- multimodal embeddings request -----")
//	req := model.MultiModalEmbeddingRequest{
//		Model: "ep-20260402205720-hctdx",
//		Input: []model.MultimodalEmbeddingInput{
//			{
//				Type:     model.MultiModalEmbeddingInputTypeImageURL,
//				ImageURL: &model.MultimodalEmbeddingImageURL{URL: "https://ark-project.tos-cn-beijing.volces.com/images/view.jpeg"},
//			},
//		},
//	}
//
//	resp, err := client.CreateMultiModalEmbeddings(ctx, req)
//	if err != nil {
//		fmt.Printf("multimodal embeddings error: %v\n", err)
//		return
//	}
//
//	s, _ := json.Marshal(resp)
//	fmt.Println(string(s))
//}

//func main() {
//	config := ark.DefaultConfig("c1caf12a-2051-4c68-b5e8-5e64a1de7549")
//	config.BaseURL = "https://ark.cn-beijing.volces.com/api/v3"
//	client := ark.NewClientWithConfig(config)
//
//	fmt.Println("----- standard request -----")
//	resp, err := client.CreateChatCompletion(
//		context.Background(),
//		ark.ChatCompletionRequest{
//			Model: "ep-20260402201453-x2tj2",
//			Messages: []ark.ChatCompletionMessage{
//				{
//					Role:    ark.ChatMessageRoleSystem,
//					Content: "你是人工智能助手",
//				},
//				{
//					Role:    ark.ChatMessageRoleUser,
//					Content: "你好",
//				},
//			},
//		},
//	)
//	if err != nil {
//		fmt.Printf("ChatCompletion error: %v\n", err)
//		return
//	}
//	fmt.Println(resp.Choices[0].Message.Content)
//
//	fmt.Println("----- streaming request -----")
//	stream, err := client.CreateChatCompletionStream(
//		context.Background(),
//		ark.ChatCompletionRequest{
//			Model: "ep-20260402201453-x2tj2",
//			Messages: []ark.ChatCompletionMessage{
//				{
//					Role:    ark.ChatMessageRoleSystem,
//					Content: "你是人工智能助手",
//				},
//				{
//					Role:    ark.ChatMessageRoleUser,
//					Content: "你好,我是一名大三的大学生,我心情很坏,你来安慰一下我",
//				},
//			},
//		},
//	)
//	if err != nil {
//		fmt.Printf("stream chat error: %v\n", err)
//		return
//	}
//	defer stream.Close()
//
//	for {
//		recv, err := stream.Recv()
//		if err == io.EOF {
//			return
//		}
//		if err != nil {
//			fmt.Printf("Stream chat error: %v\n", err)
//			return
//		}
//
//		if len(recv.Choices) > 0 {
//			fmt.Print(recv.Choices[0].Delta.Content)
//		}
//	}
//}
