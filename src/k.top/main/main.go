package main

import (
	"github.com/gpmgo/gopm/modules/log"
	"html/template"
	"net/http"
	"path/filepath"
	"sync"
)

// templ represents a single template
type templateHandler struct {
	once     sync.Once
	filename string
	templ    *template.Template
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

	http.Handle("/room", room)

	go room.run()

	println("Starting web server on", 8080)

	e := http.ListenAndServe(":8080", nil)

	if e != nil {
		log.Fatal(`http server listen and serve`, e)
	}

}
