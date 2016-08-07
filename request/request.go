package request

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

type ctxKeyType string

const (
	// GemNameKey is
	GemNameKey ctxKeyType = "name"
	// BaseURLKey is
	BaseURLKey ctxKeyType = "url"
)

// Gem is a gem.
type Gem struct {
	Name    string
	Version string
}

func extractGemName(ctx context.Context) string {
	v := ctx.Value(GemNameKey)
	return v.(string)
}

func extractBaseURL(ctx context.Context) string {
	v := ctx.Value(BaseURLKey)
	return v.(string)
}

// GetGemInfo returns the latest version of gem.
func GetGemInfo(ctx context.Context) (Gem, error) {
	name := extractGemName(ctx)
	baseURL := extractBaseURL(ctx)
	url := baseURL + "/api/v1/gems/" + name + ".json"
	tr := &http.Transport{}
	client := &http.Client{Transport: tr}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return Gem{}, err
	}
	errCh := make(chan error, 1)
	infoCh := make(chan Gem, 1)
	go func() {
		log.Printf("%s: Start request\n", name)
		resp, err := client.Do(req)
		log.Printf("%s: End request\n", name)
		if err != nil {
			errCh <- err
			return
		}
		defer resp.Body.Close()

		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			errCh <- err
			return
		}
		ret := Gem{}
		json.Unmarshal(b, &ret)
		log.Printf("%s: version %s", name, ret.Version)
		infoCh <- ret
	}()

	select {
	case err = <-errCh:
		return Gem{}, errors.Wrap(err, "Unexpected error occurred in fetching "+name+" info")
	case info := <-infoCh:
		return info, nil
	case <-ctx.Done():
		log.Println(name + ": Aborted.")
		tr.CancelRequest(req)
		return Gem{}, fmt.Errorf("Task %s is aborted.", name)
	}
}
