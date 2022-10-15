// This file is part of GoforPomodoro.
//
// GoforPomodoro is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// GoforPomodoro is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with GoforPomodoro.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"GoforPomodoro/internal/data"
	"GoforPomodoro/internal/data/persistence"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"os"
	"runtime"
)

type ErrorType int

const (
	NoAppSettings   ErrorType = 1 << iota // 1
	NoDB                                  // 2
	NoAPIConnection                       // 4
)

func main() {
	fmt.Printf("Go for Pomodoro FOSS -- sanity check.\n")
	fmt.Println("--------------------------------------------------------------")
	var noAppSettings ErrorType
	var noDb ErrorType
	var noApiConn ErrorType

	okSymbol := "✅"
	errSymbol := "❌"

	fmt.Printf("(Go runtime version: %s)\n\n", runtime.Version())

	settings, err := data.LoadAppSettings()

	var s string
	if err != nil || (len(settings.ApiToken) == 0) {
		s = errSymbol
		noAppSettings = NoAppSettings
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
		noDb = NoDB
	} else {
		s = okSymbol
	}
	fmt.Printf("- [%v] Database connected\n", s)
	if dbErr != nil {
		fmt.Printf("       A database instance is not mandatory. The bot can also run without any\n" +
			"       persistence. But keep in mind that doing so will make lose all data and\n" +
			"       irremediably lose all the sessions running after the application is\n" +
			"       shutted down.\n")
	}
	fmt.Println()

	bot, err := tgbotapi.NewBotAPI(settings.ApiToken)
	if err != nil {
		s = errSymbol
		noApiConn = NoAPIConnection
	} else {
		s = okSymbol
	}
	fmt.Printf("- [%v] Telegram API connection\n", s)
	if err == nil {
		fmt.Printf(
			"       Authorized on account %s\n", bot.Self.UserName)
	} else {
		fmt.Println(
			"       No account authorized. The application will not work without a valid API\n" +
				"       key and connection.")
	}
	fmt.Println()
	fmt.Println("--------------------------------------------------------------")
	fmt.Println()

	os.Exit(int(noAppSettings | noDb | noApiConn))
}
