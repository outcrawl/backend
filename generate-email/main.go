package main

import (
	"bytes"
	"errors"
	"html/template"
	"io/ioutil"
	"log"
	"os"

	"github.com/asciimoo/colly"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"

	"github.com/BurntSushi/toml"
	"github.com/golang-commonmark/markdown"
	"github.com/urfave/cli"
)

var (
	templateFile string
	dataFile     string
	outputFile   string
)

type Email struct {
	Text       template.HTML
	References []string
	Articles   []Article
}

type Article struct {
	Title       string
	Description template.HTML
	URL         template.URL
}

func main() {
	app := cli.NewApp()
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "tmpl",
			Usage:       "Path to template file",
			Destination: &templateFile,
		},
		cli.StringFlag{
			Name:        "data",
			Usage:       "Path to data file",
			Destination: &dataFile,
		},
		cli.StringFlag{
			Name:        "out",
			Usage:       "Output file",
			Destination: &outputFile,
		},
	}
	app.Action = func(ctx *cli.Context) error {
		if len(templateFile) == 0 {
			return errors.New("Template file not specified")
		}
		if len(dataFile) == 0 {
			return errors.New("Data file not specified")
		}
		if len(outputFile) == 0 {
			outputFile = "email_gen.html"
		}
		return generate()
	}
	app.Description = "Generate emails from templates"
	app.Version = "0.0.1"

	if err := app.Run(os.Args); err != nil {
		log.Fatalln(err)
	}
	log.Printf("Generated %s", outputFile)
}

func generate() error {
	// Parse template
	tmpl, err := template.ParseFiles(templateFile)
	if err != nil {
		return err
	}
	// Create output file
	f, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer f.Close()
	// Generate email
	email, err := parseData()
	if err != nil {
		return err
	}

	buf := bytes.NewBufferString("")
	err = tmpl.Execute(buf, email)
	if err != nil {
		return err
	}

	min := minify.New()
	min.AddFunc("text/css", css.Minify)
	min.AddFunc("text/html", html.Minify)
	if err := min.Minify("text/html", f, buf); err != nil {
		return err
	}

	return nil
}

func parseData() (*Email, error) {
	data, err := ioutil.ReadFile(dataFile)
	if err != nil {
		return nil, err
	}
	email := &Email{}
	if _, err := toml.Decode(string(data), &email); err != nil {
		return nil, err
	}
	email.Text = template.HTML(markdown.New(markdown.XHTMLOutput(true)).RenderToString([]byte(email.Text)))
	if err := scrapeReferences(email); err != nil {
		return nil, err
	}
	return email, nil
}

func scrapeReferences(email *Email) error {
	for _, ref := range email.References {
		c := colly.NewCollector()
		a := Article{URL: template.URL(ref)}
		c.OnHTML("meta", func(e *colly.HTMLElement) {
			if e.Attr("property") == "og:title" {
				a.Title = e.Attr("content")
			}
			switch e.Attr("property") {
			case "og:title":
				a.Title = e.Attr("content")
			case "og:description":
				a.Description = template.HTML(e.Attr("content"))
			}
		})
		if err := c.Visit(ref); err != nil {
			return err
		}
		email.Articles = append(email.Articles, a)
	}
	return nil
}
