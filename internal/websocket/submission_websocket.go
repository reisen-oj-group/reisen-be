package websocket

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reisen-be/internal/model"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type client struct {
	conn      *websocket.Conn
	mu        sync.Mutex
	closeChan chan struct{}
}

type broadcastMessage struct {
	submissionId model.SubmissionId
	message      model.Submission
}

type SubmissionWs struct {
	clients         map[model.SubmissionId]map[*client]bool
	clientsMux      sync.RWMutex
	broadcastChan   chan broadcastMessage
	lastMessages    map[model.SubmissionId]model.Submission
	lastMessagesMux sync.Mutex
	throttle        time.Duration
}

func NewSubmissionWs(throttle time.Duration) *SubmissionWs {
	ws := &SubmissionWs{
		clients:       make(map[model.SubmissionId]map[*client]bool),
		broadcastChan: make(chan broadcastMessage, 100),
		lastMessages:  make(map[model.SubmissionId]model.Submission),
		throttle:      throttle,
	}

	go ws.broadcastWorker()
	return ws
}

func (wm *SubmissionWs) HandleConnection(w http.ResponseWriter, r *http.Request, submissionId model.SubmissionId) error {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return fmt.Errorf("failed to upgrade connection: %v", err)
	}

	c := &client{
		conn:      conn,
		closeChan: make(chan struct{}),
	}

	wm.clientsMux.Lock()
	if _, ok := wm.clients[submissionId]; !ok {
		wm.clients[submissionId] = make(map[*client]bool)
	}
	wm.clients[submissionId][c] = true
	wm.clientsMux.Unlock()

	// 立即发送最后一条消息（如果有）
	wm.lastMessagesMux.Lock()
	if lastMsg, ok := wm.lastMessages[submissionId]; ok {
		if msg, err := json.Marshal(lastMsg); err == nil {
			c.mu.Lock()
			conn.WriteMessage(websocket.TextMessage, msg)
			c.mu.Unlock()
		}
	}
	wm.lastMessagesMux.Unlock()

	// 保持连接
	for {
		select {
		case <-c.closeChan:
			return nil
		default:
			if _, _, err := conn.NextReader(); err != nil {
				wm.clientsMux.Lock()
				delete(wm.clients[submissionId], c)
				if len(wm.clients[submissionId]) == 0 {
					delete(wm.clients, submissionId)
				}
				wm.clientsMux.Unlock()
				close(c.closeChan)
				return nil
			}
		}
	}
}

func (wm *SubmissionWs) Broadcast(submissionId model.SubmissionId, message model.Submission) {
	// 更新最后一条消息缓存
	wm.lastMessagesMux.Lock()
	wm.lastMessages[submissionId] = message
	wm.lastMessagesMux.Unlock()

	// 发送到广播通道
	wm.broadcastChan <- broadcastMessage{
		submissionId: submissionId,
		message:      message,
	}
}

func (wm *SubmissionWs) broadcastWorker() {
	ticker := time.NewTicker(wm.throttle)
	defer ticker.Stop()

	pending := make(map[model.SubmissionId]struct{})
	var pendingMux sync.Mutex

	for {
		select {
		case msg := <-wm.broadcastChan:
			pendingMux.Lock()
			pending[msg.submissionId] = struct{}{}
			pendingMux.Unlock()

		case <-ticker.C:

			pendingMux.Lock()
			submissionIds := make([]model.SubmissionId, 0, len(pending))
			for id := range pending {
				submissionIds = append(submissionIds, id)
			}
			pending = make(map[model.SubmissionId]struct{})
			pendingMux.Unlock()

			for _, id := range submissionIds {
				wm.lastMessagesMux.Lock()
				message, ok := wm.lastMessages[id]
				wm.lastMessagesMux.Unlock()

				if !ok {
					continue
				}

				wm.clientsMux.RLock()
				clients, ok := wm.clients[id]
				if !ok {
					wm.clientsMux.RUnlock()
					continue
				}

				msg, err := json.Marshal(message)
				if err != nil {
					wm.clientsMux.RUnlock()
					continue
				}

				for c := range clients {
					go func(c *client) {
						select {
						case <-c.closeChan:
							return
						default:
							c.mu.Lock()
							defer c.mu.Unlock()

							if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
								c.conn.Close()
								wm.clientsMux.Lock()
								delete(clients, c)
								if len(clients) == 0 {
									delete(wm.clients, id)
								}
								wm.clientsMux.Unlock()
							}
						}
					}(c)
				}
				wm.clientsMux.RUnlock()
			}
		}
	}
}