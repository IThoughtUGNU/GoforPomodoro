package domain

import (
	"GoforPomodoro/internal/utils"
	"fmt"
	"log"
	"time"
)

type DispatchAction struct {
	Paused       bool
	Canceled     bool
	Finished     bool
	Resumed      bool
	RestStarted  bool
	RestFinished bool
}

type SessionData struct {
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

	IsRest     bool
	IsPaused   bool
	IsCancel   bool
	IsFinished bool
}

type Session struct {
	ActionsChannel chan DispatchAction

	EndNextSprintTimestamp *time.Time
	EndNextRestTimestamp   *time.Time

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

	Data SessionData
}

func (s *Session) GetRestDuration() int {
	if s.IsFinished() {
		return 0
	}

	if s.EndNextRestTimestamp == nil || s.IsPaused() {
		log.Println("Fallback to s.RestDuration")
		return s.Data.RestDuration
	}

	return int(s.EndNextRestTimestamp.Sub(time.Now()).Seconds()) // s.PomodoroDuration

	// return s.RestDuration
}

func (s *Session) GetRestDurationSet() int {
	return s.RestDurationSet
}

func (s *Session) GetPomodoroDuration() int {
	if s.IsFinished() {
		return 0
	}

	if s.EndNextSprintTimestamp == nil || s.IsPaused() {
		log.Println("Fallback to s.PomodoroDuration")
		return s.Data.PomodoroDuration
	}

	return int(s.EndNextSprintTimestamp.Sub(time.Now()).Seconds()) // s.PomodoroDuration
}

func (s *Session) GetPomodoroDurationSet() int {
	return s.PomodoroDurationSet
}

func (s *Session) GetSprintDuration() int {
	return s.Data.SprintDuration
}

func (s *Session) GetSprintDurationSet() int {
	return s.SprintDurationSet
}

func (s *Session) IsRest() bool {
	return s.Data.IsRest
}

func DefaultSession() Session {
	return Session{
		SprintDurationSet:   4,
		PomodoroDurationSet: 25 * 60,
		RestDurationSet:     25 * 60,

		Data: SessionData{
			SprintDuration:   4,
			PomodoroDuration: 25 * 60,
			RestDuration:     25 * 60,
		},
	}
}

func (s *Session) InitChannel() *Session {
	s.ActionsChannel = make(chan DispatchAction, 10)
	return s
}

func (s *Session) AssignTimestamps() {
	s.EndNextSprintTimestamp = nil
	s.EndNextRestTimestamp = nil

	var pomodoroDurationTime time.Duration = 0
	var restDurationTime time.Duration = 0

	if s.IsRest() {
		restDurationTime = time.Second * time.Duration(s.Data.RestDuration)

		s.EndNextRestTimestamp = utils.TimePtr(time.Now().Local().Add(restDurationTime))
	} else {
		pomodoroDurationTime = time.Second * time.Duration(s.Data.PomodoroDuration)
		restDurationTime = time.Second * time.Duration(s.RestDurationSet)

		s.EndNextSprintTimestamp = utils.TimePtr(time.Now().Local().Add(pomodoroDurationTime))

		s.EndNextRestTimestamp = utils.TimePtr(time.Now().Local().Add(pomodoroDurationTime + restDurationTime))
	}
}

func (s *Session) ReadingActionChannel() <-chan DispatchAction {
	return s.ActionsChannel
}

func (s *Session) WritingActionChannel() chan<- DispatchAction {
	return s.ActionsChannel
}

func (s *Session) IsZero() bool {
	return s == nil || s.GetPomodoroDurationSet() == 0
}

func (s *Session) String() string {
	if s == nil {
		return "nil"
	}

	if s.GetPomodoroDurationSet() == 0 {
		return "No session"
	}

	var middleStr string
	sprintDuration := s.GetSprintDuration()
	if s.IsRest() {
		sprintDuration += 1

		middleStr = fmt.Sprintf("\nTime for current rest remaining: %s", utils.NiceTimeFormatting(s.GetRestDuration()))
	} else {
		middleStr = fmt.Sprintf("\nTime for current pomodoro remaining: %s", utils.NiceTimeFormatting(s.GetPomodoroDuration()))
	}

	return fmt.Sprintf("Session of %d🍅 x %dm + %dm",
		s.GetSprintDurationSet(), s.GetPomodoroDurationSet()/60, s.GetRestDurationSet()/60) +
		fmt.Sprintf("\nPomodoros remaining: %d", sprintDuration) +
		middleStr +
		fmt.Sprintf("\n\nCurrent session state: %s", s.State())
}

func (s *Session) LeftTimeMessage() string {
	if s.IsPaused() && !s.IsFinished() {
		return "Pomodoro in pause. (use /resume)"
	}
	if s.IsZero() || s.IsCanceled() || s.IsStopped() {
		return "No running pomodoros!"
	}
	if s.IsRest() {
		return "Rest for other " + utils.NiceTimeFormatting(s.GetRestDuration())
	} else {
		return "Task time: " + utils.NiceTimeFormatting(s.GetPomodoroDuration()) + " left."
	}
}

func (s *Session) IsStopped() bool {
	if s.GetPomodoroDuration() <= 0 ||
		s.GetSprintDuration() < 0 ||
		s.Data.IsPaused ||
		s.Data.IsCancel ||
		s.Data.IsFinished {
		return true
	}

	return false
}

func (s *Session) IsCanceled() bool {
	return s.Data.IsCancel
}

func (s *Session) IsPaused() bool {
	return s.Data.IsPaused
}

func (s *Session) IsFinished() bool {
	return s.Data.IsFinished
}

func (s *Session) State() string {
	var stateStr string
	if s.IsPaused() {
		if s.GetPomodoroDuration() == s.GetPomodoroDurationSet() &&
			s.GetSprintDuration() == s.GetSprintDurationSet() &&
			s.GetRestDuration() == s.GetRestDurationSet() {

			stateStr = "Pending"
		} else {
			stateStr = "Paused"
		}
	} else if s.IsCanceled() {
		stateStr = "Canceled"
	} else if s.IsFinished() {
		stateStr = "Finished"
	} else if s.IsStopped() {
		stateStr = "Stopped"
	} else {
		stateStr = "Running"
	}
	return stateStr
}
