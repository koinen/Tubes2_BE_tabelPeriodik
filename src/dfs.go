package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

var numberVisit int32

type ElementNode struct {
	Name    string
	Tier    int
	Recipes []*RecipeNode
}

type RecipeNode struct {
	Result      string
	Ingredient1 *ElementNode
	Ingredient2 *ElementNode
}

type ExportableElement struct {
	Name    string           `json:"name"`
	Tier    int              `json:"tier"`
	Recipes []ExportableRecipe `json:"recipes"`
}

type ExportableRecipe struct {
	Ingredient1 string `json:"ingredient1"`
	Ingredient2 string `json:"ingredient2"`
}


func DFS(
	current *ElementNode,
	wg *sync.WaitGroup,
	elements map[string]*ElementNode,
	recipes []RecipeNode,
	treeMu *sync.Mutex,
	cache map[string]*ElementNode,
	cacheMu *sync.Mutex,
) {
	defer wg.Done()

	count := atomic.AddInt32(&numberVisit, 1)
	fmt.Printf("Visiting node (%d): %s Tier: %d\n", count, current.Name, current.Tier)

	for _, recipe := range recipes {
		if recipe.Result != current.Name {
			continue
		}

		// Get or create Ingredient1
		cacheMu.Lock()
		ing1, found1 := cache[recipe.Ingredient1.Name]
		cacheMu.Unlock()

		if !found1 {
			base1, exists1 := elements[recipe.Ingredient1.Name]
			if !exists1 {
				fmt.Printf("Missing element: %s\n", recipe.Ingredient1.Name)
				continue
			}
			ing1 = &ElementNode{
				Name:    base1.Name,
				Tier:    base1.Tier,
				Recipes: []*RecipeNode{},
			}
			cacheMu.Lock()
			cache[ing1.Name] = ing1
			cacheMu.Unlock()
			wg.Add(1)
			go DFS(ing1, wg, elements, recipes, treeMu, cache, cacheMu)
		}

		// Get or create Ingredient2
		cacheMu.Lock()
		ing2, found2 := cache[recipe.Ingredient2.Name]
		cacheMu.Unlock()

		if !found2 {
			base2, exists2 := elements[recipe.Ingredient2.Name]
			if !exists2 {
				fmt.Printf("Missing element: %s\n", recipe.Ingredient2.Name)
				continue
			}
			ing2 = &ElementNode{
				Name:    base2.Name,
				Tier:    base2.Tier,
				Recipes: []*RecipeNode{},
			}
			cacheMu.Lock()
			cache[ing2.Name] = ing2
			cacheMu.Unlock()
			wg.Add(1)
			go DFS(ing2, wg, elements, recipes, treeMu, cache, cacheMu)
		}

		treeMu.Lock()
		current.Recipes = append(current.Recipes, &RecipeNode{
			Result:      current.Name,
			Ingredient1: ing1,
			Ingredient2: ing2,
		})
		treeMu.Unlock()
	}
}

// Converts internal ElementNode to exportable form
func ToExportableElement(node *ElementNode) ExportableElement {
	exported := ExportableElement{
		Name:    node.Name,
		Tier:    node.Tier,
		Recipes: make([]ExportableRecipe, 0, len(node.Recipes)),
	}
	for _, r := range node.Recipes {
		exported.Recipes = append(exported.Recipes, ExportableRecipe{
			Ingredient1: r.Ingredient1.Name,
			Ingredient2: r.Ingredient2.Name,
		})
	}
	return exported
}
