package main

import (
		"crypto/sha256"
    	"encoding/hex"
		"database/sql"
		"encoding/json"
		"log"
		"net"
		"time"

		_ "github.com/denisenkom/go-mssqldb" // Importa o driver SQL Server
		"golang.org/x/net/context"
		"github.com/redis/go-redis/v9"
	)

// Definição da estrutura QueryRequest
type QueryRequest struct {
	Query string `json:"query"`
}

// Função para gerar hash da query
func hashQuery(query string) string {
    h := sha256.New()
    h.Write([]byte(query))
    return hex.EncodeToString(h.Sum(nil))
}

func handleConnection(conn net.Conn, db *sql.DB, rdb *redis.Client) {
	defer conn.Close()

	var req QueryRequest
	if err := json.NewDecoder(conn).Decode(&req); err != nil {
		log.Println("Error reading query:", err)
		return
	}

	query := req.Query
	log.Printf("Received query: %s\n", query)

	// Gerar o hash da query para usar como chave no Redis
	hashKey := hashQuery(query)

	// Check Redis cache
	cachedResult, err := rdb.Get(context.Background(), hashKey).Result()
	if err == nil {
		conn.Write([]byte(cachedResult))
		return
	}

	// Execute query on SQL Server
	rows, err := db.Query(query)
	if err != nil {
		log.Println("Error executing query:", err)
		return
	}
	defer rows.Close()

	// Process rows and cache the result
	var result []map[string]interface{}
	columns, err := rows.Columns()
	if err != nil {
		log.Println("Error getting columns:", err)
		return
	}

	for rows.Next() {
		columnPointers := make([]interface{}, len(columns))
		for i := range columnPointers {
			columnPointers[i] = new(interface{})
		}
		if err := rows.Scan(columnPointers...); err != nil {
			log.Println("Error scanning row:", err)
			return
		}
		rowData := make(map[string]interface{})
		for i, col := range columns {
			val := columnPointers[i].(*interface{})
			rowData[col] = *val
		}
		result = append(result, rowData)
	}

	// Check if there are results
	if len(result) == 0 {
		log.Println("No results found for query:", query)
		return
	}

	// Store result in Redis usando o hash como chave
	resultJSON, err := json.Marshal(result)
	if err != nil {
		log.Println("Error marshaling result to JSON:", err)
		return
	}

	if err := rdb.Set(context.Background(), hashKey, resultJSON, time.Minute*5).Err(); err != nil {
		log.Println("Error storing result in Redis:", err)
		return
	}

	// Return result to client
	conn.Write(resultJSON)
}

// Função principal
func main() {
	// Configurações do banco de dados e do Redis
	connString := "server=um_servidor ;user id=um_usuario ;password=uma_senha ;database=a_base;"
	db, err := sql.Open("sqlserver", connString)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// Cria um listener para conexões
	ln, err := net.Listen("tcp", ":1433")
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	log.Println("Server listening on :1433")
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}
		go handleConnection(conn, db, rdb)
	}
}