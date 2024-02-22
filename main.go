package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"monitoring/alert"
	"monitoring/db"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type AlertMessage struct {
	MessageID    int
	ServerDomain string
}

func monitorServer(url string) bool {
	http_timeout, err := strconv.Atoi(os.Getenv("HTTP_TIMEOUT"))
	if err != nil {
		log.Printf("| Error ENV HTTP_TIMEOUT: %v\n", err)
		return true
	}
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	http.DefaultClient.Timeout = time.Duration(http_timeout) * time.Second
	response, err := http.Get(url)
	if err != nil {
		log.Printf("| Error Get monitoring service: %v\n", err)
		return true
	}
	if response.StatusCode == 200 {
		return false
	} else {
		return true
	}
}

func serverAction(action, region, name, profile string) bool {
	timestamp := time.Now().Unix()
	resp, err := http.Get(fmt.Sprintf(os.Getenv("API_RESET_SRV"), action, region, name, timestamp, profile))
	if err != nil {
		log.Printf("| Error Get Api call: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		log.Printf("| Error decoding JSON Api call: %v\n", err)
		return false
	}
	log.Printf("| Call back api service reset server : %v\n", result)
	if resp.StatusCode == 200 {
		return true
	} else {
		return false
	}

}

func main() {
	err := godotenv.Load("./vars/.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	loop_time, err := strconv.Atoi(os.Getenv("LOOP_TIME"))
	if err != nil {
		log.Printf("| Error ENV LOOP_TIME: %v\n", err)
	}
	cfg := db.DBConfig{
		Host:     os.Getenv("DB_HOST"),
		Port:     5432,
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASS"),
		DBName:   os.Getenv("DB_NAME"),
	}

	telegramToken := alert.NewTelegramConfig(os.Getenv("TELEGRAM_BOT_TOKEN"), os.Getenv("TELEGRAM_API"))
	// adminID, _ := strconv.ParseInt(os.Getenv("TELEGRAM_ADMIN_ID"), 10, 64)
	adminID, err := strconv.ParseInt(os.Getenv("TELEGRAM_ADMIN_ID"), 10, 64) //user Admin telegram ID
	if err != nil {
		fmt.Printf("| Failed load AdminID  %d of type %T , Error is %v\n", adminID, adminID, err)
	}
	var messageID int = 0

	var ServerDomain string
	var ServerPort int
	var ApiKey string
	var HealthCheck string
	var ServerRegion string
	var ServerName string
	var ErrorCount int
	var TrigerCount int
	var ResetTime int64
	var Profile string
	var Active bool

	serversMonitoring := []db.Server{}

	dbs, err := db.ConnectToDatabase(&cfg)
	if err != nil {
		fmt.Println("Failed to connect to the database:", err)
	}

	err = dbs.Ping()
	if err == nil {
		log.Println("| Read Config From Database ")
		err := db.CreateServerTable(dbs)
		if err != nil {
			fmt.Println("Failed to connect to the database:", err)
		}

		rows, err := dbs.Query("SELECT ServerDomain, ServerPort, ApiKey, HealthCheck, ServerRegion, ServerName, ErrorCount, TrigerCount, ResetTime, Profile, Active FROM server")
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		for rows.Next() {
			err := rows.Scan(&ServerDomain, &ServerPort, &ApiKey, &HealthCheck, &ServerRegion, &ServerName, &ErrorCount, &TrigerCount, &ResetTime, &Profile, &Active)
			if err != nil {
				log.Fatal(err)
			}

			serversMonitoring = append(serversMonitoring, db.Server{ServerDomain, ServerPort, ApiKey, HealthCheck, ServerRegion, ServerName, ErrorCount, TrigerCount, ResetTime, Profile, Active})
		}

	} else {

		log.Println("| Read Config From JSON file ")
		// Open our jsonFile
		serversList, err := os.ReadFile("./servers.json")
		// if we os.Open returns an error then handle it
		if err != nil {
			log.Printf("| Read Json File Error %s", err)
		}
		// we unmarshal our byteArray which contains our
		// jsonFile's content into 'users' which we defined above
		err = json.Unmarshal(serversList, &serversMonitoring)
		if err != nil {
			log.Println("| Unmarshal error:", err)
		}

		// for y := range serversMonitoring {
		// 	id := db.InsertServer(dbs, serversMonitoring[y])
		// 	log.Printf("| Inserted server with ID: %d\n", id)
		// }

	}

	fmt.Println(serversMonitoring)
	var alertMessages []AlertMessage
	for y := range serversMonitoring {
		alertMessages = append(alertMessages, AlertMessage{
			MessageID:    messageID,
			ServerDomain: serversMonitoring[y].ServerDomain,
		})
	}

	// Main loop to monitor the programs

	for {
		for i := range serversMonitoring {
			if serversMonitoring[i].Active {
				if alertMessages[i].ServerDomain == serversMonitoring[i].ServerDomain {
					//log.Printf("| Alert Domain is %s = Server Domain is %s", alertMessages[i].ServerDomain, serversMonitoring[i].ServerDomain)

					log.Printf("| %d - check service %s - errorCount: %d \n", i, serversMonitoring[i].ServerDomain, serversMonitoring[i].ErrorCount)
					healthCheckUrl := fmt.Sprintf("https://%s:%d/%s%s", serversMonitoring[i].ServerDomain, serversMonitoring[i].ServerPort, serversMonitoring[i].ApiKey, serversMonitoring[i].HealthCheck)

					if monitorServer(healthCheckUrl) {

						if alertMessages[i].MessageID != 0 {
							err := alert.DeleteMesg(telegramToken, alertMessages[i].MessageID, adminID)
							if err != nil {
								log.Printf("| Message Delete fail  %v", err)
							} else {
								log.Printf("| Message Delete successfully with ID: %d\n", alertMessages[i].MessageID)
							}
							alertMessages[i].MessageID = 0
						}

						problemMsg := fmt.Sprintf("<b>Antinone Monitoring🤖🗣</b>\n"+
							"⚠️🚽 Server Outline %s is dead ☠️⚰️‼️\n"+
							"📊 ErrCount: %d\n"+
							"🧭 Status: %d\n"+
							"🖥 ServerName: %s\n"+
							"🌎 Region: %s\n"+
							"⏰ Time: %s",
							serversMonitoring[i].ServerDomain, serversMonitoring[i].ErrorCount, 000, serversMonitoring[i].ServerName, serversMonitoring[i].ServerRegion, time.Now().Format("2006-01-02 15:04:05"))
						alertMessages[i].MessageID, err = alert.SendMesg(telegramToken, problemMsg, adminID)
						if err != nil {
							log.Printf("| Message sent fail  %v", err)
						} else {
							log.Printf("| Message sent successfully with ID: %d\n", alertMessages[i].MessageID)
						}

						if serversMonitoring[i].ErrorCount >= serversMonitoring[i].TrigerCount {
							if time.Now().Unix()-serversMonitoring[i].ResetTime > 240 {
								if serverAction("reset", serversMonitoring[i].ServerRegion, serversMonitoring[i].ServerName, serversMonitoring[i].Profile) {

									serversMonitoring[i].ResetTime = time.Now().Unix()
									log.Printf("| %d - Program reset successful server %s\n", i, serversMonitoring[i].ServerDomain)
									resetMsg := fmt.Sprintf("<b>Antinone Monitoring🤖🗣</b>\n"+
										"⚠️♻️ Server Outline %s restarted 🔎\n"+
										"🖥 ServerName: %s\n"+
										"🌎 Region: %s\n"+
										"⏰ ResetTime: %s",
										serversMonitoring[i].ServerDomain, serversMonitoring[i].ServerName, serversMonitoring[i].ServerRegion, time.Now().Format("2006-01-02 15:04:05"))
									_, err = alert.SendMesg(telegramToken, resetMsg, adminID)
									if err != nil {
										log.Printf("| Message sent fail  %v", err)
									}
								} else {
									log.Printf("| %d - Program reset failed server %s\n", i, serversMonitoring[i].ServerDomain)
								}
							} else {
								log.Printf("| %d - The server %s was reset %d minutes ago\n", i, serversMonitoring[i].ServerDomain, (time.Now().Unix()-serversMonitoring[i].ResetTime)/60)
							}

						}
						serversMonitoring[i].ErrorCount++
					} else {
						serversMonitoring[i].ErrorCount = 0
					}
				} else {
					log.Printf("| Alert Domain is %s != Server Domain is %s", alertMessages[i].ServerDomain, serversMonitoring[i].ServerDomain)
				}
			} else {
				log.Printf("| %d - Disable monitoring service %s\n", i, serversMonitoring[i].ServerDomain)
			}
		}
		log.Println("| <<<---------------END--------------->>> |")

		time.Sleep(time.Duration(loop_time) * time.Second) // Sleep for 60 seconds before checking again
	}
}
