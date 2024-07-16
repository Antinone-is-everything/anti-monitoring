package checkHost

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// Response represents the structure of the JSON response
type Response map[string][]struct {
	Address string  `json:"address"`
	Time    float64 `json:"time,omitempty"`
	Error   string  `json:"error,omitempty"`
}

func CheckHost(ctx context.Context, requestURL string) (map[string]interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		log.Printf("| -- client: could not create request: %s\n", err)
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	client := http.Client{
		Timeout: 15 * time.Second,
	}

	res, err := client.Do(req)
	if err != nil {
		//log.Printf("| -- client: error making http request: %s\n", err)
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non-200 status code: %d", res.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding JSON API call: %v", err)
	}

	return result, nil
}

func CheckResultTry(requestId string) (Response, error) {
	requestURL := fmt.Sprintf(os.Getenv("API_BLOCK_RESULT"), requestId)
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("could not create request: %v", err)
	}
	req.Header.Set("Accept", "application/json")

	client := http.Client{
		Timeout: 10 * time.Second,
	}

	// Maximum number of retries
	maxRetries := 3

	for retries := 0; retries < maxRetries; retries++ {
		res, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("error making HTTP request: %v", err)
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("non-200 status code: %d", res.StatusCode)
		}

		var response Response
		if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
			//log.Printf("Error decoding JSON response: %v. Retrying... (%d/%d)", err, retries+1, maxRetries)
			time.Sleep(3 * time.Second)
			continue // Retry the request
		} else if response["ir1.node.check-host.net"] == nil {
			//log.Printf("Error Null JSON response: %v. Retrying... (%d/%d)", response, retries+1, maxRetries)
			time.Sleep(3 * time.Second)
			continue // Retry the request
		}
		//fmt.Println("Response :", response)
		// If decoding is successful, return the response and nil error
		return response, nil
	}

	// If all retries are used up, return an error
	return nil, fmt.Errorf("maximum retries reached; unable to decode JSON response")
}
