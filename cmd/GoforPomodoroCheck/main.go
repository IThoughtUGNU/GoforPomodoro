package main

import (
	"GoforPomodoro/internal/data"
	"GoforPomodoro/internal/data/persistence"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	okSymbol := "✅"
	errSymbol := "❌"

	fmt.Printf("Go for Pomodoro FOSS -- sanity check.\n\n")

	settings, err := data.LoadAppSettings()

	var s string
	if err != nil || (len(settings.ApiToken) == 0 || len(settings.BotName) == 0) {
		s = errSymbol
	} else {
		s = okSymbol
	}
	fmt.Printf("- [%v] appsettings.toml file\n", s)
	if err != nil {
		fmt.Println("      Please create such file and provide ApiToken and BotName accordingly to\n" +
			"      the README.")
	}
	fmt.Println()

	sqliteManager := &persistence.SqliteManager{}
	dbErr := sqliteManager.OpenDatabase("./data/go4pom_data.db")
	if dbErr != nil {
		sqliteManager = nil // DB-less mode.
		s = errSymbol
	} else {
		s = okSymbol
	}
	fmt.Printf("- [%v] Database connected\n", s)
	if dbErr != nil {
		fmt.Printf("      A database instance is not mandatory. The bot can also run without any\n" +
			"      persistence. But keep in mind that doing so will make lose all data and\n" +
			"      irremediably lose all the sessions running after the application is\n" +
			"      shutted down.\n")
	}
	fmt.Println()

	bot, err := tgbotapi.NewBotAPI(settings.ApiToken)
	if err != nil {
		s = errSymbol
	} else {
		s = okSymbol
	}
	fmt.Printf("- [%v] Telegram API connection\n", s)
	if err == nil {
		fmt.Printf(
			"     Authorized on account %s\n", bot.Self.UserName)
	} else {
		fmt.Println(
			"      No account authorized. The application will not work without a valid API\n" +
				"      key and connection.")
	}
	fmt.Println()
	fmt.Println()
}
