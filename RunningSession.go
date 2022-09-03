package main

import (
	"errors"
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
	currentSession.isPaused = false
	currentSession.isCancel = false

	currentSession.SprintDuration -= 1
	/*
		nextRunSession := func() {
			gocron.NewScheduler().
		}*/
	SpawnSessionTimer(
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
	go func() {
		for {
			if currentSession.isPaused {
				pauseSessionHandler(userId, currentSession)
				return
			}
			if currentSession.isCanceled() {
				endSessionHandler(userId, currentSession, PomodoroCanceled)
				return
			}
			if currentSession.isStopped() {
				endSessionHandler(userId, currentSession, PomodoroFinished)
				return
			}
			time.Sleep(1 * time.Second)

			if currentSession.isRest {
				currentSession.RestDuration -= 1

				if currentSession.RestDuration <= 0 {
					restFinishedHandler(userId, currentSession)
					currentSession.RestDuration = currentSession.RestDurationSet
					currentSession.isRest = false
				}
			} else {
				currentSession.PomodoroDuration -= 1

				if currentSession.PomodoroDuration <= 0 && currentSession.SprintDuration > 0 {
					currentSession.SprintDuration -= 1

					if currentSession.SprintDuration < 0 {
						endSessionHandler(userId, currentSession, PomodoroFinished)
						return
					}

					restBeginHandler(userId, currentSession)
					currentSession.PomodoroDuration = currentSession.PomodoroDurationSet
					currentSession.isRest = true
				}
			}
		}
	}()
}

func PauseSession(currentSession *Session) error {
	if currentSession.isPaused {
		return errors.New("sessionDefault already paused")
	}
	currentSession.isPaused = true
	return nil
}

func CancelSession(currentSession *Session) error {
	if currentSession.isCanceled() {
		return errors.New("sessionDefault already canceled")
	}
	currentSession.isCancel = true
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
	if !currentSession.isStopped() {
		return errors.New("session already running")
	}
	if currentSession.isCancel {
		return errors.New("session was canceled")
	}
	currentSession.isPaused = false
	SpawnSessionTimer(
		userId,
		currentSession,
		restBeginHandler,
		restFinishedHandler,
		endSessionHandler,
		pauseSessionHandler,
	)
	return nil
}
