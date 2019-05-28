package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gocolly/colly"
)

const (
	DocsUrl  = "http://127.0.0.1:4567/docs/"
	CacheDir = "cache/"
)

func main() {
	c := colly.NewCollector()

	// Find and visit all doc pages
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		url := e.Attr("href")
		if !strings.HasPrefix(url, "/docs/builders") {
			return
		}
		e.Request.Visit(url)
	})

	c.OnHTML("#optional- + ul a[name]", func(e *colly.HTMLElement) {

		name := e.Attr("name")

		builder := e.Request.URL.Path[strings.Index(e.Request.URL.Path, "/builders/")+len("/builders/"):]
		builder = strings.TrimSuffix(builder, ".html")

		text := e.DOM.Parent().Text()
		text = strings.ReplaceAll(text, "\n", " ")
		text = strings.TrimSpace(text)
		text = text[strings.Index(text, ") -")+len(") -"):]

		builderPath := strings.Split(builder, "-")[0]
		// fmt.Printf("required: %25s builderPath: %20s text: %20s\n", name, builderPath, text)

		err := filepath.Walk("./builder/"+builderPath, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() || filepath.Ext(path) != ".go" {
				return nil
			}
			body, err := ioutil.ReadFile(path)
			if err != nil {
				panic(err)
			}
			regex := regexp.MustCompile(fmt.Sprintf(`(\n\s+//.*)?\n(\s*)([A-Z]\w+\s+\w+\s+.*mapstructure:"%s")(\s+required:"true")?(.*)`, name))

			replaced := regex.ReplaceAll(body, []byte("\n$2//"+text+"\n"+`$2$3 required:"false"$5`))

			if string(replaced) == string(body) {
				return nil
			}

			err = ioutil.WriteFile(path, replaced, 0)
			if err != nil {
				panic(err)
			}

			return nil
		})
		if err != nil {
			panic(err)
		}
	})

	c.CacheDir = CacheDir

	c.Visit(DocsUrl)
}
