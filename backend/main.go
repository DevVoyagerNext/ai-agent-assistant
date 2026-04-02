package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/volcengine/volcengine-go-sdk/service/arkruntime"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
)

func main() {
	client := arkruntime.NewClientWithApiKey("dd1b9139-7673-4021-a483-9c1ab1fed5e9")
	ctx := context.Background()

	fmt.Println("----- multimodal embeddings request -----")
	req := model.MultiModalEmbeddingRequest{
		Model: "ep-20260402205720-hctdx",
		Input: []model.MultimodalEmbeddingInput{
			{
				Type:     model.MultiModalEmbeddingInputTypeImageURL,
				ImageURL: &model.MultimodalEmbeddingImageURL{URL: "https://ark-project.tos-cn-beijing.volces.com/images/view.jpeg"},
			},
		},
	}

	resp, err := client.CreateMultiModalEmbeddings(ctx, req)
	if err != nil {
		fmt.Printf("multimodal embeddings error: %v\n", err)
		return
	}

	s, _ := json.Marshal(resp)
	fmt.Println(string(s))
}

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
