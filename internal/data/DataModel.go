package data

import (
	"GoforPomodoro/internal/domain"
	"GoforPomodoro/internal/utils"
	"github.com/BurntSushi/toml"
	"log"
)

func LoadAppSettings() (*domain.AppSettings, error) {
	settings := new(domain.AppSettings)
	_, err := toml.DecodeFile("appsettings.toml", settings)

	return settings, err
}

func LoadAppState() (*domain.AppState, error) {
	// temporary
	appState := new(domain.AppState)
	appState.UsersSettings = make(map[domain.ChatID]*domain.Settings)
	return appState, nil
}

func defaultUserSettingsIfNeeded(appState *domain.AppState, chatId domain.ChatID) {
	if appState.UsersSettings[chatId] == nil {
		chatSettings := new(domain.Settings)
		chatSettings.Autorun = true
		appState.UsersSettings[chatId] = chatSettings
	}
}

func AdjustChatType(appState *domain.AppState, chatId domain.ChatID, senderId domain.ChatID, isGroup bool) {
	defaultUserSettingsIfNeeded(appState, chatId)
	appState.UsersSettings[chatId].IsGroup = isGroup
}

func IsGroup(appState *domain.AppState, chatId domain.ChatID) bool {
	defaultUserSettingsIfNeeded(appState, chatId)

	return appState.UsersSettings[chatId].IsGroup
}

func GetSubscribers(appState *domain.AppState, chatId domain.ChatID) []domain.ChatID {
	defaultUserSettingsIfNeeded(appState, chatId)

	return appState.UsersSettings[chatId].Subscribers
}

func SubscribeUserInGroup(appState *domain.AppState, chatId domain.ChatID, senderId domain.ChatID) error {
	defaultUserSettingsIfNeeded(appState, chatId)

	if chatId == senderId {
		return domain.SubscriptionError{}
	}

	settings := appState.UsersSettings[chatId]
	subscribers := (*settings).Subscribers
	if !utils.Contains(subscribers, senderId) {
		(*settings).Subscribers = append(subscribers, senderId)
	} else {
		return domain.AlreadySubscribed{}
	}
	return nil
}

func UnsubscribeUser(appState *domain.AppState, chatId domain.ChatID, senderId domain.ChatID) error {
	defaultUserSettingsIfNeeded(appState, chatId)

	if chatId == senderId {
		return domain.SubscriptionError{}
	}

	settings := appState.UsersSettings[chatId]
	subscribers := (*settings).Subscribers
	if utils.Contains(subscribers, senderId) {
		newS, err := utils.AfterRemoveEl(subscribers, senderId)
		if err != nil {
			log.Printf("[UnsubscribeUser] Error while removing %d\n", senderId)
			return domain.OperationError{}
		}
		(*settings).Subscribers = newS
	} else {
		log.Printf("[UnsubscribeUser] %d was not subscribed.", senderId)
		return domain.AlreadyUnsubscribed{}
	}
	return nil
}

func CleanUserSettings(appState *domain.AppState, chatId domain.ChatID, senderId domain.ChatID) {
	appState.UsersSettings[chatId] = nil // new(Settings)
	defaultUserSettingsIfNeeded(appState, chatId)
}

func SetUserAutorun(appState *domain.AppState, chatId domain.ChatID, senderId domain.ChatID, autorun bool) {
	defaultUserSettingsIfNeeded(appState, chatId)

	appState.UsersSettings[chatId].Autorun = autorun
}

func GetUserAutorun(appState *domain.AppState, chatId domain.ChatID, senderId domain.ChatID) bool {
	defaultUserSettingsIfNeeded(appState, chatId)

	return appState.UsersSettings[chatId].Autorun
}

func UpdateUserSession(appState *domain.AppState, chatId domain.ChatID, senderId domain.ChatID, session domain.Session) {
	defaultUserSettingsIfNeeded(appState, chatId)

	settings := appState.UsersSettings[chatId]
	settings.SessionDefault = session
}

func GetUserSessionFromSettings(appState *domain.AppState, chatId domain.ChatID, senderId domain.ChatID) domain.SessionInitData {
	defaultUserSettingsIfNeeded(appState, chatId)

	session := &appState.UsersSettings[chatId].SessionDefault
	// session.Data.IsPaused = true

	sData := session.ToInitData()
	sData.IsPaused = true

	return sData // this instantiates a new session object
}

func GetNewUserSessionRunning(appState *domain.AppState, chatId domain.ChatID, senderId domain.ChatID) *domain.Session {
	defaultUserSettingsIfNeeded(appState, chatId)

	sessionDef := GetUserSessionFromSettings(appState, chatId, senderId)

	sessionDef.PomodoroDuration = sessionDef.PomodoroDurationSet
	sessionDef.SprintDuration = sessionDef.SprintDurationSet
	sessionDef.RestDuration = sessionDef.RestDurationSet
	sessionDef.IsPaused = true

	sessionRunning := sessionDef.ToSession().InitChannel()

	appState.UsersSettings[chatId].SessionRunning = sessionRunning

	return sessionRunning
}

func GetUserSessionRunning(appState *domain.AppState, chatId domain.ChatID, senderId domain.ChatID) *domain.Session {
	defaultUserSettingsIfNeeded(appState, chatId)

	var sessionRunning *domain.Session

	if appState.UsersSettings[chatId].SessionRunning == nil {
		sessionDef := GetUserSessionFromSettings(appState, chatId, senderId)

		sessionDef.PomodoroDuration = sessionDef.PomodoroDurationSet
		sessionDef.SprintDuration = sessionDef.SprintDurationSet
		sessionDef.RestDuration = sessionDef.RestDurationSet

		sessionDef.IsPaused = true

		sessionRunning = sessionDef.ToSession().InitChannel()

		appState.UsersSettings[chatId].SessionRunning = sessionRunning
		/*
			appState.UsersSettings[chatId].SessionRunning = new(domain.Session).InitChannel()

			sessionRunning = appState.UsersSettings[chatId].SessionRunning

			sessionRunning.pomodoroDurationSet = sessionDef.pomodoroDurationSet
			sessionRunning.sprintDurationSet = sessionDef.sprintDurationSet
			sessionRunning.restDurationSet = sessionDef.restDurationSet

			sessionRunning.Data.PomodoroDuration = sessionDef.pomodoroDurationSet
			sessionRunning.Data.SprintDuration = sessionDef.sprintDurationSet
			sessionRunning.Data.RestDuration = sessionDef.restDurationSet

			sessionRunning.Data.IsPaused = true*/
	} else {
		sessionRunning = appState.UsersSettings[chatId].SessionRunning

		if sessionRunning.ActionsChannel == nil {
			sessionRunning = sessionRunning.InitChannel()
		}
	}
	return sessionRunning
}
