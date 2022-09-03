package main

func LoadAppState() (*AppState, error) {
	// temporary
	appState := new(AppState)
	appState.usersSettings = make(map[ChatID]*Settings)
	return appState, nil
}

func defaultUserSettingsIfNeeded(appState *AppState, userId ChatID) {
	if appState.usersSettings[userId] == nil {
		appState.usersSettings[userId] = new(Settings)
		appState.usersSettings[userId].autorun = true
	}
}

func CleanUserSettings(appState *AppState, userId ChatID) {
	appState.usersSettings[userId] = nil // new(Settings)
	defaultUserSettingsIfNeeded(appState, userId)
}

func SetUserAutorun(appState *AppState, chatId ChatID, autorun bool) {
	defaultUserSettingsIfNeeded(appState, chatId)

	appState.usersSettings[chatId].autorun = autorun
}

func GetUserAutorun(appState *AppState, chatId ChatID) bool {
	defaultUserSettingsIfNeeded(appState, chatId)

	return appState.usersSettings[chatId].autorun
}

func UpdateUserSession(appState *AppState, userId ChatID, session Session) {
	defaultUserSettingsIfNeeded(appState, userId)

	settings := appState.usersSettings[userId]
	settings.sessionDefault = session
}

func GetUserSessionFromSettings(appState *AppState, userId ChatID) *Session {
	defaultUserSettingsIfNeeded(appState, userId)

	session := &appState.usersSettings[userId].sessionDefault
	session.isPaused = true
	return session
}

func GetNewUserSessionRunning(appState *AppState, userId ChatID) *Session {
	defaultUserSettingsIfNeeded(appState, userId)

	appState.usersSettings[userId].sessionRunning = new(Session)

	sessionDef := GetUserSessionFromSettings(appState, userId)

	sessionRunning := appState.usersSettings[userId].sessionRunning
	sessionRunning.PomodoroDurationSet = sessionDef.PomodoroDurationSet
	sessionRunning.SprintDurationSet = sessionDef.SprintDurationSet
	sessionRunning.RestDurationSet = sessionDef.RestDurationSet

	sessionRunning.PomodoroDuration = sessionDef.PomodoroDuration
	sessionRunning.SprintDuration = sessionDef.SprintDuration
	sessionRunning.RestDuration = sessionDef.RestDuration
	sessionRunning.isPaused = true

	return sessionRunning
}

func GetUserSessionRunning(appState *AppState, userId ChatID) *Session {
	defaultUserSettingsIfNeeded(appState, userId)

	var sessionRunning *Session

	if appState.usersSettings[userId].sessionRunning == nil {
		appState.usersSettings[userId].sessionRunning = new(Session)

		sessionDef := GetUserSessionFromSettings(appState, userId)

		sessionRunning = appState.usersSettings[userId].sessionRunning

		sessionRunning.PomodoroDurationSet = sessionDef.PomodoroDurationSet
		sessionRunning.SprintDurationSet = sessionDef.SprintDurationSet
		sessionRunning.RestDurationSet = sessionDef.RestDurationSet

		sessionRunning.PomodoroDuration = sessionDef.PomodoroDuration
		sessionRunning.SprintDuration = sessionDef.SprintDuration
		sessionRunning.RestDuration = sessionDef.RestDuration

		sessionRunning.isPaused = true
	} else {
		sessionRunning = appState.usersSettings[userId].sessionRunning
	}

	return sessionRunning
}
