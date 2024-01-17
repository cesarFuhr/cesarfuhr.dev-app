package main

import (
	"bytes"
	"embed"
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
)

const sourceFolder = "../../content/"
const destFolder = "../app/public/"

func main() {
	dirEntries, err := os.ReadDir(sourceFolder)
	if err != nil {
		panic(err)
	}

	// Write the about page.
	sourceBytes, err := os.ReadFile(sourceFolder + "about.md")
	if err != nil {
		panic(err)
	}

	// Starting with the about page.
	aboutPage := page{
		Title:   "About",
		Date:    time.Date(2021, 12, 20, 0, 0, 0, 0, time.UTC),
		Image:   "/images/cesar_gopher.png",
		Content: mdToHTML(sourceBytes),
	}
	aboutFile, err := os.OpenFile(destFolder+"about.html", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)

	if err != nil {
		panic(err)
	}
	defer aboutFile.Close()

	aboutFile.Write(aboutPage.build())

	var blogPosts []page
	for _, entry := range dirEntries {
		if entry.IsDir() {
			// why? why? a directory here?
			panic("we shouldn't have dir in source folder")
		}

		if entry.Name() == "about.md" {
			continue
		}

		var prev string
		// If its not the first page, it has a previous.
		if len(blogPosts) != 0 {
			prev = blogPosts[len(blogPosts)-1].Dest
		}

		sourceFileName := entry.Name()

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

		title := caser.String(strings.ReplaceAll(unformatedTitle, "_", " "))

		blogPost := page{
			Source:  sourceFileName,
			Dest:    destFileName,
			Title:   title,
			Date:    date,
			Prev:    prev,
			HasCode: true,
			// Leaving Next to the next step
			// doing it here will be too complex.
		}
		blogPosts = append(blogPosts, blogPost)
	}

	// Write the blog pages.
	for i, blogPost := range blogPosts {
		sourceBytes, err := os.ReadFile(sourceFolder + blogPost.Source)
		if err != nil {
			panic(err)
		}

		// TODO: fix preview image.
		blogPost.Image = "/images/cesar_gopher.png"
		blogPost.Content = mdToHTML(sourceBytes)

		// Spinning up an inline function to be able to defer.
		func() {
			destFile, err := os.OpenFile(destFolder+"blog/"+blogPost.Dest, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
			if err != nil {
				panic(err)
			}
			defer destFile.Close()

			if i+1 < len(blogPosts) {
				blogPost.Next = blogPosts[i+1].Dest
			}

			pageBytes := blogPost.build()
			// TODO: fix preview image.
			destFile.Write(pageBytes)

			// If its the most recent page, it should be the index.
			if len(blogPosts) == i+1 {
				indexFile, err := os.OpenFile(destFolder+"index.html", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
				if err != nil {
					panic(err)
				}
				defer indexFile.Close()

				// TODO: fix preview image.
				indexFile.Write(pageBytes)
			}
		}()
	}

	// Write the archive page.
	archiveFile, err := os.OpenFile(destFolder+"archive.html", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer archiveFile.Close()

	archivePage := page{
		Title:   "Archive",
		Date:    time.Date(2021, 12, 20, 0, 0, 0, 0, time.UTC),
		Image:   "/images/cesar_gopher.png",
		Content: archive(blogPosts),
	}
	archiveFile.Write(archivePage.build())
}

var caser = cases.Title(language.English)

type page struct {
	Title   string
	Date    time.Time
	Image   string
	Content []byte
	HasCode bool

	Source string
	Dest   string

	Prev string
	Next string
}

func mdToHTML(md []byte) []byte {
	// create markdown parser with extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock | parser.FencedCode
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)

	// create HTML renderer with extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank | html.LazyLoadImages
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	return markdown.Render(doc, renderer)
}

//go:embed templates/*
var templates embed.FS

var pageTemplate = template.Must(template.New("page.html").ParseFS(templates, "templates/page.html"))

var buf bytes.Buffer

func (p page) build() []byte {
	buf.Reset()

	args := struct {
		Title   string
		Date    string
		Image   string
		Content string
		Prev    string
		Next    string
		HasCode bool
	}{
		Title:   p.Title,
		Date:    p.Date.Format("2006-01-02"),
		Image:   p.Image,
		Content: string(p.Content),
		HasCode: p.HasCode,
	}

	if p.Prev != "" {
		args.Prev = fmt.Sprintf("<a href=\"/blog/%s\">prev</a>", p.Prev)
	}
	if p.Next != "" {
		args.Next = fmt.Sprintf("<a href=\"/blog/%s\">next</a>", p.Next)
	}

	err := pageTemplate.Execute(&buf, args)
	if err != nil {
		panic(err)
	}

	return buf.Bytes()
}

var archiveTemplate = template.Must(template.New("archive").Parse(archiveText))

func archive(pages []page) []byte {
	buf.Reset()

	type item struct {
		Date  string
		Title string
		Dest  string
	}

	args := struct{ Items []item }{Items: make([]item, len(pages))}
	for i, page := range pages {
		item := item{
			Date:  page.Date.Format("2006/01/02"),
			Title: page.Title,
			Dest:  "/blog/" + page.Dest,
		}
		args.Items[len(pages)-1-i] = item
	}

	err := archiveTemplate.Execute(&buf, args)
	if err != nil {
		panic(err)
	}

	return buf.Bytes()
}

const archiveText = `
      <header>
        <h1>Archive</h1>
      </header>

      <section class="archive">
        <ol class="archive-list">
          {{range $index, $element := .Items}}
          <li>
            <span class="date">
              {{$element.Date}} - 
            </span>
            <a href="{{$element.Dest}}">
              {{$element.Title}}
            </a>
          </li>
          {{end}}
        </ol>
      </section>
`
