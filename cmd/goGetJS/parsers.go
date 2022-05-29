package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func parseDoc(r io.Reader, baseURL string) ([]string, int, error) {
	scriptsSRC := []string{}
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return scriptsSRC, 0, fmt.Errorf("could not read HTML with goquery: %v", err)
	}

	j := 0

	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		// scripts with src
		if value, ok := s.Attr("src"); ok {
			if !strings.HasPrefix(value, "http") {
				if !strings.HasPrefix(value, "/") {
					value = fmt.Sprintf("/%s", value)
				}
				scriptsSRC = append(scriptsSRC, baseURL+value)
			} else {
				scriptsSRC = append(scriptsSRC, value)
			}
		} else {
			// scripts without src
			script := strings.TrimSpace(s.Text())

			// write to file
			scriptByte := []byte(script)
			j++
			scriptName := fmt.Sprintf("anon%s.js", strconv.Itoa(j))
			if err := os.WriteFile("data/"+scriptName, scriptByte, 0644); err != nil {
				log.Println("could not write anon script", err)
				j--
			}
		}
	})

	if len(scriptsSRC) != 0 {
		return scriptsSRC, j, nil
	}

	return scriptsSRC, j, fmt.Errorf("no src found on page")
}

func getJS(client *http.Client, url string) error {
	log.Println("getting script at:", url)
	res, err := makeRequest(url, client)
	if err != nil {
		return fmt.Errorf("could not make script request: %v", err)
	}
	err = parseScripts(res)
	if err != nil {
		return fmt.Errorf("no script available: %v", err)
	}

	return nil
}

func parseScripts(res *http.Response) error {
	defer res.Body.Close()
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return fmt.Errorf("unable to parse script page: %q", err)
	}
	script := doc.Find("body").Text()
	currentURL := *res.Request.URL
	url := currentURL.String()

	if script != "" {
		err := writeScripts(script, url)
		if err != nil {
			return fmt.Errorf("unable to write script file: %q", err)
		}
		return nil
	}

	return fmt.Errorf("no scripts at %v", url)
}
