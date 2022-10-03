package botmodule

import (
	"GoforPomodoro/internal/data"
	"GoforPomodoro/internal/domain"
	"GoforPomodoro/internal/sessionmanager"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func ActionResumeSprint(update tgbotapi.Update, appState *domain.AppState, communicator *Communicator) {
	senderId := domain.ChatID(update.Message.From.ID)
	chatId := domain.ChatID(update.Message.Chat.ID)

	session := data.GetUserSessionRunning(appState, chatId, senderId)
	communicator.SessionResumed(
		sessionmanager.ResumeSession(
			chatId,
			session,
			communicator.RestBeginHandler,
			communicator.RestFinishedHandler,
			communicator.SessionFinishedHandler,
			communicator.SessionPausedHandler,
		),
		session,
	)
}

func ActionStartSprint(update tgbotapi.Update, appState *domain.AppState, communicator *Communicator) {
	senderId := domain.ChatID(update.Message.From.ID)
	chatId := domain.ChatID(update.Message.Chat.ID)

	session := data.GetUserSessionRunning(appState, chatId, senderId)

	if !session.IsStopped() {
		communicator.SessionAlreadyRunning()
		return
	}
	session = data.GetNewUserSessionRunning(appState, chatId, senderId)

	communicator.SessionStarted(
		session,
		sessionmanager.StartSession(
			chatId,
			session,
			communicator.RestBeginHandler,
			communicator.RestFinishedHandler,
			communicator.SessionFinishedHandler,
			communicator.SessionPausedHandler,
		),
	)
}
