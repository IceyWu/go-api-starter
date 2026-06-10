package ws

import "encoding/json"

// Message 统一 WebSocket 消息格式
type Message struct {
	Type string          `json:"type"`
	ID   string          `json:"id"`
	Data json.RawMessage `json:"data"`
}

// --- Server → Hook 下行指令 ---

const (
	TypeSendTextMsg  = "send_text_msg"
	TypeGetGroupList = "get_group_list"
	TypePing         = "ping"
)

// SendTextMsgData 发送文本消息指令
type SendTextMsgData struct {
	Wxid string `json:"wxid"`
	Msg  string `json:"msg"`
}

// GetGroupListData 获取群列表指令（空）
type GetGroupListData struct{}

// --- Hook → Server 上行消息 ---

const (
	TypeAck    = "ack"
	TypeReview = "review"
	TypePong   = "pong"
)

// AckData 指令执行结果
type AckData struct {
	RefID   string          `json:"ref_id"`
	Success bool            `json:"success"`
	Error   string          `json:"error,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
}

// ReviewData 审核回调
type ReviewData struct {
	GroupUsername string `json:"group_username"`
	Phone        string `json:"phone"`
	Status       int    `json:"status"`
	Operator     string `json:"operator,omitempty"`
	Remark       string `json:"remark,omitempty"`
}
