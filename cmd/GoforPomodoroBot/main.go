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

	debugMode := settings.DebugMode

	appState, err := data.LoadAppState(sqliteManager, debugMode)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Hello from Go for Pomodoro!\n\n(debug mode set to: %v)\n\n", debugMode)

	botmodule.CommandMenuLoop(settings, appState)
}
