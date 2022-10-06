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

func LoadAppState(persistenceManager persistence.Manager) (*domain.AppState, error) {
	// temporary
	appState := new(domain.AppState)

	appState.PersistenceManager = persistenceManager

	appState.UsersSettingsLock.Lock()
	defer appState.UsersSettingsLock.Unlock()
	appState.UsersSettings = make(map[domain.ChatID]*domain.Settings)

	return appState, nil
}

func defaultUserSettingsIfNeeded(appState *domain.AppState, chatId domain.ChatID) {
	appState.UsersSettingsLock.Lock()
	defer appState.UsersSettingsLock.Unlock()

	if appState.UsersSettings[chatId] == nil {
		// Check if there is in the database, otherwise we create new settings in-place
		if appState.PersistenceManager == nil {
			chatSettings := new(domain.Settings)
			chatSettings.Autorun = true
			appState.UsersSettings[chatId] = chatSettings
		} else {
			chatSettings, err := appState.PersistenceManager.GetChatSettings(chatId)

			if err != nil {
				chatSettings := new(domain.Settings)
				chatSettings.Autorun = true
				appState.UsersSettings[chatId] = chatSettings
			} else { // err == nil

				// Load settings
				appState.UsersSettings[chatId] = chatSettings
			}
		}
	}
}

func AdjustChatType(appState *domain.AppState, chatId domain.ChatID, senderId domain.ChatID, isGroup bool) {
	defaultUserSettingsIfNeeded(appState, chatId)
	appState.UsersSettingsLock.Lock()
	appState.UsersSettings[chatId].IsGroup = isGroup
	appState.UsersSettingsLock.Unlock()
}

func IsGroup(appState *domain.AppState, chatId domain.ChatID) bool {
	defaultUserSettingsIfNeeded(appState, chatId)

	appState.UsersSettingsLock.RLock()
	defer appState.UsersSettingsLock.RUnlock()

	return appState.UsersSettings[chatId].IsGroup
}

func GetSubscribers(appState *domain.AppState, chatId domain.ChatID) []domain.ChatID {
	defaultUserSettingsIfNeeded(appState, chatId)

	appState.UsersSettingsLock.RLock()
	defer appState.UsersSettingsLock.RUnlock()

	return appState.UsersSettings[chatId].Subscribers
}

func SubscribeUserInGroup(appState *domain.AppState, chatId domain.ChatID, senderId domain.ChatID) error {
	defaultUserSettingsIfNeeded(appState, chatId)

	if chatId == senderId {
		return domain.SubscriptionError{}
	}

	appState.UsersSettingsLock.RLock()
	settings := appState.UsersSettings[chatId]
	appState.UsersSettingsLock.RUnlock()

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

	appState.UsersSettingsLock.RLock()
	settings := appState.UsersSettings[chatId]
	appState.UsersSettingsLock.RUnlock()

	subscribers := (*settings).Subscribers
	if utils.Contains(subscribers, senderId) {
		newS, err := utils.AfterRemoveEl(subscribers, senderId)
		if err != nil {
			log.Printf("[UnsubscribeUser] Error while removing %d\n", senderId)
			return domain.OperationError{}
		}
		(*settings).Subscribers = newS
	} else {
		log.Printf("[UnsubscribeUser] %d was not subscribed.", senderId)
		return domain.AlreadyUnsubscribed{}
	}
	return nil
}

func CleanUserSettings(appState *domain.AppState, chatId domain.ChatID, senderId domain.ChatID) {
	appState.UsersSettingsLock.Lock()
	appState.UsersSettings[chatId] = nil
	appState.UsersSettingsLock.Unlock()

	if appState.PersistenceManager != nil {
		err := appState.PersistenceManager.DeleteChatSettings(chatId)
		if err != nil {
			log.Printf("[DataModel::CleanUserSettings] error in deleting. (%v)\n", err.Error())
		}
	}
	defaultUserSettingsIfNeeded(appState, chatId)
}

func SetUserAutorun(appState *domain.AppState, chatId domain.ChatID, senderId domain.ChatID, autorun bool) {
	defaultUserSettingsIfNeeded(appState, chatId)

	appState.UsersSettingsLock.RLock()
	chatSettings := appState.UsersSettings[chatId]
	appState.UsersSettingsLock.RUnlock()

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

	appState.UsersSettingsLock.RLock()
	defer appState.UsersSettingsLock.RUnlock()

	return appState.UsersSettings[chatId].Autorun
}

func UpdateUserSessionRunning(appState *domain.AppState, chatId domain.ChatID) {

	appState.UsersSettingsLock.RLock()
	settings := appState.UsersSettings[chatId]
	appState.UsersSettingsLock.RUnlock()

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

	appState.UsersSettingsLock.RLock()
	settings := appState.UsersSettings[chatId]
	appState.UsersSettingsLock.RUnlock()

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

	appState.UsersSettingsLock.RLock()
	session := &appState.UsersSettings[chatId].SessionDefault
	appState.UsersSettingsLock.RUnlock()
	// session.Data.IsPaused = true

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

	appState.UsersSettingsLock.RLock()
	settings := appState.UsersSettings[chatId]
	appState.UsersSettingsLock.RUnlock()

	settings.SessionRunning = sessionRunning

	return sessionRunning
}

func GetUserSessionRunning(appState *domain.AppState, chatId domain.ChatID, senderId domain.ChatID) *domain.Session {

	// log.Printf("[NO-DB TEST] about to defaultUserSettingsIfNeeded\n")
	defaultUserSettingsIfNeeded(appState, chatId)

	// log.Printf("[NO-DB TEST] defaultUserSettingsIfNeeded done\n")
	appState.UsersSettingsLock.RLock()
	// log.Printf("[NO-DB TEST] RLock acquired\n")

	sessionRunning := appState.UsersSettings[chatId].SessionRunning
	appState.UsersSettingsLock.RUnlock()

	// var sessionRunning *domain.Session

	if sessionRunning == nil {
		sessionDef := GetUserSessionFromSettings(appState, chatId, senderId)

		sessionDef.PomodoroDuration = sessionDef.PomodoroDurationSet
		sessionDef.SprintDuration = sessionDef.SprintDurationSet
		sessionDef.RestDuration = sessionDef.RestDurationSet

		sessionDef.IsPaused = true

		sessionRunning = sessionDef.ToSession().InitChannel()

		appState.UsersSettingsLock.RLock()
		appState.UsersSettings[chatId].SessionRunning = sessionRunning
		appState.UsersSettingsLock.RUnlock()
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
		appState.UsersSettingsLock.RLock()
		sessionRunning = appState.UsersSettings[chatId].SessionRunning
		appState.UsersSettingsLock.RUnlock()

		if sessionRunning.ActionsChannel == nil {
			sessionRunning = sessionRunning.InitChannel()
		}
	}
	return sessionRunning
}
