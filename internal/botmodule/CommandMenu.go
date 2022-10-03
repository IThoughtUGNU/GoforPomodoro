package botmodule

import (
	"GoforPomodoro/internal/data"
	"GoforPomodoro/internal/domain"
	"GoforPomodoro/internal/inputprocess"
	"GoforPomodoro/internal/sessionmanager"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strings"
)

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

func CommandMenuLoop(
	settings *domain.AppSettings,
	appState *domain.AppState,
) {
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

			senderId := domain.ChatID(update.Message.From.ID)
			chatId := domain.ChatID(update.Message.Chat.ID)

			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			msgText := update.Message.Text

			// var replyMsg tgbotapi.MessageConfig
			// var replyMsgText string

			command := inputprocess.CommandFrom(settings, msgText)
			parameters := inputprocess.ParametersFrom(msgText)
			log.Printf("command: %s\n", command)

			isGroup := update.Message.Chat.IsGroup() || update.Message.Chat.IsSuperGroup()
			data.AdjustChatType(appState, chatId, senderId, isGroup)

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

				communicator.Subscribe(
					data.SubscribeUserInGroup(appState, chatId, senderId),
					update,
					senderChat.UserName,
				)
			case "/leave":
				if !isGroup {
					communicator.OnlyGroupsCommand()
					continue
				}

				communicator.Unsubscribe(data.UnsubscribeUser(appState, chatId, senderId))
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
					data.SetUserAutorun(appState, chatId, senderId, autorun)
					communicator.ReplyWith("Autorun set " + strings.ToUpper(param) + ".")
				} else {
					data.SetUserAutorun(appState, chatId, senderId, true)
					communicator.ReplyWith("Autorun set ON.")
				}
			case "/se", "/session":
				session := data.GetUserSessionRunning(appState, chatId, senderId)
				communicator.SessionState(*session)
			case "/p", "/pause":
				session := data.GetUserSessionRunning(appState, chatId, senderId)
				err := sessionmanager.PauseSession(session)
				communicator.SessionPaused(err, *session)
			case "/c", "/cancel":
				session := data.GetUserSessionRunning(appState, chatId, senderId)
				err := sessionmanager.CancelSession(session)
				communicator.SessionCanceled(err, *session)
			case "/resume":
				ActionResumeSprint(update, appState, communicator)
			case "/d", "/default":
				data.UpdateUserSession(appState, chatId, senderId, domain.DefaultSession())
				ActionStartSprint(update, appState, communicator)
			case "/s", "/start_sprint":
				ActionStartSprint(update, appState, communicator)
			case "/reset":
				data.CleanUserSettings(appState, chatId, senderId)
				communicator.DataCleaned()
			case "/help":
				communicator.Help()
			case "/info":
				communicator.Info()
			case "/clessidra":
				communicator.Hourglass()
			default:
				newSession := inputprocess.ParsePatternToSession(nil, msgText)

				if newSession != nil {
					data.UpdateUserSession(appState, chatId, senderId, *newSession)
					communicator.NewSession(*newSession)

					autorun := data.GetUserAutorun(appState, chatId, senderId)
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
				chatId := domain.ChatID(update.CallbackQuery.Message.Chat.ID)
				senderId := domain.ChatID(update.CallbackQuery.Message.From.ID)

				session := data.GetUserSessionRunning(appState, chatId, senderId)

				// We reply with a toast (callback)
				toastText := session.LeftTimeMessage()
				callback := tgbotapi.NewCallback(update.CallbackQuery.ID, toastText)
				if _, err := bot.Request(callback); err != nil {
					log.Println("[ERROR] " + err.Error())
				}

				// To reply with a message
				// sg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Data)
				// if _, err := bot.Send(msg); err != nil {
				// 	   // manage error
				// }

			}
		}
	}
}
