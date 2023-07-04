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
	"github.com/gorilla/mux"
	"github.com/gorilla/handlers"
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

var chains []BlockObject

type BlockCountData struct {
	Counts []*big.Int `json:"counts"`
	Duration int `json:"duration"`
}

type BlockObject struct {
	Blocks []*big.Int `json:"blocks"`
	MaxNumber *big.Int `json:"max"`
	Time time.Time `json:"time"`
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

func getRPC(w http.ResponseWriter, r *http.Request) {

	rpcs := []string{
		"SEQUENCE_RPC",
		"ALCHEMY_RPC",
		"QUICKNODE_RPC",
		"POLYGON_RPC",
		"ANKR_RPC",
	}

	index := getMaxIndex(chains[len(chains) - 1].Blocks)

	if index == -1 {
		index = 0;
	}

	response := map[string]interface{}{
		"provider": rpcs[index],
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func getMaxIndex(numbers []*big.Int) int {
	if len(numbers) == 0 {
		return -1
	}

	maxIndex := 0
	maxValue := numbers[0]

	for i := 1; i < len(numbers); i++ {
		if numbers[i].Cmp(maxValue) == 1 {
			maxIndex = i
			maxValue = numbers[i]
		}
	}

	return maxIndex
}

func getLastHour(w http.ResponseWriter, r *http.Request) {

	lastHour := map[string][]*big.Int{
		"0": []*big.Int{},
		"1": []*big.Int{},
		"2": []*big.Int{},
		"3": []*big.Int{},
		"4": []*big.Int{},
	}

	timeLog := []time.Time{}

	for _, blocks := range chains {
		// log.Println(blocks.Blocks[0])
		lastHour["0"] = append(lastHour["0"], blocks.Blocks[0])
		lastHour["1"] = append(lastHour["1"], blocks.Blocks[1])
		lastHour["2"] = append(lastHour["2"], blocks.Blocks[2])
		lastHour["3"] = append(lastHour["3"], blocks.Blocks[3])
		lastHour["4"] = append(lastHour["4"], blocks.Blocks[4])
		timeLog = append(timeLog, blocks.Time)
	}

	response := map[string]interface{}{
		"blocks": lastHour,
		"time": timeLog,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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

	// WebSocket endpoint
	http.HandleFunc("/counts", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Failed to upgrade WebSocket connection:", err)
			return
		}

		defer conn.Close()

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
				Duration: len(chains),
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
		for {
			if err != nil {
				log.Println("Failed to read message from WebSocket client:", err)
				break
			}
		}
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
					Time: time.Now(),
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
				Time: time.Now(),
			}

			chains = append(chains, data)
			if len(chains) > 1800 {
				chains = chains[len(chains)-1800:]
			}
		}
	}()



	go func(){

		// REST Server for most pace
		router := mux.NewRouter()
		router.HandleFunc("/api/rpc", getRPC).Methods("GET")
		router.HandleFunc("/api/1hr", getLastHour).Methods("GET")
		log.Println("REST server started")

		allowedOrigins := handlers.AllowedOrigins([]string{"http://localhost:3000"})
		allowedMethods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE"})

		log.Fatal(http.ListenAndServe(":8000", handlers.CORS(allowedOrigins, allowedMethods)(router)))
	}()

	// Websocket
	// log.Println("WebSocket server started")
	err = http.ListenAndServe(":5000", nil)
	log.Println("WebSocket server started")

	
	if err != nil {
		log.Fatal("Failed to start HTTP server:", err)
	}
}