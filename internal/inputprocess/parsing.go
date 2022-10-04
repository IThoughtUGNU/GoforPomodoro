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

	return sessionInitData.ToSession().InitChannel()
}
