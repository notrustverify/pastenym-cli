package main

import (
	"encoding/json"
	"runtime"
	"strings"
)

var Reset = "\033[0m"
var Red = "\033[31m"
var Green = "\033[32m"
var Yellow = "\033[33m"
var Blue = "\033[34m"
var Purple = "\033[35m"
var Cyan = "\033[36m"
var Gray = "\033[37m"
var White = "\033[97m"

type capababilities struct {
	IpfsHosting             bool `json:"ipfs_hosting"`
	ExpirationBitcoinHeight bool `json:"expiration_bitcoin_height"`
}

type pingMessage struct {
	Event  event  `json:"event"`
	Sender string `json:"sender"`
	Data   string `json:"data"`
}

type pingReceived struct {
	Version      string         `json:"version"`
	Capabilities capababilities `json:"capabilities"`
	Alive        bool
}

func extractLink(link string) (string, string) {
	var urlId string

	urlId = link[strings.LastIndex(link, "/")+1:]

	var key string
	if strings.Contains(urlId, "&") {
		//extract key
		key = strings.Replace(urlId[strings.LastIndex(urlId, "&")+1:], "key=", "", -1)
		urlId = urlId[:strings.Index(urlId, "&")]
	}

	return urlId, key
}

func initColor() {

	if runtime.GOOS == "windows" {
		Reset = ""
		Red = ""
		Green = ""
		Yellow = ""
		Blue = ""
		Purple = ""
		Cyan = ""
		Gray = ""
		White = ""
	}
}

func pingBackend(selfAddress string, testBackendAlive bool) pingReceived {
	pingMessage := pingMessage{
		Event:  ping,
		Sender: selfAddress,
		Data:   "emtpy",
	}

	receivedMessage := sendTextWithReply(pingMessage, 8, testBackendAlive)

	if receivedMessage.Type == "error" {
		return pingReceived{Version: "0.0.0", Alive: false}
	} else {
		messageByte := []byte(receivedMessage.Message)[9:]

		var pingData pingReceived
		err := json.Unmarshal(messageByte, &pingData)
		if err != nil {

			panic(err)
		}

		return pingData
	}
}
