package utils

import (
	"bytes"
	"fmt"
	"graduation/internal/entity"
	"text/template"
)

func GenerateHTML(event *entity.Event) (string, error) {
	htmlCode := `<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>{{.Title}}</title>
	</head>
	<body>
		<h1>{{.Title}}</h1>
		
		{{if .PhotoURLs}}
			{{range .PhotoURLs}}
				<img src="{{.}}" alt="Event Image" style="max-width: 50%;">
			{{end}}
		{{else}}
			<p>No images available</p>
		{{end}}

		<p>Description: {{.Description}}</p>
		<p>Date: {{.Date.Format "2006-01-02 15:04:05"}}</p>
		<p>Place: {{.Place}}</p>
		<p>Participants: {{.Participants}} / {{.MaxParticipants}}</p>
	</body>
	</html>
	`
	tmpl := template.New("emailTemplate")
	tmpl, err := tmpl.Parse(htmlCode)
	if err != nil {
		return "", fmt.Errorf("cannot template: %w", err)
	}

	var tplBuffer bytes.Buffer
	err = tmpl.Execute(&tplBuffer, event)
	if err != nil {
		return "", fmt.Errorf("cannot Execute: %w", err)
	}

	htmlBody := tplBuffer.String()

	return htmlBody, nil
}
