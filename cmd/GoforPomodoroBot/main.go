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
	dbErr := sqliteManager.OpenDatabase("./data/go4pom_data.db")
	if dbErr != nil {
		sqliteManager = nil // DB-less mode.
		log.Println("[main] Running bot with no database (there will be no persistence).")
		// panic(dbErr)
	}

	appState, err := data.LoadAppState(sqliteManager)
	if err != nil {
		panic(err)
	}

	fmt.Println("Hello from Go for Pomodoro!")

	botmodule.CommandMenuLoop(settings, appState)
}
