package main

import (
	"runtime"
	"strings"
)

var Reset = "\033[0m"
var Red = "\033[31m"
var Green = "\033[32m"
var Yellow = "\033[33m"
var Blue = "\033[34m"
var Purple = "\033[35m"
var Cyan = "\033[36m"
var Gray = "\033[37m"
var White = "\033[97m"

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
