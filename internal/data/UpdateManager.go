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

/*
import (
	"GoforPomodoro/internal/domain"
	"log"
	"time"
)

type DispatchUpdate struct {
	ChatId domain.ChatID
}

type UpdateManager struct {
	AppState *domain.AppState

	updateChannel chan DispatchUpdate
}

func (m *UpdateManager) WriteChannel() chan<- DispatchUpdate {
	return m.updateChannel
}

func (m *UpdateManager) StartLoop() {
mainLoop:
	for {
		select {
		case update, ok := <-m.updateChannel:
			if ok {
				log.Println("[UpdateManager] update received, calling...")
				UpdateUserSessionRunning(m.AppState, update.ChatId)
			} else {
				log.Println("updateChannel NOT ok.")
				break mainLoop
			}
		default:
			time.Sleep(500 * time.Millisecond)
		}
	}
	defer func() {
		close(m.updateChannel)
		m.updateChannel = nil
	}()
}
*/
