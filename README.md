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
  --burn
    	Specify if the text have to be deleted when read. Default is false
  --view int
    	Specify if the text have to be deleted when read.
  --debug
    	Specify if the text have to be deleted when read. Default is false
  --id string
    	Specify paste url id to retrieve. Default is empty
  --instance string
    	Instance where to get the paste from GUI (default "pastenym.ch")
  --ipfs
    	Specify if the text to share is stored on IPFS. Default is false
  --key string
    	Key for getting the plaintext
  --nymclient string
    	Nym client to connect (default "127.0.0.1:1977")
  --provider string
    	Specify the path for a file to share. Default is empty (default "6y7sSj3dKp5AESnet1RQXEHmKkEx8Bv3...")
  --public
    	Set the paste to public, i.e without encryption. Default is private
  --silent
    	Remove every output, just print data. Default is false
  --text string
    	Specify the text to share. Mandatory
  --url
    	Only print the URL. Default is false
  --ping
    	Ping the backend to see if it's alive. Return the version

```
