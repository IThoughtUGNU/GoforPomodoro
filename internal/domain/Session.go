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

const UnspecifiedSprintCardinality = -100
const DefaultRestTime = 5 * 60

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

type SessionDefaultData struct {
	SprintDurationSet   SprintDuration
	PomodoroDurationSet PomodoroDuration
	RestDurationSet     RestDuration
}

func SessionDefaultDataFromSession(s *Session) (sdd SessionDefaultData) {
	sdd.PomodoroDurationSet = s.GetPomodoroDurationSet()
	sdd.RestDurationSet = s.GetRestDurationSet()
	sdd.SprintDurationSet = s.GetSprintDurationSet()

	return
}

func (sdd SessionDefaultData) ToInitData() (sid SessionInitData) {
	sid.SprintDurationSet = sdd.SprintDurationSet
	sid.PomodoroDurationSet = sdd.PomodoroDurationSet
	sid.RestDurationSet = sdd.RestDurationSet

	sid.SprintDuration = sdd.SprintDurationSet
	sid.PomodoroDuration = sdd.PomodoroDurationSet
	sid.RestDuration = sdd.RestDurationSet

	sid.IsPaused = true
	sid.IsRest = false
	sid.IsFinished = false
	sid.IsCancel = false

	return
}

// SessionInitData represents a struct you use to initialize a session.
type SessionInitData struct {
	SprintDurationSet   SprintDuration
	PomodoroDurationSet PomodoroDuration
	RestDurationSet     RestDuration

	SprintDuration
	PomodoroDuration
	RestDuration

	EndNextSprintTimestamp time.Time
	EndNextRestTimestamp   time.Time

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

	s.data.IsCancel = sid.IsCancel
	s.data.IsPaused = sid.IsPaused
	s.data.IsRest = sid.IsRest
	s.data.IsFinished = sid.IsFinished

	if !sid.EndNextRestTimestamp.IsZero() {
		s.endNextRestTimestamp = &sid.EndNextRestTimestamp
	}
	if !sid.EndNextSprintTimestamp.IsZero() {
		s.endNextSprintTimestamp = &sid.EndNextSprintTimestamp
	}

	return
}

