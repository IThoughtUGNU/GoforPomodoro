package domain

import (
	"GoforPomodoro/internal/utils"
	"sync"
)

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

type PersistenceManager interface {
	GetChatSettings(ChatID) (*Settings, error)

	StoreChatSettings(id ChatID, settings *Settings) error
	DeleteChatSettings(id ChatID) error

	GetActiveChatSettings() []utils.Pair[ChatID, *Settings]
}

type AppState struct {
	PersistenceManager

	UsersSettings     map[ChatID]*Settings
	UsersSettingsLock sync.RWMutex
}
