package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"html"
	"os"

	"github.com/gorilla/websocket"
)

type connection struct {
	nymClient string
	provider  string
	ws        websocket.Conn
	instance  string
}

type clearObjectUser struct {
	Text string `json:"text"`
	File string `json:"file"`
}
type event string

const (
	newText event = "newText"
	getText event = "getText"
)

type userData struct {
	Text      clearObjectUser `json:"text"`
	Private   bool            `json:"private"`
	Burn      bool            `json:"burn"`
	Ipfs      bool            `json:"ipfs"`
	EncParams string          `json:"encParams"`
}

type paste struct {
	Event  event    `json:"event"`
	Sender string   `json:"sender"`
	Data   userData `json:"data"`
}

type pasteRetrieve struct {
	Event  event            `json:"event"`
	Sender string           `json:"sender"`
	Data   userDataRetrieve `json:"data"`
}

type userDataRetrieve struct {
	UrlId string `json:"urlId"`
}

type messageReceived struct {
	Type      string `json:"type"`
	Message   string `json:"message"`
	SenderTag string `json:"senderTage"`
}

type urlId struct {
	Ipfs  bool   `json:"ipfs"`
	Hash  string `json:"hash"`
	UrlId string `json:"url_id"`
}

type text struct {
	Text      string `json:"text"`
	NumView   int    `json:"num_view"`
	CreatedOn string `json:"created_on"`
	Burn      bool   `json:"is_burn"`
	Ipfs      bool   `json:"is_ipfs"`
}

const NYM_KIND_TEXT = '\x00'
const NYM_KIND_BINARY = '\x01'

const NYM_HEADER_SIZE_TEXT = '\x00' * 8 //set to 0 if it's a text
const NYM_HEADER_BINARY = '\x00' * 8    // not used now, to investigate later

var connectionData connection
var debug *bool
var silent *bool

func main() {

	// flags declaration using flag package
	text := flag.String("text", "", "Specify the text to share. Mandatory")

	// soon to be implemented
	//file := flag.String("file", "", "Specify the path for a file to share. Default is empty")

	urlId := flag.String("id", "", "Specify paste url id to retrieve. Default is empty")

	provider := flag.String("provider", "6y7sSj3dKp5AESnet1RQXEHmKkEx8Bv3RgwEJinGXv6J.FZfu6hNPi1hgQfu7crbXXUNLtr3qbKBWokjqSpBEeBMV@EBT8jTD8o4tKng2NXrrcrzVhJiBnKpT1bJy5CMeArt2w", "Specify the path for a file to share. Default is empty")
	nymClient := flag.String("nymclient", "127.0.0.1:1977", "Nym client to connect. Default 127.0.0.1:1977")
	instance := flag.String("instance", "pastenym.ch", "Instance where to get the paste from GUI")

	public := flag.Bool("public", true, "Set the paste to public, i.e without encryption. Default is private")
	ipfs := flag.Bool("ipfs", false, "Specify if the text to share is stored on IPFS. Default is false")
	burn := flag.Bool("burn", false, "Specify if the text have to be deleted when read. Default is false")
	debug = flag.Bool("debug", false, "Specify if the text have to be deleted when read. Default is false")
	silent = flag.Bool("silent", false, "Remove every output, just print data. Default is false")

	flag.Parse()

	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		*text = getFromPipe()
	} else if *text == "" && *urlId == "" {
		fmt.Println("-text or -id is mandatory")
		flag.Usage()
		os.Exit(1)
	}

	connectionData.provider = *provider
	connectionData.nymClient = *nymClient
	connectionData.instance = *instance
	connectionData.ws = *newConnection()

	if *urlId == "" {
		selfAddress := getSelfAddress()
		newPaste(*text, selfAddress, *public, *ipfs, *burn)
	} else {
		getPaste(*urlId, getSelfAddress())
	}

	defer connectionData.ws.Close()
}

func getFromPipe() string {
	var buf []byte
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		buf = append(buf, scanner.Bytes()...)
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}

	return fmt.Sprintf("%s", buf)
}

