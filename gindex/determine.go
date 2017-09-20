package gindex

import (
	"bufio"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	UKKNOWN = iota
	ODML_XML
	TEXT
)

func DetermineFileType(blob *IndexBlob) (int64, error) {
	var peekData []byte
	if blob.Size() > 1024 {
		peekData, err := bufio.NewReader(blob).Peek(1024)
		if err != nil {
			return UKKNOWN, err
		}
		peekData = peekData
	} else {
		peekData, err := ioutil.ReadAll(blob)
		if err != nil {
			return UKKNOWN, err
		}
		peekData = peekData
	}
	typeStr := http.DetectContentType(peekData)
	if strings.Contains(typeStr, "text") {
		if strings.Contains(string(peekData), "ODML") {
			return ODML_XML, nil
		}
		return TEXT, nil
	}
	return UKKNOWN, nil

}
