package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/taranpreetnatt/Discord-AI-Bot/internal/discord"
	"github.com/taranpreetnatt/Discord-AI-Bot/internal/logger"
)

type Config struct {
	APP_ID     string
	PUBLIC_KEY string
	BOT_TOKEN  string
	API_BASE   string
	ENV        string
}

func loadConfig() Config {
	env, ok := os.LookupEnv("ENV")
	if !ok {
		log.Fatal("Missing ENV environment variable")
	}

	if env != "production" {
		envErr := godotenv.Load()
		if envErr != nil {
			log.Fatalf("Could not load env from .env file: %v", envErr)
		}
	}
	return Config{
		APP_ID:     getEnv("APP_ID"),
		PUBLIC_KEY: getEnv("PUBLIC_KEY"),
		BOT_TOKEN:  getEnv("BOT_TOKEN"),
		API_BASE:   getEnv("API_BASE"),
		ENV:        env,
	}
}

func getEnv(key string) string {
	if env, ok := os.LookupEnv(key); ok {
		return env
	}
	log.Fatalf("Environment variable %s not set", key)
	return ""
}

func main() {
	config := loadConfig()

	// Initialize logger
	l, err := logger.NewZapAdapter()
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer l.Sync()

	if config.ENV == "dev" {
		l.SetLevel(logger.LevelDebug)
	}

	// Create Discord client
	client := discord.NewClient(config.BOT_TOKEN, config.API_BASE, l)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle OS signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		l.Info("Received shutdown signal")
		cancel()
	}()

	// Start the bot
	l.Info("Starting Discord bot")
	if err := client.Start(ctx); err != nil {
		l.Error("Bot stopped with error", logger.Field{Key: "error", Value: err})
		return
	}

	l.Info("Bot shutdown complete")
}
