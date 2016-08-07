package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"golang.org/x/net/context"

	"github.com/hkdnet/gom/request"
	"github.com/pkg/errors"
)

func main() {
	os.Exit(run())
}

var (
	usePrivate bool
	baseURL    string
	token      string
)

func run() int {
	/*
		flag.BoolVar(&usePrivate, "p", false, "fetch private gem version")
	*/
	flag.StringVar(&baseURL, "u", "https://rubygems.org/", "base url")
	flag.Parse()
	if usePrivate {
		config, err := newConfig()
		if err != nil {
			log.Fatal(err)
			return 1
		}
		token = config.token
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	gems := flag.Args()
	errCh := make(chan error, 1)
	doneCh := make(chan string, 1)
	for _, gemName := range gems {
		child := context.WithValue(ctx, request.GemNameKey, gemName)
		child = context.WithValue(child, request.BaseURLKey, baseURL)
		go func(child context.Context) {
			gem, err := request.GetGemInfo(child)
			if err != nil {
				errCh <- err
			} else {
				doneCh <- fmt.Sprintf(`add_dependency "%s", "%s"`, gem.Name, gem.Version)
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
