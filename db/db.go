package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

// const (
// 	host     = "localhost"
// 	port     = 5432
// 	user     = "antinone_user"
// 	password = "hey_you_are_antinone"
// 	dbname   = "antinone_db"
// )

type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

type Server struct {
	ServerDomain string
	ServerPort   int
	ApiKey       string
	HealthCheck  string
	ServerRegion string
	ServerName   string
	ErrorCount   int
	TrigerCount  int
	ResetTime    int64
	Profile      string
	Active       bool
}

func ConnectToDatabase(connection *DBConfig) (*sql.DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		connection.Host, connection.Port, connection.User, connection.Password, connection.DBName)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}
	log.Println("| Successfully connected!")
	return db, nil

}

/* Table Users
- ID
- TelegramID
- UserName
- FirstName
- LastName
- InviteCode/Referral
- Status (active/deactive)
- Score
- password
- Date Created
*/

/* Table Server
- ID
- ServerDomain
- ServerPort
- ApiKey
- HealthCheck
- ServerRegion
- ServerName
- ErrorCount
- TrigerCount
- ResetTimestamp
- Profile
- Active
*/

func CreateServerTable(db *sql.DB) error {
	query := `CREATE TABLE IF NOT EXISTS server (
		id SERIAL PRIMARY KEY,
		ServerDomain varchar(255) NOT NULL UNIQUE,
		ServerPort numeric DEFAULT 12258 ,
		ApiKey varchar(255) NOT NULL UNIQUE,
		HealthCheck varchar(255) DEFAULT '/server',
		ServerRegion varchar(255),
		ServerName varchar(255),
		ErrorCount int DEFAULT 0,
		TrigerCount int DEFAULT 3,
		ResetTime bigint,
		profile varchar(25) DEFAULT 'vahid',
		Active BOOLEAN DEFAULT true
	)`

	resp, err := db.Exec(query)
	if err != nil {
		log.Fatal(err)
		return err
	}
	log.Printf("| Create Table Server : %v", resp)

	return nil
}

func InsertServer(db *sql.DB, server Server) int {
	query := `INSERT INTO server (serverdomain , apikey, serverregion, servername, ResetTime )
	VALUES ($1, $2, $3, $4, trunc(extract(epoch from now()))) RETURNING id`
	// log.Printf("| inpute data is : %s", server)
	var pk int
	err := db.QueryRow(query, server.ServerDomain, server.ApiKey, server.ServerRegion, server.ServerName).Scan(&pk)
	if err != nil {
		log.Printf("| Error Insert to Table : %v", err)
	}
	return pk
}
