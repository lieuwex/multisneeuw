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

	room := getOrMkRoom(str[0 : len(str)-1])
	index, err := room.AddWs(ws)
	if err != nil {
		log.Println(err)
		return
	}
	defer room.RemoveWs(index)

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
				self := false
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
					for i, ws := range room.sides {
						if i != index || self {
							ws.Write([]byte(str))
						}
					}
					continue

				default:
					ws.Write([]byte(MakeWsErr("invalid-side")))
					continue
				}

				if otherIndex >= 0 && otherIndex < len(room.sides) {
					room.sides[otherIndex].Write([]byte(str))
				}
			}
		}
	}()

	<-waitchan
}
