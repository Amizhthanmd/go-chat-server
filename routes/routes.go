package routes

import (
	"fmt"
	"go-chat-server/controllers"
	"log"
	"net/http"
)

func HttpRouter(PORT string, c *controllers.Controller) {
	// Users routes
	http.HandleFunc("/api/v1/users", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		switch r.Method {
		case http.MethodGet:
			c.GetUsers(w, r)
		case http.MethodPost:
			c.JoinUser(w, r)
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
	})

	http.HandleFunc("/api/v1/users/chat", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			c.JoinChat(w, r)
		case http.MethodPost:
			c.SendChatMessage(w, r)
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
	})

	// Rooms routes
	http.HandleFunc("/api/v1/rooms", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		switch r.Method {
		case http.MethodGet:
			c.GetRooms(w, r)
		case http.MethodPost:
			c.CreateRoom(w, r)
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
	})

	http.HandleFunc("/api/v1/rooms/join", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			c.JoinRoom(w, r)
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
	})

	http.HandleFunc("/api/v1/rooms/send", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		switch r.Method {
		case http.MethodPost:
			c.SendMessage(w, r)
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
	})

	log.Printf("Listening and serving HTTP on %s", PORT)
	err := http.ListenAndServe(PORT, nil)
	if err != nil {
		fmt.Println(err)
	}
}
