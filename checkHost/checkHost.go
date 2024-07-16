package checkHost

import (
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

func CheckHost(requestURL string) (map[string]interface{}, error) {

	//requestURL := fmt.Sprintf(os.Getenv("API_BLOCK_CHECK"), serverAddress, 10901, os.Getenv("LOCATION_IR1"), os.Getenv("LOCATION_IR2"), os.Getenv("LOCATION_IR3"), os.Getenv("LOCATION_IR4"))
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		log.Printf("| client: could not create request: %s\n", err)
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	client := http.Client{
		Timeout: 5 * time.Second,
	}

	res, err := client.Do(req)
	if err != nil {
		log.Printf("| client: error making http request: %s\n", err)
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non-200 status code: %d", res.StatusCode)
	}

	// // Read response body
	// body, err := ioutil.ReadAll(res.Body)
	// if err != nil {
	// 	return nil, fmt.Errorf("error reading response body: %s", err)
	// }

	// // Print response body
	// fmt.Println("Response Body:", string(body))

	// // Attempt to decode response body as JSON
	// var result map[string]interface{}
	// if err := json.Unmarshal(body, &result); err != nil {
	// 	return nil, fmt.Errorf("error decoding JSON API call: %v", err)
	// }

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding JSON API call: %v", err)
	}

	//log.Printf("| Call back api service Check Block Host server : %v\n", result)
	// permanentLink, ok := result["permanent_link"].(string)
	// if !ok {
	// 	log.Println("| Error: permanent_link field not found or not a string value")
	// 	return false, ""
	// }

	return result, nil
}

// func CheckResult(requestId string) (Response, error) {

// 	requestURL := fmt.Sprintf(os.Getenv("API_BLOCK_RESULT"), requestId)

// 	req, err := http.NewRequest("GET", requestURL, nil)
// 	if err != nil {
// 		log.Printf("| client: could not create request: %s\n", err)
// 		return nil, err
// 	}
// 	req.Header.Set("Accept", "application/json")

// 	client := http.Client{
// 		Timeout: 30 * time.Second,
// 	}

// 	res, err := client.Do(req)
// 	if err != nil {
// 		log.Printf("| client: error making http request: %s\n", err)
// 		return nil, err
// 	}
// 	defer res.Body.Close()

// 	if res.StatusCode != http.StatusOK {
// 		return nil, fmt.Errorf("non-200 status code: %d", res.StatusCode)
// 	}

// 	// Read response body
// 	body, err := ioutil.ReadAll(res.Body)
// 	if err != nil {
// 		return nil, fmt.Errorf("error reading response body: %s", err)
// 	}

// 	// Print response body
// 	fmt.Println("Response Body:", string(body))
// 	fmt.Println("Content Length Body:", res.ContentLength)

// 	//var result map[string]interface{}
// 	// if err := json.NewDecoder(res.Body).Decode(&Response); err != nil {
// 	// 	return nil, fmt.Errorf("error decoding JSON API call: %v", err)
// 	// }

// 	// Unmarshal the JSON response into the Response struct
// 	var response Response
// 	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
// 		fmt.Printf("Error decoding JSON response: %v\n", err)
// 		return nil, err
// 	}

// 	// Iterate over the response and print the names and addresses
// 	for name, entries := range response {
// 		for _, entry := range entries {
// 			if entry.Error != "" {
// 				log.Printf("| Name: %s, Error: %s\n", name, entry.Error)
// 			} else {
// 				log.Printf("| Name: %s, Address: %s, Time: %.6f\n", name, entry.Address, entry.Time)
// 			}
// 		}
// 	}

// 	return response, nil

// }

// func CheckResultTry2(requestId string) (Response, error) {
// 	requestURL := fmt.Sprintf(os.Getenv("API_BLOCK_RESULT"), requestId)
// 	req, err := http.NewRequest("GET", requestURL, nil)
// 	if err != nil {
// 		return nil, fmt.Errorf("could not create request: %s", err)
// 	}
// 	req.Header.Set("Accept", "application/json")

// 	client := http.Client{
// 		Timeout: 30 * time.Second,
// 	}

// 	// Maximum number of retries
// 	maxRetries := 3

// 	for retries := 0; retries < maxRetries; retries++ {
// 		res, err := client.Do(req)
// 		if err != nil {
// 			return nil, fmt.Errorf("error making http request: %s", err)
// 		}

// 		// Read response body
// 		body, err := ioutil.ReadAll(res.Body)
// 		if err != nil {
// 			return nil, fmt.Errorf("error reading response body: %s", err)
// 		}

// 		// Print response body
// 		fmt.Println("Response Body:", string(body))
// 		fmt.Println("Content Length Body:", res.ContentLength)

// 		if res.StatusCode != http.StatusOK {
// 			res.Body.Close()
// 			return nil, fmt.Errorf("non-200 status code: %d", res.StatusCode)
// 		}

// 		// Check if the response body is empty or null

// 		var response Response
// 		if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
// 			res.Body.Close()
// 			return nil, fmt.Errorf("error decoding JSON response: %v", err)
// 		}

// 		// Close the response body before retrying
// 		res.Body.Close()
// 		time.Sleep(5 * time.Second)
// 	}

// 	return nil, fmt.Errorf("maximum retries reached, response body is null")
// }

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
