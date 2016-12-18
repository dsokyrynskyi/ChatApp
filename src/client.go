package main

import "github.com/gorilla/websocket"

type Client struct {
	socket *websocket.Conn
	send chan []byte // send messages
	room *room
}

func (c *Client) write(){
	for msg := range c.send{
		if err := c.socket.WriteMessage(websocket.TextMessage, msg); err != nil{
			break
		}
	}
	c.socket.Close()
}

func (c *Client) read(){
	for {
		if _, msg, err := c.socket.ReadMessage(); err == nil {
			c.room.forward <- msg
		} else {
			break
		}
	}
	c.socket.Close()
}
