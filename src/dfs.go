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
	Name       string             `json:"name"`
	Attributes string             `json:"attributes"`
	Children   []ExportableRecipe `json:"children"`
}

type ExportableRecipe struct {
	Attributes string              `json:"attributes"`
	Children   []ExportableElement `json:"children"`
}

var numberVisit int32
var visitMu sync.Mutex
var recipeLeft int32 = 4 // set this before starting

func DFS_Multiple(
	current *ElementNode,
	wg *sync.WaitGroup,
	elements map[string]*ElementNode,
	depthChan chan int,
) {
	defer wg.Done()

	visitMu.Lock()
	if current.IsVisited {
		visitMu.Unlock()
		return
	}
	visitMu.Unlock()

	if depthChan != nil {
		// Send the current depth to the channel
		depthChan <- current.Tier
	}

	if current.Tier == 0 {
		return
	}

	count := atomic.AddInt32(&numberVisit, 1)
	fmt.Printf("Visiting node Multi (%d): %s Tier: %d RecipeLeft: %d\n", count, current.Name, current.Tier, atomic.LoadInt32(&recipeLeft))

	ALLrecipes := current.Children
	current.Children = []*RecipeNode{}
	for i, recipe := range ALLrecipes {
		if recipe.Result != current.Name {
			continue
		}
		// fmt.Printf("Processing recipe %d for %s\n", i, current.Name)

		ing1 := recipe.Ingredient1
		ing2 := recipe.Ingredient2
		if ing1 == nil || ing2 == nil {
			continue
		}

		// if ing1.Tier > current.Tier || ing2.Tier > current.Tier {
		// 	fmt.Printf("Abandoning recipe for %s: ingredient tier too high (%s: %d, %s: %d > %d)\n",
		// 		current.Name, ing1.Name, ing1.Tier, ing2.Name, ing2.Tier, current.Tier)
		// 	continue
		// }

		// Only decrement global counter if this is not the first recipe
		if i != 0 {
			if atomic.LoadInt32(&recipeLeft) <= 0 {
				break
			}
			atomic.AddInt32(&recipeLeft, -1)
		}

		// Append the recipe
		visitMu.Lock()
		current.Children = append(current.Children, recipe)
		visitMu.Unlock()

		// Mark tier-0 ingredients as visited
		visitMu.Lock()
		if ing1.Tier == 0 {
			ing1.IsVisited = true
		}
		if ing2.Tier == 0 {
			ing2.IsVisited = true
		}
		visitMu.Unlock()

		// Stop if both ingredients are tier 0
		if ing1.Tier == 0 && ing2.Tier == 0 {
			continue
		}
		// Recurse deeper
		if ing1.Tier != 0 && !isVisited(ing1) {
			wg.Add(1)
			go DFS_Multiple(ing1, wg, elements, depthChan)
		}
		if ing2.Tier != 0 && !isVisited(ing2) {
			wg.Add(1)
			go DFS_Multiple(ing2, wg, elements, depthChan)
		}
	}

	// If no recipes processed (e.g. due to recipeLeft = 0), fallback to DFS_Single
	if len(current.Children) == 0 && atomic.LoadInt32(&recipeLeft) <= 0 {
		wg.Add(1)
		go DFS_Single(current, wg, elements)
	}

	visitMu.Lock()
	current.IsVisited = true
	visitMu.Unlock()
}

// Helper to check if a node is visited with mutex protection
func isVisited(node *ElementNode) bool {
	visitMu.Lock()
	defer visitMu.Unlock()
	return node.IsVisited
}

func DFS_Single(
	current *ElementNode,
	wg *sync.WaitGroup,
	elements map[string]*ElementNode,
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
	fmt.Printf("Visiting node Single (%d): %s Tier: %d\n", count, current.Name, current.Tier)

	recipes := current.Children
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
		current.Children = []*RecipeNode{recipe}
		visitMu.Unlock()

		// Stop if both ingredients are tier 0
		if ing1.Tier == 0 && ing2.Tier == 0 {
			break
		}

		// Recurse deeper if necessary
		if ing1.Tier != 0 {
			wg.Add(1)
			go DFS_Single(ing1, wg, elements)
		}
		if ing2.Tier != 0 {
			wg.Add(1)
			go DFS_Single(ing2, wg, elements)
		}
		break // Only use one recipe per element
	}
}

// Converts internal ElementNode to exportable form
func ToExportableElement(node *ElementNode, res *ExportableElement, visited map[*ElementNode]bool) {
	if node == nil || !node.IsVisited {
		return
	}

	if visited[node] {
		// prevent infinite cycle
		res.Name = node.Name
		res.Attributes = "element"
		return
	}
	visited[node] = true

	res.Name = node.Name
	res.Attributes = "element"

	if node.Tier == 0 {
		return
	}

	res.Children = make([]ExportableRecipe, len(node.Children))
	for i := range node.Children {
		ToExportableRecipe(node.Children[i], &res.Children[i], visited)
	}
}

func ToExportableRecipe(node *RecipeNode, res *ExportableRecipe, visited map[*ElementNode]bool) {
	if node == nil {
		return
	}
	res.Attributes = "recipe"
	res.Children = make([]ExportableElement, 2)
	ToExportableElement(node.Ingredient1, &res.Children[0], visited)
	ToExportableElement(node.Ingredient2, &res.Children[1], visited)
}
