package controllers

import (
	"go-chat-server/models"
	"sync"

	"go.uber.org/zap"
)

type Controller struct {
	logger          *zap.Logger
	Users           map[string]string
	Rooms           map[string]string
	RoomChannels    map[string][]chan Message
	ActiveRoomUsers map[string][]models.RoomUser
	UserChat        map[string]chan string
	Mutex           sync.Mutex
}

type Message struct {
	Name    string `json:"name"`
	Message string `json:"message"`
}

func NewController(logger *zap.Logger) *Controller {
	return &Controller{
		logger:          logger,
		Users:           make(map[string]string),
		Rooms:           make(map[string]string),
		RoomChannels:    make(map[string][]chan Message),
		ActiveRoomUsers: make(map[string][]models.RoomUser),
		UserChat:        make(map[string]chan string),
		Mutex:           sync.Mutex{},
	}
}

type HttpResponse struct {
	Message string      `json:"message"`
	Status  bool        `json:"status"`
	Data    interface{} `json:"data"`
}
