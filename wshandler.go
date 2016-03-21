package main

import (
	"bufio"
	"log"
	"strings"
	"time"

	"golang.org/x/net/websocket"
)

const (
	pingTime    = 10 // in seconds
	pingTimeout = 3  // amount of pings until timeout
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

	pingsSinceLastMessage := 0
	waitchan := make(chan int)

	go func() {
		for {
			time.Sleep(time.Second * pingTime)
			ws.Write([]byte("ping" + delim + "pong"))

			pingsSinceLastMessage++
			if pingsSinceLastMessage == pingTimeout+1 {
				log.Println("ws timed out, killing connection")
				waitchan <- 1
			}
		}
	}()

	go func() {
		for {
			str, err = reader.ReadString('\n')
			if err != nil {
				waitchan <- 1
			}

			pingsSinceLastMessage = 0

			splitted := strings.Split(str, delim)
			var otherIndex int
			switch splitted[0] {
			case "L":
				otherIndex = index - 1
			case "R":
				otherIndex = index + 1
			case "pong":
				break
			default:
				ws.Write([]byte(MakeWsErr("invalid-side")))
				continue
			}

			if otherIndex >= 0 && otherIndex < len(room.sides) {
				message := str[:len(str)-1]
				room.sides[otherIndex].Write([]byte(message))
			}
		}
	}()

	<-waitchan
}
