package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

// stringInSlice returns true if a string exists or false if not.
func stringInSlice(a string, list []string) bool {
	a = strings.ToLower(a)

	for _, b := range list {
		b = strings.ToLower(b)
		if b == a {
			return true
		}

		if strings.Contains(a, "*") {
			a = strings.Replace(a, "*", "", -1)

			if strings.Contains(b, a) {
				return true
			}
		}
	}

	return false
}

// write writes to both log and http response writer.
func write(w http.ResponseWriter, text string) {
	log.Println(text)
	fmt.Fprintf(w, text)
}
