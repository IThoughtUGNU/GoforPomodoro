package botmodule

import (
	"GoforPomodoro/internal/data"
	"GoforPomodoro/internal/domain"
	"GoforPomodoro/internal/sessionmanager"
	"GoforPomodoro/internal/utils"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strings"
)

var simpleHourglassKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("⌛", "⌛"),
	),
)

type Communicator struct {
	appState *domain.AppState
	domain.ChatID
	Bot         *tgbotapi.BotAPI
	Subscribers []domain.ChatID
	IsGroup     bool
}

func GetCommunicator(appState *domain.AppState, chatId domain.ChatID, bot *tgbotapi.BotAPI) *Communicator {
	communicator := new(Communicator)

	communicator.appState = appState
	communicator.ChatID = chatId
	communicator.Bot = bot
	communicator.Subscribers = data.GetSubscribers(appState, chatId)
	communicator.IsGroup = data.IsGroup(appState, chatId)

	return communicator
}

func (c *Communicator) subscribersAsString() string {
	bot := c.Bot

	var sb strings.Builder

	errors := 0
	for _, id := range c.Subscribers {
		subscriberName, err := bot.GetChat(tgbotapi.ChatInfoConfig{ChatConfig: tgbotapi.ChatConfig{ChatID: int64(id)}})
		if err != nil {
			errors += 1
		}
		sb.WriteString("@")
		sb.WriteString(subscriberName.UserName)
		sb.WriteString(" ")
	}

	return sb.String()
}

func (c *Communicator) toNotify(message string) string {
	// Update subscribers in case they changed
	c.Subscribers = data.GetSubscribers(c.appState, c.ChatID)

	if !c.IsGroup || len(c.Subscribers) == 0 {
		// This function is identity function if we're not in a group or there are no subscribers.
		return message
	}

	return message + "\n\n———\n" + c.subscribersAsString()
}

func (c *Communicator) Subscribe(err error, update tgbotapi.Update, username string) {
	if err != nil {
		switch err.Error() {
		case domain.AlreadySubscribed{}.Error():
			c.ReplyWith("You already subscribed this chat group.\n\n" +
				"Remember you can use /leave to cancel subscription.")
		case domain.SubscriptionError{}.Error():
			c.ReplyWith("There has been an error with this operation (subscription).")
		}
	} else {
		c.ReplyWith(fmt.Sprintf("Done! You will be tagged (@%s) in sprints' messages.", username))
	}
}

func (c *Communicator) Unsubscribe(err error) {
	if err != nil {
		switch err.Error() {
		case domain.AlreadyUnsubscribed{}.Error():
			c.ReplyWith("You are (were) not subscribed in this chat group.")
		case domain.SubscriptionError{}.Error():
			c.ReplyWith("There has been an error with this operation (subscription).")
		}
	} else {
		c.ReplyWith("Done! You no longer subscribe in this chat group thus will not be tagged in future messages.")
	}
}

func (c *Communicator) ReplyWith(text string) {
	bot := c.Bot
	chatId := int64(c.ChatID)

	msg := tgbotapi.NewMessage(chatId, text)
	_, err := bot.Send(msg)
	if err != nil {
		log.Printf("ERROR: %s", err.Error())
	}
}

func (c *Communicator) ReplyAndNotify(text string) {
	c.ReplyWith(c.toNotify(text))
}

func (c *Communicator) ReplyWithAndHourglass(text string) {
	msg := tgbotapi.NewMessage(int64(c.ChatID), text)
	msg.ReplyMarkup = simpleHourglassKeyboard
	_, err := c.Bot.Send(msg)
	if err != nil {
		log.Printf("ERROR: %s", err.Error())
	}
}

func (c *Communicator) ReplyWithAndHourglassAndNotify(text string) {
	c.ReplyWithAndHourglass(c.toNotify(text))
}

func (c *Communicator) SessionStarted(session *domain.Session, err error) {
	if err == nil {
		numberOfSprints := session.GetSprintDurationSet().ToInt()
		sessionTime := session.GetPomodoroDurationSet().Seconds() * numberOfSprints
		if numberOfSprints > 1 {
			sessionTime += session.GetRestDurationSet().Seconds() * (numberOfSprints - 1)
		}
		replyStr := fmt.Sprintf("This session will last for %s\n\nSession started!", utils.NiceTimeFormatting(sessionTime))
		c.ReplyWithAndHourglassAndNotify(replyStr)
	} else {
		c.ReplyWith("Session was not set.\nPlease set a session or use /default for classic 4x25m+25m.")
	}
}

