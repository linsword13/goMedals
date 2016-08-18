package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []Country
}

func (c *Client) writeList() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for msg := range c.send {
		err := c.conn.WriteJSON(msg)
		if err != nil {
			return
		}
	}
}

// func (c *Client) pingClient() {
// 	defer func() {
// 		c.hub.unregister <- c
// 		c.conn.Close()
// 	}()
// 	var pingPeriod = 30 * time.Second
// 	var writeWait = 5 * time.Second
// 	ticker := time.NewTicker(pingPeriod)
// 	defer ticker.Stop()
// 	for {
// 		select {
// 		case <-ticker.C:
// 			log.Println("pinging...")
// 			if err := c.conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(writeWait)); err != nil {
// 				log.Println("ping:", err)
// 				return
// 			}
// 			log.Println("finish")
// 		default:
// 		}
// 	}
// }

func chatHandler(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{hub: hub, conn: conn, send: make(chan []Country)}
	client.hub.register <- client
	client.conn.WriteJSON(CurList.Body.MedalRank.MedalsList) // initial list
	client.writeList()
	//client.pingClient()
}
