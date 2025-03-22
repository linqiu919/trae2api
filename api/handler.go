package api

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"bufio"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/trae2api/config"
	"github.com/trae2api/pkg/logger"
)

type ModelResponse struct {
	Object string  `json:"object"`
	Data   []Model `json:"data"`
}

type Model struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
}

type ChatMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

type ChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Stream      bool          `json:"stream"`
	Temperature float64       `json:"temperature,omitempty"`
}

type ContextResolver struct {
	ResolverID string `json:"resolver_id"`
	Variables  string `json:"variables"`
}

type LastLLMResponseInfo struct {
	Turn     int    `json:"turn"`
	IsError  bool   `json:"is_error"`
	Response string `json:"response"`
}

type TraeRequest struct {
	UserInput                  string               `json:"user_input"`
	IntentName                 string               `json:"intent_name"`
	Variables                  string               `json:"variables"`
	ContextResolvers           []ContextResolver    `json:"context_resolvers"`
	GenerateSuggestedQuestions bool                 `json:"generate_suggested_questions"`
	ChatHistory                []ChatHistory        `json:"chat_history"`
	SessionID                  string               `json:"session_id"`
	ConversationID             string               `json:"conversation_id"`
	CurrentTurn                int                  `json:"current_turn"`
	ValidTurns                 []int                `json:"valid_turns"`
	MultiMedia                 []interface{}        `json:"multi_media"`
	ModelName                  string               `json:"model_name"`
	LastLLMResponseInfo        *LastLLMResponseInfo `json:"last_llm_response_info,omitempty"`
}

type ChatHistory struct {
	Role      string `json:"role"`
	SessionID string `json:"session_id"`
	Locale    string `json:"locale"`
	Content   string `json:"content"`
	Status    string `json:"status"`
}

type TraeModelResponse struct {
	ModelConfigs []struct {
		CustomConfig string `json:"custom_config"`
		DisplayName  string `json:"display_name"`
		IsDefault    bool   `json:"is_default"`
		Multimodal   bool   `json:"multimodal"`
		Name         string `json:"name"`
	} `json:"model_configs"`
}

