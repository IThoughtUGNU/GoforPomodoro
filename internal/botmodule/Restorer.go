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

package botmodule

import (
	"GoforPomodoro/internal/data"
	"GoforPomodoro/internal/domain"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

func RestoreSessions(
	appState *domain.AppState,
	bot *tgbotapi.BotAPI,
) {
	if appState.PersistenceManager != nil {
		pairs, err := appState.PersistenceManager.GetActiveChatSettings()

		log.Printf("[Restorer::RestoreSessions] #sessions to restore: %v\n", len(pairs))
		if err != nil {
			log.Printf("[Restorer::RestoreSessions] error: %v\n", err.Error())
		} else {
			data.PreloadUsersSettings(appState, pairs)

			for _, pair := range pairs {
				chatId := pair.First
				settings := pair.Second

				log.Printf("[Restorer::RestoreSessions] Restoring session for chat id: %v", chatId)

				runningSession := settings.SessionRunning

				communicator := GetCommunicator(appState, chatId, bot)
				ActionRestoreSprint(chatId, appState, runningSession, communicator)
			}
		}
	}
}
