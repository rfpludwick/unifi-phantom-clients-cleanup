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
	Name string
	Mac  string
}

type unifiResponseTraffic struct {
	ClientUsageByApp []unifiResponseTrafficClientUsage `json:"client_usage_by_app"`
}

type unifiResponseTrafficClientUsage struct {
	Client     unifiResponseTrafficClient       `json:"client"`
	UsageByApp []unifiResponseTrafficUsageByApp `json:"usage_by_app"`
}

type unifiResponseTrafficClient struct {
	Fingerprint unifiResponseTrafficClientFingerprint `json:"fingerprint"`
	Hostname    string                                `json:"hostname"`
	IsWired     bool                                  `json:"is_wired"`
	Mac         string                                `json:"mac"`
	Name        string                                `json:"name"`
	Oui         string                                `json:"oui"`
	WlanconfId  string                                `json:"wlanconf_id"`
}

type unifiResponseTrafficClientFingerprint struct {
	ComputedDevId  int  `json:"computed_dev_id"`
	ComputedEngine int  `json:"computed_engine"`
	Confidence     int  `json:"confidence"`
	DevCat         int  `json:"dev_cat"`
	DevFamily      int  `json:"dev_family"`
	DevId          int  `json:"dev_id"`
	DevVendor      int  `json:"dev_vendor"`
	HasOverride    bool `json:"has_override"`
	OsName         int  `json:"os_name"`
}

type unifiResponseTrafficUsageByApp struct {
	ActivitySeconds  int `json:"activity_seconds"`
	Application      int `json:"application"`
	BytesReceived    int `json:"bytes_received"`
	BytesTransmitted int `json:"bytes_transmitted"`
	Category         int `json:"category"`
	TotalBytes       int `json:"total_bytes"`
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