func GetModels(c *gin.Context) {
	// 检查 RefreshToken 是否过期
	if config.IsRefreshTokenExpired() {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": map[string]interface{}{
				"message": "RefreshToken 已过期，请更新环境变量中的 REFRESH_TOKEN",
				"type":    "token_expired",
				"code":    http.StatusUnauthorized,
			},
		})
		return
	}

	client := &http.Client{}
	url := fmt.Sprintf("%s/api/ide/v1/model_list?type=chat", config.AppConfig.BaseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-app-id", config.AppConfig.AppID)
	req.Header.Set("x-ide-version", "1.0.10")
	req.Header.Set("x-ide-version-code", "20250303")
	req.Header.Set("x-ide-version-type", "stable")
	req.Header.Set("x-ide-token", config.GetCurrentToken())
	req.Header.Set("accept", "*/*")

	resp, err := client.Do(req)
	if err != nil {
		logger.Log.Errorf("请求模型列表失败: %v, url: %s", err, url)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		errMsg := fmt.Sprintf("API返回错误状态码 %d: %s", resp.StatusCode, string(body))
		fmt.Printf("Error: %s\n", errMsg)
		c.JSON(resp.StatusCode, gin.H{"error": errMsg})
		return
	}

	// 读取响应内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 记录原始响应
	logger.Log.WithFields(logrus.Fields{
		"response": string(body),
	}).Debug("收到原始响应")

	var traeResp TraeModelResponse
	if err := json.Unmarshal(body, &traeResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 记录解析后的响应
	logger.Log.WithFields(logrus.Fields{
		"models": traeResp,
	}).Info("模型列表解析完成")

	// 转换为OpenAI格式的响应
	var models ModelResponse
	models.Object = "list"
	models.Data = make([]Model, 0)

	// 将 Trae 的模型数据转换为 OpenAI 格式
	for _, m := range traeResp.ModelConfigs {
		if m.Name == "aws_sdk_claude37_sonnet" {
			m.Name = "claude-3-7-sonnet"
		}
		if m.Name == "claude3.5" {
			m.Name = "claude-3-5-sonnet"
		}
		models.Data = append(models.Data, Model{
			ID:      m.Name,
			Object:  "model",
			Created: time.Now().Unix(),
		})
	}

	c.JSON(http.StatusOK, models)
}

// 转换模型名称
func convertModelName(model string) string {
	switch model {
	case "claude-3-5-sonnet-20240620", "claude-3-5-sonnet-20241022", "claude-3-5-sonnet":
		return "claude3.5"
	case "claude-3-7-sonnet-20250219", "claude-3-7-sonnet", "claude-3-7":
		return "aws_sdk_claude37_sonnet"
	case "gpt-4o-mini,gpt-4o-mini-2024-07-18", "gpt-4o-latest":
		return "gpt-4o"
	case "deepseek-chat", "deepseek-coder", "deepseek-v3":
		return "deepseek-V3"
	case "deepseek-reasoner", "deepseek-r1":
		return "deepseek-R1"
	default:
		return model
	}
}

// 使用整个对话历史生成会话ID
func generateSessionIDFromMessages(messages []ChatMessage) string {
	// 将所有消息连接成一个字符串
	var conversationKey strings.Builder
	for _, msg := range messages[:1] { // 只使用第一轮对话来生成sessionID
		conversationKey.WriteString(msg.Role)
		conversationKey.WriteString(": ")
		conversationKey.WriteString(fmt.Sprintf("%v", msg.Content))
		conversationKey.WriteString("\n")
	}

	// 计算hash
	h := sha256.New()
	h.Write([]byte(conversationKey.String()))
	return fmt.Sprintf("session_%x", h.Sum(nil)[:8])
}

// 生成随机设备类型
func generateDeviceType() string {
	types := []string{"windows", "linux", "macos"}
	return types[rand.Intn(len(types))]
}

// 获取系统类型
func getSystemType(deviceType string) string {
	switch deviceType {
	case "windows":
		return "Windows"
	case "linux":
		return "Linux"
	case "macos":
		return "macOS"
	default:
		return "Windows"
	}
}

// 在设置请求头的地方修改为:
func setRequestHeaders(req *http.Request) {
	// 获取当前设备信息
	device := config.GetCurrentDevice()

	// 基础请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-app-id", config.AppConfig.AppID)
	req.Header.Set("x-ide-version", "1.2.1")
	req.Header.Set("x-ide-version-code", "20250318")
	req.Header.Set("x-ide-version-type", "stable")
	req.Header.Set("x-ide-token", config.GetCurrentToken())
	req.Header.Set("accept", "*/*")

	// 设置设备相关的请求头
	req.Header.Set("x-device-cpu", device.DeviceCPU)
	req.Header.Set("x-device-id", device.DeviceID)
	req.Header.Set("x-machine-id", device.MachineID)
	req.Header.Set("x-device-brand", device.DeviceBrand)
	req.Header.Set("x-device-type", device.DeviceType)
	req.Header.Set("x-os-version", device.OSVersion)

	// 记录设备信息
	logger.Log.WithFields(logrus.Fields{
		"deviceCPU":   device.DeviceCPU,
		"deviceID":    device.DeviceID,
		"machineID":   device.MachineID,
		"deviceBrand": device.DeviceBrand,
		"deviceType":  device.DeviceType,
		"osVersion":   device.OSVersion,
		"systemType":  device.SystemType,
		"useCount":    device.UseCount,
		"maxUses":     device.MaxUses,
	}).Info("本次请求使用的设备信息")
}

func CreateChatCompletion(c *gin.Context) {
	// 检查 RefreshToken 是否过期
	if config.IsRefreshTokenExpired() {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": map[string]interface{}{
				"message": "RefreshToken 已过期，请更新环境变量中的 REFRESH_TOKEN",
				"type":    "token_expired",
				"code":    http.StatusUnauthorized,
			},
		})
		return
	}

	var openAIReq ChatRequest
	if err := c.BindJSON(&openAIReq); err != nil {
		logger.Log.Errorf("解析请求体失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 控制台打印标准请求体Json格式数据
	reqJson, err := json.Marshal(openAIReq)
	if err != nil {
		logger.Log.Errorf("JSON编码失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	fmt.Printf("当前对话请求: %v\n", string(reqJson))

	// 添加内容格式转换逻辑
	for i, msg := range openAIReq.Messages {
		switch v := msg.Content.(type) {
		case []interface{}:
			// 如果是数组，只保留第一条消息的text内容
			contentStr := ""
			if len(v) > 0 {
				if msgObj, ok := v[0].(map[string]interface{}); ok {
					if text, ok := msgObj["text"].(string); ok {
						contentStr = text
					}
				}
			}
			openAIReq.Messages[i].Content = contentStr
		case string:
			// 如果已经是字符串，无需转换
			continue
		default:
			// 其他类型转换为字符串
			openAIReq.Messages[i].Content = fmt.Sprintf("%v", v)
		}
	}

	// 生成会话ID
	sessionID := generateSessionIDFromMessages(openAIReq.Messages)

	// 转换模型名称
	openAIReq.Model = convertModelName(openAIReq.Model)

	// 构建 context_resolvers
	contextResolvers := []ContextResolver{
		{
			ResolverID: "project-labels",
			Variables:  "{\"labels\":\"- go\\n- go.mod\"}",
		},
		{
			ResolverID: "terminal_context",
			Variables:  "{\"terminal_context\":[]}",
		},
	}

	// 获取最后一条消息的内容并转换为字符串
	lastContent := fmt.Sprintf("%v", openAIReq.Messages[len(openAIReq.Messages)-1].Content)

	// 获取随机设备类型
	deviceType := generateDeviceType()

	// 构建 variables
	variablesJSON := struct {
		Language               string `json:"language"`
		Locale                 string `json:"locale"`
		Input                  string `json:"input"`
		VersionCode            string `json:"version_code"`
		IsInlineChat           bool   `json:"is_inline_chat"`
		IsCommand              bool   `json:"is_command"`
		RawInput               string `json:"raw_input"`
		Problem                string `json:"problem"`
		CurrentFilename        string `json:"current_filename"`
		IsSelectCodeBeforeChat bool   `json:"is_select_code_before_chat"`
		LastSelectTime         int64  `json:"last_select_time"`
		LastTurnSession        string `json:"last_turn_session"`
		HashWorkspace          bool   `json:"hash_workspace"`
		HashFile               int    `json:"hash_file"`
		HashCode               int    `json:"hash_code"`
		UseFilepath            bool   `json:"use_filepath"`
		CurrentTime            string `json:"current_time"`
		BadgeClickable         bool   `json:"badge_clickable"`
		WorkspacePath          string `json:"workspace_path"`
		Brand                  string `json:"brand"`
		SystemType             string `json:"system_type"`
	}{
		Language:       "",
		Locale:         "zh-cn",
		Input:          lastContent,
		VersionCode:    "20250318",
		RawInput:       lastContent,
		IsInlineChat:   false,
		IsCommand:      false,
		UseFilepath:    true,
		CurrentTime:    time.Now().Format("20060102 15:04:05，星期二"),
		BadgeClickable: true,
		WorkspacePath:  generateRandomWorkspacePath(),
		Brand:          "trae",
		SystemType:     getSystemType(deviceType),
	}

	// 转换历史消息
	chatHistory := make([]ChatHistory, 0)
	for _, msg := range openAIReq.Messages[:len(openAIReq.Messages)-1] {
		var locale string
		if msg.Role == "assistant" {
			locale = "zh-cn"
		}

		chatHistory = append(chatHistory, ChatHistory{
			Role:      msg.Role,
			Content:   fmt.Sprintf("%v", msg.Content),
			Status:    "success",
			Locale:    locale,
			SessionID: sessionID,
		})
	}

	// 设置 LastLLMResponseInfo
	var lastLLMResponseInfo *LastLLMResponseInfo
	if len(chatHistory) > 0 {
		lastMsg := chatHistory[len(chatHistory)-1]
		if lastMsg.Role == "assistant" {
			lastLLMResponseInfo = &LastLLMResponseInfo{
				Turn:     len(chatHistory) - 1, // 修正 turn 计数
				IsError:  false,
				Response: lastMsg.Content,
			}
			variablesJSON.LastTurnSession = sessionID
		}
	}

	// 创建 ValidTurns 切片
	validTurns := make([]int, len(chatHistory))
	for i := range validTurns {
		validTurns[i] = i
	}

	variablesStr, err := json.Marshal(variablesJSON)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	traeReq := TraeRequest{
		UserInput:                  lastContent,
		IntentName:                 "general_qa_intent",
		Variables:                  string(variablesStr),
		ContextResolvers:           contextResolvers,
		GenerateSuggestedQuestions: false,
		ChatHistory:                chatHistory,
		SessionID:                  sessionID,
		ConversationID:             sessionID,
		CurrentTurn:                len(openAIReq.Messages) - 1,
		ValidTurns:                 validTurns,
		MultiMedia:                 []interface{}{},
		ModelName:                  openAIReq.Model,
		LastLLMResponseInfo:        lastLLMResponseInfo,
	}

	jsonData, err := json.Marshal(traeReq)
	if err != nil {
		errMsg := fmt.Sprintf("JSON编码失败: %v", err)
		fmt.Printf("Error: %s\n", errMsg)
		c.JSON(http.StatusInternalServerError, gin.H{"error": errMsg})
		return
	}

	// 在发送请求前记录完整的请求信息
	logger.Log.WithFields(logrus.Fields{
		"url":          fmt.Sprintf("%s/api/ide/v1/chat", config.AppConfig.BaseURL),
		"requestBody":  string(jsonData),
		"sessionID":    sessionID,
		"model":        openAIReq.Model,
		"messageCount": len(openAIReq.Messages),
	}).Info("发送聊天请求")

	url := fmt.Sprintf("%s/api/ide/v1/chat", config.AppConfig.BaseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		errMsg := fmt.Sprintf("请求失败: %v", err)
		fmt.Printf("Error: %s\n", errMsg)
		c.JSON(http.StatusInternalServerError, gin.H{"error": errMsg})
		return
	}

	setRequestHeaders(req)

	// 记录请求头
	headers := make(map[string]string)
	for k, v := range req.Header {
		headers[k] = v[0]
	}
	logger.Log.WithFields(logrus.Fields{
		"headers": headers,
	}).Debug("请求头信息")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errMsg := fmt.Sprintf("请求远端失败: %v", err)
		logger.Log.Errorf("%s", errMsg)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": map[string]interface{}{
				"message": errMsg,
				"type":    "service_unavailable",
				"code":    http.StatusServiceUnavailable,
			},
		})
		return
	}
	defer resp.Body.Close()

	// 检查响应状态码并直接返回对应的错误
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		errMsg := fmt.Sprintf("远程服务返回错误: %s", string(body))
		logger.Log.Errorf("状态码: %d, 错误信息: %s", resp.StatusCode, errMsg)

		var errorType string
		switch resp.StatusCode {
		case http.StatusBadRequest:
			errorType = "invalid_request"
		case http.StatusUnauthorized:
			errorType = "unauthorized"
		case http.StatusForbidden:
			errorType = "permission_denied"
		case http.StatusNotFound:
			errorType = "not_found"
		case http.StatusTooManyRequests:
			errorType = "rate_limit_exceeded"
		default:
			errorType = "internal_server_error"
		}

		c.JSON(resp.StatusCode, gin.H{
			"error": map[string]interface{}{
				"message": errMsg,
				"type":    errorType,
				"code":    resp.StatusCode,
			},
		})
		return
	}

	// 读取响应
	reader := bufio.NewReader(resp.Body)
	thinkStartType := new(bool)
	thinkEndType := new(bool)
	if !openAIReq.Stream {
		// 非流式响应，需要收集所有内容
		var fullResponse string
		var lastFinishReason string

		for {
			line, err := reader.ReadString('\n')
			if err == io.EOF {
				break
			}
			if err != nil {
				errMsg := fmt.Sprintf("读取响应出错: %v", err)
				fmt.Printf("Error: %s\n", errMsg)
				c.JSON(http.StatusInternalServerError, gin.H{"error": errMsg})
				return
			}

			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			if strings.HasPrefix(line, "event: ") {
				event := strings.TrimPrefix(line, "event: ")
				dataLine, err := reader.ReadString('\n')
				if err != nil {
					continue
				}
				dataLine = strings.TrimSpace(dataLine)
				if !strings.HasPrefix(dataLine, "data: ") {
					continue
				}
				data := strings.TrimPrefix(dataLine, "data: ")

				switch event {
				case "output":
					// 先打印原始数据
					fmt.Printf("原始数据: %s\n", data)
					logger.Log.WithFields(logrus.Fields{
						"rawData": data,
					}).Info("收到原始数据")

					var outputData struct {
						Response         string `json:"response"`
						ReasoningContent string `json:"reasoning_content"`
						FinishReason     string `json:"finish_reason"`
					}
					var deltaContent string
					if err := json.Unmarshal([]byte(data), &outputData); err != nil {
						logger.Log.Errorf("解析输出数据失败: %v, data: %s", err, data)
						continue
					}

					// 打印解析后的完整结构
					fmt.Printf("解析后数据: %+v\n", outputData)

					if outputData.Response == "" && outputData.ReasoningContent == "" {
						continue
					}

					// 记录最后的结束原因
					if outputData.FinishReason != "" {
						lastFinishReason = outputData.FinishReason
						logger.Log.WithFields(logrus.Fields{
							"finishReason": lastFinishReason,
							"event":        "finish_reason_update",
						}).Info("更新结束原因")

						// 直接打印到控制台
						fmt.Printf("更新结束原因: %s\n", lastFinishReason)
					}

					// thinking start
					if outputData.ReasoningContent != "" {
						if !*thinkStartType {
							deltaContent = "<think>\n\n" + outputData.ReasoningContent
							*thinkStartType = true
							*thinkEndType = false
						} else {
							deltaContent = outputData.ReasoningContent
						}
					}

					// thinking end
					if outputData.Response != "" {
						if *thinkStartType && !*thinkEndType {
							deltaContent = "</think>\n\n" + outputData.Response
							*thinkStartType = false
							*thinkEndType = true
						} else {
							deltaContent = outputData.Response
						}
					}

					fullResponse += deltaContent

				case "done":
					var doneData struct {
						FinishReason string `json:"finish_reason"`
					}

					if err := json.Unmarshal([]byte(data), &doneData); err != nil {
						logger.Log.Errorf("解析done事件数据失败: %v, data: %s", err, data)
					} else if doneData.FinishReason != "" {
						lastFinishReason = doneData.FinishReason
						logger.Log.WithFields(logrus.Fields{
							"finishReason": lastFinishReason,
							"event":        "done",
						}).Info("从done事件更新finish_reason")
					}

					// 添加更多详细信息到日志
					logger.Log.WithFields(logrus.Fields{
						"autoContinueEnabled": config.AutoContinueEnabled,
						"lastFinishReason":    lastFinishReason,
						"model":               openAIReq.Model,
						"fullResponse":        len(fullResponse),
						"event":               "done",
						"hasFinishReason":     lastFinishReason != "",
					}).Info("检查流式响应是否需要自动继续")

					// 直接打印到控制台
					fmt.Printf("完成事件: lastFinishReason=%s, responseLen=%d\n",
						lastFinishReason,
						len(fullResponse))

					// 如果启用了自动继续且是因为长度限制而结束
					if config.AutoContinueEnabled == "true" && lastFinishReason == "length" && openAIReq.Model == "aws_sdk_claude37_sonnet" {
						logger.Log.Info("流式响应触发自动继续条件，准备发起新请求")

						// 创建继续对话的请求
						continueMessages := append(openAIReq.Messages, ChatMessage{
							Role:    "assistant",
							Content: fullResponse,
						}, ChatMessage{
							Role:    "user",
							Content: "继续",
						})

						// 记录继续请求的消息数量
						logger.Log.WithFields(logrus.Fields{
							"originalMessageCount": len(openAIReq.Messages),
							"newMessageCount":      len(continueMessages),
						}).Info("创建流式继续对话的消息列表")

						// 创建新的请求对象
						continueReq := ChatRequest{
							Model:       openAIReq.Model,
							Messages:    continueMessages,
							Stream:      openAIReq.Stream,
							Temperature: openAIReq.Temperature,
						}

						// 将新请求序列化为JSON
						jsonData, err := json.Marshal(continueReq)
						if err != nil {
							logger.Log.Errorf("序列化继续请求失败: %v", err)
							return
						}

						// 创建新的请求上下文
						newContext := &gin.Context{
							Request: &http.Request{
								Method: "POST",
								Header: c.Request.Header.Clone(), // 复制原始请求的header
								Body:   io.NopCloser(bytes.NewReader(jsonData)),
							},
							Writer: c.Writer,
						}

						// 调用处理函数
						CreateChatCompletion(newContext)
						return
					}

					// 发送完成标记
					openAIResponse := map[string]interface{}{
						"id":      fmt.Sprintf("chatcmpl-%d", time.Now().Unix()),
						"object":  "chat.completion.chunk",
						"created": time.Now().Unix(),
						"model":   openAIReq.Model,
						"choices": []map[string]interface{}{
							{
								"index":         0,
								"delta":         map[string]interface{}{},
								"finish_reason": lastFinishReason,
							},
						},
					}
					responseJSON, _ := json.Marshal(openAIResponse)
					c.Writer.Write([]byte("data: " + string(responseJSON) + "\n\n"))
					c.Writer.Write([]byte("data: [DONE]\n\n"))
					c.Writer.Flush()
					return
				}
			}
		}
		return
	}

	// 流式响应处理
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")

	var lastFinishReason string
	var fullResponse string

	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			c.SSEvent("error", gin.H{"error": err.Error()})
			return
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "event: ") {
			event := strings.TrimPrefix(line, "event: ")
			dataLine, err := reader.ReadString('\n')
			if err != nil {
				continue
			}
			dataLine = strings.TrimSpace(dataLine)
			if !strings.HasPrefix(dataLine, "data: ") {
				continue
			}
			data := strings.TrimPrefix(dataLine, "data: ")

			switch event {
			case "output":
				// 先打印原始数据
				fmt.Printf("原始数据: %s\n", data)
				logger.Log.WithFields(logrus.Fields{
					"rawData": data,
				}).Info("收到原始数据")

				var outputData struct {
					Response         string `json:"response"`
					ReasoningContent string `json:"reasoning_content"`
					FinishReason     string `json:"finish_reason"`
				}
				var deltaContent string
				if err := json.Unmarshal([]byte(data), &outputData); err != nil {
					logger.Log.Errorf("解析输出数据失败: %v, data: %s", err, data)
					continue
				}

				// 打印解析后的完整结构
				fmt.Printf("解析后数据: %+v\n", outputData)

				if outputData.Response == "" && outputData.ReasoningContent == "" {
					continue
				}

				// 记录最后的结束原因
				if outputData.FinishReason != "" {
					lastFinishReason = outputData.FinishReason
					logger.Log.WithFields(logrus.Fields{
						"finishReason": lastFinishReason,
						"event":        "finish_reason_update",
					}).Info("更新结束原因")

					// 直接打印到控制台
					fmt.Printf("更新结束原因: %s\n", lastFinishReason)
				}

				// thinking start
				if outputData.ReasoningContent != "" {
					if !*thinkStartType {
						deltaContent = "<think>\n\n" + outputData.ReasoningContent
						*thinkStartType = true
						*thinkEndType = false
					} else {
						deltaContent = outputData.ReasoningContent
					}
				}

				// thinking end
				if outputData.Response != "" {
					if *thinkStartType && !*thinkEndType {
						deltaContent = "</think>\n\n" + outputData.Response
						*thinkStartType = false
						*thinkEndType = true
					} else {
						deltaContent = outputData.Response
					}
				}

				fullResponse += deltaContent

				// 转换为 OpenAI 流式格式
				openAIResponse := map[string]interface{}{
					"id":      fmt.Sprintf("chatcmpl-%d", time.Now().Unix()),
					"object":  "chat.completion.chunk",
					"created": time.Now().Unix(),
					"model":   openAIReq.Model,
					"choices": []map[string]interface{}{
						{
							"index": 0,
							"delta": map[string]interface{}{
								"content": deltaContent,
							},
							"finish_reason": nil,
						},
					},
				}
				responseJSON, _ := json.Marshal(openAIResponse)
				c.Writer.Write([]byte("data: " + string(responseJSON) + "\n\n"))
				c.Writer.Flush()

			case "done":
				// 解析 done 事件数据
				var doneData struct {
					FinishReason string `json:"finish_reason"`
				}

				if err := json.Unmarshal([]byte(data), &doneData); err != nil {
					logger.Log.Errorf("解析done事件数据失败: %v, data: %s", err, data)
				} else if doneData.FinishReason != "" {
					lastFinishReason = doneData.FinishReason
					logger.Log.WithFields(logrus.Fields{
						"finishReason": lastFinishReason,
						"event":        "done",
					}).Info("从done事件更新finish_reason")
				}

				// 添加更多详细信息到日志
				logger.Log.WithFields(logrus.Fields{
					"autoContinueEnabled": config.AutoContinueEnabled,
					"lastFinishReason":    lastFinishReason,
					"model":               openAIReq.Model,
					"fullResponse":        len(fullResponse),
					"event":               "done",
					"hasFinishReason":     lastFinishReason != "",
				}).Info("检查流式响应是否需要自动继续")

				// 直接打印到控制台
				fmt.Printf("完成事件: lastFinishReason=%s, responseLen=%d\n",
					lastFinishReason,
					len(fullResponse))

				// 如果启用了自动继续且是因为长度限制而结束
				if config.AutoContinueEnabled == "true" && lastFinishReason == "length" && openAIReq.Model == "aws_sdk_claude37_sonnet" {
					logger.Log.Info("流式响应触发自动继续条件，准备发起新请求")

					// 创建继续对话的请求
					continueMessages := append(openAIReq.Messages, ChatMessage{
						Role:    "assistant",
						Content: fullResponse,
					}, ChatMessage{
						Role:    "user",
						Content: "继续",
					})

					// 记录继续请求的消息数量
					logger.Log.WithFields(logrus.Fields{
						"originalMessageCount": len(openAIReq.Messages),
						"newMessageCount":      len(continueMessages),
					}).Info("创建流式继续对话的消息列表")

					// 创建新的请求对象
					continueReq := ChatRequest{
						Model:       openAIReq.Model,
						Messages:    continueMessages,
						Stream:      openAIReq.Stream,
						Temperature: openAIReq.Temperature,
					}

					// 将新请求序列化为JSON
					jsonData, err := json.Marshal(continueReq)
					if err != nil {
						logger.Log.Errorf("序列化继续请求失败: %v", err)
						return
					}

					// 创建新的请求上下文
					newContext := &gin.Context{
						Request: &http.Request{
							Method: "POST",
							Header: c.Request.Header.Clone(), // 复制原始请求的header
							Body:   io.NopCloser(bytes.NewReader(jsonData)),
						},
						Writer: c.Writer,
					}

					// 调用处理函数
					CreateChatCompletion(newContext)
					return
				}

				// 发送完成标记
				openAIResponse := map[string]interface{}{
					"id":      fmt.Sprintf("chatcmpl-%d", time.Now().Unix()),
					"object":  "chat.completion.chunk",
					"created": time.Now().Unix(),
					"model":   openAIReq.Model,
					"choices": []map[string]interface{}{
						{
							"index":         0,
							"delta":         map[string]interface{}{},
							"finish_reason": lastFinishReason,
						},
					},
				}
				responseJSON, _ := json.Marshal(openAIResponse)
				c.Writer.Write([]byte("data: " + string(responseJSON) + "\n\n"))
				c.Writer.Write([]byte("data: [DONE]\n\n"))
				c.Writer.Flush()
				return
			}
		}
	}
}

