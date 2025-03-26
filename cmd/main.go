package main

import (
	"fmt"

	"github.com/joho/godotenv"
	bot "github.com/taranpreetnatt/Discord-AI-Bot/internal/bot"
)

func main() {
	envErr := godotenv.Load()
	if envErr != nil {
		fmt.Errorf("Error loading env using godotenv", envErr)
	}

	err := bot.StartBot()
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}
