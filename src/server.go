package main

import (
	"net/http"
)

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func addRouteWithCORS(path string, handlerFunc http.HandlerFunc) {
	http.Handle(path, withCORS(handlerFunc))
}

func serve() {
	addRouteWithCORS("/DFS/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		elmtName := r.URL.Path[len("/DFS/"):]
		if elmtName == "" {
			http.Error(w, "Element name is required", http.StatusBadRequest)
			return
		}
	})

	addRouteWithCORS("/BFS/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		elmtName := r.URL.Path[len("/BFS/"):]
		if elmtName == "" {
			http.Error(w, "Element name is required", http.StatusBadRequest)
			return
		}
	})

	addRouteWithCORS("/Bidirect/", func(w http.ResponseWriter, r *http.Request) {
	})

	addRouteWithCORS("/example-tree-data", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[
{
  "name": "Dust",
  "attributes": "element",
  "children": [
    {
      "attributes": "recipe",
      "children": [
        {
          "name": "Earth",
          "attributes": "element",
          "children": null
        },
        {
          "name": "Air",
          "attributes": "element",
          "children": null
        }
      ]
    },
    {
      "attributes": "recipe",
      "children": [
        {
          "name": "Land",
          "attributes": "element",
          "children": [
            {
              "attributes": "recipe",
              "children": [
                {
                  "name": "Earth",
                  "attributes": "element",
                  "children": null
                },
                {
                  "name": "Earth",
                  "attributes": "element",
                  "children": null
                }
              ]
            }
          ]
        },
        {
          "name": "Air",
          "attributes": "element",
          "children": null
        }
      ]
    }
  ]
}
		]`))
	})
	http.ListenAndServe(":8080", nil)
}
