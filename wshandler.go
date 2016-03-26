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
	log.Println("new ws connection")

	reader := bufio.NewReader(ws)
	str, err := reader.ReadString('\n')
	if err != nil {
		log.Println(err)
		ws.Close()
		return
	}

	var index int
	room := getOrMkRoom(str[0 : len(str)-1])
	client, err := room.AddWs(ws)
	if err != nil {
		log.Println(err)
		return
	}
	defer room.RemoveClient(client)

	waitch := make(chan int, 4)

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
					close(waitch)
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
			case <-waitch:
				return
			default:
				str, err = reader.ReadString('\n')
				if err != nil {
					close(waitch)
					return
				}

				pingsSinceLastMessage = 0

				var otherIndex int
				self := false

				splitted := strings.Split(str, delim)
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

				case "A":
					self = true
					fallthrough
				case "B":
					for i, client := range room.clients {
						if i != index || self {
							client.ws.Write([]byte(str))
						}
					}
					continue

				default:
					ws.Write([]byte(MakeWsErr("invalid-side")))
					continue
				}

				if otherIndex >= 0 && otherIndex < len(room.clients) {
					room.clients[otherIndex].ws.Write([]byte(str))
				}
			}
		}
	}()

	<-waitch
}
