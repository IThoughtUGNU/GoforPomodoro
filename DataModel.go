package main

import "log"

func LoadAppState() (*AppState, error) {
	// temporary
	appState := new(AppState)
	appState.usersSettings = make(map[ChatID]*Settings)
	return appState, nil
}

func defaultUserSettingsIfNeeded(appState *AppState, chatId ChatID) {
	if appState.usersSettings[chatId] == nil {
		chatSettings := new(Settings)
		chatSettings.autorun = true
		appState.usersSettings[chatId] = chatSettings
	}
}

func AdjustChatType(appState *AppState, chatId ChatID, senderId ChatID, isGroup bool) {
	defaultUserSettingsIfNeeded(appState, chatId)
	appState.usersSettings[chatId].isGroup = isGroup
}

func IsGroup(appState *AppState, chatId ChatID) bool {
	defaultUserSettingsIfNeeded(appState, chatId)

	return appState.usersSettings[chatId].isGroup
}

func GetSubscribers(appState *AppState, chatId ChatID) []ChatID {
	defaultUserSettingsIfNeeded(appState, chatId)

	return appState.usersSettings[chatId].subscribers
}

func SubscribeUserInGroup(appState *AppState, chatId ChatID, senderId ChatID) error {
	defaultUserSettingsIfNeeded(appState, chatId)

	if chatId == senderId {
		return SubscriptionError{}
	}

	settings := appState.usersSettings[chatId]
	subscribers := (*settings).subscribers
	if !Contains(subscribers, senderId) {
		(*settings).subscribers = append(subscribers, senderId)
	} else {
		return AlreadySubscribed{}
	}
	return nil
}

func UnsubscribeUser(appState *AppState, chatId ChatID, senderId ChatID) error {
	defaultUserSettingsIfNeeded(appState, chatId)

	if chatId == senderId {
		return SubscriptionError{}
	}

	settings := appState.usersSettings[chatId]
	subscribers := (*settings).subscribers
	if Contains(subscribers, senderId) {
		newS, err := AfterRemoveEl(subscribers, senderId)
		if err != nil {
			log.Printf("[UnsubscribeUser] Error while removing %d\n", senderId)
			return OperationError{}
		}
		(*settings).subscribers = newS
	} else {
		log.Printf("[UnsubscribeUser] %d was not subscribed.", senderId)
		return AlreadyUnsubscribed{}
	}
	return nil
}

func CleanUserSettings(appState *AppState, chatId ChatID, senderId ChatID) {
	appState.usersSettings[chatId] = nil // new(Settings)
	defaultUserSettingsIfNeeded(appState, chatId)
}

func SetUserAutorun(appState *AppState, chatId ChatID, senderId ChatID, autorun bool) {
	defaultUserSettingsIfNeeded(appState, chatId)

	appState.usersSettings[chatId].autorun = autorun
}

func GetUserAutorun(appState *AppState, chatId ChatID, senderId ChatID) bool {
	defaultUserSettingsIfNeeded(appState, chatId)

	return appState.usersSettings[chatId].autorun
}

func UpdateUserSession(appState *AppState, chatId ChatID, senderId ChatID, session Session) {
	defaultUserSettingsIfNeeded(appState, chatId)

	settings := appState.usersSettings[chatId]
	settings.sessionDefault = session
}

func GetUserSessionFromSettings(appState *AppState, chatId ChatID, senderId ChatID) *Session {
	defaultUserSettingsIfNeeded(appState, chatId)

	session := &appState.usersSettings[chatId].sessionDefault
	session.Data.isPaused = true
	return session
}

func GetNewUserSessionRunning(appState *AppState, chatId ChatID, senderId ChatID) *Session {
	defaultUserSettingsIfNeeded(appState, chatId)

	appState.usersSettings[chatId].sessionRunning = new(Session).Init()

	sessionDef := GetUserSessionFromSettings(appState, chatId, senderId)

	sessionRunning := appState.usersSettings[chatId].sessionRunning
	sessionRunning.PomodoroDurationSet = sessionDef.PomodoroDurationSet
	sessionRunning.SprintDurationSet = sessionDef.SprintDurationSet
	sessionRunning.RestDurationSet = sessionDef.RestDurationSet

	sessionRunning.Data.PomodoroDuration = sessionDef.PomodoroDurationSet
	sessionRunning.Data.SprintDuration = sessionDef.SprintDurationSet
	sessionRunning.Data.RestDuration = sessionDef.RestDurationSet
	sessionRunning.Data.isPaused = true

	return sessionRunning
}

func GetUserSessionRunning(appState *AppState, chatId ChatID, senderId ChatID) *Session {
	defaultUserSettingsIfNeeded(appState, chatId)

	var sessionRunning *Session

	if appState.usersSettings[chatId].sessionRunning == nil {
		appState.usersSettings[chatId].sessionRunning = new(Session).Init()

		sessionDef := GetUserSessionFromSettings(appState, chatId, senderId)

		sessionRunning = appState.usersSettings[chatId].sessionRunning

		sessionRunning.PomodoroDurationSet = sessionDef.PomodoroDurationSet
		sessionRunning.SprintDurationSet = sessionDef.SprintDurationSet
		sessionRunning.RestDurationSet = sessionDef.RestDurationSet

		sessionRunning.Data.PomodoroDuration = sessionDef.PomodoroDurationSet
		sessionRunning.Data.SprintDuration = sessionDef.SprintDurationSet
		sessionRunning.Data.RestDuration = sessionDef.RestDurationSet

		sessionRunning.Data.isPaused = true
	} else {
		sessionRunning = appState.usersSettings[chatId].sessionRunning
	}

	return sessionRunning
}
