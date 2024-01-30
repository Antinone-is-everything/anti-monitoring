package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"monitoring/db"
	"net/http"
	"time"
)

func monitorServer(url string) bool {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	http.DefaultClient.Timeout = 10 * time.Second
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
	resp, err := http.Get(fmt.Sprintf("https://api.antinone.xyz/api/instance?action=%s&region=%s&name=%s&secret=%d&profile=%s", action, region, name, timestamp, profile))
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

	serversMonitoring := []db.Server{}

	dbs, err := db.ConnectToDatabase()
	if err != nil {
		fmt.Println("Failed to connect to the database:", err)
	}

	err = db.CreateServerTable(dbs)
	if err != nil {
		fmt.Println("Failed to connect to the database:", err)
	}

	// // Open our jsonFile
	// serversList, err := ioutil.ReadFile("./servers.json")
	// // if we os.Open returns an error then handle it
	// if err != nil {
	// 	log.Printf("| Read Json File Error %s", err)
	// }
	// // we unmarshal our byteArray which contains our
	// // jsonFile's content into 'users' which we defined above
	// err = json.Unmarshal(serversList, &serversMonitoring)
	// if err != nil {
	// 	log.Println("| Unmarshal error:", err)
	// }

	// for y := range serversMonitoring {
	// 	id := db.InsertServer(dbs, serversMonitoring[y])
	// 	log.Printf("| Inserted server with ID: %d\n", id)
	// }

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

	fmt.Println(serversMonitoring)

	// Main loop to monitor the programs

	for {
		for i := range serversMonitoring {
			if serversMonitoring[i].Active {
				log.Printf("| %d - check service %s - errorCount: %d \n", i, serversMonitoring[i].ServerDomain, serversMonitoring[i].ErrorCount)
				healthCheckUrl := fmt.Sprintf("https://%s:%d/%s%s", serversMonitoring[i].ServerDomain, serversMonitoring[i].ServerPort, serversMonitoring[i].ApiKey, serversMonitoring[i].HealthCheck)

				if monitorServer(healthCheckUrl) {

					if serversMonitoring[i].ErrorCount >= serversMonitoring[i].TrigerCount {
						if time.Now().Unix()-serversMonitoring[i].ResetTime > 240 {
							if serverAction("reset", serversMonitoring[i].ServerRegion, serversMonitoring[i].ServerName, serversMonitoring[i].Profile) {

								serversMonitoring[i].ResetTime = time.Now().Unix()
								log.Printf("| %d - Program reset successful server %s\n", i, serversMonitoring[i].ServerDomain)

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
				log.Printf("| %d - Disable monitoring service %s\n", i, serversMonitoring[i].ServerDomain)
			}
		}
		log.Println("| <<<---------------END--------------->>> |")
		time.Sleep(20 * time.Second) // Sleep for 60 seconds before checking again
	}
}
