package main

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func ActionStartSprint(
	bot *tgbotapi.BotAPI,
	update tgbotapi.Update,
	appState *AppState,
	userId UserID,
) {
	session := GetUserSessionRunning(appState, userId)

	if !session.isStopped() {
		ReplyWith(bot, update, "A session already running.")
		return
	}
	session = GetNewUserSessionRunning(appState, userId)
	err := StartSession(
		userId,
		session,
		// Rest begin handler
		func(id UserID, session *Session) {
			text := fmt.Sprintf(
				"Pomodoro done! Have rest for %s now.",
				NiceTimeFormatting(session.RestDurationSet),
			)

			ReplyWith(bot, update, text)
		},
		// Rest finish handler
		func(id UserID, session *Session) {
			text := fmt.Sprintf(
				"Pomodoro %s started.",
				NiceTimeFormatting(session.RestDurationSet),
			)
			ReplyWith(bot, update, text)
		},
		// End sessionDefault handler
		func(id UserID, session *Session, endKind PomodoroEndKind) {
			switch endKind {
			case PomodoroFinished:
				ReplyWith(bot, update, "Pomodoro done! The session is complete, congratulations!")
			case PomodoroCanceled:
				ReplyWith(bot, update, "Session canceled.")
			}
		},
		// Pause sessionDefault handler
		func(id UserID, session *Session) {
			ReplyWith(bot, update, "Your session has paused.")
		},
	)
	if err == nil {
		ReplyWith(bot, update, "Session started!")
	} else {
		ReplyWith(bot, update, "Session was not set.\nPlease set a session or use /default for classic 4x25m+25m.")
	}
}