func (s *Session) ToInitData() (sid SessionInitData) {
	sid.SprintDurationSet = s.sprintDurationSet
	sid.PomodoroDurationSet = s.pomodoroDurationSet
	sid.RestDurationSet = s.restDurationSet

	sid.SprintDuration = s.data.SprintDuration
	sid.PomodoroDuration = s.data.PomodoroDuration
	sid.RestDuration = s.data.RestDuration

	sid.IsPaused = s.data.IsPaused
	sid.IsRest = s.data.IsRest
	sid.IsFinished = s.data.IsFinished
	sid.IsCancel = s.data.IsCancel

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

// GetRestDuration returns how much time (in SECONDS) the actual rest
// will go on before its end.
//
// (Decreases while the rest goes on)
func (s *Session) GetRestDuration() RestDuration {
	if s.IsFinished() {
		return 0
	}

	if s.endNextRestTimestamp == nil || s.IsPaused() {
		// log.Println("Fallback to s.RestDuration")
		return s.data.RestDuration
	}

	return RestDuration(s.endNextRestTimestamp.Sub(time.Now()).Seconds()) // s.PomodoroDuration

	// return s.RestDuration
}

// GetRestDurationSet returns the time of duration of a rest (in the current
// session expressed in SECONDS.
//
// Unlike GetRestDuration, this method returns a constant value
// during all the session. If you want to know how much time is left in the
// rest (if it's rest time), use GetRestDuration instead.
func (s *Session) GetRestDurationSet() RestDuration {
	return s.restDurationSet
}

// GetPomodoroDuration returns how much time (in SECONDS) the actual sprint
// will go on before its end.
//
// (Decreases while the sprint goes on)
func (s *Session) GetPomodoroDuration() PomodoroDuration {
	if s.IsFinished() {
		return 0
	}

	if s.endNextSprintTimestamp == nil || s.IsPaused() {
		// log.Println("Fallback to s.PomodoroDuration")
		return s.data.PomodoroDuration
	}

	return PomodoroDuration(s.endNextSprintTimestamp.Sub(time.Now()).Seconds())
}

// GetPomodoroDurationSet returns the time of duration of a pomodoro
// expressed in SECONDS.
//
// Unlike GetPomodoroDuration, this method returns a constant value
// during all the session. If you want to know how much time is left in this
// sprint, use GetPomodoroDuration instead.
func (s *Session) GetPomodoroDurationSet() PomodoroDuration {
	return s.pomodoroDurationSet
}

// GetSprintDuration returns how many sprints the session are left.
//
// (Decreases while the session goes on)
func (s *Session) GetSprintDuration() SprintDuration {
	return s.data.SprintDuration
}

func (s *Session) SprintDurationFinished() bool {
	return s.data.SprintDuration > UnspecifiedSprintCardinality &&
		s.data.SprintDuration < 0
}

func (s *Session) IsSprintDurationUnspecified() bool {
	return s.data.SprintDuration <= UnspecifiedSprintCardinality
}

// GetSprintDurationSet returns how many sprints the session should have
// (independently of how many remain)
func (s *Session) GetSprintDurationSet() SprintDuration {
	return s.sprintDurationSet
}

// IsRest returns true if it is rest time for the session.
func (s *Session) IsRest() bool {
	return s.data.IsRest
}

// DefaultSession Return a default session.
//
// ActionsChannel not initialized, therefore should call .InitChannel() if you
// plan to run a session from this object's value.
func DefaultSession() SessionDefaultData {
	return SessionDefaultData{
		SprintDurationSet:   4,
		PomodoroDurationSet: 25 * 60,
		RestDurationSet:     5 * 60,
	}
}

// InitChannel initialize ActionsChannel attribute; currently done with a
// buffer of 10 elements.
func (s *Session) InitChannel() *Session {
	s.ActionsChannel = make(chan DispatchAction, 10)
	return s
}

// assignTimestamps Assign timestamp fields for integrity of Session structure.
//
// After each sprint or rest end, their fields should be updated.
//
// This method is currently called internally in Session methods and therefore
// has been made private.
func (s *Session) assignTimestamps() {
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

// ReadingActionChannel Get the ActionsChannel in receive-only mode.
func (s *Session) ReadingActionChannel() <-chan DispatchAction {
	return s.ActionsChannel
}

// WritingActionChannel Get the ActionsChannel in send-only mode.
func (s *Session) WritingActionChannel() chan<- DispatchAction {
	return s.ActionsChannel
}

// IsZero Returns true if this session object was instantiated but not
// meaningfully initialized.
func (s *Session) IsZero() bool {
	return s == nil || s.GetPomodoroDurationSet() == 0
}

// String Print the state's session in human-readable format (aimed at the
// user).
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

	var sprintDurationSetStr string
	var pomodorosRemainingStr string
	if s.IsSprintDurationUnspecified() {
		pomodorosRemainingStr = "Unspecified"
		sprintDurationSetStr = "X"
	} else {
		pomodorosRemainingStr = fmt.Sprintf("%d", sprintDuration)
		sprintDurationSetStr = fmt.Sprintf("%d", s.GetSprintDurationSet())
	}

	return fmt.Sprintf("Session of %s🍅 x %dm + %dm",
		sprintDurationSetStr, s.GetPomodoroDurationSet()/60, s.GetRestDurationSet()/60) +
		fmt.Sprintf("\nPomodoros remaining: %s", pomodorosRemainingStr) +
		middleStr +
		fmt.Sprintf("\n\nCurrent session state: %s", s.State())
}

func (sdd SessionDefaultData) String() string {
	if sdd.PomodoroDurationSet == 0 {
		return "No session"
	}

	var middleStr string
	sprintDuration := sdd.SprintDurationSet

	var sprintDurationSetStr string
	var pomodorosRemainingStr string
	if sdd.SprintDurationSet <= UnspecifiedSprintCardinality {
		pomodorosRemainingStr = "Unspecified"
		sprintDurationSetStr = "X"
	} else {
		pomodorosRemainingStr = fmt.Sprintf("%d", sprintDuration)
		sprintDurationSetStr = fmt.Sprintf("%d", sdd.SprintDurationSet)
	}

	return fmt.Sprintf("Session of %s🍅 x %dm + %dm",
		sprintDurationSetStr, sdd.PomodoroDurationSet/60, sdd.RestDurationSet/60) +
		fmt.Sprintf("\nPomodoros remaining: %s", pomodorosRemainingStr) +
		middleStr +
		fmt.Sprintf("\n\nCurrent session state: Pending")
}

// LeftTimeMessage Print in a string in human-readable format (aimed at the
// user) how much time is left either for task time or for rest.
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
		(!s.IsSprintDurationUnspecified() && s.GetSprintDuration() < 0) ||
		s.data.IsPaused ||
		s.data.IsCancel ||
		s.data.IsFinished {
		return true
	}

	return false
}

// IsCanceled returns true if Session has been canceled, otherwise false.
func (s *Session) IsCanceled() bool {
	return s.data.IsCancel
}

// IsPaused returns true if Session has been paused or never started, otherwise false.
func (s *Session) IsPaused() bool {
	return s.data.IsPaused
}

// IsFinished returns true if Session has been completed, otherwise false.
// Note that sessions are not expected to be revived after they become
// finished.
func (s *Session) IsFinished() bool {
	return s.data.IsFinished
}

// State return the Session's state as a string.
//
// # The values are
//
// "Pending" if the session was never started (and is actually on pause)
//
// "Paused" if the session is on pause, and it was started earlier.
//
// "Canceled" if the session has been canceled (s.IsCanceled() == true)
//
// "Finished" if the session is finished (s.IsFinished() == true)
//
// "Stopped" if the session is not running and none result of the above was
// the state.
//
// "Running" if the session is actually running
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

