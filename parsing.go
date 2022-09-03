package main

import (
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

func commandFrom(appSettings *AppSettings, text string) string {
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

func parametersFrom(text string) []string {
	return strings.Split(text, " ")[1:]
}

func ParsePatternToSession(r *regexp.Regexp, text string) *Session {
	if r == nil {
		r = regexp.MustCompile(BasicPattern)
	}
	matches := r.FindAllStringSubmatch(text, -1)

	var session *Session

	for _, v := range matches {
		if session == nil {
			session = new(Session)
			session.SprintDurationSet = 1
			session.SprintDuration = 1
			session.isPaused = true
		}

		// Mandatory parameter for this command.
		pomDuration, err := strconv.Atoi(v[MinutesGroup])
		if err != nil {
			return nil
		}
		session.PomodoroDurationSet = pomDuration * 60 // time from minutes to seconds.
		session.PomodoroDuration = session.PomodoroDurationSet

		// Other parameters are optional
		sprintDuration, err := strconv.Atoi(v[CardinalityGroup])
		if err == nil {
			session.SprintDurationSet = sprintDuration
			session.SprintDuration = session.SprintDurationSet
		}

		restDuration, err := strconv.Atoi(v[RestGroup])
		if err == nil {
			session.RestDurationSet = restDuration * 60
			session.RestDuration = session.RestDurationSet
		}

		break
	}
	return session
}
