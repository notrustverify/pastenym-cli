package utils

import (
	"errors"
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/gabriel-vasile/mimetype"
)

type File struct {
	Data     []byte `json:"data"`
	Filename string `json:"filename"`
	MimeType string `json:"mimeType"`
}

// This func must be Exported, Capitalized, and comment added.
func CreateFile(filename string, mimeType string, content *[]byte) (bool, string) {

	if _, err := os.Stat(filename); err == nil {

		return false, "file exists"
	} else if errors.Is(err, os.ErrNotExist) {
		f, err := os.Create(filename)
		if err != nil {
			panic(err.Error())
		}

		defer f.Close()

		f.Write(*content)
		return true, filename

	} else {
		return false, "unknown error"
	}

}

func ReadFile(filePath string) (bool, File, string) {
	filename := filePath[strings.LastIndex(filePath, "/")+1:]

	if fStat, err := os.Stat(filePath); err == nil {

		fmt.Println(fStat.Size())
		if fStat.Size() >= int64(6*math.Pow10(6)) {
			return false, File{Filename: filename}, "File too big for the mixnet"
		}
		content, err := os.ReadFile(filePath)
		mtype := mimetype.Detect(content)
		if err != nil {
			panic(err.Error())
		}

		return true, File{Data: content, Filename: filename, MimeType: mtype.String()}, ""
	} else if errors.Is(err, os.ErrNotExist) {
		return false, File{Filename: filename}, "not exists"
	} else {
		return false, File{}, "unknown error"
	}

}