// Pause Prepare a Session to be paused.
// This method modifies Session data structures, so should be used
// in a context where it is actually safe to do so.
//
// At the time of writing, each Session obj in this project is managed by one
// and only one goroutine. Pause() call is internal to such goroutine,
// therefore, it should not happen elsewhere.
func (s *Session) Pause() {
	// Cache pomodoro and rest duration. We will use them again to assign new timestamps.
	s.data.PomodoroDuration = s.GetPomodoroDuration()
	s.data.RestDuration = s.GetRestDuration()

	// Nil the timestamps (they have to be re-calculated)
	s.endNextSprintTimestamp = nil
	s.endNextRestTimestamp = nil

	s.data.IsPaused = true
}

// Cancel Set IsCancel internal attribute to true.
// This method modifies Session data structures, so should be used
// in a context where it is actually safe to do so.
func (s *Session) Cancel() {
	s.data.IsCancel = true
}

// SetFinished Set IsFinished internal attribute to true.
// This method modifies Session data structures, so should be used
// in a context where it is actually safe to do so.
func (s *Session) SetFinished() {
	s.data.IsFinished = true
}

// Resume Prepare a Session to be resumed.
// This method modifies Session data structures, so should be used
// in a context where it is actually safe to do so.
func (s *Session) Resume() {
	s.data.IsPaused = false

	s.assignTimestamps()
}

// Start Prepare a Session for the start.
// This method modifies Session data structures, so should be used
// in a context where it is actually safe to do so.
func (s *Session) Start() {
	s.data.IsPaused = false
	s.data.IsCancel = false

	s.data.SprintDuration -= 1

	s.assignTimestamps()
}

// RestStarted Prepare a Session object for rest start.
// This method modifies Session data structures, so should be used
// in a context where it is actually safe to do so.
//
// At the time of writing, each Session obj in this project is managed by one
// and only one goroutine. RestStarted() call is internal to such goroutine,
// therefore, it should not happen elsewhere.
func (s *Session) RestStarted() {
	s.data.IsRest = true
	s.data.RestDuration = s.restDurationSet
	s.assignTimestamps()
}

// RestFinished Prepare a Session object for rest end.
// This method modifies Session data structures, so should be used
// in a context where it is actually safe to do so.
//
// At the time of writing, each Session obj in this project is managed by one
// and only one goroutine. RestFinished() call is internal to such goroutine,
// therefore, it should not happen elsewhere.
func (s *Session) RestFinished() {
	s.data.IsRest = false
	s.data.PomodoroDuration = s.pomodoroDurationSet
	s.assignTimestamps()
}

// DecreaseSprintDuration Diminish by 1 the SprintDuration attribute.
// This method modifies Session data structures, so should be used
// in a context where it is actually safe to do so.
//
// At the time of writing, each Session obj in this project is managed by one
// and only one goroutine. DecreaseSprintDuration() call is internal to such
// goroutine, therefore, it should not happen elsewhere.
func (s *Session) DecreaseSprintDuration() {
	if s.data.SprintDuration == UnspecifiedSprintCardinality {
		return
	}

	s.data.SprintDuration -= 1
}

// ClearChannel close and clear (set to nil) ActionsChannel attribute.
// Call this method after a session object is discarded (its session manager
// dropped it away). Should be the session be revived (e.g., after a Resume)
// the channel field should be populated again.
func (s *Session) ClearChannel() {
	close(s.ActionsChannel)
	s.ActionsChannel = nil
}

// HasSprintEndTimePassed
// Returns true if sprint should be ended at this time, otherwise false.
// It returns false if a timestamp was not set, but this would be an error case
// and printed in the log.
func (s *Session) HasSprintEndTimePassed() bool {
	if s.endNextSprintTimestamp == nil {
		log.Println("[PROBLEM] s.endNextSprintTimestamp IS nil.")
		return false
	}

	return time.Now().Local().After(*s.endNextSprintTimestamp)
}

// HasRestEndTimePassed
// Returns true if rest should be ended at this time, otherwise false.
// It returns false if a timestamp was not set, but this would be an error case
// and printed in the log.
func (s *Session) HasRestEndTimePassed() bool {
	if s.endNextRestTimestamp == nil {
		log.Println("[PROBLEM] s.endNextRestTimestamp IS nil.")
		return false
	}

	return time.Now().Local().After(*s.endNextRestTimestamp)
}

func (s *Session) EndNextSprintTimestamp() *time.Time {
	return s.endNextSprintTimestamp
}

func (s *Session) EndNextRestTimestamp() *time.Time {
	return s.endNextRestTimestamp
}

func (s *Session) CalculateSessionTimeInSeconds() int64 {
	numberOfSprints := int64(s.GetSprintDurationSet().ToInt())
	sessionTime := int64(s.GetPomodoroDurationSet().Seconds()) * numberOfSprints
	if numberOfSprints > 1 {
		sessionTime += int64(s.GetRestDurationSet().Seconds()) * (numberOfSprints - 1)
	}
	return sessionTime
}

func (sdd SessionDefaultData) CalculateSessionTimeInSeconds() int64 {
	numberOfSprints := int64(sdd.SprintDurationSet)
	sessionTime := int64(sdd.PomodoroDurationSet) * numberOfSprints
	if numberOfSprints > 1 {
		sessionTime += int64(sdd.RestDurationSet) * (numberOfSprints - 1)
	}
	return sessionTime
}
