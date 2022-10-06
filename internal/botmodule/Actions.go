package botmodule

import (
	"GoforPomodoro/internal/data"
	"GoforPomodoro/internal/domain"
	"GoforPomodoro/internal/sessionmanager"
)

func ActionRestoreSprint(
	chatId domain.ChatID,
	appState *domain.AppState,
	session *domain.Session,
	communicator *Communicator,
) {
	go sessionmanager.SpawnSessionTimer(
		appState,
		chatId,
		session,
		communicator.RestBeginHandler,
		communicator.RestFinishedHandler,
		communicator.SessionFinishedHandler,
		communicator.SessionPausedHandler,
	)
}

func ActionResumeSprint(
	senderId domain.ChatID,
	chatId domain.ChatID,
	appState *domain.AppState,
	communicator *Communicator,
) {
	session := data.GetUserSessionRunning(appState, chatId, senderId)
	communicator.SessionResumed(
		sessionmanager.ResumeSession(
			appState,
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

func ActionStartSprint(
	senderId domain.ChatID,
	chatId domain.ChatID,
	appState *domain.AppState,
	communicator *Communicator,
) {

	session := data.GetUserSessionRunning(appState, chatId, senderId)

	if !session.IsStopped() {
		communicator.SessionAlreadyRunning()
		return
	}
	session = data.GetNewUserSessionRunning(appState, chatId, senderId)

	communicator.SessionStarted(
		session,
		sessionmanager.StartSession(
			appState,
			chatId,
			session,
			communicator.RestBeginHandler,
			communicator.RestFinishedHandler,
			communicator.SessionFinishedHandler,
			communicator.SessionPausedHandler,
		),
	)
}
