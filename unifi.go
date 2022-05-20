package main

import (
	"fmt"
)

type unifiRequestLogin struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type unifiRequestStamgr struct {
	Cmd  string   `json:"cmd"`
	Macs []string `json:"macs"`
}

type unifiResponseLogin struct {
	UniqueId string `json:"unique_id"`
}

type unifiResponseBase struct {
	Meta unifiResponseBaseMeta
}

type unifiResponseBaseMeta struct {
	Rc  string
	Msg string
}

type unifiResponseAllUser struct {
	Meta unifiResponseBaseMeta
	Data []unifiResponseAllUserClient
}

type unifiResponseAllUserClient struct {
	Name           string
	Mac            string
	TxBytes        int `json:"tx_bytes"`
	TxPackets      int `json:"tx_packets"`
	RxBytes        int `json:"rx_bytes"`
	RxPackets      int `json:"rx_packets"`
	WifiTxAttempts int `json:"wifi_tx_attempts"`
	TxRetries      int `json:"tx_retries"`
}

func unifiResponseCheckMeta(u unifiResponseBaseMeta, responseBody []byte, identifier string) error {
	if u.Rc == "error" {
		return fmt.Errorf("%s %s %s %s", "Error in UniFi", identifier, "response:", u.Msg)
	}

	if u.Rc != "ok" {
		return fmt.Errorf("%s %s %s %s", "Error with unexpected UniFi", identifier, "response:", string(responseBody))
	}

	return nil
}
