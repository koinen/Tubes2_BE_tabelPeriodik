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
	Children  []*RecipeNode
}

type RecipeNode struct {
	Result      string
	Ingredient1 *ElementNode
	Ingredient2 *ElementNode
}

type ExportableElement struct {
	Name         string           `json:"name"`
	Attributes	 string		      `json:"attributes"`
	Children     []ExportableRecipe   `json:"children"`
}

type ExportableRecipe struct {
	Attributes   string `json:"attributes"`
	Children     []ExportableElement `json:"children"`
}

var numberVisit int32
var visitMu sync.Mutex

func DFS(
	current *ElementNode,
	wg *sync.WaitGroup,
	elements map[string]*ElementNode,
	Children[]RecipeNode,
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

	for _, recipe := range Children{
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
			go DFS(base1, wg, elements, Children)
		}
		visitMu.Unlock()

		// Launch DFS on ingredient2 if not visited
		visitMu.Lock()
		if !base2.IsVisited {
			wg.Add(1)
			go DFS(base2, wg, elements, Children)
		}
		visitMu.Unlock()

		// Append recipe to current safely
		visitMu.Lock()
		current.Children= append(current.Children, &RecipeNode{
			Result:      current.Name,
			Ingredient1: base1,
			Ingredient2: base2,
		})
		visitMu.Unlock()
	}
}

// Converts internal ElementNode to exportable form
func ToExportableElement(node *ElementNode, res *ExportableElement) {
	if node == nil {
		return;
	}

	if !node.IsVisited {
		return;
	}

	if node.Tier == 0 {
		res.Name = node.Name
		res.Attributes = "element"
		return;
	}

	res.Name = node.Name
	res.Attributes = "element"
	res.Children = make([]ExportableRecipe, len(node.Children))
	for i := range node.Children {
		ToExportableRecipe(node.Children[i], &res.Children[i])
	}
}

func ToExportableRecipe(node *RecipeNode, res *ExportableRecipe) {
	if node == nil {
		return;
	}
	res.Attributes = "recipe"
	res.Children = make([]ExportableElement, 2)
	ToExportableElement(node.Ingredient1, &res.Children[0])
	ToExportableElement(node.Ingredient2, &res.Children[1])
}