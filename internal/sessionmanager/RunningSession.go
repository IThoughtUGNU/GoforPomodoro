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

package sessionmanager

import (
	"GoforPomodoro/internal/data"
	"GoforPomodoro/internal/domain"
	"errors"
	"log"
	"time"
)

type PomodoroEndKind int

const (
	PomodoroFinished PomodoroEndKind = iota
	PomodoroCanceled
)

func StartSession(
	appState *domain.AppState,
	userId domain.ChatID,
	currentSession *domain.Session,
	restBeginHandler func(id domain.ChatID, session *domain.Session),
	restFinishedHandler func(id domain.ChatID, session *domain.Session),
	endSessionHandler func(id domain.ChatID, session *domain.Session, endKind PomodoroEndKind),
	pauseSessionHandler func(id domain.ChatID, session *domain.Session),
) error {
	if currentSession.IsZero() {
		return errors.New("the session is effectively nil")
	}

	currentSession.Start()

	go SpawnSessionTimer(
		appState,
		userId,
		currentSession,
		restBeginHandler,
		restFinishedHandler,
		endSessionHandler,
		pauseSessionHandler,
	)
	return nil
}

func SpawnSessionTimer(
	appState *domain.AppState,
	chatId domain.ChatID,
	currentSession *domain.Session,
	restBeginHandler func(id domain.ChatID, session *domain.Session),
	restFinishedHandler func(id domain.ChatID, session *domain.Session),
	endSessionHandler func(id domain.ChatID, session *domain.Session, endKind PomodoroEndKind),
	pauseSessionHandler func(id domain.ChatID, session *domain.Session),
) {
	// We update session running because it started (or resumed)
	data.UpdateUserSessionRunning(appState, chatId)
mainLoop:
	for {
		select {
		case action, ok := <-currentSession.ReadingActionChannel():
			if ok {
				// The event was internal (rest started/finished)
				if action.RestStarted || action.RestFinished {
					if action.RestStarted {
						currentSession.RestStarted()
						restBeginHandler(chatId, currentSession)
					}
					if action.RestFinished {
						currentSession.RestFinished()
						restFinishedHandler(chatId, currentSession)
					}
					// We update session running because it changed state
					// (rest started or finished)
					data.UpdateUserSessionRunning(appState, chatId)
					continue mainLoop
				}

				// The event was either external (paused/canceled) or internal (finished)
				if action.Paused || action.Canceled || action.Finished {
					if action.Paused {
						currentSession.Pause()
						pauseSessionHandler(chatId, currentSession)
					} else if action.Canceled {
						currentSession.Cancel()
						endSessionHandler(chatId, currentSession, PomodoroCanceled)
					} else if action.Finished {
						currentSession.SetFinished()
						endSessionHandler(chatId, currentSession, PomodoroFinished)
					}
					// We update session running because it changed state
					// (paused, canceled or finished)
					data.UpdateUserSessionRunning(appState, chatId)
					break mainLoop
				}
			} else {
				currentSession.ActionsChannel = nil
				log.Println("Session channel is closed. Aborting main loop...")
				break mainLoop
			}
		default:
			time.Sleep(1 * time.Second)

			isRest := currentSession.IsRest()

			if !isRest && currentSession.HasSprintEndTimePassed() {
				currentSession.DecreaseSprintDuration()

				if currentSession.GetSprintDuration() < 0 {
					currentSession.WritingActionChannel() <- domain.DispatchAction{Finished: true}
					continue mainLoop
				}

				// if SprintDuration still >= 0, we have rest now
				currentSession.WritingActionChannel() <- domain.DispatchAction{RestStarted: true}
				continue mainLoop
			} else if isRest && currentSession.HasRestEndTimePassed() {

				currentSession.WritingActionChannel() <- domain.DispatchAction{RestFinished: true}
				continue mainLoop
			}
		}
	}
	defer currentSession.ClearChannel()
}

func PauseSession(currentSession *domain.Session) error {
	if currentSession.IsPaused() {
		return errors.New("sessionDefault already paused")
	}

	currentSession.WritingActionChannel() <- domain.DispatchAction{Paused: true}
	return nil
}

func CancelSession(currentSession *domain.Session) error {
	if currentSession.IsCanceled() {
		return errors.New("sessionDefault already canceled")
	}

	currentSession.WritingActionChannel() <- domain.DispatchAction{Canceled: true}
	return nil
}

func ResumeSession(
	appState *domain.AppState,
	userId domain.ChatID,
	currentSession *domain.Session,
	restBeginHandler func(id domain.ChatID, session *domain.Session),
	restFinishedHandler func(id domain.ChatID, session *domain.Session),
	endSessionHandler func(id domain.ChatID, session *domain.Session, endKind PomodoroEndKind),
	pauseSessionHandler func(id domain.ChatID, session *domain.Session),
) error {
	if currentSession.IsZero() {
		return errors.New("the session is effectively nil")
	}
	if !currentSession.IsStopped() {
		return errors.New("session already running")
	}
	if currentSession.IsCanceled() {
		return errors.New("session was canceled")
	}

	currentSession.Resume()

	go SpawnSessionTimer(
		appState,
		userId,
		currentSession,
		restBeginHandler,
		restFinishedHandler,
		endSessionHandler,
		pauseSessionHandler,
	)
	return nil
}
