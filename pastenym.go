package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	F "pastenym-cli/utils"

	"github.com/gorilla/websocket"
)

const VERSION = "1.2.5"

// event to send when query or add text
type event string

const (
	newText event = "newText"
	getText event = "getText"
	ping    event = "ping"
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
	File F.File `json:"file"`
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
	SenderTag string `json:"senderTag"`
}

type ErrorMessage struct {
	Message string
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
	file := flag.String("file", "", "Specify the path for a file to share. Default is empty")

	urlId := flag.String("id", "", "Specify paste url id to retrieve. Default is empty")
	key := flag.String("key", "", "Key for getting the plaintext")

	provider := flag.String("provider", "HWm3757chNdBq9FzKEY9j9VJ5siRxH8ukrNqYwFp9Unp.D34iYLRd5vzpCU4nZRcFVmoZpTQQMa6mws4Q65LdRosi@Fo4f4SQLdoyoGkFae5TpVhRVoXCF8UiypLVGtGjujVPf", "Specify the path for a file to share. Default is empty")
	nymClient := flag.String("nymclient", "127.0.0.1:1977", "Nym client to connect")
	instance := flag.String("instance", "pastenym.ch", "Instance where to get the paste from GUI")

	public := flag.Bool("public", false, "Set the paste to public, i.e without encryption. Default is private")
	ipfs := flag.Bool("ipfs", false, "Specify if the text to share is stored on IPFS. Default is false")
	burn := flag.Bool("burn", false, "Specify if the text have to be deleted when read. Default is false")
	burnView := flag.Int("view", 0, "Specify if the text have to be deleted when read.")
	expirationTime := flag.String("time", "", "Specify a relative time interval when the paste have to be deleted. For example 1d, 1m, 10h")
	expirationHeight := flag.Int("height", -1, "Specify a Bitcoin block height when the paste have to be deleted")
	ping := flag.Bool("ping", false, "Ping the backend to see if it's alive. Return the version")
	debug = flag.Bool("debug", false, "Specify if the text have to be deleted when read. Default is false")
	silent = flag.Bool("silent", false, "Remove every output, just print data. Default is false")
	onlyURL = flag.Bool("url", false, "Only print the URL. Default is false")

	flag.Parse()

	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		*text = getFromPipe()
	} else if !*ping && *text == "" && *urlId == "" && *file == "" {
		fmt.Printf("\nVersion: %s\n%s-text, -id or -file is mandatory%s\n", VERSION, Red, Reset)
		flag.Usage()
		os.Exit(1)
	}

	connectionData.provider = *provider
	connectionData.nymClient = *nymClient
	connectionData.instance = *instance
	connectionData.ws = *newConnection()
	defer connectionData.ws.Close()

	selfAddress := getSelfAddress()
	if *ping {
		pingData := pingBackend(selfAddress, false)
		formatPing(&pingData)
	}

	if *text != "" || *file != "" {
		if (*public || *burn) && *ipfs {
			fmt.Printf("\n%sIPFS paste cannot be public or burned%s\n", Red, Reset)
			os.Exit(1)
		}

		// create a new paste
		var userFile F.File
		var successFile bool
		var errorMsg string
		if *file != "" {
			successFile, userFile, errorMsg = F.ReadFile(*file)
			if !successFile {
				fmt.Printf("\n%sError with file %s, %s%s\n", Red, userFile.Filename, errorMsg, Reset)
				userFile = F.File{}

				connectionData.ws.Close()
				os.Exit(1)
			}
		}

		plaintext, err := json.Marshal(clearObjectUser{
			Text: *text,
			File: userFile,
		})
		if err != nil {
			panic(err.Error())
		}

		var dataUrl idNewPaste
		var key string

		pingRes := pingBackend(selfAddress, true)
		if !pingRes.Alive {
			formatPing(&pingRes)
			os.Exit(1)
		}

		if *public {
			dataUrl = newPaste(string(plaintext), encParams{}, selfAddress, *public, *ipfs, *burn, *burnView, *expirationTime, *expirationHeight)
		} else {
			var encParams encParams
			var textEncrypted string
			key, textEncrypted, encParams = encrypt(&plaintext)
			dataUrl = newPaste(textEncrypted, encParams, selfAddress, *public, *ipfs, *burn, *burnView, *expirationTime, *expirationHeight)
		}

		// show informations
		if !*silent && !*onlyURL {
			formatAddPasteVerbose(*public, dataUrl.UrlId, dataUrl.Hash, key)
		} else if *silent && !*onlyURL {
			formatAddPasteSilent(dataUrl.UrlId, key)
		} else if *onlyURL {
			formatAddPasteOnlyUrl(dataUrl.UrlId, key, *instance)
		}

	} else if *urlId != "" {

		metadata, userData := getPaste(*urlId, *key, getSelfAddress())

		if !*silent {
			formatGetPasteContentVerbose(&metadata, &userData)
		} else {
			formatGetPasteContentSilent(&userData)
		}
	}

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

func sendTextWithReply(data interface{}, timeout uint, testBackendAlive bool) messageReceived {
	//copied from https://github.com/nymtech/nym/blob/develop/clients/native/examples/go-examples/websocket/text/textsend.go

	pasteJson, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	pasteJsonStr := fmt.Sprintf("%s%s%s", ".", "{\"mimeType\":\"application/json\",\"headers\":null}", pasteJson)
	// append 7 0x00 bytes to set kind of message
	modifiedPasteJson := append(make([]byte, 7), pasteJsonStr...)

	sendRequest, err := json.Marshal(map[string]interface{}{
		"type":       "sendAnonymous",
		"recipient":  connectionData.provider,
		"message":    modifiedPasteJson,
		"replySurbs": 20,
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

	if (!*silent && !*onlyURL) && !testBackendAlive {
		fmt.Printf("waiting to receive a message from the mix network")
	}
	if timeout > 0 {
		connectionData.ws.UnderlyingConn().SetDeadline(time.Now().Add(time.Duration(timeout) * time.Second))
	}

	_, receivedMessage, err := connectionData.ws.ReadMessage()

	if err != nil && timeout > 0 {

		return messageReceived{
			Type:    "error",
			Message: "nok",
		}

	}

	if err != nil {
		panic(err)
	}

	if *debug {
		fmt.Printf("\nreceived %v from the mix network!\n", string(receivedMessage))
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

	// if a message hasn't been delivered, the nym-client forward it after pastenym-cli started again and then it panic
	var retry = 0
	var maxRetry = 25
	for {
		retry++
		err = connectionData.ws.ReadJSON(&responseJSON)
		if err == nil || retry >= maxRetry {
			break
		}

	}

	return responseJSON["address"].(string)
}
