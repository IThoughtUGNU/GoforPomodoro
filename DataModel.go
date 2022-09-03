package main

func LoadAppState() (*AppState, error) {
	// temporary
	appState := new(AppState)
	appState.usersSettings = make(map[UserID]*Settings)
	return appState, nil
}

func defaultUserSettingsIfNeeded(appState *AppState, userId UserID) {
	if appState.usersSettings[userId] == nil {
		appState.usersSettings[userId] = new(Settings)
	}
}

func CleanUserSettings(appState *AppState, userId UserID) {
	appState.usersSettings[userId] = new(Settings)
}

func UpdateUserSession(appState *AppState, userId UserID, session Session) {
	defaultUserSettingsIfNeeded(appState, userId)

	settings := appState.usersSettings[userId]
	settings.sessionDefault = session
}

func GetUserSessionFromSettings(appState *AppState, userId UserID) *Session {
	defaultUserSettingsIfNeeded(appState, userId)

	session := &appState.usersSettings[userId].sessionDefault
	session.isPaused = true
	return session
}

func GetNewUserSessionRunning(appState *AppState, userId UserID) *Session {
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

func GetUserSessionRunning(appState *AppState, userId UserID) *Session {
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
