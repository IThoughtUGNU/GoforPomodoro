package main

import (
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
	userId ChatID,
	currentSession *Session,
	restBeginHandler func(id ChatID, session *Session),
	restFinishedHandler func(id ChatID, session *Session),
	endSessionHandler func(id ChatID, session *Session, endKind PomodoroEndKind),
	pauseSessionHandler func(id ChatID, session *Session),
) error {
	if currentSession.isZero() {
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
	currentSession.Data.isPaused = false
	currentSession.Data.isCancel = false

	currentSession.Data.SprintDuration -= 1

	currentSession.AssignTimestamps()

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
	userId ChatID,
	currentSession *Session,
	restBeginHandler func(id ChatID, session *Session),
	restFinishedHandler func(id ChatID, session *Session),
	endSessionHandler func(id ChatID, session *Session, endKind PomodoroEndKind),
	pauseSessionHandler func(id ChatID, session *Session),
) {
	sData := &currentSession.Data
mainLoop:
	for {
		select {
		case action, ok := <-currentSession.ReadingActionChannel():
			if ok {
				if action.Paused || action.Canceled || action.Finished {
					if action.Paused {
						sData.isPaused = true
						pauseSessionHandler(userId, currentSession)
					} else if action.Canceled {
						sData.isCancel = true
						endSessionHandler(userId, currentSession, PomodoroCanceled)
					} else if action.Finished {
						sData.isFinished = true
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

			if sData.isRest {
				sData.RestDuration -= 1

				if sData.RestDuration <= 0 {
					restFinishedHandler(userId, currentSession)
					sData.RestDuration = currentSession.RestDurationSet
					sData.isRest = false
				}
			} else {
				sData.PomodoroDuration -= 1

				if sData.PomodoroDuration <= 0 && sData.SprintDuration > 0 {
					sData.SprintDuration -= 1

					if sData.SprintDuration < 0 {
						sData.isFinished = true
						endSessionHandler(userId, currentSession, PomodoroFinished)
						return
					}

					restBeginHandler(userId, currentSession)
					sData.PomodoroDuration = currentSession.PomodoroDurationSet
					sData.isRest = true
				}
			}
		}
	}
	defer close(currentSession.ActionsChannel)
}

func PauseSession(currentSession *Session) error {
	if currentSession.Data.isPaused {
		return errors.New("sessionDefault already paused")
	}

	currentSession.WritingActionChannel() <- DispatchAction{Paused: true}
	return nil
}

func CancelSession(currentSession *Session) error {
	if currentSession.IsCanceled() {
		return errors.New("sessionDefault already canceled")
	}

	currentSession.WritingActionChannel() <- DispatchAction{Canceled: true}
	return nil
}

func ResumeSession(
	userId ChatID,
	currentSession *Session,
	restBeginHandler func(id ChatID, session *Session),
	restFinishedHandler func(id ChatID, session *Session),
	endSessionHandler func(id ChatID, session *Session, endKind PomodoroEndKind),
	pauseSessionHandler func(id ChatID, session *Session),
) error {
	if currentSession.isZero() {
		return errors.New("the session is effectively nil")
	}
	if !currentSession.IsStopped() {
		return errors.New("session already running")
	}
	if currentSession.IsCanceled() {
		return errors.New("session was canceled")
	}

	currentSession.Data.isPaused = false

	currentSession.AssignTimestamps()

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
