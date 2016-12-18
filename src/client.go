package main

import "github.com/gorilla/websocket"

/* here channels are in-memory thread-safe message queue
where senders pass data and
receivers read data in a non-blocking, thread-safe way.
 */

type Client struct {
	socket *websocket.Conn
	send chan []byte // send messages
	room *room
}

/*continually accepts messages from the send channel
writing everything out of the socket*/

func (c *Client) write(){
	for msg := range c.send{
		if err := c.socket.WriteMessage(websocket.TextMessage, msg); err != nil{
			break
		}
	}
	c.socket.Close()
}

/* allows our client to read from the socket via the ReadMessage method,
continually sending any received messages to the forward channel on  the room type*/

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
