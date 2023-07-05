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
	"github.com/twilio/twilio-go"
	twilioApi "github.com/twilio/twilio-go/rest/api/v2010"
	"github.com/joho/godotenv"
	"github.com/gorilla/mux"
	"github.com/gorilla/handlers"
	"database/sql"
	_ "github.com/lib/pq"
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

type Result struct {
    Checks struct {
        LastBlockNum *big.Int `json:"lastBlockNum"`
    } `json:"checks"`
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

func createTableIfNotExists() {
	db, err := sql.Open("postgres", "postgres://"+os.Getenv("PG_USERNAME")+":"+os.Getenv("PG_PASSWORD")+"@localhost/monitor?sslmode=disable")

	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Prepare the SQL query to check if the table exists
	query1 := `
		SELECT EXISTS (
			SELECT 1
			FROM   pg_tables
			WHERE  schemaname = 'public'
			AND    tablename = 'occurences'
		);
	`

	// Execute the query and retrieve the result
	var exists1 bool
	err = db.QueryRow(query1).Scan(&exists1)
	if err != nil {
		log.Fatal(err)
	}

	// Check the result
	if exists1 {
		log.Println("Table of block occurrences exists!")
	} else {
		log.Println("Table does not exist!")

		createTableQuery := `
			CREATE TABLE IF NOT EXISTS occurences (
				id SERIAL PRIMARY KEY,
				behind BOOLEAN
			);
		`
		// Execute the create table query
		_, err = db.Exec(createTableQuery)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Table added!")
	}

	// Prepare the SQL query to check if the table exists
	query2 := `
		SELECT EXISTS (
			SELECT 1
			FROM   pg_tables
			WHERE  schemaname = 'public'
			AND    tablename = 'numbers'
		);
	`

	// Execute the query and retrieve the result
	var exists2 bool
	err = db.QueryRow(query2).Scan(&exists2)
	if err != nil {
		log.Fatal(err)
	}

	// Check the result
	if exists2 {
		log.Println("Table of phone numbers exists!")
	} else {
		log.Println("Table does not exist!")

		createTableQuery := `
			CREATE TABLE IF NOT EXISTS numbers (
				id SERIAL PRIMARY KEY,
				number VARCHAR
			);
		`
		// Execute the create table query
		_, err = db.Exec(createTableQuery)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Table added!")
	}
}

func addPhoneNumber(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	phoneNumber := vars["phone_number"]

	db, err := sql.Open("postgres", "postgres://"+os.Getenv("PG_USERNAME")+":"+os.Getenv("PG_PASSWORD")+"@localhost/monitor?sslmode=disable")

	// Prepare the SQL statement
	query := "INSERT INTO numbers (number) VALUES ($1);"

	// Execute the query
	_, err = db.Exec(query, phoneNumber)
	if err != nil {
		log.Fatal(err)
	}

	// Use the parameter value as needed
	log.Println("Added phone Number: %s", phoneNumber)
}

func getSequenceIndexerLatest() *big.Int {
	response, err := http.Get("https://polygon-indexer.sequence.app/status")

	if err != nil {
		log.Println(err)
		log.Println("Error occurred while making the HTTP request:", err)
		return big.NewInt(-1)
	}

	defer response.Body.Close()

	// Parse JSON response
	var result Result
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		log.Println("Error occurred while decoding JSON:", err)
		return big.NewInt(-1)
	}

	// Access the parsed data
	lastBlockNum := result.Checks.LastBlockNum

	return lastBlockNum
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
	http.HandleFunc("/live", func(w http.ResponseWriter, r *http.Request) {
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

			// make sequence indexer status request
			lastBlockNum := getSequenceIndexerLatest()
			blocks = append(blocks, lastBlockNum)

			// get max
			max := blocks[0]

			for _, num := range blocks {
				if num.Cmp(max) == 1 {
					max = num
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
		router.HandleFunc("/api/rpc/latest", getRPC).Methods("GET")
		router.HandleFunc("/api/1hr", getLastHour).Methods("GET")
		router.HandleFunc("/api/notify/{phone_number}", addPhoneNumber).Methods("GET")

		log.Println("REST server started")

		allowedOrigins := handlers.AllowedOrigins([]string{"http://137.220.54.108:2000","http://localhost:3000"})
		allowedMethods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE"})

		log.Fatal(http.ListenAndServe(":8000", handlers.CORS(allowedOrigins, allowedMethods)(router)))
	}()

	go func(){
		// create a table if one does not exist
		createTableIfNotExists()

		// connect to db
		db, _ := sql.Open("postgres", "postgres://"+os.Getenv("PG_USERNAME")+":"+os.Getenv("PG_PASSWORD")+"@localhost/monitor?sslmode=disable")

		for {
			// Prepare the SQL statement
			query := "SELECT * FROM occurences ORDER BY id DESC LIMIT 1;"

			// Execute the query
			row := db.QueryRow(query)

			// Scan the result into variables
			var id string
			var behind string
			// Add more variables to match the columns in your table

			err := row.Scan(&id, &behind)
			if err != nil {
				if err == sql.ErrNoRows {
					log.Println("No rows found in the result set.")
					// Handle the absence of rows in the result set
				} else {
					log.Println("Error occurred while scanning rows:", err)
					// Handle other errors that may have occurred
				}
			} else {
				// Process the scanned values
				if(len(chains) > 0 && behind == "false"){
					// Call your function here
					log.Println("running.")

					diff := big.NewInt(0)
					threshold := big.NewInt(20)

					diff.Sub(chains[len(chains)-1].MaxNumber, chains[len(chains)-1].Blocks[0])

					if(diff.Cmp(threshold) == 1){
						if err != nil {
							log.Fatal(err)
						}
						defer db.Close()

						// Prepare the SQL statement
						query := "INSERT INTO occurences (behind) VALUES ($1);"

						// Execute the query
						_, err = db.Exec(query, true)
						if err != nil {
							log.Fatal(err)
						}

						accountSid := os.Getenv("TWILIO_ACCOUNT_SID")
						authToken := os.Getenv("TWILIO_AUTH_TOKEN")
			
						client := twilio.NewRestClientWithParams(twilio.ClientParams{
							Username: accountSid,
							Password: authToken,
						})

						// get all phone numbers
						query1 := "SELECT * FROM numbers;"

						rows, err := db.Query(query1)
						if err != nil {
							log.Fatal(err)
						}
						defer rows.Close()

						// Iterate over the rows
						for rows.Next() {
							// Define variables to store the row values
							var id int
							var number string
							// Add more variables to match the columns in the table

							// Scan the row values into the variables
							err := rows.Scan(&id, &number)
							if err != nil {
								log.Fatal(err)
							}

							params := &twilioApi.CreateMessageParams{}
							params.SetTo(number)
							params.SetFrom("+16727020100")
							params.SetBody("Sequence Node Gateway is behind by " + diff.String())
				
							resp, err := client.Api.CreateMessage(params)
							if err != nil {
								log.Println("Error sending SMS message: " + err.Error())
							} else {
								response, _ := json.Marshal(*resp)
								log.Println("Response: " + string(response))
							}

							log.Println("Boolean value inserted successfully!")

							// Process the row data
							// Add more processing logic as per your requirements
						}

						// Check for any errors encountered during iteration
						err = rows.Err()
						if err != nil {
							log.Fatal(err)
						}

						go func() {
							time.Sleep(60 * 5 * time.Second)
							// Prepare the SQL statement
							query := "INSERT INTO occurences (behind) VALUES ($1);"

							// Execute the query
							_, err = db.Exec(query, false)
							if err != nil {
								log.Fatal(err)
							}
							log.Println("resetting")
						}()
					}

					// Sleep for 10 seconds
				}
			}


			time.Sleep(2 * time.Second)
			
		}

	}()

	// Websocket
	err = http.ListenAndServe(":5000", nil)
	log.Println("WebSocket server started")
	
	if err != nil {
		log.Fatal("Failed to start HTTP server:", err)
	}
}