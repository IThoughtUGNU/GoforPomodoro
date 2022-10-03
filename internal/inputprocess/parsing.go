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

	var session *domain.Session

	for _, v := range matches {
		if session == nil {
			session = new(domain.Session).InitChannel()
			session.SprintDurationSet = 1
			session.Data.SprintDuration = 1
			session.Data.IsPaused = true
		}

		// Mandatory parameter for this command.
		pomDuration, err := strconv.Atoi(v[MinutesGroup])
		if err != nil {
			return nil
		}
		session.PomodoroDurationSet = pomDuration * 60 // time from minutes to seconds.
		session.Data.PomodoroDuration = session.PomodoroDurationSet

		// Other parameters are optional
		sprintDuration, err := strconv.Atoi(v[CardinalityGroup])
		if err == nil {
			session.SprintDurationSet = sprintDuration
			session.Data.SprintDuration = session.SprintDurationSet

			// Default 5 minutes of rest duration in case user did not specify.
			session.RestDurationSet = 5 * 60
			session.Data.RestDuration = session.RestDurationSet
		}

		restDuration, err := strconv.Atoi(v[RestGroup])
		if err == nil {
			session.RestDurationSet = restDuration * 60
			session.Data.RestDuration = session.RestDurationSet
		}

		break
	}
	return session
}
