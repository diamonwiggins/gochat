package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/websocket"
)

var connections = make(map[*websocket.Conn]bool)
var sendChannel = make(chan Message)
var subChannel = make(chan Message)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Message struct {
	Timestamp string `json:"timestamp"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	Message   string `json:"message"`
}

func main() {
	fs := http.FileServer(http.Dir("/app/web"))
	http.Handle("/", fs)

	http.HandleFunc("/ws", webSocketHandler)

	go publishMessages()
	go receiveMessages()
	go broadcastMessages()

	log.Println("http server started on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func webSocketHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}

	connections[ws] = true

	for {
		var msg Message

		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			delete(connections, ws)
			break
		}

		sendChannel <- msg
	}
	defer ws.Close()
}

func publishMessages() {
	conn, err := redis.Dial("tcp", "sgg-redis:6379")
	if err != nil {
		log.Fatal(err)
	}
	for {
		select {
		case currentMsg := <-sendChannel:
			currentMsgMarshalled, err := json.Marshal(&currentMsg)
			if err != nil {
				log.Printf("error: %v", err)
				break
			}
			conn.Do("PUBLISH", "allmessages", currentMsgMarshalled)
		}
	}
	defer conn.Close()
}

func receiveMessages() {
	for {
		conn, err := redis.Dial("tcp", "sgg-redis:6379")
		if err != nil {
			log.Fatal(err)
		}
		psc := redis.PubSubConn{Conn: conn}
		psc.Subscribe("allmessages")

		for {
			switch v := psc.Receive().(type) {
			case redis.Message:
				b := []byte(v.Data)
				currentMsgUnmarshalled := &Message{}
				err := json.Unmarshal(b, currentMsgUnmarshalled)
				if err != nil {
					log.Printf("error: %v", err)
					break
				}
				subChannel <- *currentMsgUnmarshalled
			case redis.Subscription:
			case error:
				return
			}
		}
		defer conn.Close()
	}
}

func broadcastMessages() {
	msg := <-subChannel
	for conn := range connections {
		err := conn.WriteJSON(msg)
		if err != nil {
			log.Printf("error: %v", err)
			delete(connections, conn)
		}
	}
}
