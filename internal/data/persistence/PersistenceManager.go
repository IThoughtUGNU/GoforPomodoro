package persistence

import (
	"GoforPomodoro/internal/domain"
	"GoforPomodoro/internal/utils"
)

// Manager interface for types that want to manage persistence.
// The first three methods
//
// GetChatSettings
// StoreChatSettings
// DeleteChatSettings
//
// are classical operations of key-value stores and alike (get/update/delete).
//
// Then GetActiveChatSettings is defined for a (possibly efficient) retrieval
// of the chats that have/had a session running.
//
// Since the store is as of now thought to be key-value based, the user of this
// interface is not expected to perform complex queries, but just the minimum
// that is needed for correctly running the bot.
type Manager interface {
	// GetChatSettings get the settings for the provided chat ID
	GetChatSettings(domain.ChatID) (*domain.Settings, error)

	StoreChatSettings(id domain.ChatID, settings *domain.Settings) error
	DeleteChatSettings(id domain.ChatID) error

	GetActiveChatSettings() ([]utils.Pair[domain.ChatID, *domain.Settings], error)
}
