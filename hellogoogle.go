package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"time"
)

func main() {
	http.HandleFunc("/search", handleSearch)
	fmt.Println("Serving on http://localhost:7777/search")
	log.Fatal(http.ListenAndServe("localhost:7777", nil))

	// http.HandleFunc("/search", handleSearch)
	// fmt.Println("Serving on http://localhost:7777/hello")
	// log.Fatal(http.ListenAndServe("localhost:7777", nil))
}

type Result struct {
	Title, URL string
}

var resultsTemplate = template.Must(template.New("results").Parse(`
<html>
<head/>
<body>
  <ol>
  {{range .Results}}
    <li>{{.Title}} - <a href="{{.URL}}">{{.URL}}</a></li>
  {{end}}
  </ol>
  <p>{{len .Results}} results in {{.Elapsed}}</p>
</body>
</html>
`))

func handleSearch(w http.ResponseWriter, req *http.Request) {
	log.Println("serving", req.URL)
	// fmt.Fprintln(w, "Hello, 谷歌！")

	query := req.FormValue("q")
	if query == "" {
		http.Error(w, `missing "q" URL parameter`, http.StatusBadRequest)
		return
	}

	start := time.Now()
	results, err := Search(query)
	elapsed := time.Since(start)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Render the results.
	type templateData struct {
		Results []Result
		Elapsed time.Duration
	}
	if err := resultsTemplate.Execute(w, templateData{
		Results: results,
		Elapsed: elapsed,
	}); err != nil {
		log.Print(err)
		return
	}
}

func Search(query string) ([]Result, error) {
	u, err := url.Parse("https://www.googleapis.com/customsearch/v1?key=AIzaSyCQW0sdJJ9_omSl3fI05ruBdydH8voDq6U&cx=004414503126133779177:5vwtylddpwu")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("q", query)
	u.RawQuery = q.Encode()

	// Issue the HTTP request and handle the response
	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var jsonResponse struct {
		ResponseData struct {
			Items []struct {
				Title, Link string
			}
		}
	}
	if err := json.NewDecoder(resp.Body).Decode(&jsonResponse); err != nil {
		return nil, err
	}

	// Extract the Results from jsonResponse and return them.
	var results []Result
	for _, r := range jsonResponse.ResponseData.Items {
		results = append(results, Result{Title: r.Title, URL: r.Link})
	}
	return results, nil
}
