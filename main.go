package main

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/jomei/notionapi"
)

// Page represents the data needed for rendering the HTML page
type Page struct {
	APIKey      string
	ContentType string
	NotionID    string
}

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/export", exportHandler)

	fmt.Println("Server is running on :8080")
	http.ListenAndServe(":8080", nil)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	// Render the HTML page with the form
	tmpl, err := template.New("index").Parse(`
    
	<html>
	<head>
		<title>Notion Exporter</title>
	</head>
	<body>
		<h1>Notion Exporter</h1>
		<form action="/export" method="get">
			<label for="api_key">Notion API Key:</label>
			<input type="text" name="api_key" required><br><br>

			<label for="content_type">Select Notion Entity:</label>
			<select name="content_type" required>
				<option value="database">Database</option>
				<option value="block">Block</option>
				<option value="page">Page</option>
			</select><br><br>

			<label for="notion_id">Notion ID:</label>
			<input type="text" name="notion_id" required><br><br>

			<input type="submit" value="Export">
		</form>
	</body>
	</html>
	`)
	if err != nil {
		http.Error(w, "Error rendering HTML", http.StatusInternalServerError)
		return
	}

	// Render the HTML page
	page := Page{}
	err = tmpl.Execute(w, page)
	if err != nil {
		http.Error(w, "Error rendering HTML", http.StatusInternalServerError)
		return
	}
}

func exportHandler(w http.ResponseWriter, r *http.Request) {
	// Get the form data
	apiKey := r.FormValue("api_key")
	contentType := r.FormValue("content_type")
	notionID := r.FormValue("notion_id")

	client := notionapi.NewClient(notionapi.Token(apiKey))
	if contentType == "database" {
		db, err := client.Database.Get(context.Background(), notionapi.DatabaseID(notionID))
		if err != nil {
			http.Error(w, "Error exporting database", http.StatusInternalServerError)
			return
		}
		// return a markdown file with the contents of db
		fmt.Println(db)

	} else if contentType == "page" {
		page, err := client.Page.Get(context.Background(), notionapi.PageID(notionID))
		if err != nil {

		}
		page_url := strings.Split(page.URL, "/")
		idpart := strings.LastIndex(page_url[len(page_url)-1], "-")
		title := page_url[len(page_url)-1][:idpart]

		pagination := notionapi.Pagination{
			PageSize: 10,
		}
		children, err := client.Block.GetChildren(context.Background(), notionapi.BlockID(notionID), &pagination)
		text := []string{}
		for _, child := range children.Results {
			text = append(text, child.GetRichText())
		}
		buf := new(strings.Builder)
		_, err = buf.WriteString(strings.Join(text, "\n"))
		w.Header().Set("Content-Disposition", "attachment; filename="+title+".md")
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Length", strconv.Itoa(len(buf.String())))
		file := strings.NewReader(buf.String())

		_, err = io.Copy(w, file)
		if err != nil {
			http.Error(w, "Error copying file content", http.StatusInternalServerError)
			return
		}
	}
}
