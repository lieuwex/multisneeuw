package main

import (
	"log"
	"net/http"

	"golang.org/x/net/websocket"
)

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
	if slug == "flocon.js" {
		http.ServeFile(w, r, slug)
	} else if len(slug) == 0 {
		http.Redirect(w, r, "/"+UniqIdf(), 307)
	} else {
		http.ServeFile(w, r, "client.html")
	}
}

func main() {
	http.HandleFunc("/", frontHandler)
	http.Handle("/ws", websocket.Handler(WsHandler))
	log.Print("listening on 1337")
	log.Fatal(http.ListenAndServe(":1337", nil))
}

/*
"use strict";

let scodemap={}; //map sessioncode -> {left: websocket,right: websocket,flakes: [[x,y]]}

module.exports=function(app){
	app.param("sessioncode",function(req,res,next,value){
		if(!/^[0-9a-z]{3}$/.test(value)){
			res.end("Invalid session code!");
			return;
		}
		req.sessioncode=value;
		next();
	});

	app.get("/:sessioncode",function(req,res){
		;
	});
};
*/
