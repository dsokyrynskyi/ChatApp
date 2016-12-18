package main

import (
	"github.com/gorilla/websocket"
	"net/http"
	"log"
)

type room struct {
	// incoming messages that should be forwarded to the other clients
	forward chan []byte

	// safely add and remove clients from the clients map
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
		forward: make(chan []byte),
		join: make(chan *Client),
		leave: make(chan *Client),
	}
}
/*it will only run one  block of CASE code at a time;
our r.clients map is only ever modified by one thing at a time
*/

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
	client := &Client{
		socket:socket,
		send: make(chan []byte, messageBufferSize),
		room:r,
	}
	r.join <- client
	defer func(){ r.leave <- client }()
	go client.write()
	client.read()
}