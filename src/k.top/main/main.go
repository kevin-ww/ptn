package main

import (
	"github.com/gorilla/websocket"
	"github.com/gpmgo/gopm/modules/log"
	"html/template"
	http "net/http"
	"path/filepath"
	"sync"
)

// templ represents a single template
type templateHandler struct {
	once     sync.Once
	filename string
	templ    *template.Template
}
type client struct {
	socket *websocket.Conn
	send   chan []byte
	room   *room
}

type room struct {
	forward chan []byte

	join  chan *client
	leave chan *client

	//false means the client is left
	clients map[*client]bool
}

func (r *room) run() {
	for {
		select {
		case client := <-r.join:
			r.clients[client] = true
		case client := <-r.leave:
			delete(r.clients, client)
			close(client.send)
		case msg := <-r.forward:
			for client := range r.clients {
				client.send <- msg
			}
		}
	}
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
	_, msgAsBytes, err := c.socket.ReadMessage()
	if err != nil {
		return
	}

	c.room.forward <- msgAsBytes
}

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{ReadBufferSize: socketBufferSize,
	WriteBufferSize: socketBufferSize}

func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	conn, e := upgrader.Upgrade(w, req, nil)

	if e != nil {
		log.Fatal("ServeHTTP:", e)
		return
	}

	client := &client{
		socket: conn,
		send:   make(chan []byte, messageBufferSize),
		room:   r,
	}

	r.join <- client

	defer func() {
		r.leave <- client
	}()

	go client.write()
	client.read()
}


func NewRoom() *room{
	return &room{
		forward: make(chan []byte),
		join:    make (chan *client),
		leave:   make(chan *client),
		clients: make (map[*client]bool),
	}
}

// ServeHTTP handles the HTTP request.
func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("template",
			t.filename)))
	})
	t.templ.Execute(w, nil)
}

func main() {

	room := NewRoom()

	http.Handle("/", &templateHandler{filename: "chat.html"})

	http.Handle("/room",room)

	go room.run()

	e := http.ListenAndServe(":8080", nil)

	if e != nil {
		log.Fatal(`http server listen and serve`, e)
	}

}

//func main() {
//	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
//		w.Write([]byte(`
//         <html>
//           <head>
//             <title>Chat</title>
//           </head>
//           <body>
//             Let's chat!
//           </body>
//</html>`))
//	})
//	// start the web server
//	if err := http.ListenAndServe(":8080", nil); err != nil {
//		log.Fatal("ListenAndServe:", err)
//	}
//}
