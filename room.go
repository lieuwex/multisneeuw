package main

import (
	"log"
	"strconv"

	"golang.org/x/net/websocket"
)

// Client is a client
type Client struct {
	ws      *websocket.Conn
	indexch chan int
}

// Room is a room
type Room struct {
	id      string
	clients []*Client
	score   int
}

// AddWs adds the given websocket connection to the current room, as the left or
// right player.
func (r *Room) AddWs(ws *websocket.Conn) (*Client, error) {
	ch := make(chan int, 1)

	log.Printf("player joined room %s\n", r.id)
	client := &Client{
		ws:      ws,
		indexch: ch,
	}
	r.clients = append(r.clients, client)
	r.broadCastIndices()

	return client, nil
}

// RemoveClient removes the ws with the given index from the game and closes the
// connection.
func (r *Room) RemoveClient(client *Client) {
	log.Printf("player left room %s\n", r.id)
	client.ws.Close()

	var index int
	for i, c := range r.clients {
		if c == client {
			index = i
			break
		}
	}
	copy(r.clients[index:], r.clients[index+1:])
	r.clients[len(r.clients)-1] = nil
	r.clients = r.clients[:len(r.clients)-1]

	r.broadCastIndices()

	if len(r.clients) == 0 {
		delete(roomMap, r.id)
	}
}

func (r *Room) broadCastIndices() {
	length := strconv.Itoa(len(r.clients))
	for i, client := range r.clients {
		client.ws.Write([]byte("index" + delim + strconv.Itoa(i+1) + delim + length + "\n"))
		client.indexch <- i
	}
}
