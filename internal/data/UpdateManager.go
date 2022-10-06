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
