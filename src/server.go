package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
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

		recipeAmount := r.URL.Query().Get("recipeAmount")
		delay := r.URL.Query().Get("delay")
		val2, err := strconv.Atoi(delay)
		if err != nil {
			http.Error(w, "Invalid delay value", http.StatusBadRequest)
			return
		}
		fmt.Println("Delay value:", val2)
		val, err := strconv.Atoi(recipeAmount)
		if err != nil {
			http.Error(w, "Invalid recipe amount", http.StatusBadRequest)
			return
		}
		atomic.StoreInt32(&recipeLeft, int32(val-1))
		sem = make(chan struct{}, val-1)
		fmt.Println("Recipe left set to:", recipeLeft)

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
				ImgSrc:   el.ImgSrc,
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
		barrier := &sync.WaitGroup{}

		barrier.Add(1)
		wg.Add(1)
		go func() {
			DFS_Multiple(root, wg, elementMap, depthChan, barrier)
		}()

		go func() {
			wg.Wait()
			close(depthChan)
		}()

		for {
			if _, ok := <-depthChan; !ok {
				break
			}
			barrier.Add(1)
			time.Sleep(time.Duration(val2) * time.Millisecond) // delay
			barrier.Done()
			barrier.Wait()

			exportList := ExportableElement{
				Name:       root.Name,
				Attributes: map[string]string{"Type": "element", "Side": "Right"},
				Children:   make([]ExportableRecipe, 0, len(root.Children)),
			}
			visitedExport := make(map[*ElementNode]*ExportableElement)
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
			time.Sleep(time.Duration(val2) * time.Millisecond)
		}
		finalExport := ExportableElement{
			Name:       root.Name,
			Attributes: map[string]string{"Type": "element", "Side": "Right"},
			Children:   make([]ExportableRecipe, 0, len(root.Children)),
		}
		visitedExport := make(map[*ElementNode]*ExportableElement)
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

		atomic.StoreInt32(&recipeLeft, int32(val-1))
		sem = make(chan struct{}, val-1)
		fmt.Println("Recipe left set to:", recipeLeft)
		fmt.Println("Starting DFS for element:", elmtName)
		elementMap := make(map[string]*ElementNode)

		for _, el := range rawElements {
			elementMap[el.Name] = &ElementNode{
				Name:     el.Name,
				ImgSrc:   el.ImgSrc,
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
		barrier := &sync.WaitGroup{}
		barrier.Add(1)

		DFS_Multiple(root, wg, elementMap, nil, barrier)
		wg.Wait()

		// Export tree
		exportList := ExportableElement{
			Name:       root.Name,
			Attributes: map[string]string{"Type": "element", "Side": "Right"},
			Children:   make([]ExportableRecipe, 0, len(root.Children)),
		}
		visitedExport := make(map[*ElementNode]*ExportableElement)
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
		val, err := strconv.Atoi(recipeAmount)
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
		var allRecipes []*RecipeNode

		for _, el := range rawElements {
			elementMap[el.Name] = &ElementNode{
				Name:     el.Name,
				ImgSrc:   el.ImgSrc,
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
				recipe := &RecipeNode{
					Result:      el.Name,
					Ingredient1: ing1,
					Ingredient2: ing2,
				}
				allRecipes = append(allRecipes, recipe)
				elementMap[el.Name].Children = append(elementMap[el.Name].Children, recipe)
			}
		}

		root := elementMap[elmtName]
		bfs(root, elementMap, allRecipes, val, nil)

		// Export tree
		exportList := ExportableElement{
			Name:       root.Name,
			Attributes: map[string]string{"Type": "element", "Side": "Right"},
			Children:   make([]ExportableRecipe, 0, len(root.Children)),
		}
		visitedExport := make(map[*ElementNode]*ExportableElement)
		ToExportableElement(root, &exportList, visitedExport)

		// Write to file
		jsonOut, err := json.Marshal(exportList)
		if err != nil {
			panic(err)
		}
		fmt.Println("Exporting to JSON...")
		w.Write(jsonOut)
	})

	addRouteWithCORS("/live-BFS/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("run")
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		recipeAmount := r.URL.Query().Get("recipeAmount")
		delay := r.URL.Query().Get("delay")
		val, err := strconv.Atoi(recipeAmount)
		if err != nil {
			http.Error(w, "Invalid recipe amount", http.StatusBadRequest)
			return
		}
		val2, err := strconv.Atoi(delay)
		if err != nil {
			http.Error(w, "Invalid delay value", http.StatusBadRequest)
			return
		}
		atomic.StoreInt32(&recipeLeft, int32(val-1))

		fmt.Println("Starting live BFS stream...")

		elmtName := r.URL.Path[len("/live-BFS/"):]
		if elmtName == "" {
			http.Error(w, "Element name is required", http.StatusBadRequest)
			return
		}

		elementMap := make(map[string]*ElementNode)
		var allRecipes []*RecipeNode

		for _, el := range rawElements {
			elementMap[el.Name] = &ElementNode{
				Name:     el.Name,
				ImgSrc:   el.ImgSrc,
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
				recipe := &RecipeNode{
					Result:      el.Name,
					Ingredient1: ing1,
					Ingredient2: ing2,
				}
				allRecipes = append(allRecipes, recipe)
				elementMap[el.Name].Children = append(elementMap[el.Name].Children, recipe)
			}
		}

		// root := &ElementNode{Name: elmtName, Tier: 1, Children: []*RecipeNode{}}
		root := elementMap[elmtName]
		ch := make(chan int)
		go bfs(root, elementMap, allRecipes, val, ch)

		for _ = range ch {
			exportList := ExportableElement{
				Name:       root.Name,
				Attributes: map[string]string{"Type": "element", "Side": "Right"},
				Children:   make([]ExportableRecipe, 0, len(root.Children)),
			}
			visitedExport := make(map[*ElementNode]*ExportableElement)
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
			time.Sleep(time.Duration(val2) * time.Millisecond)
		}
		// visited := make(map[*ElementNode]*ExportableElement)
		// finalExport := ToExportableElement3(root, visited)
		finalExport := ExportableElement{
			Name:       root.Name,
			Attributes: map[string]string{"Type": "element", "Side": "Right"},
			Children:   make([]ExportableRecipe, 0, len(root.Children)),
		}
		visitedExport := make(map[*ElementNode]*ExportableElement)
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

	addRouteWithCORS("/Bidirectional/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		elmtName := r.URL.Path[len("/Bidirectional/"):]
		query := r.URL.Query().Get("recipes")
		live := r.URL.Query().Get("live")
		recipeAmount := r.URL.Query().Get("recipeAmount")
		left := r.URL.Query().Get("left")
		right := r.URL.Query().Get("right")
		fmt.Println("Right parameter:", right)
		fmt.Println("Left parameter:", left)

		fmt.Println("Query parameter:", query)
		fmt.Println("Live parameter:", live)
		val, err := strconv.Atoi(recipeAmount)
		if err != nil {
			http.Error(w, "Invalid recipe amount", http.StatusBadRequest)
			return
		}
		atomic.StoreInt32(&recipeLeft, int32(val-1))
		sem = make(chan struct{}, val-1)
		fmt.Println("Recipe left set to:", recipeLeft)
		if elmtName == "" {
			http.Error(w, "Element name is required", http.StatusBadRequest)
			return
		}
		fmt.Println("Starting Bidirect for element:", elmtName)
		elementMap := make(map[string]*ElementNode)
		var allRecipes []*RecipeNode

		for _, el := range rawElements {
			elementMap[el.Name] = &ElementNode{
				Name:     el.Name,
				ImgSrc:   el.ImgSrc,
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
				recipe := &RecipeNode{
					Result:      el.Name,
					Ingredient1: ing1,
					Ingredient2: ing2,
				}
				allRecipes = append(allRecipes, recipe)
				elementMap[el.Name].Children = append(elementMap[el.Name].Children, recipe)
			}
		}

		// root := &ElementNode{Name: elmtName, Tier: 1, Children: []*RecipeNode{}}
		root := elementMap[elmtName]
		wg := &sync.WaitGroup{}
		basic := []*ElementNode{}
		basic = append(basic, elementMap["Earth"])
		basic = append(basic, elementMap["Fire"])
		basic = append(basic, elementMap["Air"])
		basic = append(basic, elementMap["Water"])
		done := make(chan struct{})

		if right == "DFS" {
			Bidirect_Right_DFS(root, wg, elementMap, nil, done)
		} else {
			Bidirect_Right_BFS(root, val, wg, elementMap, allRecipes, nil, done)
		}
		copyAllRecipes := make([]*RecipeNode, len(allRecipes))
		copy(copyAllRecipes, allRecipes)
		if left == "BFS" {
			wg.Add(1)
			go func() {
				defer wg.Done()
				Bidirect_Left_BFS(basic, root, elementMap, copyAllRecipes, done)
				fmt.Println("[BFS Left] Done")
			}()
		} else {
			wg.Add(1)
			go func() {
				defer wg.Done()
				Bidirect_Left_DFS(basic, root, elementMap, copyAllRecipes, done)
				fmt.Println("[DFS Left] Done")
			}()
		}
		wg.Wait()
		fmt.Println("[MAIN] completed")

		exportList := ExportableElement{
			Name:       root.Name,
			Attributes: map[string]string{"Type": "element", "Side": "Right"},
			Children:   make([]ExportableRecipe, 0, len(root.Children)),
		}
		exportList.Children = make([]ExportableRecipe, 0, len(root.Children))

		visitedExport := make(map[*ElementNode]*ExportableElement)
		ToExportableElement(root, &exportList, visitedExport)

		jsonOut, err := json.Marshal(exportList)
		if err != nil {
			panic(err)
		}
		fmt.Println("Exporting to JSON...")
		w.Write(jsonOut)
	})

	addRouteWithCORS("/image", func(w http.ResponseWriter, r *http.Request) {
		// Shuffle rawElements
		rand.Seed(time.Now().UnixNano())
		shuffled := make([]Element, len(rawElements))
		copy(shuffled, rawElements)
		rand.Shuffle(len(shuffled), func(i, j int) {
			shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
		})

		// Pick up to 50 elements with non-empty image URLs
		imageMap := make(map[string]string)
		count := 0
		for _, el := range shuffled {
			if el.ImgSrc != "" {
				imageMap[el.Name] = el.ImgSrc
				count++
			}
			if count >= 50 {
				break
			}
		}

		// Write JSON response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(imageMap); err != nil {
			http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
		}
	})

	// addRouteWithCORS("/live-Bidirectional/", func(w http.ResponseWriter, r *http.Request) {
	// 	w.Header().Set("Content-Type", "text/event-stream")
	// 	w.Header().Set("Cache-Control", "no-cache")
	// 	w.Header().Set("Connection", "keep-alive")

	// 	recipeAmount := r.URL.Query().Get("recipeAmount")
	// 	left := r.URL.Query().Get("left")
	// 	right := r.URL.Query().Get("right")
	// 	delay := r.URL.Query().Get("delay")
	// 	val2, err := strconv.Atoi(delay)
	// 	if err != nil {
	// 		http.Error(w, "Invalid delay value", http.StatusBadRequest)
	// 		return
	// 	}
	// 	fmt.Println("Delay value:", val2)
	// 	val, err := strconv.Atoi(recipeAmount)
	// 	if err != nil {
	// 		http.Error(w, "Invalid recipe amount", http.StatusBadRequest)
	// 		return
	// 	}
	// 	atomic.StoreInt32(&recipeLeft, int32(val - 1))
	// 	sem = make(chan struct{}, val-1)
	// 	fmt.Println("Recipe left set to:", recipeLeft)

	// 	elmtName := r.URL.Path[len("/live-Bidirectional/"):]
	// 	if elmtName == "" {
	// 		http.Error(w, "Element name is required", http.StatusBadRequest)
	// 		return
	// 	}
	// 	fmt.Println("Starting Bidirect Live for element:", elmtName)
	// 	elementMap := make(map[string]*ElementNode)
	// 	var allRecipes []*RecipeNode

	// 	for _, el := range rawElements {
	// 		elementMap[el.Name] = &ElementNode{
	// 			Name:     el.Name,
	// 			Tier:     el.Tier,
	// 			Children: []*RecipeNode{},
	// 		}
	// 	}

	// 	for _, el := range rawElements {
	// 		for _, r := range el.Recipes {
	// 			ing1 := elementMap[r[0]]
	// 			ing2 := elementMap[r[1]]
	// 			if ing1 == nil || ing2 == nil {
	// 				// fmt.Printf("Skipping invalid recipe for %s: missing ingredient(s) %s or %s\n", el.Name, r[0], r[1])
	// 				continue
	// 			}
	// 			recipe := &RecipeNode{
	// 				Result:      el.Name,
	// 				Ingredient1: ing1,
	// 				Ingredient2: ing2,
	// 			}
	// 			allRecipes = append(allRecipes, recipe)
	// 			elementMap[el.Name].Children = append(elementMap[el.Name].Children, recipe)
	// 		}
	// 	}

	// 	// root := &ElementNode{Name: elmtName, Tier: 1, Children: []*RecipeNode{}}
	// 	root := elementMap[elmtName]
	// 	wg := &sync.WaitGroup{}
	// 	basic := []*ElementNode{}
	// 	basic = append(basic, elementMap["Earth"])
	// 	basic = append(basic, elementMap["Fire"])
	// 	basic = append(basic, elementMap["Air"])
	// 	basic = append(basic, elementMap["Water"])
	// 	done := make(chan struct{})
	// 	depthChan := make(chan int)

	// 	copyAllRecipes := make([]*RecipeNode, len(allRecipes))
	// 	copy(copyAllRecipes, allRecipes)
	// 	// depthChanLeft := make(chan int)
	// 	if left == "BFS" {
	// 		wg.Add(1)
	// 		go func() {
	// 			defer wg.Done()
	// 			Bidirect_Left_BFS(basic, root, elementMap, copyAllRecipes, done)
	// 			fmt.Println("[BFS Left] Done")
	// 		}()
	// 	} else {
	// 		wg.Add(1)
	// 		go func() {
	// 			defer wg.Done()
	// 			Bidirect_Left_DFS(basic, root, elementMap, copyAllRecipes, done)
	// 			fmt.Println("[DFS Left] Done")
	// 		}()
	// 	}
	// 	if right == "DFS" {
	// 		Bidirect_Right_DFS(root, wg, elementMap, depthChan, done)
	// 	} else {
	// 		Bidirect_Right_BFS(root, wg, elementMap, allRecipes, depthChan, done)
	// 	}
	// 	go func() {
	// 		wg.Wait()
	// 		close(depthChan)
	// 		// close(depthChanLeft)
	// 		fmt.Println("[MAIN] completed")
	// 	}()

	// 	for _ = range depthChan {
	// 		// exportList := ExportableElement{
	// 		// 	Name:       root.Name,
	// 		// 	Attributes: map[string]string{"Type": "element", "Side": "Right"},
	// 		// 	Children:   make([]ExportableRecipe, 0, len(root.Children)),
	// 		// }
	// 		// visitedExport := make(map[*ElementNode]*ExportableElement)
	// 		// ToExportableElement(root, &exportList, visitedExport)
	// 		visited := make(map[*ElementNode]*ExportableElement)
	// 		export := ToExportableElement3(root, visited)
	// 		// Write to file
	// 		payload := map[string]any{
	// 			"depth": export, // or "data": exportList
	// 		}
	// 		wrapped, err := json.Marshal(payload)
	// 		if err != nil {
	// 			panic(err)
	// 		}
	// 		fmt.Fprintf(w, "data: %s\n\n", wrapped)
	// 		w.(http.Flusher).Flush()
	// 		time.Sleep(time.Duration(val2) * time.Millisecond)
	// 	}
	// 	// finalExport := ExportableElement{
	// 	// 	Name:       root.Name,
	// 	// 	Attributes: map[string]string{"Type": "element", "Side": "Right"},
	// 	// 	Children:   make([]ExportableRecipe, 0, len(root.Children)),
	// 	// }
	// 	visited := make(map[*ElementNode]*ExportableElement)
	// 	finalExport := ToExportableElement3(root, visited)

	// 	finalPayload := map[string]any{
	// 		"depth": finalExport,
	// 	}
	// 	finalWrapped, err := json.Marshal(finalPayload)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	fmt.Fprintf(w, "data: %s\n\n", finalWrapped)
	// 	w.(http.Flusher).Flush()
	// })

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
