# CLI for pastenym

## Prerequisites

* A running [Nym-client](https://nymtech.net/docs/stable/integrations/websocket-client)

## Installation

```bash
git clone https://github.com/notrustverify/pastenym-cli/
cd pastenym-cli
go build
```


## Quick start

### Share a paste
```bash
./pastenym -text "my text"
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
Usage of ./pastenym:
  -burn
    	Specify if the text have to be deleted when read. Default is false
  -debug
    	Specify if the text have to be deleted when read. Default is false
  -id string
    	Specify paste url id to retrieve. Default is empty
  -instance string
    	Instance where to get the paste from GUI (default "pastenym.ch")
  -ipfs
    	Specify if the text to share is stored on IPFS. Default is false
  -key string
    	Key for getting the plaintext
  -nymclient string
    	Nym client to connect. Default 127.0.0.1:1977 (default "127.0.0.1:1977")
  -provider string
    	Specify the path for a file to share. Default is empty (default "6y7sSj3dKp5AESnet1RQXEHmKkEx8Bv3RgwEJinGXv6J.FZfu6hNPi1hgQfu7crbXXUNLtr3qbKBWokjqSpBEeBMV@EBT8jTD8o4tKng2NXrrcrzVhJiBnKpT1bJy5CMeArt2w")
  -public
    	Set the paste to public, i.e without encryption. Default is private
  -silent
    	Remove every output, just print data. Default is false
  -text string
    	Specify the text to share. Mandatory
```
