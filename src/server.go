package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
	"sync/atomic"
  "strconv"
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

		recipeAmount := r.URL.Query().Get("recipeAmount")
		val, err := strconv.Atoi(recipeAmount)
		if err != nil {
			http.Error(w, "Invalid recipe amount", http.StatusBadRequest)
			return
		}
		atomic.StoreInt32(&recipeLeft, int32(val - 1))

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
			time.Sleep(500 * time.Millisecond)
		}
		finalExport := ExportableElement{
			Name:       root.Name,
			Attributes: "element",
			Children:   make([]ExportableRecipe, 0, len(root.Children)),
		}
		visitedExport := make(map[*ElementNode]bool)
		ToExportableElement(root, &finalExport, visitedExport)

		finalPayload := map[string]any{
			"depth": finalExport,
		}
		finalWrapped, err := json.Marshal(finalPayload)
		if err != nil {
			panic(err)
		}
		fmt.Fprintf(w, "data: %s\n\n", finalWrapped)
		w.(http.Flusher).Flush()
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
		val, err := strconv.Atoi(recipeAmount)
		if err != nil {
		http.Error(w, "Invalid recipe amount", http.StatusBadRequest)
		return
		}
		if elmtName == "" {
			http.Error(w, "Element name is required", http.StatusBadRequest)
			return
		}

		atomic.StoreInt32(&recipeLeft, int32(val - 1))
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
		root, exists := elementMap[elmtName]
		if !exists {
			http.Error(w, "Element not found", http.StatusNotFound)
			return
		}
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
		query := r.URL.Query().Get("recipes")
		live := r.URL.Query().Get("live")
		recipeAmount := r.URL.Query().Get("recipeAmount")
		fmt.Println("Query parameter:", query)
		fmt.Println("Live parameter:", live)
		fmt.Println("Recipe amount parameter:", recipeAmount)
		// val, err := strconv.Atoi(recipeAmount)
		if err != nil {
			http.Error(w, "Invalid recipe amount", http.StatusBadRequest)
			return
		}
		if elmtName == "" {
			http.Error(w, "Element name is required", http.StatusBadRequest)
			return
		}
		fmt.Println("Starting FS for element:", elmtName)
		elementMap := make(map[string]*ElementNode)
		var allRecipes []RecipeNode

		for _, el := range rawElements {
			elementMap[el.Name] = &ElementNode{
				Name:     el.Name,
				Tier:     el.Tier,
				Children: []*RecipeNode{},
			}
			for _, r := range el.Recipes {
				if len(r) == 2 {
					allRecipes = append(allRecipes, RecipeNode{
						Result:      el.Name,
						Ingredient1: &ElementNode{Name: r[0], IsVisited: false, Children: []*RecipeNode{}},
						Ingredient2: &ElementNode{Name: r[1], IsVisited: false, Children: []*RecipeNode{}},
					})
				}
			}
		}


		// Build tree from user input
		root := &ElementNode{Name: elmtName, Tier: 1, Children: []*RecipeNode{}}
		root.Tier = elementMap[elmtName].Tier
		// if !exists {
		// 	http.Error(w, "Element not found", http.StatusNotFound)
		// 	return
		// }
		bfs(root, elementMap, allRecipes)

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
