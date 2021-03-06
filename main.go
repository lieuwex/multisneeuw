package main

import (
	"log"
	"net/http"
	"os"

	"golang.org/x/net/websocket"
)

const defaultPort = "1337"

var roomMap = make(map[string]*Room)

func getOrMkRoom(id string) *Room {
	room, exists := roomMap[id]
	if exists {
		return room
	}

	room = &Room{
		id: id,
	}
	roomMap[id] = room
	return room
}

func frontHandler(w http.ResponseWriter, r *http.Request) {
	slug := r.URL.Path[1:]
	switch slug {
	case "flocon.js", "client.js", "client.css":
		http.ServeFile(w, r, "client/"+slug)
	case "":
		http.Redirect(w, r, "/"+UniqIdf(), 307)
	default:
		http.ServeFile(w, r, "client/client.html")
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	http.HandleFunc("/", frontHandler)
	http.Handle("/ws", websocket.Handler(WsHandler))
	log.Print("listening on " + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
