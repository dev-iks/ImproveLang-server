package main

import (
	"fmt"
	"net/http"

	"gopkg.in/olivere/elastic.v6"

	r "github.com/dancannon/gorethink"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  2048,
	WriteBufferSize: 2048,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// Handler function accepts as an argument in Handle function
type Handler func(*Client, interface{})

// Router struct that stores rules to WebSocket
type Router struct {
	rules         map[string]Handler
	session       *r.Session
	elasticClient *elastic.Client
}

// NewRouter inciated Router
func NewRouter(session *r.Session, elasticClient *elastic.Client) *Router {
	return &Router{
		rules:         make(map[string]Handler),
		session:       session,
		elasticClient: elasticClient,
	}
}

// Handle registers the handler for the given message name and which func to handle it
func (r *Router) Handle(messageName string, handler Handler) {
	r.rules[messageName] = handler
}

func (r *Router) FindHandler(msgName string) (Handler, bool) {
	handler, found := r.rules[msgName]
	return handler, found
}

func (e *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	socket, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err.Error())
		return
	}
	client := NewClient(socket, e.FindHandler, e.session, e.elasticClient)
	defer client.Close()
	go client.Write()
	client.Read()
}
