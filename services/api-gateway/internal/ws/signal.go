package ws

import (
	"encoding/json"
	"fmt"
)

// signalMessage is the envelope sent over RabbitMQ to the SFU.
type signalMessage struct {
	Action string          `json:"action"`
	RoomID string          `json:"room_id"`
	UserID string          `json:"user_id"`
	Data   json.RawMessage `json:"data"`
}

// PublishSignal publishes a WS client message to the RabbitMQ signal exchange.
func PublishSignal(publisher func(exchange, routingKey string, body []byte), userID, roomID, action string, data json.RawMessage) error {
	if publisher == nil {
		return fmt.Errorf("no publisher configured")
	}

	msg := signalMessage{
		Action: action,
		RoomID: roomID,
		UserID: userID,
		Data:   data,
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal signal message: %w", err)
	}

	publisher(signalExchange, "signal.room."+roomID, body)
	return nil
}

// IncomingMessage types that get forwarded to the SFU.
const (
	MsgTypeJoinRoom     = "join_room"
	MsgTypeSDPAnswer    = "sdp_answer"
	MsgTypeICECandidate = "ice_candidate"
	MsgTypePublish      = "publish"
	MsgTypeUnpublish    = "unpublish"
	MsgTypeLeaveRoom    = "leave_room"
)
