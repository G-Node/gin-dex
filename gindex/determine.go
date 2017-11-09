package gindex

import (
	"bufio"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/G-Node/gogs/pkg/tool"
)

const (
	UKKNOWN = iota
	ANNEX
	ODML_XML
	TEXT
)

func DetermineFileType(peekData []byte) (int64, error) {
	if tool.IsAnnexedFile(peekData){
		return ANNEX,nil
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
func BlobFileType(blob *IndexBlob) (int64, error) {
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
	return DetermineFileType(peekData)

}
