package models

type Route struct {
	AppGuid string `json:"app_guid" db:"app_guid"`
	Fqdn    string `json:"fqdn" db:"fqdn"`
}
