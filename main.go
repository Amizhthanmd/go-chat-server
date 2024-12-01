package main

import (
	"fmt"
	"go-chat-server/controllers"
	"go-chat-server/helpers"
	"go-chat-server/loaders"
	"go-chat-server/routes"
	"os"
)

func main() {
	// Load ENV
	helpers.LoadEnv()

	var (
		PORT     = fmt.Sprintf(":%s", os.Getenv("PORT"))
		LOG_MODE = os.Getenv("LOG_MODE")
	)

	// Initialize logger
	logger := loaders.ZapLogger(LOG_MODE)
	defer logger.Sync()

	controller := controllers.NewController(logger)

	// Start server
	routes.HttpRouter(PORT, controller)

}
