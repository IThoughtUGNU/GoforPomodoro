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

type SprintDuration int
type PomodoroDuration int64
type RestDuration int64

func (d SprintDuration) ToInt() int {
	return int(d)
}

func (d PomodoroDuration) Seconds() int {
	return int(d)
}

func (d RestDuration) Seconds() int {
	return int(d)
}

type SessionData struct {
	// SprintDuration captures the number of sprints that the session has yet.
	// It can be updated during the run of a session.
	//
	// Use SprintDurationSet if you want to refer to the total number of
	// sprints.
	SprintDuration

	// PomodoroDuration is used with reference to how much time is left for the
	// current pomodoro in run.
	//
	// Use PomodoroDurationSet if you want to refer to the defined pomodoro
	// duration for the session.
	PomodoroDuration

	// RestDuration is used with reference to how much time is left for the
	// current pomodoro rest in run.
	//
	// Use RestDurationSet if you want to refer to the defined pomodoro rest
	// duration for the session.
	RestDuration

	IsRest     bool
	IsPaused   bool
	IsCancel   bool
	IsFinished bool
}

// SessionInitData represents a struct you use to initialize a session.
type SessionInitData struct {
	SprintDurationSet   SprintDuration
	PomodoroDurationSet PomodoroDuration
	RestDurationSet     RestDuration

	SprintDuration
	PomodoroDuration
	RestDuration

	IsRest     bool
	IsPaused   bool
	IsCancel   bool
	IsFinished bool
}

func (sid SessionInitData) ToSession() (s *Session) {
	s = new(Session)

	s.sprintDurationSet = sid.SprintDurationSet
	s.pomodoroDurationSet = sid.PomodoroDurationSet
	s.restDurationSet = sid.RestDurationSet

	s.data.SprintDuration = sid.SprintDuration
	s.data.PomodoroDuration = sid.PomodoroDuration
	s.data.RestDuration = sid.RestDuration

	return
}

func (s *Session) ToInitData() (sid SessionInitData) {
	sid.SprintDurationSet = s.sprintDurationSet
	sid.PomodoroDurationSet = s.pomodoroDurationSet
	sid.RestDurationSet = s.restDurationSet

	sid.SprintDuration = s.data.SprintDuration
	sid.PomodoroDuration = s.data.PomodoroDuration
	sid.RestDuration = s.data.RestDuration

	return
}

type Session struct {
	ActionsChannel chan DispatchAction

	endNextSprintTimestamp *time.Time
	endNextRestTimestamp   *time.Time

	// sprintDurationSet represents how many sprints the session is the session
	// intended to have.
	//
	// For example, sprintDurationSet == 4 means that there will be 4 different
	// sprints in the session, separated by (4-1) rests.
	//
	// Unlike SprintDuration, this variable is intended to be kept constant
	// during all the session. SprintDuration is used with reference to how
	// many sprints are left.
	sprintDurationSet SprintDuration

	// pomodoroDurationSet represents the time of duration of a pomodoro
	// expressed in SECONDS.
	//
	// Unlike PomodoroDuration, this variable is intended to be kept constant
	// during all the session. PomodoroDuration is used with reference to how
	// much time is left for the current pomodoro in run.
	pomodoroDurationSet PomodoroDuration

	// restDurationSet represents the time of rest duration of a pomodoro
	// expressed in SECONDS.
	//
	// Unlike RestDuration, this variable is intended to be kept constant
	// during all the session. RestDuration is used with reference to how
	// much time is left for the current pomodoro rest in run.
	restDurationSet RestDuration

	data SessionData
}

func (s *Session) GetRestDuration() RestDuration {
	if s.IsFinished() {
		return 0
	}

	if s.endNextRestTimestamp == nil || s.IsPaused() {
		log.Println("Fallback to s.RestDuration")
		return s.data.RestDuration
	}

	return RestDuration(s.endNextRestTimestamp.Sub(time.Now()).Seconds()) // s.PomodoroDuration

	// return s.RestDuration
}

func (s *Session) GetRestDurationSet() RestDuration {
	return s.restDurationSet
}

func (s *Session) GetPomodoroDuration() PomodoroDuration {
	if s.IsFinished() {
		return 0
	}

	if s.endNextSprintTimestamp == nil || s.IsPaused() {
		log.Println("Fallback to s.PomodoroDuration")
		return s.data.PomodoroDuration
	}

	return PomodoroDuration(s.endNextSprintTimestamp.Sub(time.Now()).Seconds())
}

