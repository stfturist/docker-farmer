package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/kardianos/osext"
	"github.com/stfturist/docker-farmer/config"
	"github.com/stfturist/docker-farmer/handlers"
)

var (
	configFlag = flag.String("config", "config.json", "Path to config file")
	publicFlag = flag.String("public", "public", "Path to public directory")
)

// path will return right path for file, looks at the
// given file first and then looks in the executable folder.
func realpath(file string) string {
	if _, err := os.Stat(file); err == nil {
		return file
	}

	path, _ := osext.ExecutableFolder()

	if _, err := os.Stat(path + "/" + file); os.IsNotExist(err) {
		return file
	}

	return path + "/" + file
}

// ipAllowed will determine if the request ip is allowed
// or not.
func ipAllowedMiddleware(c *config.Config, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var ip string

		if o1 := r.Header.Get("X-Forwarded-For"); o1 != "" {
			ip = o1
		} else if o2 := r.Header.Get("x-forwarded-for"); o2 != "" {
			ip = o2
		} else if o3 := r.Header.Get("X-FORWARDED-FOR"); o3 != "" {
			ip = o3
		}

		if len(c.AllowedIPs) == 0 {
			next.ServeHTTP(w, r)
			return
		}

		for _, i := range c.AllowedIPs {
			if ip == i {
				next.ServeHTTP(w, r)
				return
			}
		}

		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(http.StatusText(http.StatusUnauthorized)))
	})
}

func main() {
	flag.Parse()

	// Init config.
	config.Init(realpath(*configFlag))
	c := config.Get()

	if c.Listen[0] == ':' {
		c.Listen = "0.0.0.0" + c.Listen
	}

	// Index route.
	http.Handle("/", ipAllowedMiddleware(c, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		links, err := json.Marshal(c.Links)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		var styles string
		for key := range c.Style {
			styles += fmt.Sprintf("%s { %s }\n", key, c.Style[key])
		}

		templates := template.Must(template.ParseFiles(realpath(*publicFlag) + "/index.html"))
		err = templates.ExecuteTemplate(w, "index.html", map[string]interface{}{
			"Config": c,
			"Links":  template.JS(string(links)),
			"Styles": template.CSS(styles),
		})

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})))

	routes := map[string]http.Handler{
		"/assets/":            http.StripPrefix("/assets/", http.FileServer(http.Dir(realpath(*publicFlag)+"/assets"))),
		"/api/config":         http.HandlerFunc(handlers.ConfigHandler),
		"/api/containers":     http.HandlerFunc(handlers.ContainersHandler),
		"/api/database":       http.HandlerFunc(handlers.DatabaseHandler),
		"/services/bitbucket": http.HandlerFunc(handlers.BitbucketHandler),
		"/services/github":    http.HandlerFunc(handlers.GithubHandler),
		"/services/gitlab":    http.HandlerFunc(handlers.GitlabHandler),
		"/services/jira":      http.HandlerFunc(handlers.JiraHandler),
		"/services/test":      http.HandlerFunc(handlers.TestHandler),
	}

	for path, handler := range routes {
		http.Handle(path, ipAllowedMiddleware(c, handler))
	}

	fmt.Printf("Listening to http://%s\n", c.Listen)

	// Listen to port.
	log.Fatal(http.ListenAndServe(c.Listen, nil))
}
