package main

import (
	"github.com/emvi/pirsch"
	"net/http"
)

func main() {
	http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" { // don't track resources
			go pirsch.SaveHit(r)
		}

		w.Write([]byte("<h1>Hello World!</h1>"))
	}))
	http.ListenAndServe(":8080", nil)
}
