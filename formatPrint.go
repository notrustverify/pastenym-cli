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

func formatGetPasteContentVerbose(metadata *textRetrieved, userData *clearObjectUser) {

	fmt.Printf("\n\nCreated on: %s", metadata.CreatedOn)

	if metadata.NumView > 0 && !metadata.Burn {
		fmt.Printf(" - Num views: %d - ", metadata.NumView)
	} else {
		fmt.Printf(" - ")
	}
	fmt.Printf("ipfs: %t", metadata.Ipfs)

	if metadata.Burn {
		fmt.Printf("\n%sThe paste is now deleted%s\n", Yellow, Reset)
	}

	fmt.Printf("%s\nPaste content%s\n%s\n", Green, Reset, userData.Text)
}

func formatGetPasteContentSilent(userData *clearObjectUser) {
	fmt.Printf("%s", userData.Text)

}
