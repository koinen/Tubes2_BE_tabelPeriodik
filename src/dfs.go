package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

type ElementNode struct {
	IsVisited bool
	Name      string
	Tier      int
	Recipes   []*RecipeNode
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

var numberVisit int32
var visitMu sync.Mutex

func DFS(
	current *ElementNode,
	wg *sync.WaitGroup,
	elements map[string]*ElementNode,
	recipes []RecipeNode,
) {
	defer wg.Done()

	// Locking visit check
	visitMu.Lock()
	if current.IsVisited {
		visitMu.Unlock()
		return
	}
	current.IsVisited = true
	visitMu.Unlock()

	count := atomic.AddInt32(&numberVisit, 1)
	fmt.Printf("Visiting node (%d): %s Tier: %d\n", count, current.Name, current.Tier)

	for _, recipe := range recipes {
		if recipe.Result != current.Name {
			continue
		}

		// Get pointers to real ingredients
		base1, ok1 := elements[recipe.Ingredient1.Name]
		base2, ok2 := elements[recipe.Ingredient2.Name]
		if !ok1 || !ok2 {
			fmt.Printf("Missing ingredients for %s: %s or %s\n", current.Name, recipe.Ingredient1.Name, recipe.Ingredient2.Name)
			continue
		}

		// Launch DFS on ingredient1 if not visited
		visitMu.Lock()
		if !base1.IsVisited {
			wg.Add(1)
			go DFS(base1, wg, elements, recipes)
		}
		visitMu.Unlock()

		// Launch DFS on ingredient2 if not visited
		visitMu.Lock()
		if !base2.IsVisited {
			wg.Add(1)
			go DFS(base2, wg, elements, recipes)
		}
		visitMu.Unlock()

		// Append recipe to current safely
		visitMu.Lock()
		current.Recipes = append(current.Recipes, &RecipeNode{
			Result:      current.Name,
			Ingredient1: base1,
			Ingredient2: base2,
		})
		visitMu.Unlock()
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