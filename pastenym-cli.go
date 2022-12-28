package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"html"
	"os"
	"runtime"

	"github.com/gorilla/websocket"
)

// handle connection parameters
type connection struct {
	nymClient string
	provider  string
	ws        websocket.Conn
	instance  string
}

// store payload data user
type clearObjectUser struct {
	Text string `json:"text"`
	File string `json:"file"`
}

// event to send when query or add text
type event string

const (
	newText event = "newText"
	getText event = "getText"
)

// to add a paste
type pasteAdd struct {
	Event  event       `json:"event"`
	Sender string      `json:"sender"`
	Data   userDataAdd `json:"data"`
}

// informations to set for adding a paste
type userDataAdd struct {
	Text      clearObjectUser `json:"text"`
	Private   bool            `json:"private"`
	Burn      bool            `json:"burn"`
	Ipfs      bool            `json:"ipfs"`
	EncParams string          `json:"encParams"`
}

type idNewPaste struct {
	Ipfs  bool   `json:"ipfs"`
	Hash  string `json:"hash"`
	UrlId string `json:"url_id"`
}

// to retrieve a paste
type pasteRetrieve struct {
	Event  event            `json:"event"`
	Sender string           `json:"sender"`
	Data   userDataRetrieve `json:"data"`
}

// informations needed to retrieve a paste
type userDataRetrieve struct {
	UrlId string `json:"urlId"`
}

type textRetrieved struct {
	Text      string `json:"text"`
	NumView   int    `json:"num_view"`
	CreatedOn string `json:"created_on"`
	Burn      bool   `json:"is_burn"`
	Ipfs      bool   `json:"is_ipfs"`
}

// informations received query or add a paste
type messageReceived struct {
	Type      string `json:"type"`
	Message   string `json:"message"`
	SenderTag string `json:"senderTage"`
}

var connectionData connection
var debug *bool
var silent *bool

var Reset = "\033[0m"
var Red = "\033[31m"
var Green = "\033[32m"
var Yellow = "\033[33m"
var Blue = "\033[34m"
var Purple = "\033[35m"
var Cyan = "\033[36m"
var Gray = "\033[37m"
var White = "\033[97m"

func main() {

	initColor()
	// flags declaration using flag package
	text := flag.String("text", "", "Specify the text to share. Mandatory")

	// to be implemented
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
		fmt.Printf("%s-text or -id is mandatory%s", Red, Reset)
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
		defer conn.Close()
		fmt.Printf("%sError with nym-client connection %s%s. Is it started ?\n", Red, uri, Reset)
		panic(err)
	}

	return conn
}

func newPaste(text string, selfAddress string, public bool, ipfs bool, burn bool) {

	paste := pasteAdd{
		Event:  newText,
		Sender: selfAddress,
		Data: userDataAdd{
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

	var dataUrl idNewPaste
	err := json.Unmarshal(messageByte, &dataUrl)
	if err != nil {
		panic(err)
	}
	if !*silent {

		fmt.Printf("%sURL ID is %s", Green, dataUrl.UrlId)
		fmt.Printf("\nLink: https://%s/#/%s%s", connectionData.instance, dataUrl.UrlId, Reset)
		if dataUrl.Ipfs {
			fmt.Printf("\n%sipfs://%s%s", Green, dataUrl.Hash, Reset)
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
	var textData textRetrieved
	err := json.Unmarshal(messageByte, &textData)
	if err != nil {
		panic(err)
	}

	decodedText := html.UnescapeString(textData.Text)

	content := []byte(decodedText)
	var clearObjectUser clearObjectUser
	err = json.Unmarshal(content, &clearObjectUser)
	if err != nil {
		fmt.Printf("%sFile are not supported in pastenym CLI%s\n", Red, Reset)
	}

	if !*silent {

		fmt.Printf("%sPaste text\n%s%s", Green, clearObjectUser.Text, Reset)
	} else {
		fmt.Printf("%s", clearObjectUser.Text)
	}

}

func sendTextWithReply(data interface{}) messageReceived {
	//copied from https://github.com/nymtech/nym/blob/develop/clients/native/examples/go-examples/websocket/text/textsend.go

	pasteJson, err := json.Marshal(data)
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
