package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"os"

	"github.com/pborman/getopt/v2"
)

type configuration struct {
	Host     string
	Site     string
	Username string
	Password string
}

var (
	showHelp              = false
	configurationFilename = "configuration.json"
)

func init() {
	getopt.FlagLong(&showHelp, "help", 'h', "Show help")
	getopt.FlagLong(&configurationFilename, "config", 'c', "Path to the configuration file")
}

func main() {
	os.Exit(exec())
}

func exec() int {
	getopt.Parse()

	if showHelp {
		getopt.Usage()

		return 0
	}

	configurationFileData, err := ioutil.ReadFile(configurationFilename)

	if err != nil {
		fmt.Fprintln(os.Stderr, "Error reading configuration JSON:", err)

		return 1
	}

	var c configuration

	err = json.Unmarshal(configurationFileData, &c)

	if err != nil {
		fmt.Fprintln(os.Stderr, "Error decoding configuration JSON:", err)

		return 1
	}

	cookieJar, err := cookiejar.New(nil)

	if err != nil {
		fmt.Fprintln(os.Stderr, "Error setting up HTTP cookie jar:", err)

		return 1
	}

	httpClient := &http.Client{
		Jar: cookieJar,
	}

	requestBodyLogin, err := json.Marshal(unifiRequestLogin{
		Username: c.Username,
		Password: c.Password,
	})

	if err != nil {
		fmt.Fprintln(os.Stderr, "Error encoding JSON for UniFi login:", err)

		return 1
	}

	httpResponse, err := httpClient.Post(c.Host+"/api/login", "application/json", bytes.NewBuffer(requestBodyLogin))

	if err != nil {
		fmt.Fprintln(os.Stderr, "Error communicating host for UniFi login:", err)

		return 1
	}

	defer httpResponse.Body.Close()

	responseBodyLogin, err := ioutil.ReadAll(httpResponse.Body)

	if err != nil {
		fmt.Fprintln(os.Stderr, "Error reading UniFi login response:", err)

		return 1
	}

	responseLogin := unifiResponseBase{}

	err = json.Unmarshal(responseBodyLogin, &responseLogin)

	if err != nil {
		fmt.Fprintln(os.Stderr, "Error decoding UniFi login response:", err)

		return 1
	}

	err = unifiResponseCheckMeta(responseLogin.Meta, responseBodyLogin)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)

		return 1
	}

	httpResponse, err = httpClient.Get(c.Host + "/api/s/" + c.Site + "/stat/alluser")

	if err != nil {
		fmt.Fprintln(os.Stderr, "Error communicating host UniFi alluser:", err)

		return 1
	}

	defer httpResponse.Body.Close()

	responseBodyAllUser, err := ioutil.ReadAll(httpResponse.Body)

	if err != nil {
		fmt.Fprintln(os.Stderr, "Error reading UniFi allsuer response:", err)

		return 1
	}

	responseAllUser := unifiResponseAllUser{}

	err = json.Unmarshal(responseBodyAllUser, &responseAllUser)

	if err != nil {
		fmt.Fprintln(os.Stderr, "Error decoding UniFi alluser response:", err)

		return 1
	}

	err = unifiResponseCheckMeta(responseAllUser.Meta, responseBodyAllUser)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)

		return 1
	}

	var macsToForget []string

	for _, user := range responseAllUser.Data {
		var sum = len(user.Name) + user.TxBytes + user.TxPackets + user.RxBytes + user.RxPackets + user.WifiTxAttempts + user.TxRetries

		if sum == 0 {
			macsToForget = append(macsToForget, user.Mac)
		}
	}

	if len(macsToForget) > 0 {
		requestStamgrForget := unifiRequestStamgrForget{
			Macs: macsToForget,
		}

		requestStamgrForget.Init()

		requestBodyStamgrForget, err := json.Marshal(requestStamgrForget)

		if err != nil {
			fmt.Fprintln(os.Stderr, "Error encoding JSON for UniFi stamgr forget:", err)

			return 1
		}

		httpResponse, err := httpClient.Post(c.Host+"/api/s/"+c.Site+"/cmd/stamgr", "application/json", bytes.NewBuffer(requestBodyStamgrForget))

		if err != nil {
			fmt.Fprintln(os.Stderr, "Error communicating host for UniFi stamgr forget:", err)

			return 1
		}

		defer httpResponse.Body.Close()

		responseBodyStamgrForget, err := ioutil.ReadAll(httpResponse.Body)

		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading UniFi stamgr forget response:", err)

			return 1
		}

		responseStamgrForget := unifiResponseBase{}

		err = json.Unmarshal(responseBodyStamgrForget, &responseStamgrForget)

		if err != nil {
			fmt.Fprintln(os.Stderr, "Error decoding UniFi stamgr forget response:", err)

			return 1
		}

		err = unifiResponseCheckMeta(responseStamgrForget.Meta, responseBodyStamgrForget)

		if err != nil {
			fmt.Fprintln(os.Stderr, err)

			return 1
		}
	}

	return 0
}
