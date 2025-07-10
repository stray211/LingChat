package service

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/sashabaranov/go-openai"

	"LingChat/api/routes/common"
	"LingChat/api/routes/v1/response"
	"LingChat/api/routes/ws/types"
	"LingChat/internal/clients/VitsTTS"
	"LingChat/internal/clients/emotionPredictor"
	"LingChat/internal/clients/llm"
	"LingChat/internal/data/ent/ent"
	"LingChat/internal/data/ent/ent/conversationmessage"
)

type LingChatService struct {
	emotionPredictorClient *emotionPredictor.Client
	VitsTTSClient          *VitsTTS.Client
	llmClient              *llm.LLMClient
	conversationService    *ConversationService
	ConfigModel            string
	tempFilePath           string
	voiceStatusTracker     *VoiceStatusTracker
}

func NewLingChatService(
	epClient *emotionPredictor.Client,
	vtClient *VitsTTS.Client,
	llmClient *llm.LLMClient,
	conversationService *ConversationService,
	configModel string,
	path string,
) (*LingChatService, error) {

	statusTracker, err := NewVoiceStatusTracker()
	if err != nil {
		return nil, fmt.Errorf("failed to create voice status tracker: %w", err)
	}

	return &LingChatService{
		emotionPredictorClient: epClient,
		VitsTTSClient:          vtClient,
		llmClient:              llmClient,
		conversationService:    conversationService,
		ConfigModel:            configModel,
		tempFilePath:           path,
		voiceStatusTracker:     statusTracker,
	}, nil
}

func (l *LingChatService) LingChat(ctx context.Context, message string, conversationID, prevMessageID string, characterID string) (*response.CompletionResponse, error) {

	useLegacyTempChatContext := common.UseLegacyTempChatContext(ctx)
	var (
		conv       *ent.Conversation
		userMsgObj *ent.ConversationMessage
		messages   []openai.ChatCompletionMessage
		respMsg    *ent.ConversationMessage
		err        error
	)

	if useLegacyTempChatContext {
		l.conversationService.legacyTempChatContext.AddMessage(openai.ChatCompletionMessage{
			Role:    string(conversationmessage.RoleUser),
			Content: message,
		})
		messages = l.conversationService.legacyTempChatContext.DumpMessage()
	} else {
		// 记录会话和消息
		conv, userMsgObj, err = l.conversationService.RecordConversationAndMessage(ctx, message, conversationID, prevMessageID, characterID)
		if err != nil {
			return nil, err
		}

		// 获取消息链
		messages, err = l.conversationService.GetChatContext(ctx, userMsgObj.ID)
		if err != nil {
			return nil, err
		}
	}

	// 调用LLM获取回复
	rawLLMResp, err := l.llmClient.Chat(ctx, messages, l.ConfigModel)
	if err != nil {
		err = fmt.Errorf("LLM Chat error: %w", err)
		return nil, err
	}

	if useLegacyTempChatContext {
		l.conversationService.legacyTempChatContext.AddMessage(openai.ChatCompletionMessage{
			Role:    string(conversationmessage.RoleAssistant),
			Content: rawLLMResp,
		})
	} else {
		// 将助手回复保存到数据库
		respMsg, err = l.conversationService.SaveAssistantMessage(ctx, userMsgObj.ID, rawLLMResp)
		if err != nil {
			log.Printf("保存助手回复失败: %s", err)
			return nil, err
		}
	}

	emotionSegments := AnalyzeEmotions(rawLLMResp)
	emotionSegments = GenerateVoiceFileNames(emotionSegments, "wav", l.voiceStatusTracker)

	// TODO: 这里两条会耦合使用emotionSegments的字段，后面要改
	l.GenerateVoice(ctx, emotionSegments, l.tempFilePath, true)
	emotionSegments = l.EmoPredictBatch(ctx, emotionSegments)

	convID := 0
	respMsgID := 0
	if conv != nil {
		convID = int(conv.ID)
	}
	if respMsg != nil {
		respMsgID = int(respMsg.ID)
	}
	return &response.CompletionResponse{
		ConversationID: strconv.Itoa(convID),
		MessageID:      strconv.Itoa(respMsgID),
		Messages:       l.CreateResponse(emotionSegments, message),
	}, nil
}

func (l *LingChatService) GetChatHistory(ctx context.Context) []openai.ChatCompletionMessage {
	return l.conversationService.GetChatHistory(ctx)
}

