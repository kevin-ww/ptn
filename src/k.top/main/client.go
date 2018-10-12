package main

import "github.com/gorilla/websocket"

type client struct {
	socket *websocket.Conn
	send   chan []byte
	room   *room
}

func (c *client) write() {
	defer c.socket.Close()

	for msg := range c.send {
		e := c.socket.WriteMessage(websocket.TextMessage, msg)
		if e != nil {
			return
		}
	}
}

func (c *client) read() {
	defer c.socket.Close()

	for {
		_, msgAsBytes, err := c.socket.ReadMessage()
		if err != nil {
			return
		}

		c.room.forward <- msgAsBytes
	}
}
