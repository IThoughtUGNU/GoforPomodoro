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

package data

import (
	"GoforPomodoro/internal/data/persistence"
	"GoforPomodoro/internal/domain"
	"GoforPomodoro/internal/utils"
	"github.com/BurntSushi/toml"
	"log"
)

func PreloadUsersSettings(
	appState *domain.AppState,
	pairs []utils.Pair[domain.ChatID, *domain.Settings],
) {
	appState.UsersSettingsLock.Lock()
	defer appState.UsersSettingsLock.Unlock()

	for _, pair := range pairs {
		chatId := pair.First
		settings := pair.Second

		appState.UsersSettings[chatId] = settings
	}
}

func LoadAppSettings() (*domain.AppSettings, error) {
	settings := new(domain.AppSettings)
	_, err := toml.DecodeFile("appsettings.toml", settings)

	return settings, err
}

func LoadAppState(persistenceManager persistence.Manager, debugMode bool) (*domain.AppState, error) {
	appState := new(domain.AppState)

	appState.DebugMode = debugMode

	appState.PersistenceManager = persistenceManager

	appState.UsersSettingsLock.Lock()
	appState.UsersSettings = make(map[domain.ChatID]*domain.Settings)
	appState.UsersSettingsLock.Unlock()

	return appState, nil
}

func defaultUserSettingsIfNeeded(appState *domain.AppState, chatId domain.ChatID) {
	if appState.ReadSettings(chatId) == nil {
		// Check if there is in the database, otherwise we create new settings in-place
		if appState.PersistenceManager == nil {
			chatSettings := new(domain.Settings)
			chatSettings.Autorun = true
			appState.WriteSettings(chatId, chatSettings)
		} else {
			chatSettings, err := appState.PersistenceManager.GetChatSettings(chatId)

			if err != nil {
				chatSettings := new(domain.Settings)
				chatSettings.Autorun = true
				appState.WriteSettings(chatId, chatSettings)
			} else { // err == nil

				appState.WriteSettings(chatId, chatSettings)
			}
		}
	}
}

func AdjustChatType(appState *domain.AppState, chatId domain.ChatID, senderId domain.ChatID, isGroup bool) {
	defaultUserSettingsIfNeeded(appState, chatId)

	appState.ReadSettings(chatId).IsGroup = isGroup
}

func IsGroup(appState *domain.AppState, chatId domain.ChatID) bool {
	defaultUserSettingsIfNeeded(appState, chatId)

	return appState.ReadSettings(chatId).IsGroup
}

func GetSubscribers(appState *domain.AppState, chatId domain.ChatID) []domain.ChatID {
	defaultUserSettingsIfNeeded(appState, chatId)

	return appState.ReadSettings(chatId).Subscribers
}

func SubscribeUserInGroup(appState *domain.AppState, chatId domain.ChatID, senderId domain.ChatID) error {
	defaultUserSettingsIfNeeded(appState, chatId)

	if chatId == senderId {
		return domain.SubscriptionError{}
	}

	settings := appState.ReadSettings(chatId)

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

	settings := appState.ReadSettings(chatId)

	subscribers := (*settings).Subscribers
	if utils.Contains(subscribers, senderId) {
		newS, err := utils.AfterRemoveEl(subscribers, senderId)
		if err != nil {
			if appState.DebugMode {
				log.Printf("[UnsubscribeUser] Error while removing %d\n", senderId)
			}
			return domain.OperationError{}
		}
		(*settings).Subscribers = newS
	} else {
		if appState.DebugMode {
			log.Printf("[UnsubscribeUser] %d was not subscribed.", senderId)
		}
		return domain.AlreadyUnsubscribed{}
	}
	return nil
}

func CleanUserSettings(appState *domain.AppState, chatId domain.ChatID, senderId domain.ChatID) {
	appState.WriteSettings(chatId, nil)

	if appState.PersistenceManager != nil {
		err := appState.PersistenceManager.DeleteChatSettings(chatId)
		if err != nil {
			log.Printf("[DataModel::CleanUserSettings] error in deleting. (%v)\n", err.Error())
		}
	}
	// defaultUserSettingsIfNeeded(appState, chatId)
}

func SetUserAutorun(appState *domain.AppState, chatId domain.ChatID, senderId domain.ChatID, autorun bool) {
	defaultUserSettingsIfNeeded(appState, chatId)

	chatSettings := appState.ReadSettings(chatId)

	chatSettings.Autorun = autorun

	if appState.PersistenceManager != nil {
		err := appState.PersistenceManager.StoreChatSettings(chatId, chatSettings)
		if err != nil {
			log.Printf("[DataModel::SetUserAutorun] error in storing. (%v)\n", err.Error())
		}
	}
}

