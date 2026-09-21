package ws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/google/uuid"
	"go-api-starter/internal/platform/logger"
)

var (
	ErrNoConnection = errors.New("wechat hook 未连接")
	ErrTimeout      = errors.New("指令响应超时")
)

// Hub 管理 wechat_hook 的 WebSocket 连接
type Hub struct {
	mu   sync.RWMutex
	conn *websocket.Conn

	// Write serialization — WebSocket connections don't support concurrent writes
	writeMu sync.Mutex

	// pending 存放等待 ack 的请求
	pendingMu sync.Mutex
	pending   map[string]chan *AckData

	// 上行消息处理器
	reviewHandler func(data *ReviewData)

	// 配置
	ackTimeout   time.Duration
	pingInterval time.Duration
}

func NewHub() *Hub {
	return &Hub{
		pending:      make(map[string]chan *AckData),
		ackTimeout:   15 * time.Second,
		pingInterval: 30 * time.Second,
	}
}

// SetReviewHandler 设置审核消息处理回调
func (h *Hub) SetReviewHandler(fn func(data *ReviewData)) {
	h.reviewHandler = fn
}

// IsConnected 是否有活跃连接
func (h *Hub) IsConnected() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.conn != nil
}

// Register 注册新连接（替换旧连接）
func (h *Hub) Register(conn *websocket.Conn) {
	h.mu.Lock()
	old := h.conn
	h.conn = conn
	h.mu.Unlock()

	if old != nil {
		old.CloseNow()
	}

	logger.Log.Info("wechat_hook connected")
	go h.readPump(conn)
	go h.pingLoop(conn)
}

// Send 发送指令并等待 ack
func (h *Hub) Send(msgType string, data interface{}) (*AckData, error) {
	h.mu.RLock()
	conn := h.conn
	h.mu.RUnlock()

	if conn == nil {
		return nil, ErrNoConnection
	}

	id := uuid.New().String()
	logger.Log.Info("websocket command sent", "type", msgType, "id", id[:8])

	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("序列化 data 失败: %w", err)
	}

	msg := Message{
		Type: msgType,
		ID:   id,
		Data: dataBytes,
	}

	// 注册 pending
	ackCh := make(chan *AckData, 1)
	h.pendingMu.Lock()
	h.pending[id] = ackCh
	h.pendingMu.Unlock()

	defer func() {
		h.pendingMu.Lock()
		delete(h.pending, id)
		h.pendingMu.Unlock()
	}()

	// 发送（使用 writeMu 序列化写操作）
	h.writeMu.Lock()
	dataBytes, err = json.Marshal(msg)
	if err == nil {
		writeCtx, cancel := context.WithTimeout(context.Background(), h.ackTimeout)
		err = conn.Write(writeCtx, websocket.MessageText, dataBytes)
		cancel()
	}
	h.writeMu.Unlock()
	if err != nil {
		return nil, fmt.Errorf("发送失败: %w", err)
	}

	// 等待 ack
	select {
	case ack := <-ackCh:
		logger.Log.Info("websocket ack received", "type", msgType, "success", ack.Success)
		return ack, nil
	case <-time.After(h.ackTimeout):
		logger.Log.Warn("websocket command timed out", "type", msgType, "id", id[:8])
		return nil, ErrTimeout
	}
}

// SendNoWait 发送指令不等待 ack（用于 ping 等）
func (h *Hub) SendNoWait(msgType string, data interface{}) error {
	h.mu.RLock()
	conn := h.conn
	h.mu.RUnlock()

	if conn == nil {
		return ErrNoConnection
	}

	id := uuid.New().String()

	dataBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("序列化 data 失败: %w", err)
	}

	msg := Message{
		Type: msgType,
		ID:   id,
		Data: dataBytes,
	}

	h.writeMu.Lock()
	dataBytes, err = json.Marshal(msg)
	if err == nil {
		writeCtx, cancel := context.WithTimeout(context.Background(), h.ackTimeout)
		err = conn.Write(writeCtx, websocket.MessageText, dataBytes)
		cancel()
	}
	h.writeMu.Unlock()
	return err
}

// readPump 读取上行消息
func (h *Hub) readPump(conn *websocket.Conn) {
	defer func() {
		h.mu.Lock()
		if h.conn == conn {
			h.conn = nil
		}
		h.mu.Unlock()
		conn.CloseNow()
		logger.Log.Info("wechat_hook disconnected")
	}()

	for {
		_, data, err := conn.Read(context.Background())
		if err != nil {
			logger.Log.Warn("websocket read ended", "error", err)
			return
		}

		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			logger.Log.Warn("websocket message decode failed", "error", err)
			continue
		}

		switch msg.Type {
		case TypeAck:
			h.handleAck(msg)
		case TypeReview:
			h.handleReview(msg)
		case TypePong:
			// 心跳回复，已通过 SetReadDeadline 处理
		default:
			logger.Log.Warn("unknown websocket message type", "type", msg.Type)
		}
	}
}

// handleAck 处理 ack 消息
func (h *Hub) handleAck(msg Message) {
	var ack AckData
	if err := json.Unmarshal(msg.Data, &ack); err != nil {
		logger.Log.Warn("websocket ack decode failed", "error", err)
		return
	}

	refID := ack.RefID
	if refID == "" {
		refID = msg.ID // 兼容：有些实现可能把 ref_id 放在外层 id
	}

	h.pendingMu.Lock()
	ch, ok := h.pending[refID]
	h.pendingMu.Unlock()

	if ok {
		ch <- &ack
	}
}

// handleReview 处理审核消息
func (h *Hub) handleReview(msg Message) {
	var data ReviewData
	if err := json.Unmarshal(msg.Data, &data); err != nil {
		logger.Log.Warn("websocket review decode failed", "error", err)
		return
	}

	if h.reviewHandler != nil {
		go h.reviewHandler(&data)
	}
}

// pingLoop 定时发送 ping
func (h *Hub) pingLoop(conn *websocket.Conn) {
	ticker := time.NewTicker(h.pingInterval)
	defer ticker.Stop()

	for range ticker.C {
		h.mu.RLock()
		current := h.conn
		h.mu.RUnlock()

		if current != conn {
			return // 连接已被替换
		}

		h.writeMu.Lock()
		data, marshalErr := json.Marshal(Message{
			Type: TypePing,
			ID:   uuid.New().String(),
			Data: json.RawMessage("{}"),
		})
		if marshalErr != nil {
			h.writeMu.Unlock()
			return
		}
		writeCtx, cancel := context.WithTimeout(context.Background(), h.ackTimeout)
		err := conn.Write(writeCtx, websocket.MessageText, data)
		cancel()
		h.writeMu.Unlock()

		if err != nil {
			return
		}
	}
}
