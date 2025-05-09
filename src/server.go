package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
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

func serve(jsonBytes []byte) {
	var rawElements []Element
	err := json.Unmarshal(jsonBytes, &rawElements)
	if err != nil {
		panic(err)
	}

	addRouteWithCORS("/live-DFS/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		fmt.Println("Starting live DFS stream...")

		elmtName := r.URL.Path[len("/live-DFS/"):]
		if elmtName == "" {
			http.Error(w, "Element name is required", http.StatusBadRequest)
			return
		}
		fmt.Println("Starting DFS for element:", elmtName)
		elementMap := make(map[string]*ElementNode)
		for _, el := range rawElements {
			elementMap[el.Name] = &ElementNode{
				Name:     el.Name,
				Tier:     el.Tier,
				Children: []*RecipeNode{},
			}
		}
		for _, el := range rawElements {
			for _, r := range el.Recipes {
				ing1 := elementMap[r[0]]
				ing2 := elementMap[r[1]]
				if ing1 == nil || ing2 == nil {
					continue
				}
				recipe := RecipeNode{
					Result:      el.Name,
					Ingredient1: ing1,
					Ingredient2: ing2,
				}
				elementMap[el.Name].Children = append(elementMap[el.Name].Children, &recipe)
			}
		}
		root := elementMap[elmtName]
		wg := &sync.WaitGroup{}
		depthChan := make(chan int)
		wg.Add(1)
		go func() {
			DFS_Multiple(root, wg, elementMap, depthChan)
		}()

		go func() {
			wg.Wait()
			close(depthChan)
		}()

		for _ = range depthChan {
			exportList := ExportableElement{
				Name:       root.Name,
				Attributes: "element",
				Children:   make([]ExportableRecipe, 0, len(root.Children)),
			}
			visitedExport := make(map[*ElementNode]bool)
			ToExportableElement(root, &exportList, visitedExport)
			// Write to file
			payload := map[string]any{
				"depth": exportList, // or "data": exportList
			}
			wrapped, err := json.Marshal(payload)
			if err != nil {
				panic(err)
			}
			fmt.Fprintf(w, "data: %s\n\n", wrapped)
			w.(http.Flusher).Flush()
			time.Sleep(200 * time.Millisecond)
		}
	})

	addRouteWithCORS("/DFS/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		elmtName := r.URL.Path[len("/DFS/"):]
		query := r.URL.Query().Get("recipes")
		live := r.URL.Query().Get("live")
		recipeAmount := r.URL.Query().Get("recipeAmount")
		fmt.Println("Query parameter:", query)
		fmt.Println("Live parameter:", live)
		fmt.Println("Recipe amount parameter:", recipeAmount)
		if elmtName == "" {
			http.Error(w, "Element name is required", http.StatusBadRequest)
			return
		}
		fmt.Println("Starting DFS for element:", elmtName)
		elementMap := make(map[string]*ElementNode)

		for _, el := range rawElements {
			elementMap[el.Name] = &ElementNode{
				Name:     el.Name,
				Tier:     el.Tier,
				Children: []*RecipeNode{},
			}
		}

		for _, el := range rawElements {
			for _, r := range el.Recipes {
				ing1 := elementMap[r[0]]
				ing2 := elementMap[r[1]]
				if ing1 == nil || ing2 == nil {
					// fmt.Printf("Skipping invalid recipe for %s: missing ingredient(s) %s or %s\n", el.Name, r[0], r[1])
					continue
				}
				recipe := RecipeNode{
					Result:      el.Name,
					Ingredient1: ing1,
					Ingredient2: ing2,
				}
				elementMap[el.Name].Children = append(elementMap[el.Name].Children, &recipe)
			}
		}

		// Build tree from user input
		// root := &ElementNode{Name: elmtName, Tier: elementTier, Children: []*RecipeNode{}}
		root := elementMap[elmtName]
		wg := &sync.WaitGroup{}

		wg.Add(1)
		DFS_Multiple(root, wg, elementMap, nil)
		wg.Wait()

		// Export tree
		exportList := ExportableElement{
			Name:       root.Name,
			Attributes: "element",
			Children:   make([]ExportableRecipe, 0, len(root.Children)),
		}
		visitedExport := make(map[*ElementNode]bool)
		ToExportableElement(root, &exportList, visitedExport)

		// Write to file
		jsonOut, err := json.Marshal(exportList)
		if err != nil {
			panic(err)
		}
		fmt.Println("Exporting to JSON...")
		w.Write(jsonOut)
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

	addRouteWithCORS(("/example-stream"), func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		fmt.Println("Starting stream...")
		for i := 0; i < 10; i++ {
			fmt.Fprintf(w, "data: %s\n\n", fmt.Sprintf("{\"Event\": %d}", i))
			fmt.Println("Sending event:", i)
			time.Sleep(2 * time.Second)
			w.(http.Flusher).Flush()
		}

		<-r.Context().Done()
	})

	addRouteWithCORS("/example-tree-data", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[
{
  "name": "Sword",
  "attributes": "element",
  "children": [
    {
      "attributes": "recipe",
      "children": [
        {
          "name": "Blade",
          "attributes": "element",
          "children": [
            {
              "attributes": "recipe",
              "children": [
                {
                  "name": "Stone",
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
                          "name": "Pressure",
                          "attributes": "element",
                          "children": [
                            {
                              "attributes": "recipe",
                              "children": [
                                {
                                  "name": "Air",
                                  "attributes": "element",
                                  "children": null
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
                      ]
                    }
                  ]
                },
                {
                  "name": "Metal",
                  "attributes": "element",
                  "children": [
                    {
                      "attributes": "recipe",
                      "children": [
                        {
                          "name": "Stone",
                          "attributes": "element",
                          "children": null
                        },
                        {
                          "name": "Fire",
                          "attributes": "element",
                          "children": null
                        }
                      ]
                    },
                    {
                      "attributes": "recipe",
                      "children": [
                        {
                          "name": "Stone",
                          "attributes": "element",
                          "children": null
                        },
                        {
                          "name": "Heat",
                          "attributes": "element",
                          "children": [
                            {
                              "attributes": "recipe",
                              "children": [
                                {
                                  "name": "Air",
                                  "attributes": "element",
                                  "children": null
                                },
                                {
                                  "name": "Energy",
                                  "attributes": "element",
                                  "children": [
                                    {
                                      "attributes": "recipe",
                                      "children": [
                                        {
                                          "name": "Fire",
                                          "attributes": "element",
                                          "children": null
                                        },
                                        {
                                          "name": "Fire",
                                          "attributes": "element",
                                          "children": null
                                        }
                                      ]
                                    }
                                  ]
                                }
                              ]
                            }
                          ]
                        }
                      ]
                    }
                  ]
                }
              ]
            }
          ]
        },
        {
          "name": "Metal",
          "attributes": "element",
          "children": null
        }
      ]
    }
  ]
}
		]`))
	})

	fmt.Println("Now serving in port 8080...")
	// Wrap the default mux with CORS so all routes (including 404) get CORS headers
	http.ListenAndServe(":8080", withCORS(http.DefaultServeMux))
}
