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

package persistence

import (
	"GoforPomodoro/internal/domain"
	"GoforPomodoro/internal/utils"
	"database/sql"
	"encoding/json"
	"log"
	_ "modernc.org/sqlite"
	"os"
	"sync"
	"time"
)

type SqliteManager struct {
	db *sql.DB

	dbLock sync.RWMutex

	// getChatSettingsItem 1 parameter (chat_id)
	getChatSettingsItem *sql.Stmt

	getActiveChatsSettings *sql.Stmt

	// upsertChatSettingsItem all parameters (chat_id, ...)
	upsertChatSettingsItem *sql.Stmt

	// deleteChatSettingsItem 1 parameter (chat_id)
	deleteChatSettingsItem *sql.Stmt
}

var _ Manager = &SqliteManager{}

// var _ Manager = SqliteManager{}

func (m *SqliteManager) OpenDatabase(dataSourceName string) error {
	if _, err := os.Stat(dataSourceName); err != nil {
		// file does not exist or is not available.
		return err
	}

	db, err := sql.Open("sqlite", dataSourceName)

	if err != nil {
		log.Println("[SqliteManager] ERROR AT OPENING DATABASE")
	} else {
		m.db = db
		m.InitializePreparedStatements()
	}

	return err
}

func (m *SqliteManager) InitializePreparedStatements() {
	var err error

	m.getChatSettingsItem, err = m.db.Prepare(`
		SELECT * 
		FROM chat_settings
		WHERE chat_id = ?`)
	if err != nil {
		log.Printf("[SQLITE MANAGER] ERROR IN PREPARING STATEMENTS (SELECT)! (%s)\n", err.Error())
		panic(err)
	}

	m.getActiveChatsSettings, err = m.db.Prepare(`
		SELECT *
		FROM chat_settings
		WHERE active = true`)

	m.upsertChatSettingsItem, err = m.db.Prepare(`
		INSERT INTO chat_settings 
		    (chat_id,                       
			default_sprint_duration_set,   
			default_pomodoro_duration_set, 
			default_rest_duration_set,     
			running_sprint_duration_set,   
			running_pomodoro_duration_set, 
			running_rest_duration_set,     
			running_sprint_duration,       
			running_pomodoro_duration,     
			running_rest_duration,         
			running_end_next_sprint_ts,    
			running_end_next_rest_ts,      
			running_is_cancel,             
			running_is_paused,             
			running_is_rest,               
			running_is_finished,           
			autorun,                       
			is_group,                      
			subscribers,                   
			active)
			VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
			ON CONFLICT (chat_id) DO UPDATE SET
			default_sprint_duration_set = ?,   
			default_pomodoro_duration_set = ?, 
			default_rest_duration_set = ?,     
			running_sprint_duration_set = ?,   
			running_pomodoro_duration_set = ?, 
			running_rest_duration_set = ?,     
			running_sprint_duration = ?,       
			running_pomodoro_duration = ?,     
			running_rest_duration = ?,         
			running_end_next_sprint_ts = ?,    
			running_end_next_rest_ts = ?,      
			running_is_cancel = ?,             
			running_is_paused = ?,             
			running_is_rest = ?,               
			running_is_finished = ?,           
			autorun = ?,                       
			is_group = ?,                      
			subscribers = ?,                   
			active = ?
		WHERE chat_id = ?
	`)
	if err != nil {
		log.Printf("[SQLITE MANAGER] ERROR IN PREPARING STATEMENTS (INSERT)! (%s)\n", err.Error())
		panic(err)
	}
	/*
			m.updateChatSettingsItem, err = m.db.Prepare(`
				UPDATE chat_settings SET
		            chat_id = ?,
					default_sprint_duration_set = ?,
					default_pomodoro_duration_set = ?,
					default_rest_duration_set = ?,
					running_sprint_duration_set = ?,
					running_pomodoro_duration_set = ?,
					running_rest_duration_set = ?,
					running_sprint_duration = ?,
					running_pomodoro_duration = ?,
					running_rest_duration = ?,
					running_end_next_sprint_ts = ?,
					running_end_next_rest_ts = ?,
					running_is_cancel = ?,
					running_is_paused = ?,
					running_is_rest = ?,
					running_is_finished = ?,
					autorun = ?,
					is_group = ?,
					subscribers = ?,
					active = ?
				WHERE chat_id = ?;`)
			if err != nil {
				log.Printf("[SQLITE MANAGER] ERROR IN PREPARING STATEMENTS (UPDATE)! (%s)\n", err.Error())
				panic(err)
			}
	*/
	m.deleteChatSettingsItem, err = m.db.Prepare(`
		DELETE FROM chat_settings 
		WHERE chat_id = ?`)
	if err != nil {
		log.Printf("[SqliteManager] ERROR IN PREPARING STATEMENTS (DELETE)! (%s)\n", err.Error())
		panic(err)
	}
}

