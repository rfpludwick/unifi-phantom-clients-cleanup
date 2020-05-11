package main

import (
	"fmt"
)

type unifiRequestLogin struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type unifiRequestStamgrForget struct {
	Cmd  string   `json:"cmd"`
	Macs []string `json:"macs"`
}

type unifiResponseBase struct {
	Meta unifiResponseBaseMeta `json:"meta"`
}

type unifiResponseBaseMeta struct {
	Rc  string `json:"rc"`
	Msg string `json:"msg"`
}

type unifiResponseAllUser struct {
	Meta unifiResponseBaseMeta        `json:"meta"`
	Data []unifiResponseAllUserClient `json:"data"`
}

type unifiResponseAllUserClient struct {
	Name           string `json:"name"`
	Mac            string `json:"mac"`
	TxBytes        int    `json:"tx_bytes"`
	TxPackets      int    `json:"tx_packets"`
	RxBytes        int    `json:"rx_bytes"`
	RxPackets      int    `json:"rx_packets"`
	WifiTxAttempts int    `json:"wifi_tx_attempts"`
	TxRetries      int    `json:"tx_retries"`
}

func unifiResponseCheckMeta(u unifiResponseBaseMeta, responseBody []byte) error {
	if u.Rc == "error" {
		return fmt.Errorf("%s %s", "Error in UniFi response:", u.Msg)
	}

	if u.Rc != "ok" {
		return fmt.Errorf("%s %s", "Error with unexpected UniFi response:", string(responseBody))
	}

	return nil
}

func (u *unifiRequestStamgrForget) Init() {
	u.Cmd = "forget-sta"
}
