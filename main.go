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

/*
var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonURL("1.com", "http://1.com"),
		tgbotapi.NewInlineKeyboardButtonData("2", "2"),
		tgbotapi.NewInlineKeyboardButtonData("3", "3"),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("4", "4"),
		tgbotapi.NewInlineKeyboardButtonData("5", "5"),
		tgbotapi.NewInlineKeyboardButtonData("6", "6"),
	),
)*/

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
				ActionResumeSprint(bot, update, appState, chatId)
			case "/d", "/default":
				UpdateUserSession(appState, chatId, DefaultSession())
				ActionStartSprint(bot, update, appState, chatId)
			case "/s", "/start_sprint":
				ActionStartSprint(bot, update, appState, chatId)
			case "/reset":
				CleanUserSettings(appState, chatId)
				ReplyWith(bot, update, "Your data has been cleaned.")
			case "/help":
				ReplyWith(bot, update, "E.g.\n/25for4rest5 --> 4 🍅, 25 minutes + 5m for rest.\n"+
					"The latter is also achieved with /default.\n"+
					"/30for4 --> 4 🍅, 30 minutes (default: +5m for rest).\n"+
					"/25 --> 1 🍅, 25 minutes (single pomodoro sprint)\n\n"+
					"(/s) /start_sprint to start (if /autorun is set off)\n"+
					"(/p) /pause to pause a session in run\n"+
					"(/c) /cancel to cancel a session\n"+
					"/resume to resume a paused session.\n"+
					"(/se) /session to check your session settings and status.\n"+
					"/reset to reset your profile/chat settings.\n"+
					"/info to have some info on this bot.")
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
					/*
						switch update.Message.Text {
						case "open":
							replyMsgText = "Keyboard test"
							replyMsg = tgbotapi.NewMessage(update.Message.Chat.ID, replyMsgText)

							replyMsg.ReplyMarkup = simpleHourglassKeyboard
							_, err := bot.Send(replyMsg)
							if err != nil {
								log.Printf("ERROR: %s", err.Error())
							}
						}*/

					// replyMsg.ReplyToMessageID = update.Message.MessageID

					// replyMsgText = "Can't manage this command right now."
				}
			}
		} else if update.CallbackQuery != nil {
			// Respond to the callback query, telling Telegram to show the user
			// a message with the data received.

			switch update.CallbackQuery.Data {
			case "⌛":
				chatId := ChatID(update.CallbackQuery.Message.Chat.ID)
				session := GetUserSessionRunning(appState, chatId)
				toastText := session.LeftTimeMessage()
				callback := tgbotapi.NewCallback(update.CallbackQuery.ID, toastText)
				if _, err := bot.Request(callback); err != nil {
					panic(err)
				}
			}
			/*
				// And finally, send a message containing the data received.
				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Data)
				if _, err := bot.Send(msg); err != nil {
					panic(err)
				}*/
		}
	}
}