type Scannable interface {
	Err() error
	Scan(dest ...any) error
}

func (m *SqliteManager) getChatSettings(chatId *domain.ChatID, row Scannable) (*domain.Settings, error) {
	if row.Err() != nil {
		log.Printf("[SqliteManager] ERROR AT RETRIEVING CHAT ID (%v), error: %v\n", chatId, row.Err())
		return nil, row.Err()
	}

	autorun := false
	isGroup := false

	var subscribers []domain.ChatID
	var subscribersText string
	var active bool

	defaultS := domain.SessionDefaultData{}

	runningS := domain.SessionInitData{}

	var endNextSprintTimestamp *time.Time
	var endNextRestTimestamp *time.Time

	var _chatId domain.ChatID
	scanErr := row.Scan(
		&_chatId,
		&defaultS.SprintDurationSet,
		&defaultS.PomodoroDurationSet,
		&defaultS.RestDurationSet,

		&runningS.SprintDurationSet,
		&runningS.PomodoroDurationSet,
		&runningS.RestDurationSet,

		&runningS.SprintDuration,
		&runningS.PomodoroDuration,
		&runningS.RestDuration,

		&endNextSprintTimestamp,
		&endNextRestTimestamp,

		&runningS.IsCancel,
		&runningS.IsPaused,
		&runningS.IsRest,
		&runningS.IsFinished,
		&autorun,
		&isGroup,
		&subscribersText,
		&active,
	)

	if *chatId == 0 {
		*chatId = _chatId
	} else if *chatId != _chatId {
		log.Println("[SqliteManager] This condition should have never happened.")
	}

	if endNextSprintTimestamp != nil {
		runningS.EndNextSprintTimestamp = *endNextSprintTimestamp
	}

	if endNextRestTimestamp != nil {
		runningS.EndNextRestTimestamp = *endNextRestTimestamp
	}

	if scanErr != nil {
		log.Printf("[SqliteManager] ERROR IN SCANNING (%v)\n", scanErr.Error())

		return nil, scanErr
	}

	if subscribersText != "" {
		jsonErr := json.Unmarshal([]byte(subscribersText), &subscribers)
		if jsonErr != nil {

			log.Printf("[SqliteManager] ERROR AT DECODING JSON FROM (%v)\n", subscribersText)

			return nil, jsonErr
		}
	}

	settings := &domain.Settings{
		SessionDefault: defaultS,
		SessionRunning: runningS.ToSession(),
		Autorun:        autorun,
		IsGroup:        isGroup,
		Subscribers:    subscribers,
	}
	return settings, nil
}

func (m *SqliteManager) GetChatSettings(chatId domain.ChatID) (*domain.Settings, error) {
	row := m.getChatSettingsItem.QueryRow(chatId)

	m.dbLock.RLock()
	defer m.dbLock.RUnlock()

	return m.getChatSettings(&chatId, row)
}

