package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"

	"LingChat/api/routes"
	v1 "LingChat/api/routes/v1"
	"LingChat/api/routes/ws"
	"LingChat/internal/clients/VitsTTS"
	"LingChat/internal/clients/emotionPredictor"
	"LingChat/internal/clients/llm"
	"LingChat/internal/config"
	"LingChat/internal/cron"
	"LingChat/internal/data"
	"LingChat/internal/service"
	"LingChat/pkg/jwt"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// init Config
	err := godotenv.Load()
	if err != nil && os.Getenv("LINGCHAT_ENV_EXISTS") != "true" {
		log.Fatal("无法加载 .env 文件: ", err)
	}
	conf := config.GetConfigFromEnv()

	// init pkg instances
	secBytes, err := base64.StdEncoding.DecodeString(conf.Server.JWTSecret)
	if err != nil {
		log.Fatal("decode jwt secret failed: ", err)
	}
	j := jwt.NewJWT(secBytes, "LingChat-Backend")

	// init Clients
	emotionPredictorClient := emotionPredictor.NewClient(conf.Emotion.URL)
	vitsTTSClient := VitsTTS.NewClient(conf.Vits.APIURL, conf.TempDirs.VoiceDir, conf.Vits.SpeakerID)
	llmClient := llm.NewLLMClient(conf.Chat.BaseURL, conf.Chat.APIKey)

	// init Cron Client
	cronClient := cron.NewClient(conf.TempDirs.VoiceDir)
	if err := cronClient.StartAsync(); err != nil {
		log.Fatal("Failed to start cron scheduler: ", err)
	}
	defer cronClient.Stop()

	// init Data & Repos
	entClient, err := data.NewEntClient(ctx, conf.Data.DataBase.Driver, conf.Data.DataBase.Source, conf.Data.DataBase.AutoMigrate)
	if err != nil {
		log.Fatal(err)
	}
	d, cleanup, err := data.NewData(entClient, nil)
	defer cleanup()
	if err != nil {
		log.Fatal(err)
	}
	userRepo := data.NewUserRepo(d)
	conversationRepo := data.NewConversationRepo(d)
	characterRepo := data.NewCharacterRepo(d)
	legacyTempChatContext := data.NewLegacyTempChatContext()

	// init Service
	userService := service.NewUserService(userRepo, j)
	conversationService := service.NewConversationService(conversationRepo, characterRepo, legacyTempChatContext, conf.Chat.Model)
	characterService := service.NewCharacterService(characterRepo, conf)
	backgroundService := service.NewBackgroundService(conf)
	chatService, err := service.NewLingChatService(emotionPredictorClient, vitsTTSClient, llmClient, conversationService, conf.Chat.Model, conf.TempDirs.VoiceDir)
	if err != nil {
		log.Fatal("Failed to create LingChatService: ", err)
	}
	defer func() {
		if err := chatService.Close(); err != nil {
			log.Printf("Failed to close LingChatService: %v", err)
		}
	}()

	// init HTTP server
	chatRoute := v1.NewChatRoute(chatService, conversationService, characterService, backgroundService, userRepo, j)
	userRoute := v1.NewUserRoute(userService, userRepo, j)
	webRoute := v1.NewWebRoute(conf.Backend.StaticDir, conf.TempDirs.VoiceDir)
	httpEngine := routes.NewHTTPEngine(
		// Server Addr
		fmt.Sprintf("%s:%d", conf.Backend.BindAddr, conf.Backend.Port),
		// WebSocket Handler
		ws.NewWebSocketHandler(ws.NewWebSocketEngine(chatService).LingChatHandler),
		// HTTP REST API routes
		chatRoute, userRoute, webRoute,
	)

	// Start server
	if _, err := httpEngine.Run(); err != nil {
		log.Fatal(err)
	}
}
