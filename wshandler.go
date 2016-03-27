package main

import (
	"bufio"
	"log"
	"strings"
	"time"

	"golang.org/x/net/websocket"
)

const (
	pingTime    = 5 // in seconds
	pingTimeout = 1 // amount of pings until timeout
)

func WsHandler(ws *websocket.Conn) {
	log.Printf("new ws connection from %s\n", ws.RemoteAddr().String())
	reader := bufio.NewReader(ws)

	roomid, err := reader.ReadString('\n')
	if err != nil {
		log.Println(err)
		ws.Close()
		return
	}

	room := getOrMkRoom(roomid[0 : len(roomid)-1])

	client, err := room.AddWs(ws)
	if err != nil {
		log.Println(err)
		return
	}
	defer room.RemoveClient(client)

	waitch := make(chan int)

	var index int
	go func() {
		for {
			select {
			case index = <-client.indexch:
			case <-waitch:
				return
			}
		}
	}()

	pingsSinceLastMessage := 0

	go func() {
		for {
			select {
			case <-waitch:
				return
			default:
				time.Sleep(time.Second * pingTime)

				if pingsSinceLastMessage == pingTimeout {
					log.Println("ws timed out, killing connection")
					ws.Close()
					return
				}

				ws.Write([]byte("ping" + delim + "pong\n"))
				pingsSinceLastMessage++
			}
		}
	}()

	go func() {
		for {
			str, err := reader.ReadString('\n')
			if err != nil {
				close(waitch)
				return
			}

			pingsSinceLastMessage = 0

			splitted := strings.Split(str, delim)
			switch splitted[0] {
			case "ping":
				ws.Write([]byte("pong" + delim + "ping\n"))
			case "pong":

			case "L", "R":
				var otherIndex int
				if splitted[0] == "L" {
					otherIndex = index - 1
				} else {
					otherIndex = index + 1
				}

				if otherIndex >= 0 && otherIndex < len(room.clients) {
					room.clients[otherIndex].ws.Write([]byte(str))
				}

			case "A", "B":
				self := splitted[0] == "A"

				for i, client := range room.clients {
					if i != index || self {
						client.ws.Write([]byte(str))
					}
				}

			default:
				ws.Write(MakeWsErr("invalid-side"))
			}
		}
	}()

	<-waitch
}
