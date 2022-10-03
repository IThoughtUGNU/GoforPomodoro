package main

type AppSettings struct {
	ApiToken string
	BotName  string
}

type ChatID int64

type Settings struct {
	sessionDefault Session
	sessionRunning *Session
	autorun        bool
	isGroup        bool
	subscribers    []ChatID
}

type AppState struct {
	usersSettings map[ChatID]*Settings
}
