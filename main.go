package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	// "strconv"
	"github.com/fjw95/link-checker/util"
	"sync"
)

var mu sync.Mutex
var count int

func main() {
	seedUrls := os.Args[1:]
	var wg sync.WaitGroup

	for _, url := range seedUrls {
		checkLink(url, &wg)
	}

	wg.Wait()

	if count > 0 {
		fmt.Println("\n---Found", count, "Dead Link URL---\n")
	} else {
		fmt.Println("\n---Not Found Dead Link URL---\n")
	}
}

func checkLink(url string, wg *sync.WaitGroup) {

	client := &http.Client{}
	resp, err := client.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	p := `<a\s+(?:[^>]*?\s+)?href="([^"]*)"`
	pattern, err := regexp.Compile(p)
	if err != nil {
		log.Fatal(err)
	}

	urls := []string{}
	found := pattern.FindAllStringSubmatch(string(body), -1)

	for _, url := range found {
		urls = append(urls, url[1])
	}

	urls = util.RemoveDuplicates(urls)

	wg.Add(len(urls))
	for _, urlLink := range urls {
		//fmt.Println(urlLink)
		go func(url string, wg *sync.WaitGroup) {
			defer wg.Done()

			client := &http.Client{}
			resp, err := client.Get(url)
			if err != nil {
				return
			}

			statusCode := resp.StatusCode

			if statusCode != 200 {
				mu.Lock()
				count++
				fmt.Println("URL : "+url+"\n Status Code : ", statusCode, "\n")
				mu.Unlock()
			}

		}(urlLink, wg)
	}

}