func (s *Session) GetPomodoroDurationSet() PomodoroDuration {
	return s.pomodoroDurationSet
}

func (s *Session) GetSprintDuration() SprintDuration {
	return s.data.SprintDuration
}

func (s *Session) GetSprintDurationSet() SprintDuration {
	return s.sprintDurationSet
}

func (s *Session) IsRest() bool {
	return s.data.IsRest
}

func DefaultSession() Session {
	return Session{
		sprintDurationSet:   4,
		pomodoroDurationSet: 25 * 60,
		restDurationSet:     25 * 60,

		data: SessionData{
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
	s.endNextSprintTimestamp = nil
	s.endNextRestTimestamp = nil

	var pomodoroDurationTime time.Duration = 0
	var restDurationTime time.Duration = 0

	if s.IsRest() {
		restDurationTime = time.Second * time.Duration(s.data.RestDuration)

		s.endNextRestTimestamp = utils.TimePtr(time.Now().Local().Add(restDurationTime))
	} else {
		pomodoroDurationTime = time.Second * time.Duration(s.data.PomodoroDuration)
		restDurationTime = time.Second * time.Duration(s.restDurationSet)

		s.endNextSprintTimestamp = utils.TimePtr(time.Now().Local().Add(pomodoroDurationTime))

		s.endNextRestTimestamp = utils.TimePtr(time.Now().Local().Add(pomodoroDurationTime + restDurationTime))
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

		middleStr = fmt.Sprintf("\nTime for current rest remaining: %s", utils.NiceTimeFormatting(s.GetRestDuration().Seconds()))
	} else {
		middleStr = fmt.Sprintf("\nTime for current pomodoro remaining: %s", utils.NiceTimeFormatting(s.GetPomodoroDuration().Seconds()))
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
		return "Rest for other " + utils.NiceTimeFormatting(s.GetRestDuration().Seconds())
	} else {
		return "Task time: " + utils.NiceTimeFormatting(s.GetPomodoroDuration().Seconds()) + " left."
	}
}

func (s *Session) IsStopped() bool {
	if s.GetPomodoroDuration() <= 0 ||
		s.GetSprintDuration() < 0 ||
		s.data.IsPaused ||
		s.data.IsCancel ||
		s.data.IsFinished {
		return true
	}

	return false
}

func (s *Session) IsCanceled() bool {
	return s.data.IsCancel
}

func (s *Session) IsPaused() bool {
	return s.data.IsPaused
}

func (s *Session) IsFinished() bool {
	return s.data.IsFinished
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

func (s *Session) Pause() {
	// Cache pomodoro and rest duration. We will use them again to assign new timestamps.
	s.data.PomodoroDuration = s.GetPomodoroDuration()
	s.data.RestDuration = s.GetRestDuration()

	s.data.IsPaused = true
}

func (s *Session) Cancel() {
	s.data.IsCancel = true
}

func (s *Session) SetFinished() {
	s.data.IsFinished = true
}

func (s *Session) Resume() {
	s.data.IsPaused = false

	s.AssignTimestamps()
}

func (s *Session) Start() {
	s.data.IsPaused = false
	s.data.IsCancel = false

	s.data.SprintDuration -= 1

	s.AssignTimestamps()
}

func (s *Session) RestStarted() {
	s.data.IsRest = true
	s.data.RestDuration = s.restDurationSet
	s.AssignTimestamps()
}

func (s *Session) RestFinished() {
	s.data.IsRest = false
	s.data.PomodoroDuration = s.pomodoroDurationSet
	s.AssignTimestamps()
}

func (s *Session) DecreaseSprintDuration() {
	s.data.SprintDuration -= 1
}

func (s *Session) ClearChannel() {
	close(s.ActionsChannel)
	s.ActionsChannel = nil
}

func (s *Session) HasSprintEndTimePassed() bool {
	if s.endNextSprintTimestamp == nil {
		log.Println("[PROBLEM] s.endNextSprintTimestamp IS nil.")
		return false
	}

	return time.Now().Local().After(*s.endNextSprintTimestamp)
}

func (s *Session) HasRestEndTimePassed() bool {
	if s.endNextRestTimestamp == nil {
		log.Println("[PROBLEM] s.endNextRestTimestamp IS nil.")
		return false
	}

	return time.Now().Local().After(*s.endNextRestTimestamp)
}
