package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type idNewPaste struct {
	Ipfs  bool   `json:"ipfs"`
	Hash  string `json:"hash"`
	UrlId string `json:"url_id"`
}

// to add a paste
type pasteAdd struct {
	Event  event       `json:"event"`
	Sender string      `json:"sender"`
	Data   userDataAdd `json:"data"`
}

// informations to set for adding a paste
type userDataAdd struct {
	Text      string    `json:"text"`
	Private   bool      `json:"private"`
	Burn      bool      `json:"burn"`
	Ipfs      bool      `json:"ipfs"`
	EncParams encParams `json:"encParams"`
}

func newPaste(text string, encryptionParams encParams, selfAddress string, public bool, ipfs bool, burn bool) idNewPaste {

	paste := pasteAdd{
		Event:  newText,
		Sender: selfAddress,
		Data: userDataAdd{
			Text:      text,
			Private:   public,
			Burn:      burn,
			Ipfs:      ipfs,
			EncParams: encryptionParams,
		},
	}

	receivedMessage := sendTextWithReply(&paste)
	messageByte := []byte(receivedMessage.Message)[9:]

	if strings.Contains(strings.ToLower(receivedMessage.Message), "error") {
		errMsg := strings.ToLower(receivedMessage.Message)
		fmt.Printf("\n\n%s%s%s\n", Red, strings.Replace(errMsg, "\"", "", -1), Reset)
		os.Exit(1)
	}

	var dataUrl idNewPaste
	err := json.Unmarshal(messageByte, &dataUrl)
	if err != nil {

		panic(err)
	}

	return dataUrl
}
