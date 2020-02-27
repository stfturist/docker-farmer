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

	// Assets route.
	http.Handle("/assets/", ipAllowedMiddleware(c, http.StripPrefix("/assets/", http.FileServer(http.Dir(realpath(*publicFlag)+"/assets")))))

	// Config api route.
	http.Handle("/api/config", ipAllowedMiddleware(c, http.HandlerFunc(handlers.ConfigHandler)))

	// Containers api route.
	http.Handle("/api/containers", ipAllowedMiddleware(c, http.HandlerFunc(handlers.ContainersHandler)))

	// Database api route.
	http.Handle("/api/database", ipAllowedMiddleware(c, http.HandlerFunc(handlers.DatabaseHandler)))

	// BitBucket service route.
	http.Handle("/services/bitbucket", ipAllowedMiddleware(c, http.HandlerFunc(handlers.BitbucketHandler)))

	// GitHub service route.
	http.Handle("/services/github", ipAllowedMiddleware(c, http.HandlerFunc(handlers.GithubHandler)))

	// GitLab service route.
	http.Handle("/services/gitlab", ipAllowedMiddleware(c, http.HandlerFunc(handlers.GitlabHandler)))

	// Jira service route.
	http.Handle("/services/jira", ipAllowedMiddleware(c, http.HandlerFunc(handlers.JiraHandler)))

	// Test service route.
	http.Handle("/services/test", ipAllowedMiddleware(c, http.HandlerFunc(handlers.TestHandler)))

	fmt.Printf("Listening to http://%s\n", c.Listen)

	// Listen to port.
	log.Fatal(http.ListenAndServe(c.Listen, nil))
}
