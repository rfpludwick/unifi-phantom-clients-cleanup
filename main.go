package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type configuration struct {
	Host     string
	Site     string
	Username string
	Password string
}

func main() {
	configurationFileData, err := ioutil.ReadFile("configuration.json")

	if err != nil {
		fmt.Println("Error reading configuration JSON:", err)

		return
	}

	c := configuration{}

	err = json.Unmarshal(configurationFileData, &c)

	if err != nil {
		fmt.Println("Error decoding configuration JSON:", err)

		return
	}

	requestBody, err := json.Marshal(map[string]string{
		"username": c.Username,
		"password": c.Password,
	})

	if err != nil {
		fmt.Println("Error preparing UniFi request:", err)

		return
	}

	httpResponse, err := http.Post(c.Host+"/api/login", "application/json", bytes.NewBuffer(requestBody))

	if err != nil {
		fmt.Println("Error communicating with UniFi host:", err)

		return
	}

	defer httpResponse.Body.Close()

	responseBody, err := ioutil.ReadAll(httpResponse.Body)

	if err != nil {
		fmt.Println("Error reading UniFi response:", err)

		return
	}

	var response interface{}

	err = json.Unmarshal(responseBody, &response)

	if err != nil {
		fmt.Println("Error decoding UniFi response:", err)

		return
	}

	meta, ok := response.(map[string]interface{})["meta"]

	if ok == false {
		fmt.Println("Error with unexpected UniFi response; response:", string(responseBody))

		return
	}

	rc, ok := meta.(map[string]interface{})["rc"]

	if ok == false {
		fmt.Println("Error with unexpected UniFi response; response:", string(responseBody))

		return
	}

	if rc != "ok" {
		fmt.Println("UniFi login failed; response:", string(responseBody))

		return
	}

	// todo need cookiejar from here
}
