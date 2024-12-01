package models

type User struct {
	DisplayName string `json:"display_name"`
}
type JoinRoom struct {
	RoomId string `json:"room_id"`
	UserId string `json:"user_id"`
}

type RoomUser struct {
	UserId string `json:"user_id"`
	Name   string `json:"name"`
}
