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

package inputprocess

import (
	"GoforPomodoro/internal/domain"
	"regexp"
	"strconv"
	"strings"
)

const BasicPattern = `\/([1-9]\d*)(for([1-9]\d*)(rest([1-9]\d*))?)?` // `\/([1-9]\d*)`

const (
	MinutesGroup     = 1
	CardinalityGroup = 3
	RestGroup        = 5
)

func CommandFrom(appSettings *domain.AppSettings, text string) string {
	command := strings.Split(text, " ")[0]

	botName := appSettings.BotName
	if !strings.HasPrefix(botName, "@") {
		botName = "@" + botName
	}

	if strings.HasSuffix(command, botName) {
		command = strings.Split(command, "@")[0]
	}

	return command
}

func ParametersFrom(text string) []string {
	return strings.Split(text, " ")[1:]
}

func ParsePatternToSession(r *regexp.Regexp, text string) *domain.Session {
	if r == nil {
		r = regexp.MustCompile(BasicPattern)
	}
	matches := r.FindAllStringSubmatch(text, -1)

	var sessionInitData domain.SessionInitData

	match := false
	for _, v := range matches {
		match = true

		sessionInitData.SprintDurationSet = 1
		sessionInitData.SprintDuration = 1
		sessionInitData.IsPaused = true

		// Mandatory parameter for this command.
		pomDuration, err := strconv.Atoi(v[MinutesGroup])
		if err != nil {
			return nil
		}
		sessionInitData.PomodoroDurationSet = domain.PomodoroDuration(pomDuration * 60) // time from minutes to seconds.
		sessionInitData.PomodoroDuration = sessionInitData.PomodoroDurationSet

		// Other parameters are optional
		sprintDuration, err := strconv.Atoi(v[CardinalityGroup])
		if err == nil {
			sessionInitData.SprintDurationSet = domain.SprintDuration(sprintDuration)
			sessionInitData.SprintDuration = sessionInitData.SprintDurationSet

			// Default 5 minutes of rest duration in case user did not specify.
			sessionInitData.RestDurationSet = 5 * 60
			sessionInitData.RestDuration = sessionInitData.RestDurationSet
		}

		restDuration, err := strconv.Atoi(v[RestGroup])
		if err == nil {
			sessionInitData.RestDurationSet = domain.RestDuration(restDuration * 60)
			sessionInitData.RestDuration = sessionInitData.RestDurationSet
		}

		break
	}

	if !match {
		return nil
	}

	return sessionInitData.ToSession()
}
