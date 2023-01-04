package main

import (
	"fmt"
	"os"
	F "pastenym-cli/utils"
)

func formatAddPasteVerbose(public bool, urlId string, ipfsHash string, key string) {

	fmt.Printf("\n\n%sID: %s", Green, urlId)

	var link string
	if !public {
		fmt.Printf("%s\nKey: %s\n", Green, key)
		link = fmt.Sprintf("https://%s/#/%s&key=%s", connectionData.instance, urlId, key)

	} else {
		link = fmt.Sprintf("https://%s/#/%s", connectionData.instance, urlId)
	}
	fmt.Printf("\nLink: %s%s\n", link, Reset)

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

	if userData.File.Filename != "" {
		currDir, _ := os.Getwd()

		success, filename := F.CreateFile(userData.File.Filename, userData.File.MimeType, &userData.File.Data)
		if success {
			fmt.Printf("%s\nFile created:%s %s/%s", Green, Reset, currDir, filename)
		} else {
			fmt.Printf("%s\nFile already exists:%s %s", Red, Reset, userData.File.Filename)
		}
	}
	if userData.Text != "" {
		fmt.Printf("%s\nPaste content%s\n%s", Green, Reset, userData.Text)
	}
	fmt.Printf("\n")
}

func formatGetPasteContentSilent(userData *clearObjectUser) {
	fmt.Printf("%s", userData.Text)

}
