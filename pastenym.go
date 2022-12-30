package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/gorilla/websocket"
)

const VERSION = "1.0.0"

// event to send when query or add text
type event string

const (
	newText event = "newText"
	getText event = "getText"
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
	File File   `json:"file"`
}

type File struct {
	Data     []byte `json:"data"`
	Filename string `json:"filename"`
	MimeType string `json:"mimeType"`
}

type encParams struct {
	Salt   string `json:"salt"`
	Adata  string `json:"adata"`
	Iv     string `json:"iv"`
	Ks     uint32 `json:"ks"`
	V      uint8  `json:"v"`
	Ts     uint32 `json:"ts"`
	Mode   string `json:"mode"`
	Cipher string `json:"cipher"`
	Iter   uint32 `json:"iter"`
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
var onlyURL *bool

func main() {

	initColor()

	// flags declaration using flag package
	text := flag.String("text", "", "Specify the text to share. Mandatory")

	// to be implemented
	//file := flag.String("file", "", "Specify the path for a file to share. Default is empty")

	urlId := flag.String("id", "", "Specify paste url id to retrieve. Default is empty")
	key := flag.String("key", "", "Key for getting the plaintext")

	provider := flag.String("provider", "6y7sSj3dKp5AESnet1RQXEHmKkEx8Bv3RgwEJinGXv6J.FZfu6hNPi1hgQfu7crbXXUNLtr3qbKBWokjqSpBEeBMV@EBT8jTD8o4tKng2NXrrcrzVhJiBnKpT1bJy5CMeArt2w", "Specify the path for a file to share. Default is empty")
	nymClient := flag.String("nymclient", "127.0.0.1:1977", "Nym client to connect. Default 127.0.0.1:1977")
	instance := flag.String("instance", "pastenym.ch", "Instance where to get the paste from GUI")

	public := flag.Bool("public", false, "Set the paste to public, i.e without encryption. Default is private")
	ipfs := flag.Bool("ipfs", false, "Specify if the text to share is stored on IPFS. Default is false")
	burn := flag.Bool("burn", false, "Specify if the text have to be deleted when read. Default is false")
	debug = flag.Bool("debug", false, "Specify if the text have to be deleted when read. Default is false")
	silent = flag.Bool("silent", false, "Remove every output, just print data. Default is false")
	onlyURL = flag.Bool("url", false, "Only print the URL. Default is false")

	flag.Parse()

	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		*text = getFromPipe()
	} else if *text == "" && *urlId == "" {
		fmt.Printf("\nVersion: %s\n%s-text or -id is mandatory%s\n", VERSION, Red, Reset)
		flag.Usage()
		os.Exit(1)
	}

	connectionData.provider = *provider
	connectionData.nymClient = *nymClient
	connectionData.instance = *instance
	connectionData.ws = *newConnection()

	if *text != "" {
		// create a new paste

		selfAddress := getSelfAddress()

		plaintext, err := json.Marshal(clearObjectUser{
			Text: *text,
			File: File{},
		})
		if err != nil {
			panic(err.Error())
		}

		var dataUrl idNewPaste
		var key string

		if *public {
			dataUrl = newPaste(string(plaintext), encParams{}, selfAddress, *public, *ipfs, *burn)
		} else {
			var encParams encParams
			var textEncrypted string
			key, textEncrypted, encParams = encrypt(&plaintext)
			dataUrl = newPaste(textEncrypted, encParams, selfAddress, *public, *ipfs, *burn)
		}

		// show informations
		if !*silent && !*onlyURL {
			formatAddPasteVerbose(*public, dataUrl.UrlId, dataUrl.Hash, key)
		} else if *silent && !*onlyURL {
			formatAddPasteSilent(*urlId, key)
		} else if *onlyURL {
			formatAddPasteOnlyUrl(*urlId, key, *instance)
		}

	} else if *urlId != "" {

		data := getPaste(*urlId, *key, getSelfAddress())

		if !*silent {
			formatGetPasteContentVerbose(&data)
		} else {
			formatGetPasteContentSilent(&data)
		}
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
		fmt.Printf("%s\nError: No connection to nym-client %s%s\n\nIs it started ?\nHow to run one https://nymtech.net/docs/stable/integrations/websocket-client\n\n", Red, uri, Reset)
		if *debug {
			panic(err)
		}
		os.Exit(1)

	}

	return conn
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
		fmt.Printf("sending '%s' over the mix network\n", pasteJson)
	}

	if err = connectionData.ws.WriteMessage(websocket.TextMessage, []byte(sendRequest)); err != nil {
		panic(err)
	}

	if !*silent && !*onlyURL {
		fmt.Printf("waiting to receive a message from the mix network")
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
