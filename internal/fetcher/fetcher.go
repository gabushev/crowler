package fetcher

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type WebFetcher struct {
	acceptableMimeType map[string]bool
}

func NewWebFetcher(mimetypes []string) *WebFetcher {
	acceptableMime := make(map[string]bool)
	for _, mime := range mimetypes {
		acceptableMime[mime] = true
	}
	return &WebFetcher{
		acceptableMimeType: acceptableMime,
	}
}

func contains(list map[string]bool, item string) bool {
	for i := range list {
		if strings.Contains(item, i) {
			return true
		}
	}
	return false
}
func (wf WebFetcher) Download(urlString string) ([]byte, error) {
	client := &http.Client{
		CheckRedirect: noRedirect,
	}
	response, err := client.Get(urlString)
	if err != nil || response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unable to reach the address, %v", err)
	}

	if !contains(wf.acceptableMimeType, response.Header.Get("Content-Type")) {
		return nil, fmt.Errorf("unacceptable mime type: %s", response.Header.Get("Content-Type"))
	}

	defer response.Body.Close()

	return io.ReadAll(response.Body)
}

func noRedirect(req *http.Request, via []*http.Request) error {
	return errors.New("you shall not pass!")
}
