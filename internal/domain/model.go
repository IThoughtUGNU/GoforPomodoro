package domain

type AppSettings struct {
	ApiToken string
	BotName  string
}

type ChatID int64

type Settings struct {
	SessionDefault Session
	SessionRunning *Session
	Autorun        bool
	IsGroup        bool
	Subscribers    []ChatID
}

type AppState struct {
	UsersSettings map[ChatID]*Settings
}
