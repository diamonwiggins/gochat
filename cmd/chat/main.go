package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/websocket"
)

type Message struct {
	Timestamp string `json:"timestamp"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	Message   string `json:"message"`
}

func main() {
	var connections = make(map[*websocket.Conn]bool)
	var sendChannel = make(chan Message)
	var subChannel = make(chan Message)

	fs := http.FileServer(http.Dir("/app/web"))
	http.Handle("/", fs)

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		webSocketHandler(w, r, sendChannel, connections)
	})

	go publishMessages(sendChannel)
	go receiveMessages(subChannel)
	go broadcastMessages(subChannel, connections)

	log.Println("http server started on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func webSocketHandler(w http.ResponseWriter, r *http.Request, senChan chan Message, conn map[*websocket.Conn]bool) {
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}

	conn[ws] = true

	for {
		var msg Message

		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			delete(conn, ws)
			break
		}

		senChan <- msg
	}
	defer ws.Close()
}

func publishMessages(senChan chan Message) {
	conn, err := redis.Dial("tcp", "redis:6379")
	if err != nil {
		log.Fatal(err)
	}
	for {
		select {
		case currentMsg := <-senChan:
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

func receiveMessages(subChan chan Message) {
	for {
		conn, err := redis.Dial("tcp", "redis:6379")
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
				subChan <- *currentMsgUnmarshalled
			case redis.Subscription:
			case error:
				return
			}
		}
		defer conn.Close()
	}
}

func broadcastMessages(subChan chan Message, c map[*websocket.Conn]bool) {
	for {
		msg := <-subChan
		for conn := range c {
			err := conn.WriteJSON(msg)
			if err != nil {
				log.Printf("error: %v", err)
				delete(c, conn)
			}
		}
	}
}
