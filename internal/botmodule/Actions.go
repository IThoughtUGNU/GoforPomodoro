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

func ActionCancelSprint(
	senderId domain.ChatID,
	chatId domain.ChatID,
	appState *domain.AppState,
	communicator *Communicator,
) {
	session := data.GetUserSessionRunning(appState, chatId, senderId)

	var err error
	if !session.IsPaused() {
		err = sessionmanager.CancelSession(session)
	} else {
		session.Cancel()
		communicator.SessionFinishedHandler(chatId, session, sessionmanager.PomodoroCanceled)
	}

	communicator.SessionCanceled(err, *session)
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

	// log.Printf("[NO-DB TEST] ActionStartSprint!!\n")
	session := data.GetUserSessionRunning(appState, chatId, senderId)

	// log.Printf("[NO-DB TEST] data.GetUserSessionRunning succeded\n")
	if !session.IsStopped() {
		communicator.SessionAlreadyRunning()
		// log.Printf("[NO-DB TEST] session already running: stopping\n")
		return
	}
	session = data.GetNewUserSessionRunning(appState, chatId, senderId)

	// log.Printf("[NO-DB TEST] new session running: %v\n", session)
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
