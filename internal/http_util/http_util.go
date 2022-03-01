package http_util

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
)

var (
	httpClient  *http.Client
	BlankHeader = map[string]string{}
)

const urlPrefix = "http://"

func init() {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = 300
	t.MaxConnsPerHost = 300
	t.MaxIdleConnsPerHost = 100
	httpClient = &http.Client{Transport: t}
}

func get(remoteUrl string, headers map[string]string, respStructPointer interface{}) error {
	if !strings.HasPrefix(remoteUrl, urlPrefix) {
		remoteUrl = urlPrefix + remoteUrl
	}
	req, _ := http.NewRequest("GET", remoteUrl, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	httpResp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()
	body, _ := ioutil.ReadAll(httpResp.Body)
	err = json.Unmarshal(body, respStructPointer)
	if err != nil {
		return err
	}
	return nil
}

func Get(remoteUrl string, headers map[string]string, respStructPointer interface{}) (err error) {
	for retriedAmount := 0; retriedAmount < 3; retriedAmount++ {
		err = get(remoteUrl, headers, respStructPointer)
		if err == nil {
			return
		}
	}
	return
}

func post(remoteUrl string, headers map[string]string, bodyByte []byte) ([]byte, error) {
	if !strings.HasPrefix(remoteUrl, urlPrefix) {
		remoteUrl = urlPrefix + remoteUrl
	}
	req, err := http.NewRequest("POST", remoteUrl, bytes.NewBuffer(bodyByte))
	if err != nil {
		return []byte{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := httpClient.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return []byte{}, err
	}
	return ioutil.ReadAll(resp.Body)
}

func PostJson(
	remoteUrl string, headers map[string]string,
	bodyByte []byte, respIns interface{}) error {
	if headers == nil {
		headers = make(map[string]string)
	}
	headers["content-type"] = "application/json"

	respBodyByte, err := post(remoteUrl, headers, bodyByte)
	if err != nil {
		return err
	}
	return json.Unmarshal(respBodyByte, respIns)
}
