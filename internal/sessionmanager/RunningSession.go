package sessionmanager

import (
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

	// How a sessionDefault is defined:
	// SprintDuration   int
	// PomodoroDuration int
	// RestDuration     int

	// We want
	// 1. Decrease Sprint duration by 1
	// 2. set a timer of PomodoroDuration minutes
	// 3. at its end, set a timer of PomodoroDuration seconds
	// 4. at its end, check if Sprint duration is >0. If so, go to 1, otherwise isPaused.

	currentSession.Start()

	go SpawnSessionTimer(
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
	userId domain.ChatID,
	currentSession *domain.Session,
	restBeginHandler func(id domain.ChatID, session *domain.Session),
	restFinishedHandler func(id domain.ChatID, session *domain.Session),
	endSessionHandler func(id domain.ChatID, session *domain.Session, endKind PomodoroEndKind),
	pauseSessionHandler func(id domain.ChatID, session *domain.Session),
) {
mainLoop:
	for {
		select {
		case action, ok := <-currentSession.ReadingActionChannel():
			if ok {
				// The event was internal (rest started/finished)
				if action.RestStarted || action.RestFinished {
					if action.RestStarted {
						currentSession.RestStarted()
						restBeginHandler(userId, currentSession)
					}
					if action.RestFinished {
						currentSession.RestFinished()
						restFinishedHandler(userId, currentSession)
					}
					continue mainLoop
				}

				// The event was either external (paused/canceled) or internal (finished)
				if action.Paused || action.Canceled || action.Finished {
					if action.Paused {
						currentSession.Pause()
						pauseSessionHandler(userId, currentSession)
					} else if action.Canceled {
						currentSession.Cancel()
						endSessionHandler(userId, currentSession, PomodoroCanceled)
					} else if action.Finished {
						currentSession.SetFinished()
						endSessionHandler(userId, currentSession, PomodoroFinished)
					}
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
		userId,
		currentSession,
		restBeginHandler,
		restFinishedHandler,
		endSessionHandler,
		pauseSessionHandler,
	)
	return nil
}
