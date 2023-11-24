package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Logger struct represents a simple logging utility.
type Logger struct{}

// Setup prints a message when the server is running.
func (l *Logger) Setup() {
	fmt.Println("\x1b[37m\x1b[42m\n Server is running \x1b[0m")
}

// Success logs a success message with optional IP information.
func (l *Logger) Success(message string, ip string) {
	fmt.Printf("\x1b[42mSuccess!\x1b[0m %s %s\n", message, getIPInfo(ip))
}

// Warning logs a warning message with optional IP information.
func (l *Logger) Warning(message string, ip string) {
	fmt.Printf("\x1b[43mWarning!\x1b[0m %s %s\n", message, getIPInfo(ip))
}

// Error logs an error message with optional IP information.
func (l *Logger) Error(message string, ip string) {
	fmt.Printf("\x1b[41mError!\x1b[0m %s %s\n", message, getIPInfo(ip))
}

// Reader struct represents a simple reader utility for decoding messages.
type Reader struct {
	packet []byte
}

// Decode converts the packet bytes to a string.
func (r *Reader) Decode() string {
	return string(r.packet)
}

var clients = make(map[string]*websocket.Conn)
var mutex = &sync.Mutex{}

func generateClientId() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%s", randString(9))
}

func randString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func getIPInfo(ip string) string {
	if ip != "" {
		return "from " + ip
	}
	return ""
}

func handleConnection(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	clientID := generateClientId()

	mutex.Lock()
	clients[clientID] = conn
	mutex.Unlock()

	logger := &Logger{}
	logger.Success(fmt.Sprintf("Connect client ID: %s", clientID), r.RemoteAddr)

	conn.WriteJSON(map[string]interface{}{
		"type":   0,
		"status": 200,
		"id":     clientID,
	})

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			delete(clients, clientID)
			logger.Error("Connection closed", r.RemoteAddr)
			break
		}

		message := Reader{packet: data}.Decode()
		fmt.Println(message)
	}
}

func main() {
	http.HandleFunc("/", handleConnection)
	port := 3000
	fmt.Printf("Server is running on port %d...\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
