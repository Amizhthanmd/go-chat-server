package controllers

import (
	"encoding/json"
	"fmt"
	"go-chat-server/helpers"
	"go-chat-server/models"

	"net/http"

	"go.uber.org/zap"
)

func (c *Controller) CreateRoom(w http.ResponseWriter, r *http.Request) {
	var room models.Room
	err := json.NewDecoder(r.Body).Decode(&room)
	if err != nil {
		c.logger.Error("Error decoding body", zap.Error(err))
		http.Error(w, "Error decoding body", http.StatusBadRequest)
		return
	}

	UniqueID, _ := helpers.GenerateRoomID()
	room.RoomId = UniqueID

	c.Rooms[UniqueID] = room.RoomName

	response := HttpResponse{
		Message: "Room created successfully",
		Status:  true,
		Data:    room,
	}
	c.logger.Info("Room created successfully", zap.String("room", room.RoomName))
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(&response)
	if err != nil {
		c.logger.Error("Error encoding response", zap.Error(err))
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

func (c *Controller) GetRooms(w http.ResponseWriter, r *http.Request) {
	if len(c.Rooms) == 0 {
		http.Error(w, "No rooms found", http.StatusNotFound)
		return
	}
	res := HttpResponse{
		Message: "Rooms fetched successfully",
		Status:  true,
		Data:    c.Rooms,
	}
	w.WriteHeader(http.StatusOK)
	c.logger.Info("Rooms fetched successfully")
	err := json.NewEncoder(w).Encode(&res)
	if err != nil {
		c.logger.Error("Error encoding response", zap.Error(err))
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

func (c *Controller) JoinRoom(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked")

	var room models.JoinRoom
	err := json.NewDecoder(r.Body).Decode(&room)
	if err != nil {
		c.logger.Error("Error decoding body", zap.Error(err))
		http.Error(w, "Error decoding body", http.StatusBadRequest)
		return
	}

	if _, exists := c.Users[room.UserId]; !exists {
		c.logger.Error("User not found", zap.String("user_id", room.UserId))
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	if _, exists := c.Rooms[room.RoomId]; !exists {
		c.logger.Error("Room not found", zap.String("room_id", room.RoomId))
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	msgChannel := make(chan map[string]string)
	c.RoomChannels[room.RoomId] = append(c.RoomChannels[room.RoomId], msgChannel)
	c.ActiveRoomUsers[room.RoomId] = append(c.ActiveRoomUsers[room.RoomId], models.RoomUser{
		UserId: room.UserId,
		Name:   c.Users[room.UserId],
	})

	defer func() {
		for i, client := range c.RoomChannels[room.RoomId] {
			if client == msgChannel {
				c.RoomChannels[room.RoomId] = append(c.RoomChannels[room.RoomId][:i], c.RoomChannels[room.RoomId][i+1:]...)
				break
			}
		}
		for i, v := range c.ActiveRoomUsers[room.RoomId] {
			if v.UserId == room.UserId {
				c.ActiveRoomUsers[room.RoomId] = append(c.ActiveRoomUsers[room.RoomId][:i], c.ActiveRoomUsers[room.RoomId][i+1:]...)
				break
			}
		}
	}()

	for {
		select {
		case <-r.Context().Done():
			return
		case msg := <-msgChannel:
			messageData := fmt.Sprintf("data: {\"name\": \"%s\", \"message\": \"%s\"}\n\n", msg["name"], msg["message"])
			_, err := w.Write([]byte(messageData))
			if err != nil {
				c.logger.Error("Error writing to response", zap.Error(err))
				fmt.Println("Error writing to response:", err)
				return
			}
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
		}
	}
}

func (c *Controller) SendMessage(w http.ResponseWriter, r *http.Request) {
	var roomId string
	roomId = r.URL.Query().Get("room_id")
	if roomId == "" {
		c.logger.Error("Room ID not found in URL query")
		http.Error(w, "Room ID not found in URL query", http.StatusBadRequest)
		return
	}
	if _, exists := c.RoomChannels[roomId]; !exists {
		c.logger.Error("Room not found", zap.String("room_id", roomId))
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	var message struct {
		UserId  string `json:"user_id"`
		Message string `json:"message"`
	}
	err := json.NewDecoder(r.Body).Decode(&message)
	if err != nil {
		c.logger.Error("Error decoding body", zap.Error(err))
		http.Error(w, "Error decoding body", http.StatusBadRequest)
		return
	}
	if _, exists := c.Users[message.UserId]; !exists {
		c.logger.Error("User not found", zap.String("user_id", message.UserId))
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	for _, roomUser := range c.ActiveRoomUsers[roomId] {
		if roomUser.UserId != message.UserId {
			c.logger.Error("You are not in the room", zap.String("room name", c.Rooms[roomId]))
			http.Error(w, "You are not in the room", http.StatusNotFound)
			return
		}
	}

	for _, roomChannels := range c.RoomChannels[roomId] {
		roomChannels <- map[string]string{
			"name":    c.Users[message.UserId],
			"message": message.Message,
		}
	}

	res := HttpResponse{
		Message: "Message sent successfully",
		Status:  true,
		Data:    message,
	}
	c.logger.Info("Message sent successfully")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(&res)
	if err != nil {
		c.logger.Error("Error encoding response", zap.Error(err))
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

func (c *Controller) GetRoomsUsers(w http.ResponseWriter, r *http.Request) {
	roomId := r.URL.Query().Get("room_id")
	user_Id := r.URL.Query().Get("user_id")
	if roomId == "" || user_Id == "" {
		c.logger.Error("Valid query not found in URL")
		http.Error(w, "Valid query not found in URL", http.StatusBadRequest)
		return
	}
	for _, user := range c.ActiveRoomUsers[roomId] {
		if user.UserId != user_Id {
			c.logger.Error("You are not in the room", zap.String("room name", c.Rooms[roomId]))
			http.Error(w, "You are not in the room", http.StatusNotFound)
			return
		}
	}

	if _, exists := c.ActiveRoomUsers[roomId]; !exists {
		c.logger.Error("Room not found", zap.String("room id", roomId))
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	res := HttpResponse{
		Message: "Room users fetched successfully",
		Status:  true,
		Data:    c.ActiveRoomUsers[roomId],
	}
	c.logger.Info("Room users fetched successfully")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(&res)
	if err != nil {
		c.logger.Error("Error encoding response", zap.Error(err))
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}
