package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// 配置 Kimi (Moonshot AI)
const (
	API_KEY = "sk-DZmdKHgsqXyVyAYY8rflyVGofvFe2sNlDSHB0KHsfZpFfPbf" 
	
	//Kimi 的官方接口地址
	API_URL = "https://api.moonshot.cn/v1/chat/completions"
	//模型名称
	MODEL_NAME = "moonshot-v1-8k"
)

//定义请求结构 (兼容 OpenAI 格式)
type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"` //控制回答的随机性
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

//定义响应结构
type ChatResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

//AskNPC 发送请求给AI
func AskNPC(npcName string, npcPersona string, userContent string) string {
	fmt.Printf("[AI] 正在请求 Kimi: %s 说: %s\n", npcName, userContent)

	//构造 Prompt (人设 + 剧情)
	systemPrompt := fmt.Sprintf(`你现在是文字MUD游戏里的NPC [%s]。
你的设定是：%s。
玩家对你说：“%s”。
请用符合你身份的口语回答，不要太长，50字以内。
不要说“我是AI”之类的话，要完全沉浸在角色里。`, npcName, npcPersona, userContent)

	reqBody := ChatRequest{
		Model: MODEL_NAME,
		Messages: []Message{
			{Role: "system", Content: "你是一个专业的角色扮演辅助AI。"},
			{Role: "user", Content: systemPrompt},
		},
		Temperature: 0.7,
	}

	jsonData, _ := json.Marshal(reqBody)

	//发送 HTTP 请求
	req, _ := http.NewRequest("POST", API_URL, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+API_KEY)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("[AI] 请求失败:", err)
		return "（NPC 似乎没听清你说什么...）"
	}
	defer resp.Body.Close()

	// 3. 解析结果
	body, _ := io.ReadAll(resp.Body)
	
	var chatResp ChatResponse
	json.Unmarshal(body, &chatResp)

	// 错误处理
	if chatResp.Error.Message != "" {
		fmt.Println("[AI] API 返回错误:", chatResp.Error.Message)
		return "（NPC 此时头有点痛，不想理你）"
	}

	if len(chatResp.Choices) > 0 {
		return chatResp.Choices[0].Message.Content
	}
	return "..."
}