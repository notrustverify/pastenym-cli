package main

import (
	"encoding/json"
	"fmt"
	"html"
	"os"
	"strings"
)

// informations needed to retrieve a paste
type userDataRetrieve struct {
	UrlId string `json:"urlId"`
}

// to retrieve a paste
type pasteRetrieve struct {
	Event  event            `json:"event"`
	Sender string           `json:"sender"`
	Data   userDataRetrieve `json:"data"`
}

type textRetrieved struct {
	Text      string    `json:"text"`
	NumView   int       `json:"num_view"`
	CreatedOn string    `json:"created_on"`
	Burn      bool      `json:"is_burn"`
	Ipfs      bool      `json:"is_ipfs"`
	EncParams encParams `json:"encParams"`
}

func getPaste(urlId string, key string, selfAddress string) clearObjectUser {

	var urlIdData userDataRetrieve
	urlIdData.UrlId = urlId

	var userKey string
	userKey = key

	// if url is paste extract urlId and key
	if strings.Contains(urlId, "http") {
		urlId, key := extractLink(urlId)
		urlIdData.UrlId = urlId
		userKey = key
	}
	textToGet := pasteRetrieve{
		Event:  getText,
		Sender: selfAddress,
		Data:   urlIdData,
	}

	receivedMessage := sendTextWithReply(&textToGet)
	messageByte := []byte(receivedMessage.Message)[9:]
	var textData textRetrieved
	err := json.Unmarshal(messageByte, &textData)
	if err != nil {
		panic(err)
	}

	decodedText := html.UnescapeString(textData.Text)
	var content []byte

	if decodedText == "" {
		fmt.Printf("%sText not found%s\n", Red, Reset)
		os.Exit(1)
	}

	if userKey != "" {
		encParams := textData.EncParams

		content = []byte(decrypt(userKey, &decodedText, encParams))
	} else {
		content = []byte(decodedText)
	}
	var clearObjectUser clearObjectUser
	err = json.Unmarshal(content, &clearObjectUser)
	if err != nil {
		fmt.Printf("\n%sError with text retrieved. --key is needed for decryption%s\n", Red, Reset)

		if *debug {
			panic(err.Error())
		}
		os.Exit(1)
	}
	if clearObjectUser.File.Filename != "" {
		fmt.Printf("%sFile are not supported in pastenym CLI%s\n", Red, Reset)
	}

	return clearObjectUser

}