/*
func (c *Communicator) SessionFinished() {

}

func (c *Communicator) SessionResumed() {

}

func (c *Communicator) SessionPaused() {

}*/

func (c *Communicator) SessionFinishedHandler(id domain.ChatID, session *domain.Session, endKind sessionmanager.PomodoroEndKind) {
	switch endKind {
	case sessionmanager.PomodoroFinished:
		c.ReplyAndNotify("Pomodoro done! The session is complete, congratulations!")
	case sessionmanager.PomodoroCanceled:
		c.ReplyAndNotify("Session canceled.")
	}
}

func (c *Communicator) SessionPausedHandler(id domain.ChatID, session *domain.Session) {
	c.ReplyAndNotify("Your session has paused.")
}

func (c *Communicator) RestFinishedHandler(id domain.ChatID, session *domain.Session) {
	text := fmt.Sprintf(
		"Pomodoro %s started.",
		utils.NiceTimeFormatting(session.GetPomodoroDurationSet().Seconds()),
	)
	c.ReplyWithAndHourglassAndNotify(text)
}

func (c *Communicator) RestBeginHandler(id domain.ChatID, session *domain.Session) {
	text := fmt.Sprintf(
		"Pomodoro done! Have rest for %s now.",
		utils.NiceTimeFormatting(session.GetRestDurationSet().Seconds()),
	)

	c.ReplyAndNotify(text)
}

func (c *Communicator) SessionAlreadyRunning() {
	c.ReplyWith("A session already running.")
}

func (c *Communicator) SessionResumed(err error, session *domain.Session) {
	if err != nil {
		if session.IsZero() {
			c.ReplyWith("Session was not set.")
		} else if session.IsCanceled() {
			c.ReplyWith("Last session was canceled.")
		} else if !session.IsStopped() {
			c.ReplyWith("Session is already running.")
		} else {
			c.ReplyWith("Server error.")
		}
		return
	}

	c.ReplyWithAndHourglassAndNotify("Session resumed!")
}

func (c *Communicator) OnlyGroupsCommand() {
	c.ReplyWith("This command works only in groups, sorry.")
}

func (c *Communicator) NewSession(session domain.Session) {
	c.ReplyWith(fmt.Sprintf("New session!\n\n%s", session.String()))
}

func (c *Communicator) Info() {
	c.ReplyWith("I am a pomodoro bot written in Go.\n\nExperimental Version for new concurrency and persistence.")
}

func (c *Communicator) DataCleaned() {
	c.ReplyWith("Your data has been cleaned.")
}

func (c *Communicator) Help() {
	c.ReplyWith("E.g.\n/25for4rest5 --> 4 🍅, 25 minutes + 5m for rest.\n" +
		"The latter is also achieved with /default.\n" +
		"/30for4 --> 4 🍅, 30 minutes (default: +5m for rest).\n" +
		"/25 --> 1 🍅, 25 minutes (single pomodoro sprint)\n\n" +
		"(/s) /start_sprint to start (if /autorun is set off)\n" +
		"(/p) /pause to pause a session in run\n" +
		"(/c) /cancel to cancel a session\n" +
		"/resume to resume a paused session.\n" +
		"(/se) /session to check your session settings and status.\n" +
		"/reset to reset your profile/chat settings.\n" +
		"/info to have some info on this bot.")
}

func (c *Communicator) SessionPaused(err error, session domain.Session) {
	if err != nil {
		if !session.IsStopped() {
			c.ReplyWith("Session was not running.")
		} else {
			c.ReplyWith("Server error.")
		}
	}
}

func (c *Communicator) SessionCanceled(err error, session domain.Session) {
	if err != nil {
		if session.IsStopped() {
			c.ReplyWith("Session was not running.")
		} else {
			c.ReplyWith("Server error.")
		}
	}
}

func (c *Communicator) SessionState(session domain.Session) {
	var stateStr = session.State()

	var replyMsgText string
	if session.IsCanceled() {
		replyMsgText = fmt.Sprintf("Your session state: %s.", stateStr)
	} else {
		replyMsgText = session.String()
	}
	c.ReplyWith(replyMsgText)
}

func (c *Communicator) CommandError() {
	c.ReplyWith("Command error.")
}

func (c *Communicator) Hourglass() {
	c.ReplyWithAndHourglass("Here is an hourglass")
}
