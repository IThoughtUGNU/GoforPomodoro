package main

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func ActionResumeSprint(
	bot *tgbotapi.BotAPI,
	update tgbotapi.Update,
	appState *AppState,
	chatId ChatID,
) {
	session := GetUserSessionRunning(appState, chatId)
	err := ResumeSession(
		chatId,
		session,
		// Rest begin handler
		func(id ChatID, session *Session) {
			text := fmt.Sprintf(
				"Pomodoro done! Have rest for %s now.",
				NiceTimeFormatting(session.RestDurationSet),
			)

			ReplyWith(bot, update, text)
		},
		// Rest finish handler
		func(id ChatID, session *Session) {
			text := fmt.Sprintf(
				"Pomodoro %s started.",
				NiceTimeFormatting(session.PomodoroDurationSet),
			)
			ReplyWithAndHourglass(bot, update, text)
		},
		// End sessionDefault handler
		func(id ChatID, session *Session, endKind PomodoroEndKind) {
			switch endKind {
			case PomodoroFinished:
				ReplyWith(bot, update, "Pomodoro done! The session is complete, congratulations!")
			case PomodoroCanceled:
				ReplyWith(bot, update, "Session canceled.")
			}
		},
		// Pause sessionDefault handler
		func(id ChatID, session *Session) {
			ReplyWith(bot, update, "Your session has paused.")
		},
	)
	if err != nil {
		if session.isZero() {
			ReplyWith(bot, update, "Session was not set.")
		} else if session.isCancel {
			ReplyWith(bot, update, "Last session was canceled.")
		} else if !session.isStopped() {
			ReplyWith(bot, update, "Session is already running.")
		} else {
			ReplyWith(bot, update, "Server error.")
		}
		return
	}
	ReplyWithAndHourglass(bot, update, "Session resumed!")
}

func ActionStartSprint(
	bot *tgbotapi.BotAPI,
	update tgbotapi.Update,
	appState *AppState,
	chatId ChatID,
) {
	session := GetUserSessionRunning(appState, chatId)

	if !session.isStopped() {
		ReplyWith(bot, update, "A session already running.")
		return
	}
	session = GetNewUserSessionRunning(appState, chatId)
	err := StartSession(
		chatId,
		session,
		// Rest begin handler
		func(id ChatID, session *Session) {
			text := fmt.Sprintf(
				"Pomodoro done! Have rest for %s now.",
				NiceTimeFormatting(session.RestDurationSet),
			)

			ReplyWith(bot, update, text)
		},
		// Rest finish handler
		func(id ChatID, session *Session) {
			text := fmt.Sprintf(
				"Pomodoro %s started.",
				NiceTimeFormatting(session.PomodoroDurationSet),
			)
			ReplyWithAndHourglass(bot, update, text)
		},
		// End sessionDefault handler
		func(id ChatID, session *Session, endKind PomodoroEndKind) {
			switch endKind {
			case PomodoroFinished:
				ReplyWith(bot, update, "Pomodoro done! The session is complete, congratulations!")
			case PomodoroCanceled:
				ReplyWith(bot, update, "Session canceled.")
			}
		},
		// Pause sessionDefault handler
		func(id ChatID, session *Session) {
			ReplyWith(bot, update, "Your session has paused.")
		},
	)
	if err == nil {
		ReplyWithAndHourglass(bot, update, "Session started!")
	} else {
		ReplyWith(bot, update, "Session was not set.\nPlease set a session or use /default for classic 4x25m+25m.")
	}
}
