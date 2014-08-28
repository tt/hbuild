package hbuild

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
)

type herokuClient struct {
	httpClient *http.Client
	token      string
	url        *url.URL
	version    string
	userAgent  string
}

type herokuRequest struct {
	method string
	path   string
	body   interface{}
}

func newHerokuClient(token string) herokuClient {
	herokuUrl, _ := resolveHerokuClientUrl()

	return herokuClient{
		httpClient: http.DefaultClient,
		token:      token,
		url:        herokuUrl,
		version:    "application/vnd.heroku+json; version=edge",
		userAgent:  "hbuild/1",
	}
}

func resolveHerokuClientUrl() (*url.URL, error) {
	if s := os.Getenv("HEROKU_API_URL"); s != "" {
		return url.Parse(s)
	}

	return url.Parse("https://api.heroku.com")
}

func (hc herokuClient) request(hrequest herokuRequest, v interface{}) (err error) {
	var requestBody io.Reader

	url := hc.url.String() + hrequest.path

	if hrequest.body != nil {
		requestJson, err := json.Marshal(hrequest.body)
		if err != nil {
			return err
		}
		requestBody = bytes.NewReader(requestJson)
	}

	request, err := http.NewRequest(hrequest.method, url, requestBody)
	if err != nil {
		return
	}

	request.SetBasicAuth("", hc.token)
	request.Header.Set("Accept", hc.version)
	request.Header.Set("User-Agent", hc.userAgent)
	if hrequest.body != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	response, err := hc.httpClient.Do(request)
	if err != nil {
		return
	}

	if response.StatusCode/100 != 2 {
		var herr HerokuJsonError
		err = json.NewDecoder(response.Body).Decode(&herr)
		if err != nil {
			return err
		}
		return HerokuError{errors.New(herr.Message), herr.Id, herr.URL}
	}

	err = json.NewDecoder(response.Body).Decode(&v)
	return
}

type HerokuJsonError struct {
	Message string
	Id      string
	URL     string `json:"url"`
}

type HerokuError struct {
	error
	Id  string
	URL string
}
