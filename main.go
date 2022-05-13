package main

import (
	"bytes"
	"crypto/tls"
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
	ApiHost  string
	Site     string
	Username string
	Password string
	Udmp     bool
}

var (
	showHelp              = false
	configurationFilename = "configuration.json"
	noCheckCert           = false
	verbose               = false
)

func init() {
	getopt.FlagLong(&showHelp, "help", 'h', "Show help")
	getopt.FlagLong(&configurationFilename, "config", 'c', "Path to the configuration file")
	getopt.FlagLong(&noCheckCert, "no-check-cert", 0, "Don't check server TLS certificate")
	getopt.FlagLong(&verbose, "verbose", 'v', "Verbose output")
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

	c.ApiHost = c.Host

	if c.Udmp {
		c.ApiHost = c.ApiHost + "/proxy/network"
	}

	cookieJar, err := cookiejar.New(nil)

	if err != nil {
		fmt.Fprintln(os.Stderr, "Error setting up HTTP cookie jar:", err)

		return 1
	}

	// Don't verify TLS certs...
	tls := &tls.Config{}
	if noCheckCert {
		tls.InsecureSkipVerify = true
	}

	// Get TLS transport
	tr := &http.Transport{TLSClientConfig: tls}

	httpClient := &http.Client{
		Jar:       cookieJar,
		Transport: tr,
	}

	requestBodyLogin, err := json.Marshal(unifiRequestLogin{
		Username: c.Username,
		Password: c.Password,
	})

	if err != nil {
		fmt.Fprintln(os.Stderr, "Error encoding JSON for UniFi login:", err)

		return 1
	}

	loginPath := c.Host

	if c.Udmp {
		loginPath = loginPath + "/api/auth/login"
	} else {
		loginPath = loginPath + "/api/login"
	}

	httpResponse, err := httpClient.Post(loginPath, "application/json", bytes.NewBuffer(requestBodyLogin))

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

	responseLogin := unifiResponseLogin{}

	err = json.Unmarshal(responseBodyLogin, &responseLogin)

	if err != nil {
		fmt.Fprintln(os.Stderr, "Error decoding UniFi login response:", err)

		return 1
	}

	if responseLogin.UniqueId == "" {
		fmt.Fprintln(os.Stderr, "Error determining UniFi unique ID")

		return 1
	}

	csrfTokenLogin := httpResponse.Header.Get("X-Csrf-Token")

	if csrfTokenLogin == "" {
		fmt.Fprintln(os.Stderr, "Error determining UniFi CSRF token")

		return 1
	}

	if verbose {
		fmt.Fprintln(os.Stdout, "Login complete")
	}

	httpResponse, err = httpClient.Get(c.ApiHost + "/api/s/" + c.Site + "/stat/alluser")

	if err != nil {
		fmt.Fprintln(os.Stderr, "Error communicating host UniFi alluser:", err)

		return 1
	}

	defer httpResponse.Body.Close()

	responseBodyAllUser, err := ioutil.ReadAll(httpResponse.Body)

	if err != nil {
		fmt.Fprintln(os.Stderr, "Error reading UniFi alluser response:", err)

		return 1
	}

	responseAllUser := unifiResponseAllUser{}

	err = json.Unmarshal(responseBodyAllUser, &responseAllUser)

	if err != nil {
		fmt.Fprintln(os.Stderr, "Error decoding UniFi alluser response:", err)

		return 1
	}

	err = unifiResponseCheckMeta(responseAllUser.Meta, responseBodyAllUser, "alluser")

	if err != nil {
		fmt.Fprintln(os.Stderr, err)

		return 1
	}

	if verbose {
		fmt.Fprintln(os.Stdout, "Alluser retrieval complete")
	}

	var macsToForget []string

	numberMacs := 0

	for _, user := range responseAllUser.Data {
		var sum = len(user.Name) + user.TxBytes + user.TxPackets + user.RxBytes + user.RxPackets + user.WifiTxAttempts + user.TxRetries

		if sum == 0 {
			macsToForget = append(macsToForget, user.Mac)
		}
		numberMacs++
	}

	if verbose {
		fmt.Fprintf(os.Stdout, "%d MACs found\n", numberMacs)
	}

	numberMacsToForget := len(macsToForget)

	if verbose {
		fmt.Fprintf(os.Stdout, "%d MACs to forget\n", numberMacsToForget)
	}

	if numberMacsToForget > 0 {
		pageSize := 25
		lowBound := 0
		highBound := pageSize

		for lowBound < numberMacsToForget {
			if highBound > numberMacsToForget {
				highBound = numberMacsToForget
			}

			requestStamgrForget := unifiRequestStamgr{
				Cmd:  "forget-sta",
				Macs: macsToForget[lowBound:highBound],
			}

			requestBodyStamgrForget, err := json.Marshal(requestStamgrForget)

			if err != nil {
				fmt.Fprintln(os.Stderr, "Error encoding JSON for UniFi stamgr:", err)

				return 1
			}

			httpRequest, err := http.NewRequest("POST", c.ApiHost+"/api/s/"+c.Site+"/cmd/stamgr", bytes.NewBuffer(requestBodyStamgrForget))

			if err != nil {
				fmt.Fprintln(os.Stderr, "Error preparing request for UniFi stamgr:", err)

				return 1
			}

			httpRequest.Header.Add("Content-Type", "application/json")
			httpRequest.Header.Add("X-CSRF-Token", csrfTokenLogin)

			httpResponse, err := httpClient.Do(httpRequest)

			if err != nil {
				fmt.Fprintln(os.Stderr, "Error communicating host for UniFi stamgr:", err)

				return 1
			}

			defer httpResponse.Body.Close()

			responseBodyStamgr, err := ioutil.ReadAll(httpResponse.Body)

			if err != nil {
				fmt.Fprintln(os.Stderr, "Error reading UniFi stamgr response:", err)

				return 1
			}

			responseStamgr := unifiResponseBase{}

			err = json.Unmarshal(responseBodyStamgr, &responseStamgr)

			if err != nil {
				fmt.Fprintln(os.Stderr, "Error decoding UniFi stamgr response:", err, ";", string(responseBodyStamgr))

				return 1
			}

			err = unifiResponseCheckMeta(responseStamgr.Meta, responseBodyStamgr, "stamgr")

			if err != nil {
				fmt.Fprintln(os.Stderr, err)

				return 1
			}

			if verbose {
				fmt.Fprintf(os.Stdout,
					"Sta-forget complete for %d - %d\n",
					lowBound, highBound)
			}

			lowBound = lowBound + pageSize
			highBound = lowBound + pageSize
		}

		if verbose {
			fmt.Printf("%s %d %s", "Forgot", numberMacsToForget, "devices")
		}
	}

	fmt.Fprintln(os.Stdout, "Successful completion")

	return 0
}
