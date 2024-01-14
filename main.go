package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Program represents a monitored program
type Server struct {
	ServerDomain   string
	ServerPort     int
	ApiKey         string
	HealthCheck    string
	ServerRegion   string
	ServerName     string
	ErrorCount     int
	TrigerCount    int
	ResetTimestamp int64
	Profile        string
	Active         bool
}

func monitorServer(url string) bool {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	http.DefaultClient.Timeout = 5 * time.Second
	response, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
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
		fmt.Printf("Error: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		fmt.Printf("Error decoding JSON: %v\n", err)
		return false
	}
	fmt.Printf("Reset Status Server : %v\n", result)
	if resp.StatusCode == 200 {
		return true
	} else {
		return false
	}

}

func main() {

	serversMonitoring := []Server{
		{
			ServerDomain:   "example.antinone.xyz",
			ServerPort:     8080,
			ApiKey:         "abcdefgh12344567889",
			HealthCheck:    "/server",
			ServerRegion:   "ca-central-1",
			ServerName:     "Ubuntu-server-name",
			ErrorCount:     0,
			TrigerCount:    3,
			ResetTimestamp: 0,
			Profile:        "profile-name",
			Active:         true,
		},
	}

	// Main loop to monitor the programs

	for {
		for i := range serversMonitoring {
			if serversMonitoring[i].Active {
				fmt.Printf("check service %s errorCount : %d \n", serversMonitoring[i].ServerDomain, serversMonitoring[i].ErrorCount)
				healthCheckUrl := fmt.Sprintf("https://%s:%d/%s%s", serversMonitoring[i].ServerDomain, serversMonitoring[i].ServerPort, serversMonitoring[i].ApiKey, serversMonitoring[i].HealthCheck)

				if monitorServer(healthCheckUrl) {
					serversMonitoring[i].ErrorCount++
					if serversMonitoring[i].ErrorCount >= serversMonitoring[i].TrigerCount {
						if time.Now().Unix()-serversMonitoring[i].ResetTimestamp > 500 {
							if serverAction("reset", serversMonitoring[i].ServerRegion, serversMonitoring[i].ServerName, serversMonitoring[i].Profile) {

								serversMonitoring[i].ResetTimestamp = time.Now().Unix()
								fmt.Printf("Program reset successful for %s\n", serversMonitoring[i].ServerDomain)

							} else {
								fmt.Printf("Program reset failed for %s\n", serversMonitoring[i].ServerDomain)
							}
						} else {
							fmt.Printf("The server %s was reset %d minutes ago\n", serversMonitoring[i].ServerDomain, (time.Now().Unix()-serversMonitoring[i].ResetTimestamp)/60)
						}

					}
				} else {
					serversMonitoring[i].ErrorCount = 0
				}
			} else {
				fmt.Printf("Disable monitoring service %s\n", serversMonitoring[i].ServerDomain)
			}
		}
		time.Sleep(10 * time.Second) // Sleep for 60 seconds before checking again
	}
}
