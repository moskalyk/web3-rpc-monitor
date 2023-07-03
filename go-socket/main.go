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
	"github.com/joho/godotenv"
)

var (
	upgrader  = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			// allowedOrigin := "http://localhost:3000"
			return true
		},
	}
	clients   = make(map[*websocket.Conn]bool)
	clientsBlockCounts   = make(map[*websocket.Conn]bool)
	broadcast = make(chan []byte)
)

type BlockCountData struct {
	Counts []*big.Int `json:"counts"`
}

type BlockObject struct {
	Blocks []*big.Int `json:"blocks"`
	MaxNumber *big.Int `json:"max"`
}

func calculateDifferences(blockObjects []BlockObject) []*big.Int {
	differences := make([]*big.Int, 0)

	// init new big.int
	diffs_sequence := big.NewInt(0)
	diffs_alchemy := big.NewInt(0)
	diffs_quicknode := big.NewInt(0)
	diffs_polygon := big.NewInt(0)
	diffs_ankr := big.NewInt(0)

	for _, blockObj := range blockObjects {
		var diffs []*big.Int

		// init new big.int
		diffs_max_sequence := big.NewInt(0)
		diffs_max_alchemy := big.NewInt(0)
		diffs_max_quicknode := big.NewInt(0)
		diffs_max_polygon := big.NewInt(0)
		diffs_max_ankr := big.NewInt(0)

		// sub max - block
		diffs_max_sequence.Sub(blockObj.MaxNumber, blockObj.Blocks[0])
		diffs_max_alchemy.Sub(blockObj.MaxNumber, blockObj.Blocks[1])
		diffs_max_quicknode.Sub(blockObj.MaxNumber, blockObj.Blocks[2])
		diffs_max_polygon.Sub(blockObj.MaxNumber, blockObj.Blocks[3])
		diffs_max_ankr.Sub(blockObj.MaxNumber, blockObj.Blocks[4])

		// diffs += max_diffs
		diffs_sequence.Add(diffs_sequence, diffs_max_sequence)
		diffs_alchemy.Add(diffs_alchemy, diffs_max_alchemy)
		diffs_quicknode.Add(diffs_quicknode, diffs_max_quicknode)
		diffs_polygon.Add(diffs_polygon, diffs_max_polygon)
		diffs_ankr.Add(diffs_ankr, diffs_max_ankr)

		// create array
		diffs = append(diffs, diffs_sequence)
		diffs = append(diffs, diffs_alchemy)
		diffs = append(diffs, diffs_quicknode)
		diffs = append(diffs, diffs_polygon)
		diffs = append(diffs, diffs_ankr)

		// return
		differences = diffs
	}

	return differences
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	myArray := []string{
		os.Getenv("SEQUENCE_RPC"),
		os.Getenv("ALCHEMY_RPC"),
		os.Getenv("QUICKNODE_RPC"),
		os.Getenv("POLYGON_RPC"),
		os.Getenv("ANKR_RPC"),
	}
	
	var chains []BlockObject

	// WebSocket endpoint
	http.HandleFunc("/counts", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Failed to upgrade WebSocket connection:", err)
			return
		}

		defer conn.Close()

		log.Println("WebSocket client connected block counts:", conn.RemoteAddr())

		// Add new client to the list
		clientsBlockCounts[conn] = true

		for {
			if err != nil {
				log.Println("Failed to read message from WebSocket client:", err)
				break
			}
		}
		delete(clientsBlockCounts, conn)
	})

	go func() {
		for {

			differences := calculateDifferences(chains)

			data := BlockCountData{
				Counts: differences,
			}

			jsonData, err := json.Marshal(data)
			if err != nil {

			}

			// Iterate over connected clients and send the message
			for client := range clientsBlockCounts {
				err := client.WriteMessage(websocket.TextMessage, jsonData)
				if err != nil {
					log.Println("Failed to send message to WebSocket client:", err)
					client.Close()
					delete(clientsBlockCounts, client)
				}
			}

			time.Sleep(2*time.Second)
		}
	}()


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
		delete(clients, conn)
	})

	// Start a goroutine to handle broadcasting messages to clients
	go func() {
		for {
			if(len(chains) > 0){

				blocks := chains[len(chains)-1].Blocks
				max := blocks[0] // Assume the first element as the initial maximum

				for _, num := range blocks {
					if num.Cmp(max) == 1 {
						max = num // Update the maximum value
					}
				}

				data := BlockObject{
					Blocks: chains[len(chains)-1].Blocks,
					MaxNumber: max,
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

	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			var blocks [] *big.Int
			for _, value := range myArray {
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
			}

			max := blocks[0] // Assume the first element as the initial maximum

			for _, num := range blocks {
				if num.Cmp(max) == 1 {
					max = num // Update the maximum value
				}
			}

			data := BlockObject{
				Blocks: blocks,
				MaxNumber: max,
			}

			chains = append(chains, data)
			if len(chains) > 1800 {
				chains = chains[len(chains)-1800:]
			}
		}
	}()

	err = http.ListenAndServe(":5000", nil)
	log.Println("WebSocket server started")
	
	if err != nil {
		log.Fatal("Failed to start HTTP server:", err)
	}
}