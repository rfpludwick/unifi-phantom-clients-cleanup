package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"os"
	"time"
)

func init() {
	initConfiguration()
}

func main() {
	err := exec()

	if err != nil {
		fmt.Fprintln(os.Stderr, err)

		os.Exit(1)
	}

	os.Exit(0)
}

func exec() error {
	cF, err := processConfiguration()

	if err != nil {
		return err
	}

	for _, cfSite := range cF.Sites {
		fmt.Println("Executing for host", cfSite.Host)

		cfSite.ApiHost = cfSite.Host

		if cfSite.Udmp {
			cfSite.ApiHost = cfSite.ApiHost + "/proxy/network"
		}

		cookieJar, err := cookiejar.New(nil)

		if err != nil {
			return fmt.Errorf("%s %s", "Error setting up HTTP cookie jar:", err)
		}

		// Don't verify TLS certs...
		tls := &tls.Config{}

		tls.InsecureSkipVerify = !cfSite.ValidateCertificate

		// Get TLS transport
		tr := &http.Transport{TLSClientConfig: tls}

		httpClient := &http.Client{
			Jar:       cookieJar,
			Transport: tr,
		}

		requestBodyLogin, err := json.Marshal(unifiRequestLogin{
			Username: cfSite.Username,
			Password: cfSite.Password,
		})

		if err != nil {
			return fmt.Errorf("%s %s", "Error encoding JSON for UniFi login:", err)
		}

		loginPath := cfSite.Host

		if cfSite.Udmp {
			loginPath = loginPath + "/api/auth/login"
		} else {
			loginPath = loginPath + "/api/login"
		}

		httpResponse, err := httpClient.Post(loginPath, "application/json", bytes.NewBuffer(requestBodyLogin))

		if err != nil {
			return fmt.Errorf("%s %s", "Error communicating host for UniFi login:", err)
		}

		err = logHttpCall(cF, requestBodyLogin, httpResponse)

		if err != nil {
			return err
		}

		defer httpResponse.Body.Close()

		responseBodyLogin, err := io.ReadAll(httpResponse.Body)

		if err != nil {
			return fmt.Errorf("%s %s", "Error reading UniFi login response:", err)
		}

		responseLogin := unifiResponseLogin{}

		err = json.Unmarshal(responseBodyLogin, &responseLogin)

		if err != nil {
			return fmt.Errorf("%s %s", "Error decoding UniFi login response:", err)
		}

		if responseLogin.UniqueId == "" {
			return fmt.Errorf("%s", "Error determining UniFi unique ID")
		}

		csrfTokenLogin := httpResponse.Header.Get("X-Csrf-Token")

		if csrfTokenLogin == "" {
			return fmt.Errorf("%s", "Error determining UniFi CSRF token")
		}

		if flagVerboseOutput {
			fmt.Fprintln(os.Stdout, "Login complete")
		}

		httpResponse, err = httpClient.Get(cfSite.ApiHost + "/api/s/" + cfSite.Site + "/stat/alluser")

		if err != nil {
			return fmt.Errorf("%s %s", "Error communicating host UniFi alluser:", err)
		}

		err = logHttpCall(cF, []byte{}, httpResponse)

		if err != nil {
			return err
		}

		defer httpResponse.Body.Close()

		responseBodyAllUser, err := io.ReadAll(httpResponse.Body)

		if err != nil {
			return fmt.Errorf("%s %s", "Error reading UniFi alluser response:", err)
		}

		responseAllUser := unifiResponseAllUser{}

		err = json.Unmarshal(responseBodyAllUser, &responseAllUser)

		if err != nil {
			return fmt.Errorf("%s %s", "Error decoding UniFi alluser response:", err)
		}

		err = unifiResponseCheckMeta(responseAllUser.Meta, responseBodyAllUser, "alluser")

		if err != nil {
			return err
		}

		if flagVerboseOutput {
			fmt.Fprintln(os.Stdout, "Alluser retrieval complete")
		}

		var clientsToForget []string

		numberClients := 0

		for _, user := range responseAllUser.Data {
			var sum = len(user.Name) + user.TxBytes + user.TxPackets + user.RxBytes + user.RxPackets + user.WifiTxAttempts + user.TxRetries

			if sum == 0 {
				clientsToForget = append(clientsToForget, user.Mac)
			}
			numberClients++
		}

		if flagVerboseOutput {
			fmt.Printf("%d client(s) found\n", numberClients)
		}

		numberClientsToForget := len(clientsToForget)

		if flagVerboseOutput {
			fmt.Printf("%d client(s) to forget\n", numberClientsToForget)
		}

		if numberClientsToForget > 0 {
			pageSize := 25
			lowBound := 0
			highBound := pageSize

			for lowBound < numberClientsToForget {
				if highBound > numberClientsToForget {
					highBound = numberClientsToForget
				}

				requestStamgrForget := unifiRequestStamgr{
					Cmd:  "forget-sta",
					Macs: clientsToForget[lowBound:highBound],
				}

				requestBodyStamgrForget, err := json.Marshal(requestStamgrForget)

				if err != nil {
					return fmt.Errorf("%s %s", "Error encoding JSON for UniFi stamgr:", err)
				}

				httpRequest, err := http.NewRequest("POST", cfSite.ApiHost+"/api/s/"+cfSite.Site+"/cmd/stamgr", bytes.NewBuffer(requestBodyStamgrForget))

				if err != nil {
					return fmt.Errorf("%s %s", "Error preparing request for UniFi stamgr:", err)
				}

				httpRequest.Header.Add("Content-Type", "application/json")
				httpRequest.Header.Add("X-CSRF-Token", csrfTokenLogin)

				httpResponse, err := httpClient.Do(httpRequest)

				if err != nil {
					return fmt.Errorf("%s %s", "Error communicating host for UniFi stamgr:", err)
				}

				err = logHttpCall(cF, requestBodyStamgrForget, httpResponse)

				if err != nil {
					return err
				}

				defer httpResponse.Body.Close()

				responseBodyStamgr, err := io.ReadAll(httpResponse.Body)

				if err != nil {
					return fmt.Errorf("%s %s", "Error reading UniFi stamgr response:", err)
				}

				responseStamgr := unifiResponseBase{}

				err = json.Unmarshal(responseBodyStamgr, &responseStamgr)

				if err != nil {
					return fmt.Errorf("%s %s %s %s", "Error decoding UniFi stamgr response:", err, ";", string(responseBodyStamgr))
				}

				err = unifiResponseCheckMeta(responseStamgr.Meta, responseBodyStamgr, "stamgr")

				if err != nil {
					return err
				}

				if flagVerboseOutput {
					fmt.Printf("Sta-forget complete for %d - %d\n", lowBound, highBound)
				}

				lowBound = lowBound + pageSize
				highBound = lowBound + pageSize
			}

			if flagVerboseOutput {
				fmt.Printf("%s %d %s\n", "Forgot", numberClientsToForget, "devices")
			}
		}

		fmt.Println("Successful completion")
	}

	return nil
}

func logHttpCall(cF *ConfigurationFile, httpRequestBody []byte, httpResponse *http.Response) error {
	if cF.HttpLogDirectory != "" {
		file, err := os.Create(fmt.Sprintf("%s/%s.log", cF.HttpLogDirectory, time.Now().Format("2006-01-02_15-04-05_-0700")))

		if err != nil {
			return fmt.Errorf("%s %s", "Error writing HTTP log file:", err)
		}

		defer file.Close()

		responseBody, err := io.ReadAll(httpResponse.Body)

		if err != nil {
			return fmt.Errorf("%s %s", "Error reading HTTP response body:", err)
		}

		fmt.Fprintf(file, "Request\n\n%+v\n\n", httpResponse.Request)
		fmt.Fprintf(file, "Request Body\n\n%+v\n\n", string(httpRequestBody))
		fmt.Fprintf(file, "Response\n\n%+v\n\n", httpResponse)
		fmt.Fprintf(file, "Response Body\n\n%+v\n\n", string(responseBody))
		fmt.Fprintf(file, "TLS\n\n%+v", httpResponse.TLS)

		httpResponse.Body.Close()
		httpResponse.Body = io.NopCloser(bytes.NewBuffer(responseBody))
	}

	return nil
}
