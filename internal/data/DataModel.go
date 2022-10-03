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

func GetUserSessionFromSettings(appState *domain.AppState, chatId domain.ChatID, senderId domain.ChatID) *domain.Session {
	defaultUserSettingsIfNeeded(appState, chatId)

	session := &appState.UsersSettings[chatId].SessionDefault
	session.Data.IsPaused = true
	return session
}

func GetNewUserSessionRunning(appState *domain.AppState, chatId domain.ChatID, senderId domain.ChatID) *domain.Session {
	defaultUserSettingsIfNeeded(appState, chatId)

	appState.UsersSettings[chatId].SessionRunning = new(domain.Session).Init()

	sessionDef := GetUserSessionFromSettings(appState, chatId, senderId)

	sessionRunning := appState.UsersSettings[chatId].SessionRunning
	sessionRunning.PomodoroDurationSet = sessionDef.PomodoroDurationSet
	sessionRunning.SprintDurationSet = sessionDef.SprintDurationSet
	sessionRunning.RestDurationSet = sessionDef.RestDurationSet

	sessionRunning.Data.PomodoroDuration = sessionDef.PomodoroDurationSet
	sessionRunning.Data.SprintDuration = sessionDef.SprintDurationSet
	sessionRunning.Data.RestDuration = sessionDef.RestDurationSet
	sessionRunning.Data.IsPaused = true

	return sessionRunning
}

func GetUserSessionRunning(appState *domain.AppState, chatId domain.ChatID, senderId domain.ChatID) *domain.Session {
	defaultUserSettingsIfNeeded(appState, chatId)

	var sessionRunning *domain.Session

	if appState.UsersSettings[chatId].SessionRunning == nil {
		appState.UsersSettings[chatId].SessionRunning = new(domain.Session).Init()

		sessionDef := GetUserSessionFromSettings(appState, chatId, senderId)

		sessionRunning = appState.UsersSettings[chatId].SessionRunning

		sessionRunning.PomodoroDurationSet = sessionDef.PomodoroDurationSet
		sessionRunning.SprintDurationSet = sessionDef.SprintDurationSet
		sessionRunning.RestDurationSet = sessionDef.RestDurationSet

		sessionRunning.Data.PomodoroDuration = sessionDef.PomodoroDurationSet
		sessionRunning.Data.SprintDuration = sessionDef.SprintDurationSet
		sessionRunning.Data.RestDuration = sessionDef.RestDurationSet

		sessionRunning.Data.IsPaused = true
	} else {
		sessionRunning = appState.UsersSettings[chatId].SessionRunning
	}

	return sessionRunning
}
