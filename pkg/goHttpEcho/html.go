package goHttpEcho

import (
	"bytes"
	"fmt"
	"html/template"
)

// PageData holds all the data needed to render an HTML page.
type PageData struct {
	Title       string
	Description string
	Language    string
	Theme       string
	Content     []ContentBlock
}

// ContentBlock defines a generic block of content.
type ContentBlock struct {
	Type  string // e.g., "heading", "paragraph", "list"
	Value string
}

const tmpl = `<!DOCTYPE html>
<html lang="{{.Language}}" data-theme="light">
<head>
    <meta charset="UTF-8"/>
    <meta name="viewport" content="width=device-width, initial-scale=1.0"/>
    <meta name="description" content="{{.Description}}">
    <title>{{.Title}}</title>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@picocss/pico@2/css/pico.{{.Theme}}.min.css">
</head>
<body>
    <main class="container">
       <h4>{{.Title}}</h4>
       <section>
       {{range .Content}}
           {{if eq .Type "heading"}}
               <h5>{{.Value}}</h5>
           {{else if eq .Type "paragraph"}}
               <p>{{.Value}}</p>
           {{end}}
       {{end}}
       </section>
    </main>
</body>
</html>`

func GetHtmlPage(data PageData) (string, error) {
	t, err := template.New("page").Parse(tmpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func GetHtmlError(msg string) string {
	const tmpl = `<!DOCTYPE html><html lang="en"><head><meta charset="UTF-8"/><meta name="viewport" content="width=device-width, initial-scale=1.0"/><title>error occurred</title><link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@picocss/pico@2/css/pico.{{.Theme}}.min.css"></head>
<body><main class="container"><h4>An error occurred</h4><section><p>%s</p></section></main></body></html>`
	return fmt.Sprintf(tmpl, msg)
}
