package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

func loadAppSettings() (*AppSettings, error) {
	settings := new(AppSettings)
	_, err := toml.DecodeFile("appsettings.toml", settings)

	return settings, err
}

func ReplyWith(bot *tgbotapi.BotAPI, update tgbotapi.Update, text string) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	_, err := bot.Send(msg)
	if err != nil {
		log.Printf("ERROR: %s", err.Error())
	}
}

func main() {
	settings, err := loadAppSettings()
	if err != nil {
		log.Fatal(err)
	}

	appState, err := LoadAppState()
	if err != nil {
		panic(err)
	}

	fmt.Println("Hello from Go for Pomodoro!")

	fmt.Printf("%s\n\n", settings.ApiToken)
	bot, err := tgbotapi.NewBotAPI(settings.ApiToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil { // If we got a message

			userId := UserID(update.Message.From.ID)

			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			msgText := update.Message.Text

			var replyMsg tgbotapi.MessageConfig
			var replyMsgText string

			switch msgText {
			case "/se", "/session":
				session := GetUserSessionRunning(appState, userId)
				var stateStr = session.State()

				if session.isCancel {
					replyMsgText = fmt.Sprintf("Your session state: %s.", stateStr)
				} else {
					replyMsgText = session.String()
					/*replyMsgText = fmt.Sprintf(
					"Your session state: %s,\n"+
						"pomodoro time left: %s", stateStr, NiceTimeFormatting(session.PomodoroDuration))*/
				}
				ReplyWith(bot, update, replyMsgText)
				continue
			case "/p", "/pause":
				session := GetUserSessionRunning(appState, userId)
				err := PauseSession(session)
				if err != nil {
					if !session.isStopped() {
						ReplyWith(bot, update, "Session was not running.")
					} else {
						ReplyWith(bot, update, "Server error.")
					}
					continue
				}
				// ReplyWith(bot, update, "Session paused!")
				continue
			case "/c", "/cancel":
				session := GetUserSessionRunning(appState, userId)
				err := CancelSession(session)
				if err != nil {
					if session.isStopped() {
						ReplyWith(bot, update, "Session was not running.")
					} else {
						ReplyWith(bot, update, "Server error.")
					}
					continue
				}
			case "/resume":
				session := GetUserSessionRunning(appState, userId)
				err := ResumeSession(
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
					continue
				}
				ReplyWith(bot, update, "Session resumed!")
				continue
			case "/d", "/default":
				UpdateUserSession(appState, userId, DefaultSession())
				ActionStartSprint(bot, update, appState, userId)
			case "/s", "/start_sprint":
				ActionStartSprint(bot, update, appState, userId)
				continue
			case "/reset":

				ReplyWith(bot, update, "Your data has been cleaned.")
			case "/info":
				ReplyWith(bot, update, "I am a pomodoro bot written in Go.")
				continue
			default:
				newSession := ParsePatternToSession(nil, msgText)

				if newSession != nil {
					replyMsgText = fmt.Sprintf("New session!\n\n%s", newSession.String())

					UpdateUserSession(appState, UserID(userId), *newSession)
					replyMsg = tgbotapi.NewMessage(update.Message.Chat.ID, replyMsgText)

					_, err := bot.Send(replyMsg)
					if err != nil {
						log.Printf("ERROR: %s", err.Error())
					}
				} else {
					// replyMsg.ReplyToMessageID = update.Message.MessageID

					// replyMsgText = "Can't manage this command right now."
				}
			}
		}
	}
}
