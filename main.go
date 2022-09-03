package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strings"
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

			chatId := ChatID(update.Message.Chat.ID)

			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			msgText := update.Message.Text

			var replyMsg tgbotapi.MessageConfig
			var replyMsgText string

			command := commandFrom(settings, msgText)
			parameters := parametersFrom(msgText)
			log.Printf("command: %s\n", command)

			switch command {
			case "/autorun":
				if len(parameters) > 0 {
					param := parameters[0]
					var autorun bool
					if param == "on" {
						autorun = true
					} else if param == "off" {
						autorun = false
					} else {
						ReplyWith(bot, update, "Command error.")
						continue
					}
					SetUserAutorun(appState, chatId, autorun)
					ReplyWith(bot, update, "Autorun set "+strings.ToUpper(param)+".")
				} else {
					SetUserAutorun(appState, chatId, true)
					ReplyWith(bot, update, "Autorun set ON.")
				}
			case "/se", "/session":
				session := GetUserSessionRunning(appState, chatId)
				var stateStr = session.State()

				if session.isCancel {
					replyMsgText = fmt.Sprintf("Your session state: %s.", stateStr)
				} else {
					replyMsgText = session.String()
				}
				ReplyWith(bot, update, replyMsgText)
			case "/p", "/pause":
				session := GetUserSessionRunning(appState, chatId)
				err := PauseSession(session)
				if err != nil {
					if !session.isStopped() {
						ReplyWith(bot, update, "Session was not running.")
					} else {
						ReplyWith(bot, update, "Server error.")
					}
					continue
				}
			case "/c", "/cancel":
				session := GetUserSessionRunning(appState, chatId)
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
							NiceTimeFormatting(session.RestDurationSet),
						)
						ReplyWith(bot, update, text)
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
					continue
				}
				ReplyWith(bot, update, "Session resumed!")
			case "/d", "/default":
				UpdateUserSession(appState, chatId, DefaultSession())
				ActionStartSprint(bot, update, appState, chatId)
			case "/s", "/start_sprint":
				ActionStartSprint(bot, update, appState, chatId)
			case "/reset":
				CleanUserSettings(appState, chatId)
				ReplyWith(bot, update, "Your data has been cleaned.")
			case "/info":
				ReplyWith(bot, update, "I am a pomodoro bot written in Go.")
			default:
				newSession := ParsePatternToSession(nil, msgText)

				if newSession != nil {
					replyMsgText = fmt.Sprintf("New session!\n\n%s", newSession.String())

					UpdateUserSession(appState, chatId, *newSession)
					replyMsg = tgbotapi.NewMessage(update.Message.Chat.ID, replyMsgText)

					_, err := bot.Send(replyMsg)
					if err != nil {
						log.Printf("ERROR: %s", err.Error())
					}

					autorun := GetUserAutorun(appState, chatId)
					if autorun {
						ActionStartSprint(bot, update, appState, chatId)
					}
				} else {
					// replyMsg.ReplyToMessageID = update.Message.MessageID

					// replyMsgText = "Can't manage this command right now."
				}
			}
		}
	}
}
