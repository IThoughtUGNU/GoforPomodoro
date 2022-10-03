package main

import (
	"GoforPomodoro/internal/botmodule"
	"GoforPomodoro/internal/data"
	"fmt"
	"log"
)

func main() {
	settings, err := data.LoadAppSettings()
	if err != nil {
		log.Fatal(err)
	}

	appState, err := data.LoadAppState()
	if err != nil {
		panic(err)
	}

	fmt.Println("Hello from Go for Pomodoro!")

	botmodule.CommandMenuLoop(settings, appState)
}
