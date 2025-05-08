package main

import (
	"net/http"
)

func serve() {
	http.HandleFunc("/DFS/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		elmtName := r.URL.Path[len("/DFS/"):]
		if elmtName == "" {
			http.Error(w, "Element name is required", http.StatusBadRequest)
			return
		}
	})

	http.HandleFunc("/BFS/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		elmtName := r.URL.Path[len("/BFS/"):]
		if elmtName == "" {
			http.Error(w, "Element name is required", http.StatusBadRequest)
			return
		}
	})

	http.HandleFunc("/Bidirect/", func(w http.ResponseWriter, r *http.Request) {
	})

	http.ListenAndServe(":8080", nil)
}
