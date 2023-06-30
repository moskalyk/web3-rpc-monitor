package main

import (
	"os"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"
	"github.com/gorilla/websocket"
	"math/big"
	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	upgrader  = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			// Replace with the origin of your React app
			allowedOrigin := "http://localhost:3000"
	
			return r.Header.Get("Origin") == allowedOrigin
		},
	}
	clients   = make(map[*websocket.Conn]bool)
	broadcast = make(chan []byte)
)

type Data struct {
	Blocks []*big.Int `json:"blocks"`
	Max *big.Int `json:"max"`
}

func main() {
	myArray := []string{
		os.Getenv("SEQUENCE_RPC"),
		os.Getenv("ALCHEMY_RPC"),
		os.Getenv("QUICKNODE_RPC"),
		os.Getenv("POLYGON_RPC"),
		os.Getenv("ANKR_RPC"),
	}
	
	var chains [][]*big.Int

	// WebSocket endpoint
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Failed to upgrade WebSocket connection:", err)
			return
		}
		defer conn.Close()

		log.Println("WebSocket client connected:", conn.RemoteAddr())

		// Add new client to the list
		clients[conn] = true

		// Handle incoming messages from the WebSocket client
		for {
			// _, msg, err := conn.ReadMessage()
			if err != nil {
				log.Println("Failed to read message from WebSocket client:", err)
				break
			}

			// Broadcast the received message to all connected clients
			message := []byte("Hello, clients!")
			broadcast <- message
			time.Sleep(2 * time.Second)
		}

		// Remove the client from the list when connection is closed
		delete(clients, conn)
	})

	// Start a goroutine to handle broadcasting messages to clients
	go func() {
		for {
			if(len(chains) > 0){

				blocks := chains[len(chains)-1]
				max := blocks[0] // Assume the first element as the initial maximum

				for _, num := range blocks {
					if num.Cmp(max) == 1 {
						max = num // Update the maximum value
					}
				}

				data := Data{
					Blocks: chains[len(chains)-1],
					Max: max,
				}

				jsonData, err := json.Marshal(data)

				// err := json.Unmarshal({msg: "yes"}, &message)
				if err != nil {
					log.Println("Failed to unmarshal JSON message:", err)
					continue
				}

				// Iterate over connected clients and send the message
				for client := range clients {
					err := client.WriteMessage(websocket.TextMessage, jsonData)
					if err != nil {
						log.Println("Failed to send message to WebSocket client:", err)
						client.Close()
						delete(clients, client)
					}
				}
				time.Sleep(2*time.Second)
			}
		}
	}()

	log.Println("WebSocket server started")
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			var blocks [] *big.Int
			for index, value := range myArray {
				log.Println("Index:", index, "Value:", value)
				client, err := ethclient.Dial(value)

				if err != nil {
					log.Println("Failed to connect to the Ethereum client:", err)
					continue
				}

				header, err := client.HeaderByNumber(context.Background(), nil)
				if err != nil {
					log.Println("Failed to retrieve the latest block header:", err)
					continue
				}

				blockNumber := header.Number
				blocks = append(blocks, blockNumber)
				log.Println("Latest block number:", blockNumber)
				log.Println("Array:", chains)
			}
			chains = append(chains, blocks)
			if len(chains) > 4 {
				chains = chains[len(chains)-4:]
			}
		}
	}()

	err := http.ListenAndServe(":5000", nil)
	if err != nil {
		log.Fatal("Failed to start HTTP server:", err)
	}
}