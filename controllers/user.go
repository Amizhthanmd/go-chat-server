package controllers

import (
	"encoding/json"
	"fmt"
	"go-chat-server/models"
	"net/http"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (c *Controller) JoinUser(w http.ResponseWriter, r *http.Request) {
	var User models.User
	err := json.NewDecoder(r.Body).Decode(&User)
	if err != nil {
		c.logger.Error("Error decoding body", zap.Error(err))
		http.Error(w, "Error decoding body", http.StatusBadRequest)
		return
	}

	UniqueID := uuid.New().String()
	c.Users[UniqueID] = User.DisplayName

	response := HttpResponse{
		Message: "User joined successfully",
		Status:  true,
		Data:    UniqueID,
	}
	c.logger.Info("User joined successfully", zap.String("user", User.DisplayName))
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(&response)
	if err != nil {
		c.logger.Error("Error encoding response", zap.Error(err))
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

func (c *Controller) GetUsers(w http.ResponseWriter, r *http.Request) {
	if len(c.Users) == 0 {
		c.logger.Error("No users found")
		http.Error(w, "No users found", http.StatusNotFound)
		return
	}

	res := HttpResponse{
		Message: "Users fetched successfully",
		Status:  true,
		Data:    c.Users,
	}
	c.logger.Info("Users fetched successfully")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(&res)
	if err != nil {
		c.logger.Error("Error encoding response", zap.Error(err))
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

func (c *Controller) JoinChat(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked")
	var userID string
	userID = r.URL.Query().Get("user_id")
	if userID == "" {
		c.logger.Error("User ID not found in URL query")
		http.Error(w, "User ID not found in URL query", http.StatusBadRequest)
		return
	}

	c.Mutex.Lock()
	if _, exists := c.Users[userID]; !exists {
		c.logger.Error("User not found", zap.String("user_id", userID))
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	messageChan := make(chan string)
	c.UserChat[userID] = messageChan
	c.Mutex.Unlock()

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}
	_, err := w.Write([]byte("data: {\"message\": \"Connected to chat\"}\n\n"))
	if err != nil {
		c.logger.Error("Error sending initial message", zap.Error(err))
		return
	}
	flusher.Flush()

	for {
		select {
		case <-r.Context().Done():
			// Remove the user in the chat is disconnect
			c.Mutex.Lock()
			delete(c.UserChat, userID)
			c.Mutex.Unlock()
			return
		case message := <-messageChan:
			_, err := w.Write([]byte("data: " + message + "\n\n"))
			if err != nil {
				c.logger.Error("Error writing to response", zap.Error(err))
				return
			}
			flusher.Flush()
		}
	}
}

func (c *Controller) SendChatMessage(w http.ResponseWriter, r *http.Request) {
	var reqPayload struct {
		SenderID   string `json:"sender_id"`
		ReceiverID string `json:"receiver_id"`
		Message    string `json:"message"`
	}

	err := json.NewDecoder(r.Body).Decode(&reqPayload)
	if err != nil {
		c.logger.Error("Error decoding body", zap.Error(err))
		http.Error(w, "Error decoding body", http.StatusBadRequest)
		return
	}

	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	receiverChannel, exists := c.UserChat[reqPayload.ReceiverID]
	if !exists {
		c.logger.Error("Receiver not found", zap.String("receiver_id", reqPayload.ReceiverID))
		http.Error(w, "Receiver not found", http.StatusNotFound)
		return
	}

	// Send message the channel of the user
	message := fmt.Sprintf(`{"name": "%s", "message": "%s"}`, c.Users[reqPayload.SenderID], reqPayload.Message)
	receiverChannel <- message

	res := HttpResponse{
		Message: "Message sent successfully",
		Status:  true,
		Data:    message,
	}
	c.logger.Info("Message sent successfully")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(&res)
	if err != nil {
		c.logger.Error("Error encoding respons", zap.Error(err))
		http.Error(w, "Error encoding respons", http.StatusInternalServerError)
		return
	}
}
