package main

import (
	"log"

	"stack/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatalf("Ошибка работы приложения: %v", err)
	}
}
