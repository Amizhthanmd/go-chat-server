package models

type Room struct {
	RoomName string `json:"room_name" binding:"required"`
	RoomId   string `json:"room_id" binding:"required"`
}
