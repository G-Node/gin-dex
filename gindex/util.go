package gindex

import (
	"net/http"
	"io/ioutil"
	"encoding/json"
	"io"
	"fmt"
	log "github.com/Sirupsen/logrus"
)

func getParsedBody(r *http.Request, obj interface{}) error {
	data, err := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		log.Debugf("Could not read request body: %+v", err)
		return err
	}
	err = json.Unmarshal(data, obj)
	if err != nil {
		log.Debugf("Could not unmarshal request: %+v, %s", err, string(data))
		return err
	}
	return nil
}

func getParsedResponse(resp *http.Response, obj interface{}) error {
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, obj)
}

func getParsedHttpCall(method, path string, body io.Reader, token, csrfT string, obj interface{}) error {
	client := &http.Client{}
	req, _ := http.NewRequest(method, path, body)
	req.Header.Set("Cookie", fmt.Sprintf("i_like_gogits=%s; _csrf=%S", token, csrfT))
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if (resp.StatusCode != http.StatusOK) {
		return fmt.Errorf("Not Authorized")
	}
	return getParsedResponse(resp, obj)
}

// Encodes a given map into a struct.
// Lazyly Uses json package instead of reflecting directly
func map2struct(in interface{}, out interface{}) error {
	data, err := json.Marshal(in)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, out)
}
