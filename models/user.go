package models

type User struct {
	DisplayName string `json:"display_name" binding:"required"`
}
type JoinRoom struct {
	RoomId string `json:"room_id" binding:"required"`
	UserId string `json:"user_id" binding:"required"`
}

type RoomUser struct {
	UserId string `json:"user_id"`
	Name   string `json:"name"`
}
