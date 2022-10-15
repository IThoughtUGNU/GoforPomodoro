// This file is part of GoforPomodoro.
//
// GoforPomodoro is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// GoforPomodoro is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with GoforPomodoro.  If not, see <http://www.gnu.org/licenses/>.

package domain

import (
	"GoforPomodoro/internal/utils"
	"sync"
)

type PrivacySettingsVersion int

type PrivacySettingsType int

const (
	AcceptedEssential PrivacySettingsType = 1 << iota // 1
	AcceptedAll                                       // 2
)

func (privacy PrivacySettingsType) IsZero() bool {
	return privacy == 0
}

func (privacy PrivacySettingsType) HasAcceptedEssential() bool {
	return privacy&AcceptedEssential != 0 ||
		privacy&AcceptedAll != 0
}

func (privacy PrivacySettingsType) HasAcceptedAll() bool {
	return privacy&AcceptedEssential != 0 ||
		privacy&AcceptedAll != 0
}

type AppSettings struct {
	ApiToken             string
	BotName              string
	DebugMode            bool
	AdminIds             []ChatID
	ListenAddressPrivate string
	ListenPortPrivate    int
}

type AppVariables struct {
	PrivacyPolicy1 string
	OpenSource1    string
	PrivacySettingsVersion
	PrivacyPolicyEnabled bool
}

func (v AppVariables) IsPrivacyPolicyVersionUpdated(version PrivacySettingsVersion) bool {
	return version == v.PrivacySettingsVersion
}

type ChatID int64

type Settings struct {
	SessionDefault  SessionDefaultData
	SessionRunning  *Session
	Autorun         bool
	IsGroup         bool
	Subscribers     []ChatID
	PrivacySettings PrivacySettingsType
	PrivacySettingsVersion
}

type PersistenceManager interface {
	GetChatSettings(ChatID) (*Settings, error)

	StoreChatSettings(id ChatID, settings *Settings) error
	DeleteChatSettings(id ChatID) error

	GetActiveChatSettings() ([]utils.Pair[ChatID, *Settings], error)

	LockDB()
	UnlockDB()
}

type AppState struct {
	DebugMode bool

	PersistenceManager PersistenceManager

	UsersSettings     map[ChatID]*Settings
	UsersSettingsLock sync.RWMutex
}

func (appState *AppState) ReadSettings(
	chatId ChatID,
) *Settings {
	appState.UsersSettingsLock.RLock()
	defer appState.UsersSettingsLock.RUnlock()

	return appState.UsersSettings[chatId]
}

func (appState *AppState) WriteSettings(
	chatId ChatID,
	settings *Settings,
) {
	appState.UsersSettingsLock.Lock()
	defer appState.UsersSettingsLock.Unlock()

	appState.UsersSettings[chatId] = settings
}

//type DispatchServerAction struct {
//	Shutdown bool
//}
