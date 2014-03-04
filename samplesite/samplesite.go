package samplesite

import (
	"html/template"
	"net/http"
	"net/http/httptest"
)

type page struct {
	Title string
	Links []string
}

var URL string

var robotsBody = []byte(`# robots.txt
User-Agent: Googlebot
Disallow: /

User-Agent: GoSpiderBot
Disallow: /nospider

User-Agent: *
Disallow: /contact
`)

var baseTmpl = `<!DOCTYPE html>
<html>
<head>
	<title>{{ .Title }}</title>
</head>
<body>
	<header>
		<h1>{{ .Title }}</h1>
		<nav>
			<ul>
				<li><a href="/">Home</a></li>
				<li><a href="/latest">Latest News</a></li>
				<li><a href="/contact">Contact Us</a></li>
			</ul>
		</nav>
	</header>

	<article>{{ range .Links }}
		<p><a href="{{ . }}">{{ . }}</a></p>
	{{ end }}</article>

	<footer>
		<nav>
			<ul>
				<li><a href="/">Home</a></li>
				<li><a href="/latest">Latest News</a></li>
				<li><a href="/contact">Contact Us</a></li>
			</ul>
		</nav>
	</footer>
</body>
</html>`

func init() {
	base, err := template.New("base").Parse(baseTmpl)
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) { w.Write(robotsBody) })

	pages := map[string]page{
		"/":         {"Index", []string{"/article1", "/nospider", "/article2", "/article3"}},
		"/latest":   {"Index", []string{"/article1", "/article2"}},
		"/article1": {"Article 1", []string{"/article2", "/article3"}},
		"/article2": {"Article 1", []string{"/article1", "/article3", "/nospider"}},
		"/article3": {"Article 1", []string{"/article1", "/article2"}},
		"/nospider": {"Don't Spider Me!", []string{}},
		"/contact":  {"Contact Us", []string{}},
	}

	for path, content := range pages {
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			if err := base.Execute(w, content); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		})
	}

	server := httptest.NewServer(mux)
	URL = server.URL
}
