package main

import (
	"fmt"
)

func formatAddPasteVerbose(public bool, urlId string, ipfsHash string, key string) {

	fmt.Printf("%sID: %s", Green, urlId)

	if !public {
		fmt.Printf("%s\nKey: %s\n", Green, key)
		fmt.Printf("\nLink: https://%s/#/%s&key=%s%s", connectionData.instance, urlId, key, Reset)
	} else {
		fmt.Printf("\nLink: https://%s/#/%s%s", connectionData.instance, urlId, Reset)
	}

	if ipfsHash != "" {
		fmt.Printf("\n%sipfs://%s%s", Green, ipfsHash, Reset)
	}
}

func formatAddPasteSilent(urlId string, key string) {
	fmt.Printf("%s %s", urlId, key)
}

func formatAddPasteOnlyUrl(urlId string, key string, instance string) {
	fmt.Printf("https://%s/#/%s", instance, urlId)

	if key != "" {
		fmt.Printf("&key=%s", key)

	}
}

func formatGetPasteContentVerbose(data *clearObjectUser) {
	fmt.Printf("\nPaste content\n%s%s%s", Green, data.Text, Reset)
}

func formatGetPasteContentSilent(data *clearObjectUser) {
	fmt.Printf("%s", data.Text)

}