func (l *LingChatService) LoadChatHistory(ctx context.Context, msg []openai.ChatCompletionMessage) []openai.ChatCompletionMessage {
	return l.conversationService.LoadChatHistory(ctx, msg)
}

func (l *LingChatService) CreateResponse(results []Result, userMessage string) []types.Response {
	var resp []types.Response
	for i, result := range results {
		resp = append(resp, types.Response{
			Type:            "reply",
			Emotion:         result.Predicted,
			OriginalTag:     result.OriginalTag,
			Message:         result.FollowingText,
			MotionText:      result.MotionText,
			AudioFile:       result.VoiceFile,
			OriginalMessage: userMessage,
			IsMultiPart:     true,
			PartIndex:       i,
			TotalParts:      len(results),
		})
	}
	return resp
}

func (l *LingChatService) EmoPredictBatch(ctx context.Context, results []Result) []Result {
	// 创建4秒超时的上下文
	timeoutCtx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	resultsChannel := make(chan struct {
		index      int
		Predicted  string
		Confidence float64
	}, len(results))
	for i, result := range results {
		wg.Add(1)
		go func(index int, result Result) {
			defer wg.Done()
			resp, err := l.emotionPredictorClient.Predict(timeoutCtx, result.OriginalTag, 0.08)
			if err != nil {
				log.Printf("Failed to predict emotion: %v", err)
				resultsChannel <- struct {
					index      int
					Predicted  string
					Confidence float64
				}{
					index, "unknown", 0.0,
				}
			} else {
				resultsChannel <- struct {
					index      int
					Predicted  string
					Confidence float64
				}{
					index, resp.Label, resp.Confidence,
				}
			}
		}(i, result)
	}

	go func() {
		wg.Wait()
		close(resultsChannel)
	}()

	for result := range resultsChannel {
		index := result.index
		results[index].Confidence = result.Confidence
		results[index].Predicted = result.Predicted
	}
	return results
}

func (l *LingChatService) GenerateVoice(ctx context.Context, textSegments []Result, tempFilePath string, saveFile bool) {
	// 为每个文本片段启动一个独立的goroutine，立即返回
	for i, segment := range textSegments {
		go func(idx int, text, voiceFileName string, statusTracker *VoiceStatusTracker) {
			var status string

			// 调用VITS TTS服务生成语音
			audioData, err := l.VitsTTSClient.VoiceVITS(ctx, text)
			if err != nil {
				log.Printf("Failed to generate voice for segment %d: %v", idx, err)
				status = StatusFailed
			} else {
				// 如果需要保存文件且有音频数据
				if saveFile && len(audioData) != 0 {
					// 根据当前时间的小时数奇偶性决定子目录（以便定时清理）
					hour := time.Now().Hour()
					subDir := "even"
					if hour%2 == 1 {
						subDir = "odd"
					}

					// 拼接完整的文件路径：tempFilePath/odd或even/voiceFileName
					fullVoicePath := filepath.Join(tempFilePath, subDir, voiceFileName)

					// 确保目录存在
					dir := filepath.Dir(fullVoicePath)
					if err := os.MkdirAll(dir, 0755); err != nil {
						log.Printf("Failed to create directory %s: %v", dir, err)
						status = StatusFailed
					} else {
						// 写入文件
						if err := os.WriteFile(fullVoicePath, audioData, 0644); err != nil {
							log.Printf("Failed to write file %s: %v", fullVoicePath, err)
							status = StatusFailed
						} else {
							status = StatusReady
						}
					}
				} else {
					status = StatusReady // 即使不保存文件，生成成功也算ready
				}
			}

			// 更新状态
			if statusTracker != nil {
				if err := statusTracker.UpdateStatus(voiceFileName, status); err != nil {
					log.Printf("Failed to update status for %s to %s: %v", voiceFileName, status, err)
				}
			}
		}(i, segment.JapaneseText, segment.VoiceFile, l.voiceStatusTracker)
	}
}

// GetVoiceFileStatus 根据文件名查询语音文件状态
func (l *LingChatService) GetVoiceFileStatus(filename string) VoiceFileStatus {
	return l.voiceStatusTracker.GetStatus(filename)
}

// Close 关闭服务，清理资源
func (l *LingChatService) Close() error {
	return l.voiceStatusTracker.Close()
}
