package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	bot "github.com/taranpreetnatt/Discord-AI-Bot/internal/bot"
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

	err := bot.StartBot(config.BOT_TOKEN, config.API_BASE)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}
