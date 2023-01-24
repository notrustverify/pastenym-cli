# CLI for pastenym

[![asciicast](https://asciinema.org/a/548628.svg)](https://asciinema.org/a/548628)

## Prerequisites

* A running [Nym-client](https://nymtech.net/docs/stable/integrations/websocket-client)

## Installation

Grab the latest release: https://github.com/notrustverify/pastenym-cli/releases/latest

or compile it yourself

1. Install [Go](https://go.dev/doc/install)
2. Get the repo and compile
```bash
git clone https://github.com/notrustverify/pastenym-cli/
cd pastenym-cli
make
```


## Quick start

### Share a paste
```bash
./pastenym -text "my text"
```

#### With file
```bash
./pastenym -text "my text" -file mixnethowto.pdf
```

### Pipe a paste

```bash
echo "my text" | ./pastenym
```

### Get a paste

#### From key and id

```bash
./pastenym -id x4jO7s4W -key b5hstfjtd6ojkuwsj9a46di964qreocf
```
#### From URL

```bash
./pastenym -id https://pastenym.ch/#/x4jO7s4W&key=b5hstfjtd6ojkuwsj9a46di964qreocf
```


## Usage


```
Usage of pastenym:
  -burn
    	Specify if the text have to be deleted when read. Default is false
  -debug
    	Specify if the text have to be deleted when read. Default is false
  -file string
    	Specify the path for a file to share. Default is empty
  -height int
    	Specify a Bitcoin block height when the paste have to be deleted (default -1)
  -id string
    	Specify paste url id to retrieve. Default is empty
  -instance string
    	Instance where to get the paste from GUI (default "pastenym.ch")
  -ipfs
    	Specify if the text to share is stored on IPFS. Default is false
  -key string
    	Key for getting the plaintext
  -nymclient string
    	Nym client to connect (default "127.0.0.1:1977")
  -ping
    	Ping the backend to see if it's alive. Return the version
  -provider string
    	Specify the provider. (default "HWm3757chNdBq9FzKEY9j9VJ5siRxH8ukrNqYwFp9Unp.D34iYLRd5vzpCU4nZRcFVmoZpTQQMa6mws4Q65LdRosi@Fo4f4SQLdoyoGkFae5TpVhRVoXCF8UiypLVGtGjujVPf")
  -public
    	Set the paste to public, i.e without encryption. Default is private
  -silent
    	Remove every output, just print data. Default is false
  -text string
    	Specify the text to share. Mandatory
  -time string
    	Specify a relative time interval when the paste have to be deleted. For example 1d, 1m, 10h
  -url
    	Only print the URL. Default is false
  -view int
    	Specify if the text have to be deleted when read.

```
