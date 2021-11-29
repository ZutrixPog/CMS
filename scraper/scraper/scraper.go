package scraper

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/geziyor/geziyor"
	"github.com/geziyor/geziyor/client"
)

type Scraper struct {
	Urls    []string
	Plugins []string
}

func validUrls(urls []string) (bool, []string) {
	failedUrls := []string{}
	r, _ := regexp.Compile(`https?:\/\/(www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_\+.~#?&//=]*)`)
	for i := range urls {
		if !(r.MatchString(urls[i])) {
			failedUrls = append(failedUrls, urls[i])
		}
	}
	if failedUrls == nil {
		return true, failedUrls
	}
	return false, failedUrls
}

func New(urls []string) (*Scraper, error) {
	areValid, failedUrls := validUrls(urls)
	if !areValid {
		fmt.Printf("Some of your Urls are wrong Please fix these and submit again! : %v", failedUrls)
		return nil, errors.New("Please enter all valid Urls!\n")
	}
	return &Scraper{
		Urls:    urls,
		Plugins: []string{},
	}, nil
}

func (s *Scraper) Initialize() {
	geziyor.NewGeziyor(&geziyor.Options{
		StartURLs: s.Urls,
		ParseFunc: s.parserFunc,
	}).Start()
}

func (s *Scraper) parserFunc(g *geziyor.Geziyor, r *client.Response) {
	r.HTMLDoc.Find("")
}
