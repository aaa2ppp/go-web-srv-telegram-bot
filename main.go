package main

import (
	"context"
	"log"
	"os"
)

func main() {
	const op = "main"

	var ok bool
	BotToken, ok = os.LookupEnv("TOKEN")
	if !ok {
		log.Fatalf("%s: env TOKEN required", op)
	}

	err := startTaskBot(context.Background(), "")
	if err != nil {
		log.Fatalln(err)
	}
}
