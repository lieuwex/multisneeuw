package main

import (
	"bufio"
	"log"
	"strings"

	"golang.org/x/net/websocket"
)

func WsHandler(ws *websocket.Conn) {
	log.Println("new ws connection")

	reader := bufio.NewReader(ws)
	str, err := reader.ReadString('\n')
	if err != nil {
		log.Println(err)
		ws.Close()
		return
	}

	room := getOrMkRoom(str[0 : len(str)-1])
	index, err := room.AddWs(ws)
	if err != nil {
		log.Println(err)
		return
	}
	defer room.RemoveWs(index)

	for {
		str, err = reader.ReadString('\n')
		if err != nil {
			log.Println(err)
			return
		}

		splitted := strings.Split(str, delim)
		var otherIndex int
		switch splitted[0] {
		case "L":
			otherIndex = index - 1
		case "R":
			otherIndex = index + 1
		default:
			ws.Write([]byte(MakeWsErr("invalid-side")))
			continue
		}

		if otherIndex >= 0 && otherIndex < len(room.sides) {
			message := str[:len(str)-1]
			room.sides[otherIndex].Write([]byte(message))
		}
	}
}
