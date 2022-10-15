// This file is part of GoforPomodoro.
//
// GoforPomodoro is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// GoforPomodoro is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with GoforPomodoro.  If not, see <http://www.gnu.org/licenses/>.

package botmodule

import (
	"GoforPomodoro/internal/data"
	"GoforPomodoro/internal/domain"
	"GoforPomodoro/internal/inputprocess"
	"GoforPomodoro/internal/sessionmanager"
	"GoforPomodoro/internal/utils"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"io"
	"log"
	"net/http"
	"os"
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

func ListenPrivateHTTP(appState *domain.AppState, address string, port int) {
	http.HandleFunc("/hello", getHello)
	http.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request) {
		// dispatchServerAction <- domain.DispatchServerAction{Shutdown: true}
		log.Println("[ListenPrivateHTTP] Shutdown request from HTTP.")
		data.PrepareForShutdown(
			appState,
			func() {
				log.Println("[ListenPrivateHTTP] DB lock acquired.")
				_, _ = io.WriteString(w, "shutting down\n")
				go os.Exit(0)
			},
		)
	})

	err := http.ListenAndServe(fmt.Sprintf("%s:%d", address, port), nil)
	if err != nil {
		log.Fatal(err)
	}
}

func getHello(w http.ResponseWriter, r *http.Request) {
	_ = r
	fmt.Printf("got /hello request\n")
	_, err := io.WriteString(w, "Hello, HTTP!\n")
	if err != nil {
		log.Println("getHello err:", err)
	}
}

func CommandMenuLoop(
	settings *domain.AppSettings,
	appVariables *domain.AppVariables,
	appState *domain.AppState,
) {
	bot, err := tgbotapi.NewBotAPI(settings.ApiToken)
	if err != nil {
		log.Panic(err)
	}

	settings.BotName = bot.Self.UserName

	debugMode := settings.DebugMode
	bot.Debug = debugMode

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	RestoreSessions(appState, appVariables, bot)

	PrivacyPolicyEnabled := appVariables.PrivacyPolicyEnabled
	privacyVersion := appVariables.PrivacySettingsVersion

	updates := bot.GetUpdatesChan(u)

mainLoop:
	for update := range updates {
		if update.Message != nil { // If we got a message
			senderId := domain.ChatID(update.Message.From.ID)
			chatId := domain.ChatID(update.Message.Chat.ID)

			newChat := data.IsThisNewUser(appState, chatId)

			if debugMode {
				log.Printf("[%s] %s\n", update.Message.From.UserName, update.Message.Text)
				log.Printf("New chat? | %v\n", utils.YesNo(newChat))
			}

			msgText := update.Message.Text

			// var replyMsg tgbotapi.MessageConfig
			// var replyMsgText string

			command := inputprocess.CommandFrom(settings, msgText)
			parameters := inputprocess.ParametersFrom(msgText)

			if debugMode {
				log.Printf("command: %s\n", command)
			}

			isGroup := update.Message.Chat.IsGroup() || update.Message.Chat.IsSuperGroup()
			data.AdjustChatType(appState, chatId, senderId, isGroup)

			communicator := GetCommunicator(appState, appVariables, chatId, bot)

			if PrivacyPolicyEnabled {
				// Check privacy policy agreement
				userPrivacy, userPrivacyVersion := data.GetUserPrivacyPolicy(appState, chatId)
				if userPrivacy.IsZero() || privacyVersion > userPrivacyVersion {
					// The user has no privacy policy set (or it is too old).

					// If the user is changing privacy now, we manage the change.
					if inputprocess.IsPrivacySettingsCommand(command) {
						switch command {
						case "/accept_essential":
							data.SetUserPrivacyPolicy(appState, chatId, domain.AcceptedEssential, privacyVersion)
						case "/accept_all":
							data.SetUserPrivacyPolicy(appState, chatId, domain.AcceptedAll, privacyVersion)
						}
						communicator.PrivacySettingsUpdated()

						data.DefaultUserSettingsIfNeeded(appState, chatId)
					} else {
						// Otherwise, must show privacy policy
						communicator.ShowPrivacyPolicy()
						communicator.ShowLicenseNotice()
					}
					continue
				}
			} else {
				if newChat {
					communicator.Info()
					communicator.Help()
					data.DefaultUserSettingsIfNeeded(appState, chatId)
				}
			}

			switch command {
			// Admin commands
			case "/shutdown":
				isAdmin := utils.Contains(settings.AdminIds, senderId)
				if isAdmin {
					communicator.ReplyWith("Soft shutting down...")
					data.PrepareForShutdown(
						appState,
						func() {
							communicator.ReplyWith("DB lock acquired.")
							os.Exit(0)
						},
					)
					break mainLoop
				}
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
				ActionResumeSprint(senderId, chatId, appState, communicator)
			case "/d", "/default":
				data.UpdateDefaultUserSession(appState, chatId, senderId, domain.DefaultSession())
				ActionStartSprint(senderId, chatId, appState, communicator)
			case "/s", "/start_sprint":
				ActionStartSprint(senderId, chatId, appState, communicator)
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
				sessionDataOpt := inputprocess.ParsePatternToSession(nil, msgText)
				sessionData, err := sessionDataOpt.GetValue()
				if err != nil {
					// Session wasn't parsed
					continue
				}
				_, err = inputprocess.ValidateSessionParsed(sessionData)
				if err == nil {
					data.UpdateDefaultUserSession(appState, chatId, senderId, sessionData)
					communicator.NewSession(sessionData)
					autorun := data.GetUserAutorun(appState, chatId, senderId)
					if autorun {
						ActionStartSprint(senderId, chatId, appState, communicator)
					}
				} else {
					communicator.ErrorSessionTooLong()
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
