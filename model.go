package main

import (
	"fmt"
	"time"
)

type AppSettings struct {
	ApiToken string
	BotName  string
}

type ChatID int64

type Session struct {
	EndNextSprintTimestamp time.Time
	EndNextRestTimestamp   time.Time

	// SprintDurationSet represents how many sprints the session is the session
	// intended to have.
	//
	// For example, SprintDurationSet == 4 means that there will be 4 different
	// sprints in the session, separated by (4-1) rests.
	//
	// Unlike SprintDuration, this variable is intended to be kept constant
	// during all the session. SprintDuration is used with reference to how
	// many sprints are left.
	SprintDurationSet int

	// PomodoroDurationSet represents the time of duration of a pomodoro
	// expressed in SECONDS.
	//
	// Unlike PomodoroDuration, this variable is intended to be kept constant
	// during all the session. PomodoroDuration is used with reference to how
	// much time is left for the current pomodoro in run.
	PomodoroDurationSet int

	// RestDurationSet represents the time of rest duration of a pomodoro
	// expressed in SECONDS.
	//
	// Unlike RestDuration, this variable is intended to be kept constant
	// during all the session. RestDuration is used with reference to how
	// much time is left for the current pomodoro rest in run.
	RestDurationSet int

	// SprintDuration captures the number of sprints that the session has yet.
	// It can be updated during the run of a session.
	//
	// Use SprintDurationSet if you want to refer to the total number of
	// sprints.
	SprintDuration int

	// PomodoroDuration is used with reference to how much time is left for the
	// current pomodoro in run.
	//
	// Use PomodoroDurationSet if you want to refer to the defined pomodoro
	// duration for the session.
	PomodoroDuration int

	// RestDuration is used with reference to how much time is left for the
	// current pomodoro rest in run.
	//
	// Use RestDurationSet if you want to refer to the defined pomodoro rest
	// duration for the session.
	RestDuration int

	isRest     bool
	isPaused   bool
	isCancel   bool
	isFinished bool
}

func DefaultSession() Session {
	return Session{
		SprintDurationSet:   4,
		PomodoroDurationSet: 25 * 60,
		RestDurationSet:     25 * 60,

		SprintDuration:   4,
		PomodoroDuration: 25 * 60,
		RestDuration:     25 * 60,
	}
}

func a() {

	// timestamp := time.Unix(time.Now().Unix(), 0)
}

func (s *Session) isZero() bool {
	return s == nil || s.PomodoroDurationSet == 0
}

func (s *Session) String() string {
	if s == nil {
		return "nil"
	}

	if s.PomodoroDurationSet == 0 {
		return "No session"
	}

	sprintDuration := s.SprintDuration
	if s.isRest {
		sprintDuration += 1
	}

	return fmt.Sprintf("Session of %d🍅 x %dm + %dm", s.SprintDurationSet, s.PomodoroDurationSet/60, s.RestDurationSet/60) +
		fmt.Sprintf("\nPomodoros remaining: %d", sprintDuration) +
		fmt.Sprintf("\nTime for current pomodoro remaining: %s", NiceTimeFormatting(s.PomodoroDuration)) +
		fmt.Sprintf("\nRest time: %s", NiceTimeFormatting(s.RestDuration)) +
		fmt.Sprintf("\n\nCurrent session state: %s", s.State())
}

func (s *Session) LeftTimeMessage() string {
	if s.isZero() || s.isCancel || s.isFinished {
		return "No running pomodoros!"
	}
	if s.isRest {
		return "Rest for other " + NiceTimeFormatting(s.RestDuration)
	} else {
		return "Task time: " + NiceTimeFormatting(s.PomodoroDuration) + " left."
	}
}

func (s *Session) isStopped() bool {
	if s.PomodoroDuration <= 0 || s.SprintDuration < 0 || s.isPaused || s.isCancel || s.isFinished {
		return true
	}
	return false
}

func (s *Session) isCanceled() bool {
	return s.isCancel
}

func (s *Session) State() string {
	var stateStr string
	if s.isPaused {
		if s.PomodoroDuration == s.PomodoroDurationSet &&
			s.SprintDuration == s.SprintDurationSet &&
			s.RestDuration == s.RestDurationSet {

			stateStr = "Pending"
		} else {
			stateStr = "Paused"
		}
	} else if s.isCancel {
		stateStr = "Canceled"
	} else if s.isFinished {
		stateStr = "Finished"
	} else if s.isStopped() {
		stateStr = "Stopped"
	} else {
		stateStr = "Running"
	}
	return stateStr
}

type Settings struct {
	sessionDefault Session
	sessionRunning *Session
	autorun        bool
	isGroup        bool
	subscribers    []ChatID
}

type AppState struct {
	usersSettings map[ChatID]*Settings
}
