package main

import (
	"github.com/gorilla/websocket"
	"net/http"
	"log"
	"github.com/stretchr/objx"
)

type room struct {
	forward chan *message
	join chan *Client
	leave chan *Client
	clients map[*Client]bool
}

func (r *room) run() {
	for {
		select {
		case Client := <-r.join: r.clients[Client] = true 				//update the r.clients map to keep a reference of the client that has joined the room
		case Client := <-r.leave: delete(r.clients, Client); close(Client.send)		//delete the client type from the map, and close its send channel
		case msg := <-r.forward:							//iterate over all the clients and send the message down each client's send channel
			for Client := range r.clients{
				select {
				case Client.send <- msg: // send message
				default: delete(r.clients, Client); close(Client.send)
				}
			}
		}
	}
}

func newRoom() *room {
	return &room{
		clients: make(map[*Client]bool),
		forward: make(chan *message),
		join: make(chan *Client),
		leave: make(chan *Client),
	}
}

const(
	socketBufferSize  = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{ReadBufferSize: socketBufferSize, WriteBufferSize:socketBufferSize}

func (r *room) ServeHTTP (w http.ResponseWriter, req *http.Request){
	socket, err := upgrader.Upgrade(w,req,nil)
	if err != nil{
		log.Fatal("ServeHTTP: ", err)
		return
	}
	authCookie, err := req.Cookie("auth")
	if err != nil{
		log.Fatal("Failed to get auth cookie: ", err)
		return
	}
	client := &Client{
		socket:socket,
		send: make(chan *message, messageBufferSize),
		room:r,
		userData: objx.MustFromBase64(authCookie.Value),
	}
	r.join <- client
	defer func(){ r.leave <- client }()
	go client.write()
	client.read()
}