func (m *SqliteManager) StoreChatSettings(chatId domain.ChatID, settings *domain.Settings) error {
	if chatId == 0 {
		return nil
	}

	m.dbLock.Lock()
	defer m.dbLock.Unlock()

	sessionRunning := settings.SessionRunning
	if sessionRunning == nil {
		sessionRunning = new(domain.Session)
	}

	defaultSprintDurationSet := settings.SessionDefault.SprintDurationSet
	defaultPomodoroDurationSet := settings.SessionDefault.PomodoroDurationSet
	defaultRestDurationSet := settings.SessionDefault.RestDurationSet

	runningSprintDurationSet := sessionRunning.GetSprintDurationSet()
	runningPomodoroDurationSet := sessionRunning.GetPomodoroDurationSet()
	runningRestDurationSet := sessionRunning.GetRestDurationSet()

	runningSprintDuration := sessionRunning.GetSprintDuration()
	runningPomodoroDuration := sessionRunning.GetPomodoroDuration()
	runningRestDuration := sessionRunning.GetRestDuration()

	endNextSprintTs := sessionRunning.EndNextSprintTimestamp()
	endNextRestTs := sessionRunning.EndNextRestTimestamp()

	runningIsCancel := sessionRunning.IsCanceled()
	runningIsPaused := sessionRunning.IsPaused()
	runningIsRest := sessionRunning.IsRest()
	runningIsFinished := sessionRunning.IsFinished()
	autorun := settings.Autorun
	isGroup := settings.IsGroup
	subscribers, errM := json.Marshal(settings.Subscribers)
	if errM != nil {
		subscribers = nil
		log.Printf("[SqliteManager] ERROR AT ENCODING JSON FROM (%v)\n", settings.Subscribers)
	}
	active := sessionRunning.State() == "Running"

	_, err := m.upsertChatSettingsItem.Exec(chatId,
		defaultSprintDurationSet,
		defaultPomodoroDurationSet,
		defaultRestDurationSet,
		runningSprintDurationSet,
		runningPomodoroDurationSet,
		runningRestDurationSet,
		runningSprintDuration,
		runningPomodoroDuration,
		runningRestDuration,
		endNextSprintTs,
		endNextRestTs,
		runningIsCancel,
		runningIsPaused,
		runningIsRest,
		runningIsFinished,
		autorun,
		isGroup,
		subscribers,
		active,
		defaultSprintDurationSet,
		defaultPomodoroDurationSet,
		defaultRestDurationSet,
		runningSprintDurationSet,
		runningPomodoroDurationSet,
		runningRestDurationSet,
		runningSprintDuration,
		runningPomodoroDuration,
		runningRestDuration,
		endNextSprintTs,
		endNextRestTs,
		runningIsCancel,
		runningIsPaused,
		runningIsRest,
		runningIsFinished,
		autorun,
		isGroup,
		subscribers,
		active,
		chatId,
	)

	if err != nil {
		log.Printf("[SqliteManager] ERROR AT STORING RECORD! (%v)\n", err.Error())
	}

	return err
}

func (m *SqliteManager) DeleteChatSettings(chatId domain.ChatID) error {
	m.dbLock.Lock()
	defer m.dbLock.Unlock()

	_, err := m.deleteChatSettingsItem.Exec(chatId)

	return err
}

func (m *SqliteManager) GetActiveChatSettings() ([]utils.Pair[domain.ChatID, *domain.Settings], error) {
	m.dbLock.RLock()

	rows, err := m.getActiveChatsSettings.Query()
	if err != nil {
		return nil, err
	}

	defer func() {
		err := rows.Close()
		defer m.dbLock.RUnlock()
		if err != nil {
			log.Printf("[GetActiveChatSettings] err at Close(): %v\n", err.Error())
		}
	}()

	var pairs []utils.Pair[domain.ChatID, *domain.Settings]

	for rows.Next() {
		var chatId domain.ChatID

		settings, scanErr := m.getChatSettings(&chatId, rows)

		if scanErr != nil {
			log.Println("[GetActiveChatSettings] internal scan error.")
			continue
		}

		newPair := utils.Pair[domain.ChatID, *domain.Settings]{
			First:  chatId,
			Second: settings,
		}
		pairs = append(pairs, newPair)
	}
	return pairs, nil
}
