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

func removeWsAtIndex(room *Room, id string, index int) {
	if room.RemoveWs(index) {
		delete(roomMap, id)
	}
}

func WsHandler(ws *websocket.Conn) {
	log.Println("new ws connection")

	reader := bufio.NewReader(ws)
	str, err := reader.ReadString('\n')
	if err != nil {
		log.Println(err)
		ws.Close()
		return
	}

	id := str[0 : len(str)-1]
	room := getOrMkRoom(id)
	index, err := room.AddWs(ws)
	if err != nil {
		log.Println(err)
		return
	}
	defer removeWsAtIndex(room, id, index)

	pingsSinceLastMessage := 0

	const channelCount = 3
	waitchan := make(chan int, channelCount)
	quit := func() {
		for i := 0; i < channelCount; i++ {
			waitchan <- 1
		}
	}

	go func() {
		for {
			select {
			case <-waitchan:
				return
			default:
				time.Sleep(time.Second * pingTime)

				if pingsSinceLastMessage == pingTimeout {
					log.Println("ws timed out, killing connection")
					quit()
					return
				}

				ws.Write([]byte("ping" + delim + "pong\n"))
				pingsSinceLastMessage++
			}
		}
	}()

	go func() {
		for {
			select {
			case <-waitchan:
				return
			default:
				str, err = reader.ReadString('\n')
				if err != nil {
					quit()
					return
				}

				pingsSinceLastMessage = 0

				splitted := strings.Split(str, delim)
				var otherIndex int
				switch splitted[0] {
				case "ping":
					ws.Write([]byte("pong" + delim + "ping\n"))
					continue
				case "pong":
					continue

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
	}()

	<-waitchan
}
