package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// Program represents a monitored program
type Server struct {
	ServerDomain   string `json:"ServerDomain"`
	ServerPort     int    `json:"ServerPort"`
	ApiKey         string `json:"ApiKey"`
	HealthCheck    string `json:"HealthCheck"`
	ServerRegion   string `json:"ServerRegion"`
	ServerName     string `json:"ServerName"`
	ErrorCount     int    `json:"ErrorCount"`
	TrigerCount    int    `json:"TrigerCount"`
	ResetTimestamp int64  `json:"ResetTimestamp"`
	Profile        string `json:"Profile"`
	Active         bool   `json:"Active"`
}

func monitorServer(url string) bool {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	http.DefaultClient.Timeout = 5 * time.Second
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

	// serversMonitoring := []Server{
	// 	{
	// 		ServerDomain:   "example.antinone.xyz",
	// 		ServerPort:     8080,
	// 		ApiKey:         "abcdefgh12344567889",
	// 		HealthCheck:    "/server",
	// 		ServerRegion:   "ca-central-1",
	// 		ServerName:     "Ubuntu-server-name",
	// 		ErrorCount:     0,
	// 		TrigerCount:    3,
	// 		ResetTimestamp: 0,
	// 		Profile:        "profile-name",
	// 		Active:         true,
	// 	},
	// }

	// Open our jsonFile
	serversList, err := ioutil.ReadFile("./servers.json")
	// if we os.Open returns an error then handle it
	if err != nil {
		log.Printf("| Read Json File Error %s", err)
	}

	// we initialize our Users array
	var serversMonitoring []Server

	// we unmarshal our byteArray which contains our
	// jsonFile's content into 'users' which we defined above
	err = json.Unmarshal(serversList, &serversMonitoring)
	if err != nil {
		log.Println("| Unmarshal error:", err)
	}

	// Main loop to monitor the programs

	for {
		for i := range serversMonitoring {
			if serversMonitoring[i].Active {
				log.Printf("| %d - check service %s - errorCount: %d \n", i, serversMonitoring[i].ServerDomain, serversMonitoring[i].ErrorCount)
				healthCheckUrl := fmt.Sprintf("https://%s:%d/%s%s", serversMonitoring[i].ServerDomain, serversMonitoring[i].ServerPort, serversMonitoring[i].ApiKey, serversMonitoring[i].HealthCheck)

				if monitorServer(healthCheckUrl) {
					serversMonitoring[i].ErrorCount++
					if serversMonitoring[i].ErrorCount >= serversMonitoring[i].TrigerCount {
						if time.Now().Unix()-serversMonitoring[i].ResetTimestamp > 240 {
							if serverAction("reset", serversMonitoring[i].ServerRegion, serversMonitoring[i].ServerName, serversMonitoring[i].Profile) {

								serversMonitoring[i].ResetTimestamp = time.Now().Unix()
								log.Printf("| Program reset successful server %s\n", serversMonitoring[i].ServerDomain)

							} else {
								log.Printf("| Program reset failed server %s\n", serversMonitoring[i].ServerDomain)
							}
						} else {
							log.Printf("| The server %s was reset %d minutes ago\n", serversMonitoring[i].ServerDomain, (time.Now().Unix()-serversMonitoring[i].ResetTimestamp)/60)
						}

					}
				} else {
					serversMonitoring[i].ErrorCount = 0
				}
			} else {
				log.Printf("| Disable monitoring service %s\n", serversMonitoring[i].ServerDomain)
			}
		}
		log.Println("| <<<---------------END--------------->>> |")
		time.Sleep(20 * time.Second) // Sleep for 60 seconds before checking again
	}
}
