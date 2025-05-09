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

func DFS_Multiple(
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

		visitMu.Lock()
		if !base1.IsVisited {
			wg.Add(1)
			go DFS_Multiple(base1, wg, elements, Children)
		}
		visitMu.Unlock()

		visitMu.Lock()
		if !base2.IsVisited {
			wg.Add(1)
			go DFS_Multiple(base2, wg, elements, Children)
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

func DFS_Single(
	current *ElementNode,
	wg *sync.WaitGroup,
	elements map[string]*ElementNode,
	recipes []RecipeNode,
) {
	defer wg.Done()

	// Prevent re-visiting
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

		ing1, ok1 := elements[recipe.Ingredient1.Name]
		ing2, ok2 := elements[recipe.Ingredient2.Name]
		if !ok1 || !ok2 {
			continue
		}

		if ing1.Tier > current.Tier || ing2.Tier > current.Tier {
			fmt.Printf("Abandoning recipe for %s: ingredient tier too high (%s: %d, %s: %d > %d)\n",
				current.Name, ing1.Name, ing1.Tier, ing2.Name, ing2.Tier, current.Tier)
			continue
		}

		visitMu.Lock()
		if ing1.Tier == 0 {
			ing1.IsVisited = true
		}
		if ing2.Tier == 0 {
			ing2.IsVisited = true
		}
		visitMu.Unlock()

		// Add the recipe
		visitMu.Lock()
		current.Children = append(current.Children, &RecipeNode{
			Result:      current.Name,
			Ingredient1: ing1,
			Ingredient2: ing2,
		})
		visitMu.Unlock()

		// Stop if both ingredients are tier 0
		if ing1.Tier == 0 && ing2.Tier == 0 {
			break
		}

		// Recurse deeper if necessary
		if ing1.Tier != 0 {
			wg.Add(1)
			go DFS_Single(ing1, wg, elements, recipes)
		}
		if ing2.Tier != 0 {
			wg.Add(1)
			go DFS_Single(ing2, wg, elements, recipes)
		}
		break // Only use one recipe per element
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