package main

import (
	"github.com/gorilla/websocket"
	"time"
)

type Client struct {
	socket *websocket.Conn
	send chan *message // send messages
	room *room
	userData map[string]interface{}
}

func (c *Client) write(){
	for msg := range c.send{
		if err := c.socket.WriteJSON(msg); err != nil{
			break
		}
	}
	c.socket.Close()
}

func (c *Client) read(){
	for {
		var msg *message
		if err := c.socket.ReadJSON(&msg); err == nil{
			msg.When = time.Now()
			msg.Name = c.userData["name"].(string)
			c.room.forward <- msg
		} else {
			break
		}
	}
	c.socket.Close()
}
