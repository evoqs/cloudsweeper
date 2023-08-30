package utils

import (
	"errors"
	"net/url"
	"strings"
)

func GetDBUrl(c *Config) (string, error) {
	var mongoDBUrl string
	mongoDBUrl = "mongodb://"
	if c.Database.Username != "" {
		if c.Database.Password != "" {
			mongoDBUrl = mongoDBUrl + strings.TrimSpace(c.Database.Username) + ":" + url.QueryEscape(strings.TrimSpace(c.Database.Password)) + "@"
		} else {
			return "", errors.New("Empty Password for mongodb")
		}

	}

	if c.Database.Host != "" {
		mongoDBUrl = mongoDBUrl + strings.TrimSpace(c.Database.Host)
	} else {
		return "", errors.New("Empty Hostname for mongodb")
	}

	if c.Database.Port != "" {
		mongoDBUrl = mongoDBUrl + ":" + strings.TrimSpace(c.Database.Port)
	}

	if c.Database.Name != "" {
		mongoDBUrl = mongoDBUrl + "/" + strings.TrimSpace(c.Database.Name)
	}

	return mongoDBUrl, nil
}
