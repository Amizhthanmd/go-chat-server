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

	c.Mutex.Lock()
	// Check if the user is exist in the users list
	if _, exists := c.Users[room.UserId]; !exists {
		c.Mutex.Unlock()
		c.logger.Error("User not found", zap.String("user_id", room.UserId))
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	// Check if the joining room is present in the rooms list
	if _, exists := c.Rooms[room.RoomId]; !exists {
		c.Mutex.Unlock()
		c.logger.Error("Room not found", zap.String("room_id", room.RoomId))
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	msgChannel := make(chan map[string]string)
	c.RoomChannels[room.RoomId] = append(c.RoomChannels[room.RoomId], msgChannel)
	// Add the users to the rooms list
	c.ActiveRoomUsers[room.RoomId] = append(c.ActiveRoomUsers[room.RoomId], models.RoomUser{
		UserId: room.UserId,
		Name:   c.Users[room.UserId],
	})
	c.Mutex.Unlock()

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Send the initial sse connection message
	_, err = w.Write([]byte("data: {\"message\": \"Connected to room\"}\n\n"))
	if err != nil {
		c.logger.Error("Error sending initial message", zap.Error(err))
		return
	}
	flusher.Flush()

	for {
		select {
		case <-r.Context().Done():
			// Remove users in the room if disconnected
			c.Mutex.Lock()
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
			c.Mutex.Unlock()
			return
		case msg, ok := <-msgChannel:
			if !ok {
				return
			}
			messageData := fmt.Sprintf("data: {\"name\": \"%s\", \"message\": \"%s\"}\n\n", msg["name"], msg["message"])
			_, err := w.Write([]byte(messageData))
			if err != nil {
				c.logger.Error("Error writing to response", zap.Error(err))
				return
			}
			flusher.Flush()
		}
	}
}

func (c *Controller) SendMessage(w http.ResponseWriter, r *http.Request) {
	roomId := r.URL.Query().Get("room_id")
	if roomId == "" {
		c.logger.Error("Room ID not found in URL query")
		http.Error(w, "Room ID not found in URL query", http.StatusBadRequest)
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

	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	userName, userExists := c.Users[message.UserId]
	if !userExists {
		c.logger.Error("User not found", zap.String("user_id", message.UserId))
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if _, roomExists := c.RoomChannels[roomId]; !roomExists {
		c.logger.Error("Room not found", zap.String("room_id", roomId))
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	userInRoom := false
	for _, roomUser := range c.ActiveRoomUsers[roomId] {
		if roomUser.UserId == message.UserId {
			userInRoom = true
			break
		}
	}
	if !userInRoom {
		c.logger.Error("User not in room", zap.String("user_id", message.UserId), zap.String("room_id", roomId))
		http.Error(w, "User not in the room", http.StatusForbidden)
		return
	}
	// Send message to all the users in the room
	for _, roomChannel := range c.RoomChannels[roomId] {
		roomChannel <- map[string]string{
			"name":    userName,
			"message": message.Message,
		}
	}

	res := HttpResponse{
		Message: "Message sent successfully",
		Status:  true,
		Data:    message,
	}
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
	userId := r.URL.Query().Get("user_id")
	if roomId == "" || userId == "" {
		c.logger.Error("Valid query not found in URL")
		http.Error(w, "Valid query not found in URL", http.StatusBadRequest)
		return
	}

	// Check users present in the current room
	var userFound = false
	for _, user := range c.ActiveRoomUsers[roomId] {
		if user.UserId == userId {
			userFound = true
			break
		}
	}
	if !userFound {
		c.logger.Error("You are not in the room", zap.String("room name", c.Rooms[roomId]))
		http.Error(w, "You are not in the room", http.StatusNotFound)
		return
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
