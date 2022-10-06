package main

import (
	"GoforPomodoro/internal/botmodule"
	"GoforPomodoro/internal/data"
	"GoforPomodoro/internal/data/persistence"
	"fmt"
	"log"
)

func main() {
	settings, err := data.LoadAppSettings()
	if err != nil {
		log.Fatal(err)
	}

	sqliteManager := &persistence.SqliteManager{}
	sqliteManager.OpenDatabase("./data/go4pom_data.db")

	appState, err := data.LoadAppState(sqliteManager)
	if err != nil {
		panic(err)
	}

	fmt.Println("Hello from Go for Pomodoro!")

	botmodule.CommandMenuLoop(settings, appState)
}
