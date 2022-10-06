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
