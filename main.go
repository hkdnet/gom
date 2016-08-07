package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/net/context"

	"github.com/pkg/errors"
)

func main() {
	os.Exit(run())
}

func run() int {
	_, err := newConfig()
	if err != nil {
		log.Fatal(err)
		return 1
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	gems := []string{"rails", "activesupport", "rspec"}
	errCh := make(chan error, 1)
	doneCh := make(chan string, 1)
	for _, gem := range gems {
		child := context.WithValue(ctx, "name", gem)
		go func(child context.Context) {
			info, err := getGemInfo(child)
			if err != nil {
				errCh <- err
			} else {
				doneCh <- fmt.Sprintf(`add_dependency "%s", "%s"`, info.Name, info.Version)
			}
		}(child)
	}
	for i := 0; i < len(gems); i++ {
		select {
		case err := <-errCh:
			if err != nil {
				log.Fatal(errors.Wrap(err, "Something happened. Abort all requests."))
				cancel()
				return 1
			}
		case str := <-doneCh:
			fmt.Println(str)
		}
	}

	return 0
}

type gemInfo struct {
	Name    string
	Version string
}

func getGemInfo(ctx context.Context) (gemInfo, error) {
	v := ctx.Value("name")
	name := v.(string)
	url := "https://rubygems.org/api/v1/gems/" + name + ".json"
	tr := &http.Transport{}
	client := &http.Client{Transport: tr}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return gemInfo{}, err
	}
	errCh := make(chan error, 1)
	infoCh := make(chan gemInfo, 1)
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
		ret := gemInfo{}
		json.Unmarshal(b, &ret)
		log.Printf("%s: version %s", name, ret.Version)
		infoCh <- ret
	}()

	select {
	case err = <-errCh:
		return gemInfo{}, errors.Wrap(err, "Unexpected error occurred in fetching "+name+" info")
	case info := <-infoCh:
		return info, nil
	case <-ctx.Done():
		log.Println(name + ": Aborted.")
		tr.CancelRequest(req)
		return gemInfo{}, fmt.Errorf("Task %s is aborted.", name)
	}
}
