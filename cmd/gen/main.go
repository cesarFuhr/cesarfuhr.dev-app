package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	_ "embed"
)

func main() {
	sourceFolder := "../../content/"
	sourceFileName := "2022_02_07-distributed_rate_limiting_in_go.md"
	destFolder := "../app/public/md/"

	sourceBytes, err := os.ReadFile(sourceFolder + sourceFileName)
	if err != nil {
		panic(err)
	}

	dateString, titleString, found := strings.Cut(sourceFileName, "-")
	if !found {
		panic("wrong file format")
	}

	date, err := time.Parse("2006_01_02", dateString)
	if err != nil {
		panic(err)
	}

	unformatedTitle := strings.TrimSuffix(titleString, ".md")
	destFileName := unformatedTitle + ".html"

	destFile, err := os.OpenFile(destFolder+destFileName, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}

	caser := cases.Title(language.English)
	title := caser.String(strings.ReplaceAll(unformatedTitle, "_", " "))

	destFile.Write(header(title, date.Format("2006-01-02"), "/images/token_bucket.svg"))
	destFile.Write(mdToHTML(sourceBytes))
	destFile.Write(footer("simple-rules-to-avoid-some-range-for-loop-pitfalls.html", ""))
}

func mdToHTML(md []byte) []byte {
	// create markdown parser with extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock | parser.FencedCode
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)

	// create HTML renderer with extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	return markdown.Render(doc, renderer)
}

var headerTemplate = template.Must(template.New("header").Parse(headerText))

var buf bytes.Buffer

func header(title, date, imagePath string) []byte {
	buf.Reset()

	args := struct {
		Title string
		Date  string
		Image string
	}{
		Title: title,
		Date:  date,
		Image: imagePath,
	}
	err := headerTemplate.Execute(&buf, args)
	if err != nil {
		panic(err)
	}

	return buf.Bytes()
}

const headerText = `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link rel="stylesheet" type="text/css" href="/style-md.css" />
    <link rel="stylesheet" type="text/css" href="/prism.css" />

    <title>{{.Title}} - cesarFuhr.dev</title>
    <meta name="author" content="César Fuhr">
    <meta name="image" property="og:image" content="{{.Image}}">
    <meta name="publish_date" property="og:publish_date" content="{{.Date}}">
    <link rel="icon" href="/images/cesar_gopher.ico">
  </head>

  <body>
    <a id="top"></a>
    <nav class="navbar-wrapper">
      <ul class="navbar" id="navbar">
        <li class="navbar-header">
          <div class="navbar-brand">
            <a class="nav-link" href="/">
              <img src="/images/cesar_gopher.png" id="gopher"/>
            </a>
            <a class="nav-link" href="/">cesarfuhr.dev</a>
          </div>
          <a class="nav-icon" href="javascript:void(0)" onclick="dropMenu()">||</a>
        </li>
        <li class="nav-item">
          <a class="nav-link" href="/archive.html">Archive</a>
        </li>
        <li class="nav-item">
          <a class="nav-link" href="/about.html">About</a>
        </li>
        <!--- <li class="nav-item">
          <a class="nav-link" href="cesarfuhr.rss">RSS</a>
          </li> --->
      </ul>
    </nav>

    <main>
`

var footerTemplate = template.Must(template.New("footer").Parse(footerText))

func footer(prev, next string) []byte {
	buf.Reset()

	var args struct {
		Prev string
		Next string
	}

	if prev != "" {
		args.Prev = fmt.Sprintf("<a href=\"/blog/%s\">prev</a>", prev)
	}
	if next != "" {
		args.Next = fmt.Sprintf("<a href=\"/blog/%s\">next</a>", next)
	}

	err := footerTemplate.Execute(&buf, args)
	if err != nil {
		panic(err)
	}

	return buf.Bytes()
}

const footerText = `<footer>
        {{.Prev}}
        <a href="#top">top</a>
        {{.Next}}
      </footer>
    </main>

    <script src="/js/dropMenu.js"  type="text/javascript"></script>
    <script src="/js/prism.js"     type="text/javascript"></script>
  </body>
</html>
`
