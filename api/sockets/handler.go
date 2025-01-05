package sockets

import (
	"fmt"
	"github.com/semaphoreui/semaphore/db"
	"net/http"
	"time"

	"github.com/gorilla/context"
	"github.com/gorilla/websocket"
	"github.com/semaphoreui/semaphore/util"
	log "github.com/sirupsen/logrus"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

type connection struct {
	ws     *websocket.Conn
	send   chan []byte
	userID int
}

// readPump pumps messages from the websocket connection to the hub.
func (c *connection) readPump() {
	defer func() {
		h.unregister <- c
		_ = c.ws.Close()
		//util.LogErrorWithFields(c.ws.Close(), log.Fields{"error": "Error closing websocket"})
	}()

	c.ws.SetReadLimit(maxMessageSize)

	//util.LogErrorWithFields(c.ws.SetReadDeadline(time.Now().Add(pongWait)), log.Fields{"error": "Socket state corrupt"})
	//
	//c.ws.SetPongHandler(func(string) error {
	//	util.LogErrorWithFields(c.ws.SetReadDeadline(time.Now().Add(pongWait)), log.Fields{"error": "Socket state corrupt"})
	//	return nil
	//})

	for {
		_, message, err := c.ws.ReadMessage()
		fmt.Println(string(message))

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				util.LogError(err)
			}
			break
		}
	}
}

// write writes a message with the given message type and payload.
func (c *connection) write(mt int, payload []byte) error {

	err := c.ws.SetWriteDeadline(time.Now().Add(writeWait))

	util.LogErrorF(err, log.Fields{
		"error": "Cannot set write deadline",
	})

	return c.ws.WriteMessage(mt, payload)
}

// writePump pumps messages from the hub to the websocket connection.
func (c *connection) writePump() {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		_ = c.ws.Close()
		//util.LogError(c.ws.Close())
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				util.LogErrorF(c.write(websocket.CloseMessage, []byte{}), log.Fields{
					"error": "Cannot send close message",
				})
				return
			}
			if err := c.write(websocket.TextMessage, message); err != nil {
				util.LogErrorF(err, log.Fields{
					"error": "Cannot send text message",
				})
				return
			}
		case <-ticker.C:
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				util.LogErrorF(err, log.Fields{
					"error": "Cannot send ping message",
				})
				return
			}
		}
	}
}

// Handler is used by the router to handle the /ws endpoint
func Handler(w http.ResponseWriter, r *http.Request) {
	usr := context.Get(r, "user")
	if usr == nil {
		return
	}

	user := usr.(*db.User)
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		panic(err)
	}

	c := &connection{
		send:   make(chan []byte, 256),
		ws:     ws,
		userID: user.ID,
	}

	h.register <- c

	go c.writePump()
	c.readPump()
}

// Message allows a message to be sent to the websockets, called in API task logging
func Message(userID int, message []byte) {
	h.broadcast <- &sendRequest{
		userID: userID,
		msg:    message,
	}
}
