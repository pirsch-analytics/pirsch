package main

import (
	"database/sql"
	"github.com/emvi/pirsch"
	_ "github.com/lib/pq"
	"log"
	"net/http"
)

func main() {
	db, err := sql.Open("postgres", "host=localhost port=5432 user=postgres password=postgres dbname=pirsch sslmode=disable")

	if err != nil {
		log.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	pirsch.SetStore(pirsch.NewPostgresStore(db))
	http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" { // don't track resources
			go pirsch.SaveHit(r)
		}

		w.Write([]byte("<h1>Hello World!</h1>"))
	}))
	log.Println("starting server...")
	http.ListenAndServe(":8080", nil)
}
