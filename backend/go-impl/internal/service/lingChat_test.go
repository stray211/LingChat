package service

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"

	"LingChat/internal/clients/VitsTTS"
	"LingChat/internal/clients/emotionPredictor"
	"LingChat/internal/clients/llm"
	"LingChat/internal/config"
	"LingChat/internal/data"
)

var conf *config.Config
var service *LingChatService
var ctx = context.Background()

func init() {
	fmt.Println(os.Getwd())
	err := godotenv.Load()
	if err != nil {
		log.Fatal("无法加载 .env 文件: ", err)
	}
	conf = config.GetConfigFromEnv()
	fmt.Println(conf.Chat.BaseURL)

	emotionPredictorClient := emotionPredictor.NewClient(conf.Emotion.URL)
	vitsTTSClient := VitsTTS.NewClient(conf.Vits.APIURL, conf.TempDirs.VoiceDir, 0)
	llmClient := llm.NewLLMClient(conf.Chat.BaseURL, conf.Chat.APIKey)

	// init Data & Repos
	entClient, err := data.NewEntClient(ctx, conf.Data.DataBase.Driver, conf.Data.DataBase.Source, conf.Data.DataBase.AutoMigrate)
	if err != nil {
		log.Fatal(err)
	}
	d, _, err := data.NewData(entClient, nil)
	conversationRepo := data.NewConversationRepo(d)
	legacyTempChatContext := data.NewLegacyTempChatContext()

	conversationService := NewConversationService(conversationRepo, legacyTempChatContext, conf.Chat.Model)
	service = NewLingChatService(emotionPredictorClient, vitsTTSClient, llmClient, conversationService, conf.Chat.Model, conf.TempDirs.VoiceDir)
}

func Test_ChatAndParse(t *testing.T) {
	rawResp, err := service.llmClient.Chat(ctx, []openai.ChatCompletionMessage{
		{
			Role:    "system",
			Content: data.SystemPrompt,
		}, {
			Role:    "user",
			Content: "你好",
		},
	}, "deepseek-chat")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(AnalyzeEmotions(rawResp, service.tempFilePath, "wav"))
}

func Test_LingChat(t *testing.T) {
	fmt.Println(service.LingChat(ctx, "你好",
		"", "",
	))
}
