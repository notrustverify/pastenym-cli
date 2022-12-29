package main

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"html"
	"io"
	"math/big"
	"os"
	"runtime"
	"strings"

	"golang.org/x/crypto/pbkdf2"

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
	File File   `json:"file"`
}

type File struct {
	Data     []byte `json:"data"`
	Filename string `json:"filename"`
	MimeType string `json:"mimeType"`
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
	Text      string    `json:"text"`
	Private   bool      `json:"private"`
	Burn      bool      `json:"burn"`
	Ipfs      bool      `json:"ipfs"`
	EncParams encParams `json:"encParams"`
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
	Text      string    `json:"text"`
	NumView   int       `json:"num_view"`
	CreatedOn string    `json:"created_on"`
	Burn      bool      `json:"is_burn"`
	Ipfs      bool      `json:"is_ipfs"`
	EncParams encParams `json:"encParams"`
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
	key := flag.String("key", "", "Key for getting the plaintext")

	provider := flag.String("provider", "6y7sSj3dKp5AESnet1RQXEHmKkEx8Bv3RgwEJinGXv6J.FZfu6hNPi1hgQfu7crbXXUNLtr3qbKBWokjqSpBEeBMV@EBT8jTD8o4tKng2NXrrcrzVhJiBnKpT1bJy5CMeArt2w", "Specify the path for a file to share. Default is empty")
	nymClient := flag.String("nymclient", "127.0.0.1:1977", "Nym client to connect. Default 127.0.0.1:1977")
	instance := flag.String("instance", "pastenym.ch", "Instance where to get the paste from GUI")

	public := flag.Bool("public", false, "Set the paste to public, i.e without encryption. Default is private")
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
			key, textEncrypted, encParams = encrypt(plaintext)
			dataUrl = newPaste(textEncrypted, encParams, selfAddress, *public, *ipfs, *burn)
		}

		if !*silent {

			fmt.Printf("%sID: %s", Green, dataUrl.UrlId)

			if !*public {
				fmt.Printf("%s\nKey: %s\n", Green, key)
				fmt.Printf("\nLink: https://%s/#/%s&key=%s%s", connectionData.instance, dataUrl.UrlId, key, Reset)
			} else {
				fmt.Printf("\nLink: https://%s/#/%s%s", connectionData.instance, dataUrl.UrlId, Reset)
			}

			if dataUrl.Ipfs {
				fmt.Printf("\n%sipfs://%s%s", Green, dataUrl.Hash, Reset)
			}
		} else {
			fmt.Printf("%s", dataUrl.UrlId)
		}
	} else {

		data := getPaste(*urlId, *key, getSelfAddress())

		if !*silent {
			fmt.Printf("\nPaste content\n%s%s%s", Green, data.Text, Reset)
		} else {
			fmt.Printf("%s", data.Text)
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
		fmt.Printf("%sError with nym-client connection %s%s. Is it started ?\n", Red, uri, Reset)
		panic(err)
	}

	return conn
}

func newPaste(text string, encryptionParams encParams, selfAddress string, public bool, ipfs bool, burn bool) idNewPaste {

	var paste pasteAdd
	if encryptionParams.Salt == "" {
		paste = pasteAdd{
			Event:  newText,
			Sender: selfAddress,
			Data: userDataAdd{
				Text:    text,
				Private: !public,
				Burn:    burn,
				Ipfs:    ipfs,
			},
		}
	} else {
		paste = pasteAdd{
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
	}

	receivedMessage := sendTextWithReply(&paste)
	messageByte := []byte(receivedMessage.Message)[9:]

	var dataUrl idNewPaste
	err := json.Unmarshal(messageByte, &dataUrl)
	if err != nil {
		panic(err)
	}

	return dataUrl
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

		content = []byte(decrypt(userKey, decodedText, encParams))
	} else {
		content = []byte(decodedText)
	}
	var clearObjectUser clearObjectUser
	err = json.Unmarshal(content, &clearObjectUser)
	if err != nil {
		panic(err.Error())
	}
	if clearObjectUser.File.Filename != "" {
		fmt.Printf("%sFile are not supported in pastenym CLI%s\n", Red, Reset)
	}

	return clearObjectUser

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

	if !*silent {

		fmt.Printf("waiting to receive a message from the mix network\n")
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

func encrypt(plaintext []byte) (string, string, encParams) {
	passphrase, key, salt := genKey(32)

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}

	aesGCM, err := cipher.NewGCM(block)

	if err != nil {
		panic(err.Error())
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}

	ciphertextBytes := aesGCM.Seal(nonce, nonce, plaintext, nil)

	cipherTextString := base64.StdEncoding.EncodeToString(ciphertextBytes[aesGCM.NonceSize():])
	//keyEncoded := base64.StdEncoding.EncodeToString(key)

	// parameters used by sjcl
	encParams := encParams{
		Iv:     base64.StdEncoding.EncodeToString(nonce),
		Salt:   base64.StdEncoding.EncodeToString(salt),
		Ks:     256,
		V:      1,
		Mode:   "gcm",
		Cipher: "aes",
		Iter:   10000,
		Ts:     128,
	}

	return passphrase, cipherTextString, encParams
}

func decrypt(passphrase string, encryptedString string, encParams encParams) string {

	decodeSalt, _ := base64.StdEncoding.DecodeString(encParams.Salt)
	key, _ := deriveKey(passphrase, []byte(decodeSalt))
	encryptedText, err := base64.StdEncoding.DecodeString(encryptedString)
	if err != nil {
		panic(err.Error())
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}

	aesGCM, err := cipher.NewGCM(block)

	if err != nil {
		fmt.Printf("\n%sUnable to decrypt%s\n", Red, Reset)
		if *debug {
			panic(err.Error())
		}
	}

	//adataDecode, err := base64.StdEncoding.DecodeString(encParams.Adata)
	ivDecode, _ := base64.StdEncoding.DecodeString(encParams.Iv)
	if err != nil {
		panic(err.Error())
	}

	plaintext, err := aesGCM.Open(nil, ivDecode, encryptedText, nil)
	if err != nil {
		fmt.Printf("\n%sUnable to decrypt%s\n", Red, Reset)
		if *debug {
			panic(err.Error())
		}
	}

	return string(plaintext)

}

func genKey(size uint8) (string, []byte, []byte) {
	const letters = "0123456789abcdefghijklmnopqrstuvwxyz"
	passphrase := make([]byte, size)

	var i uint8
	for i = 0; i < size; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			panic(err.Error())
		}
		passphrase[i] = letters[num.Int64()]
	}

	key, salt := deriveKey(string(passphrase), nil)
	return string(passphrase), key, salt //encode key in bytes to string for saving

}

func deriveKey(passphrase string, salt []byte) ([]byte, []byte) {
	if salt == nil {
		salt = make([]byte, 8)
		// http://www.ietf.org/rfc/rfc2898.txt
		// Salt.
		rand.Read(salt)
	}
	return pbkdf2.Key([]byte(passphrase), salt, 10000, 32, sha256.New), salt
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
