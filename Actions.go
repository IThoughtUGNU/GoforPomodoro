package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func ActionResumeSprint(update tgbotapi.Update, appState *AppState, communicator *Communicator) {
	senderId := ChatID(update.Message.From.ID)
	chatId := ChatID(update.Message.Chat.ID)

	session := GetUserSessionRunning(appState, chatId, senderId)
	communicator.SessionResumed(
		ResumeSession(
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

func ActionStartSprint(update tgbotapi.Update, appState *AppState, communicator *Communicator) {
	senderId := ChatID(update.Message.From.ID)
	chatId := ChatID(update.Message.Chat.ID)

	session := GetUserSessionRunning(appState, chatId, senderId)

	if !session.isStopped() {
		communicator.SessionAlreadyRunning()
		return
	}
	session = GetNewUserSessionRunning(appState, chatId, senderId)

	communicator.SessionStarted(
		session,
		StartSession(
			chatId,
			session,
			communicator.RestBeginHandler,
			communicator.RestFinishedHandler,
			communicator.SessionFinishedHandler,
			communicator.SessionPausedHandler,
		),
	)
}