func newConnection() *websocket.Conn {
	uri := "ws://" + connectionData.nymClient
	conn, _, err := websocket.DefaultDialer.Dial(uri, nil)
	if err != nil {
		panic(err)
	}

	return conn
}

func newPaste(text string, selfAddress string, public bool, ipfs bool, burn bool) {

	paste := paste{
		Event:  newText,
		Sender: selfAddress,
		Data: userData{
			Text: clearObjectUser{
				Text: text,
				File: "",
			},
			Private: !public,
			Burn:    burn,
			Ipfs:    ipfs,
		},
	}

	receivedMessage := sendTextWithReply(&paste)
	messageByte := []byte(receivedMessage.Message)[9:]

	var dataUrl urlId
	err := json.Unmarshal(messageByte, &dataUrl)
	if err != nil {
		panic(err)
	}
	if !*silent {

		fmt.Printf("URL ID is %s", dataUrl.UrlId)
		fmt.Printf("\nLink: https://%s/#/%s", connectionData.instance, dataUrl.UrlId)
		if dataUrl.Ipfs {
			fmt.Printf("\nipfs://%s", dataUrl.Hash)
		}
	} else {
		fmt.Printf("%s", dataUrl.UrlId)
	}
}

func getPaste(urlId string, selfAddress string) {
	textToGet := pasteRetrieve{
		Event:  getText,
		Sender: selfAddress,
		Data: userDataRetrieve{
			UrlId: urlId,
		},
	}

	receivedMessage := sendTextWithReply(&textToGet)
	messageByte := []byte(receivedMessage.Message)[9:]
	var textData text
	err := json.Unmarshal(messageByte, &textData)
	if err != nil {
		panic(err)
	}

	decodedText := html.UnescapeString(textData.Text)

	content := []byte(decodedText)
	var clearObjectUser clearObjectUser
	err = json.Unmarshal(content, &clearObjectUser)
	if err != nil {
		panic(err)
	}

	if !*silent {
		fmt.Printf("Paste text: %s", clearObjectUser.Text)
	} else {
		fmt.Printf("%s", clearObjectUser.Text)
	}

}

func sendTextWithReply(paste interface{}) messageReceived {
	//copied from https://github.com/nymtech/nym/blob/develop/clients/native/examples/go-examples/websocket/text/textsend.go

	pasteJson, err := json.Marshal(paste)
	if err != nil {
		panic(err)
	}

	// append 9 0x00 bytes to set kind of message
	modifiedPasteJson := append(make([]byte, 9), pasteJson...)

	sendRequest, err := json.Marshal(map[string]interface{}{
		"type":      "send",
		"recipient": connectionData.provider,
		"message":   modifiedPasteJson,
	})
	if err != nil {
		panic(err)
	}

	if *debug {
		fmt.Printf("sending '%s' over the mix network...\n", pasteJson)
	}

	if err = connectionData.ws.WriteMessage(websocket.TextMessage, []byte(sendRequest)); err != nil {
		panic(err)
	}

	if !*silent {

		fmt.Printf("waiting to receive a message from the mix network...\n")
	}
	_, receivedMessage, err := connectionData.ws.ReadMessage()
	if err != nil {
		panic(err)
	}

	if *debug {
		fmt.Printf("received %v from the mix network!\n", string(receivedMessage))
	}

	var receivedMessageJSON messageReceived
	err = json.Unmarshal(receivedMessage, &receivedMessageJSON)
	if err != nil {
		panic(err)
	}

	return receivedMessageJSON
}

func getSelfAddress() string {
	//copied from https://github.com/nymtech/nym/blob/develop/clients/native/examples/go-examples/websocket/text/textsend.go

	selfAddressRequest, err := json.Marshal(map[string]string{"type": "selfAddress"})
	if err != nil {
		panic(err)
	}

	if err = connectionData.ws.WriteMessage(websocket.TextMessage, []byte(selfAddressRequest)); err != nil {
		panic(err)
	}

	responseJSON := make(map[string]interface{})
	err = connectionData.ws.ReadJSON(&responseJSON)
	if err != nil {
		panic(err)
	}

	return responseJSON["address"].(string)
}
