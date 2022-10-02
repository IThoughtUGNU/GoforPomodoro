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

			senderId := ChatID(update.Message.From.ID)
			chatId := ChatID(update.Message.Chat.ID)

			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			msgText := update.Message.Text

			// var replyMsg tgbotapi.MessageConfig
			// var replyMsgText string

			command := commandFrom(settings, msgText)
			parameters := parametersFrom(msgText)
			log.Printf("command: %s\n", command)

			isGroup := update.Message.Chat.IsGroup() || update.Message.Chat.IsSuperGroup()
			AdjustChatType(appState, chatId, senderId, isGroup)

			communicator := GetCommunicator(appState, chatId, bot)
			switch command {
			// Group commands
			case "/join":
				if !isGroup {
					communicator.ReplyWith("This command works only in groups, sorry.")
					continue
				}
				senderChat, err := bot.GetChat(tgbotapi.ChatInfoConfig{ChatConfig: tgbotapi.ChatConfig{ChatID: int64(senderId)}})
				if err != nil {
					communicator.ReplyWith("Error with your account.")
					continue
				}

				// ReplyWith(bot, update, "Your username: @"+senderChat.UserName)

				communicator.Subscribe(
					SubscribeUserInGroup(appState, chatId, senderId),
					update,
					senderChat.UserName,
				)
			case "/leave":
				if !isGroup {
					communicator.OnlyGroupsCommand()
					continue
				}

				communicator.Unsubscribe(UnsubscribeUser(appState, chatId, senderId))
			// Personal commands
			case "/autorun":
				if len(parameters) > 0 {
					param := parameters[0]
					var autorun bool
					if param == "on" {
						autorun = true
					} else if param == "off" {
						autorun = false
					} else {
						communicator.CommandError()
						continue
					}
					SetUserAutorun(appState, chatId, senderId, autorun)
					ReplyWith(bot, update, "Autorun set "+strings.ToUpper(param)+".")
				} else {
					SetUserAutorun(appState, chatId, senderId, true)
					ReplyWith(bot, update, "Autorun set ON.")
				}
			case "/se", "/session":
				session := GetUserSessionRunning(appState, chatId, senderId)
				communicator.SessionState(*session)
			case "/p", "/pause":
				session := GetUserSessionRunning(appState, chatId, senderId)
				err := PauseSession(session)
				communicator.SessionPaused(err, *session)
			case "/c", "/cancel":
				session := GetUserSessionRunning(appState, chatId, senderId)
				err := CancelSession(session)
				communicator.SessionCanceled(err, *session)
			case "/resume":
				ActionResumeSprint(update, appState, communicator)
			case "/d", "/default":
				UpdateUserSession(appState, chatId, senderId, DefaultSession())
				ActionStartSprint(update, appState, communicator)
			case "/s", "/start_sprint":
				ActionStartSprint(update, appState, communicator)
			case "/reset":
				CleanUserSettings(appState, chatId, senderId)
				communicator.DataCleaned()
			case "/help":
				communicator.Help()
			case "/info":
				communicator.Info()
			default:
				newSession := ParsePatternToSession(nil, msgText)

				if newSession != nil {
					UpdateUserSession(appState, chatId, senderId, *newSession)
					communicator.NewSession(*newSession)

					autorun := GetUserAutorun(appState, chatId, senderId)
					if autorun {
						ActionStartSprint(update, appState, communicator)
					}
				}
			}
		} else if update.CallbackQuery != nil {
			// Respond to the callback query, telling Telegram to show the user
			// a message with the data received.

			switch update.CallbackQuery.Data {
			case "⌛":
				chatId := ChatID(update.CallbackQuery.Message.Chat.ID)
				senderId := ChatID(update.CallbackQuery.Message.From.ID)

				session := GetUserSessionRunning(appState, chatId, senderId)
				toastText := session.LeftTimeMessage()
				callback := tgbotapi.NewCallback(update.CallbackQuery.ID, toastText)
				if _, err := bot.Request(callback); err != nil {
					// panic(err)
					log.Println("[ERROR] " + err.Error())
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
