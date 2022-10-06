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

	GetActiveChatSettings() ([]utils.Pair[ChatID, *Settings], error)
}

type AppState struct {
	PersistenceManager PersistenceManager

	UsersSettings     map[ChatID]*Settings
	UsersSettingsLock sync.RWMutex
}
