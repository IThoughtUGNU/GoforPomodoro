package main

import "fmt"

type AppSettings struct {
	ApiToken string
	BotName  string
}

type ChatID int64

type Session struct {
	SprintDurationSet   int
	PomodoroDurationSet int
	RestDurationSet     int

	SprintDuration   int
	PomodoroDuration int
	RestDuration     int

	isRest   bool
	isPaused bool
	isCancel bool
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

	return fmt.Sprintf("Session of %d🍅 x %dm + %dm", s.SprintDurationSet, s.PomodoroDurationSet/60, s.RestDurationSet/60) +
		fmt.Sprintf("\nPomodoros remaining: %d", s.SprintDuration) +
		fmt.Sprintf("\nTime for current pomodoro remaining: %s", NiceTimeFormatting(s.PomodoroDuration)) +
		fmt.Sprintf("\nRest time: %s", NiceTimeFormatting(s.RestDuration)) +
		fmt.Sprintf("\n\nCurrent session state: %s", s.State())
}

func (s *Session) LeftTimeMessage() string {
	if s.isZero() {
		return "No running pomodoros!"
	}
	if s.isRest {
		return "Rest for other " + NiceTimeFormatting(s.RestDuration)
	} else {
		return "Task time: " + NiceTimeFormatting(s.PomodoroDuration) + " left."
	}
}

func (s *Session) isStopped() bool {
	if s.PomodoroDuration <= 0 || s.SprintDuration < 0 || s.isPaused || s.isCancel {
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
