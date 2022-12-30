package main

import (
	"fmt"
)

func formatAddPasteVerbose(public bool, urlId string, ipfsHash string, key string) {

	fmt.Printf("\n\n%sID: %s", Green, urlId)

	if !public {
		fmt.Printf("%s\nKey: %s\n", Green, key)
		fmt.Printf("\nLink: https://%s/#/%s&key=%s%s", connectionData.instance, urlId, key, Reset)
	} else {
		fmt.Printf("\nLink: https://%s/#/%s%s", connectionData.instance, urlId, Reset)
	}

	if ipfsHash != "" {
		fmt.Printf("\nipfs://%s\n", ipfsHash)
	}

	fmt.Printf("\n")
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
	fmt.Printf("\n\n%sPaste content%s\n%s\n", Green, Reset, data.Text)
}

func formatGetPasteContentSilent(data *clearObjectUser) {
	fmt.Printf("%s", data.Text)

}
