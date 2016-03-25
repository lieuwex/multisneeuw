package main

import (
	"errors"
	"log"
	"strconv"

	"golang.org/x/net/websocket"
)

const maxRoomSize = 2

// Room is a room
type Room struct {
	id    string
	sides []*websocket.Conn
}

// AddWs adds the given websocket connection to the current room, as the left or
// right player.
func (r *Room) AddWs(ws *websocket.Conn) (int, error) {
	if len(r.sides) == maxRoomSize {
		s := MakeWsErr("room-full")
		n, err := ws.Write([]byte(s))
		if n < len(s) || err != nil {
			ws.Close()
			return -1, errors.New("error while writing to websocket")
		}

		log.Printf("player joined room %s but room was full\n", r.id)
		return -1, errors.New("room full")
	}

	log.Printf("player joined room %s\n", r.id)
	index := len(r.sides)
	r.sides = append(r.sides, ws)

	r.broadCastIndices()
	return index, nil
}

// RemoveWs removes the ws with the given index from the game and closes the
// connection.
func (r *Room) RemoveWs(i int) {
	log.Printf("player left room %s\n", r.id)
	if i >= 0 && i < len(r.sides) {
		r.sides[i].Close()

		copy(r.sides[i:], r.sides[i+1:])
		r.sides[len(r.sides)-1] = nil
		r.sides = r.sides[:len(r.sides)-1]

		r.broadCastIndices()

		if len(r.sides) == 0 {
			delete(roomMap, r.id)
		}
	}
}

func (r *Room) broadCastIndices() {
	length := strconv.Itoa(len(r.sides))
	for i, ws := range r.sides {
		ws.Write([]byte("index" + delim + strconv.Itoa(i+1) + delim + length + "\n"))
	}
}
