package main

import (
	//"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
	"log"
	"net/http"
	"websocket/PubSubServer/pubsub"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func autoId() string{
	return uuid.Must(uuid.NewV4()).String()
}

var ps = &pubsub.PubSub{}

func webscoketHandler(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin =func(r *http.Request) bool {
		return true;
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println("New client is connected")

	client := pubsub.Client{
		Id:         autoId(),
		Connection: conn,
	}
	// add this client into the list
	ps.AddClient(client)
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println("Something went wrong", err)
			ps.RemoveClient(client)
			log.Println("Total clients and subscriptions", len(ps.Clients), len(ps.Subscriptions))

			return
		}
		ps.HandleReceiveMessage(client,messageType,p)
	}
}
func main(){

	http.HandleFunc("/", func(w http.ResponseWriter,r *http.Request){
		http.ServeFile(w,r,"static")
	})

	http.HandleFunc("/ws", webscoketHandler)

	fmt.Println("Server is running")
	http.ListenAndServe(":8080", nil)
}