func GetUserAutorun(appState *domain.AppState, chatId domain.ChatID, senderId domain.ChatID) bool {
	defaultUserSettingsIfNeeded(appState, chatId)

	return appState.ReadSettings(chatId).Autorun
}

func UpdateUserSessionRunning(appState *domain.AppState, chatId domain.ChatID) {

	settings := appState.ReadSettings(chatId)

	// TODO: Dispatch this call to a goroutine using a channel instead of spawning a go-func
	go func() {
		if appState.PersistenceManager != nil {
			err := appState.PersistenceManager.StoreChatSettings(chatId, settings)
			if err != nil {
				log.Printf("[DataModel::UpdateUserSessionRunning] error in storing. (%v)\n", err.Error())
			}
		}
	}()
}

func UpdateUserSession(appState *domain.AppState, chatId domain.ChatID, senderId domain.ChatID, session domain.Session) {
	defaultUserSettingsIfNeeded(appState, chatId)

	settings := appState.ReadSettings(chatId)

	settings.SessionDefault = session

	if appState.PersistenceManager != nil {
		err := appState.PersistenceManager.StoreChatSettings(chatId, settings)
		if err != nil {
			log.Printf("[DataModel::UpdateUserSession] error in storing. (%v)\n", err.Error())
		}
	}
}

func GetUserSessionFromSettings(appState *domain.AppState, chatId domain.ChatID, senderId domain.ChatID) domain.SessionInitData {
	defaultUserSettingsIfNeeded(appState, chatId)

	session := &appState.ReadSettings(chatId).SessionDefault

	sData := session.ToInitData()
	sData.IsPaused = true

	return sData // this instantiates a new session object
}

func GetNewUserSessionRunning(appState *domain.AppState, chatId domain.ChatID, senderId domain.ChatID) *domain.Session {
	defaultUserSettingsIfNeeded(appState, chatId)

	sessionDef := GetUserSessionFromSettings(appState, chatId, senderId)

	sessionDef.PomodoroDuration = sessionDef.PomodoroDurationSet
	sessionDef.SprintDuration = sessionDef.SprintDurationSet
	sessionDef.RestDuration = sessionDef.RestDurationSet
	sessionDef.IsPaused = true

	sessionRunning := sessionDef.ToSession().InitChannel()

	settings := appState.ReadSettings(chatId)

	settings.SessionRunning = sessionRunning

	return sessionRunning
}

func GetUserSessionRunning(appState *domain.AppState, chatId domain.ChatID, senderId domain.ChatID) *domain.Session {
	defaultUserSettingsIfNeeded(appState, chatId)

	sessionRunning := appState.ReadSettings(chatId).SessionRunning

	// var sessionRunning *domain.Session

	if sessionRunning == nil {
		sessionDef := GetUserSessionFromSettings(appState, chatId, senderId)

		sessionDef.PomodoroDuration = sessionDef.PomodoroDurationSet
		sessionDef.SprintDuration = sessionDef.SprintDurationSet
		sessionDef.RestDuration = sessionDef.RestDurationSet

		sessionDef.IsPaused = true

		sessionRunning = sessionDef.ToSession().InitChannel()

		appState.ReadSettings(chatId).SessionRunning = sessionRunning

		/*
			appState.UsersSettings[chatId].SessionRunning = new(domain.Session).InitChannel()

			sessionRunning = appState.UsersSettings[chatId].SessionRunning

			sessionRunning.pomodoroDurationSet = sessionDef.pomodoroDurationSet
			sessionRunning.sprintDurationSet = sessionDef.sprintDurationSet
			sessionRunning.restDurationSet = sessionDef.restDurationSet

			sessionRunning.Data.PomodoroDuration = sessionDef.pomodoroDurationSet
			sessionRunning.Data.SprintDuration = sessionDef.sprintDurationSet
			sessionRunning.Data.RestDuration = sessionDef.restDurationSet

			sessionRunning.Data.IsPaused = true*/
	} else {
		sessionRunning = appState.ReadSettings(chatId).SessionRunning

		if sessionRunning.ActionsChannel == nil {
			sessionRunning = sessionRunning.InitChannel()
		}
	}
	return sessionRunning
}