// LoginPageHandler 处理登录页面请求
func LoginPageHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{
		"error": "",
	})
}

// VerifyPasswordHandler 处理密码验证
func VerifyPasswordHandler(c *gin.Context) {
	password := c.PostForm("password")

	// 固定密码为 linuxdo
	if password == "linuxdo" {
		// 设置 Cookie 标记已验证
		c.SetCookie("authenticated", "true", 3600*24, "/", "", false, true)

		// 重定向到首页
		c.Redirect(http.StatusFound, "/index")
	} else {
		// 密码错误，显示错误信息
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{
			"error": "密码错误，请重试",
		})
	}
}

// IndexHandler 登录验证检查
func IndexHandler(c *gin.Context) {
	// 检查 Cookie 是否已验证
	authenticated, _ := c.Cookie("authenticated")
	if authenticated != "true" {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// 已验证，显示正常页面
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title": "Trae2API - 配置指南",
	})
}

// generateRandomWorkspacePath 生成随机工作空间路径
func generateRandomWorkspacePath() string {
	// 定义可能的根目录列表
	rootDirs := []string{
		"/User",
		"/home",
		"/workspace",
		"/data",
		"/opt",
		"/var/lib",
	}

	dirs := []string{"projects", "workspace", "dev", "code", "work"}

	rand.Seed(time.Now().UnixNano())

	// 生成随机用户名（5-8位，字母开头）
	username := generateRandomUsername(5 + rand.Intn(4))

	// 生成8-12位随机项目名
	projectName := generateRandomString(8 + rand.Intn(5))

	return filepath.Join(
		rootDirs[rand.Intn(len(rootDirs))],
		username,
		"Documents",
		dirs[rand.Intn(len(dirs))],
		"project-"+projectName,
	)
}

// 随机用户名生成函数（字母开头）
func generateRandomUsername(length int) string {
	const (
		letters = "abcdefghijklmnopqrstuvwxyz"
		charset = letters + "0123456789"
	)

	if length < 2 {
		length = 2
	}

	b := make([]byte, length)
	// 首字符必须是字母
	b[0] = letters[rand.Intn(len(letters))]

	// 剩余字符可以是字母或数字
	for i := 1; i < length; i++ {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// 随机字符串函数
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
