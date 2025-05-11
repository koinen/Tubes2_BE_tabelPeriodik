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
var sem chan struct{}
var recipeLeft int32

func DFS_Multiple(
	current *ElementNode,
	wg *sync.WaitGroup,
	elements map[string]*ElementNode,
	depthChan chan int,
) {
	defer func (){
		if depthChan != nil {
			depthChan <- current.Tier
		}
	}()
	visitMu.Lock()
	if current.IsVisited {
		visitMu.Unlock()
		return
	}
	current.IsVisited = true
	visitMu.Unlock()

	// if depthChan != nil {
	// 	depthChan <- current.Tier
	// }

	if current.Tier == 0 {
		return
	}

	count := atomic.AddInt32(&numberVisit, 1)
	fmt.Printf("Visiting node Multi (%d): %s Tier: %d RecipeLeft: %d\n", count, current.Name, current.Tier, atomic.LoadInt32(&recipeLeft))

	ALLrecipes := make([]*RecipeNode, len(current.Children))
	copy(ALLrecipes, current.Children)
	current.Children = []*RecipeNode{}

	fistAdd := true
	for _, recipe := range ALLrecipes {
		if recipe.Result != current.Name {
			continue
		}
		// fmt.Printf("Processing recipe %d for %s\n", i, current.Name)

		ing1 := recipe.Ingredient1
		ing2 := recipe.Ingredient2
		if ing1 == nil || ing2 == nil {
			continue
		}

		if ing1.Tier >= current.Tier || ing2.Tier >= current.Tier {
			continue
		}

		if !fistAdd {
			if atomic.LoadInt32(&recipeLeft) <= 0 {
				continue
			}
			atomic.AddInt32(&recipeLeft, -1)
		}
		fistAdd = false

		visitMu.Lock()
		current.Children = append(current.Children, recipe)
		fmt.Printf("Appending recipe Multi for %s, %s + %s\n", current.Name, ing1.Name, ing2.Name)
		visitMu.Unlock()

		visitMu.Lock()
		if ing1.Tier == 0 {
			ing1.IsVisited = true
		}
		if ing2.Tier == 0 {
			ing2.IsVisited = true
		}
		visitMu.Unlock()

		if ing1.Tier == 0 && ing2.Tier == 0 {
			continue
		}

		if depthChan != nil {
			depthChan <- current.Tier
		}

		select {
		case sem <- struct{}{}:
			wg.Add(1)
			go func(n *ElementNode) {
				defer wg.Done()
				DFS_Multiple(n, wg, elements, depthChan)
				<-sem // release slot
			}(ing1)
		default:
			DFS_Multiple(ing1, wg, elements, depthChan)
		}
		select {
		case sem <- struct{}{}:
			wg.Add(1)
			go func(n *ElementNode) {
				defer wg.Done()
				DFS_Multiple(n, wg, elements, depthChan)
				<-sem // release slot
			}(ing2)
		default:
			DFS_Multiple(ing2, wg, elements, depthChan)
		}
	}

	if len(current.Children) == 0 {
		DFS_Single(current, wg, elements, depthChan)
	}
	visitMu.Lock()
	current.IsVisited = true
	visitMu.Unlock()
}

func DFS_Single(
	current *ElementNode,
	wg *sync.WaitGroup,
	elements map[string]*ElementNode,
	depthChan chan int,
) {
	defer func (){
		if depthChan != nil {
			depthChan <- current.Tier
		}
	}()

	visitMu.Lock()
	if current.IsVisited {
		visitMu.Unlock()
		return
	}
	current.IsVisited = true
	visitMu.Unlock()

	// if depthChan != nil {
	// 	depthChan <- current.Tier
	// }

	count := atomic.AddInt32(&numberVisit, 1)
	fmt.Printf("Visiting node Single (%d): %s Tier: %d\n", count, current.Name, current.Tier)

	ALLrecipes := make([]*RecipeNode, len(current.Children))
	copy(ALLrecipes, current.Children)
	current.Children = []*RecipeNode{}

	for _, recipe := range ALLrecipes {
		if recipe.Result != current.Name {
			continue
		}

		ing1, ok1 := elements[recipe.Ingredient1.Name]
		ing2, ok2 := elements[recipe.Ingredient2.Name]
		if !ok1 || !ok2 {
			continue
		}

		if ing1.Tier >= current.Tier || ing2.Tier >= current.Tier {
			fmt.Printf("Abandoning recipe Single for %s: ingredient tier too high (%s: %d, %s: %d > %d)\n",
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

		visitMu.Lock()
		current.Children = append(current.Children, recipe)
		fmt.Printf("Appending recipe for %s, %s + %s\n", current.Name, ing1.Name, ing2.Name)
		visitMu.Unlock()

		if ing1.Tier == 0 && ing2.Tier == 0 {
			break
		}

		if depthChan != nil {
			depthChan <- current.Tier
		}

		select {
		case sem <- struct{}{}:
			wg.Add(1)
			go func(n *ElementNode) {
				defer wg.Done()
				DFS_Multiple(n, wg, elements, depthChan)
				<-sem // release slot
			}(ing1)
		default:
			DFS_Multiple(ing1, wg, elements, depthChan)
		}
		select {
		case sem <- struct{}{}:
			wg.Add(1)
			go func(n *ElementNode) {
				defer wg.Done()
				DFS_Multiple(n, wg, elements, depthChan)
				<-sem // release slot
			}(ing2)
		default:
			DFS_Multiple(ing2, wg, elements, depthChan)
		}
		break // Only use one recipe per element
	}
}

// Converts internal ElementNode to exportable form
func ToExportableElement(node *ElementNode, res *ExportableElement, visited map[*ElementNode]*ExportableElement) {
	if node == nil || !node.IsVisited {
		return
	}

	if _, exists := visited[node]; exists {
		// prevent infinite cycle
		res.Name = visited[node].Name
		res.Attributes = "element"
		res.Children = visited[node].Children
		return
	}

	visited[node] = res

	res.Name = node.Name
	res.Attributes = "element"

	if node.Tier == 0 {
		return
	}

	res.Children = make([]ExportableRecipe, len(node.Children))
	for i := range node.Children {
		ToExportableRecipe(node.Children[i], &res.Children[i], visited)
	}

	// if node.Name == "Pressure" {
	// 	fmt.Printf("Children of Pres: %d\n", len(node.Children))
	// 	for _, child := range node.Children {
	// 		fmt.Printf("Child: %s\n", child.Ingredient1.Name)
	// 		fmt.Printf("Child: %s\n", child.Ingredient2.Name)
	// 	}
	// }
}

func ToExportableRecipe(node *RecipeNode, res *ExportableRecipe, visited map[*ElementNode]*ExportableElement) {
	if node == nil {
		return
	}
	res.Attributes = "recipe"
	res.Children = make([]ExportableElement, 2)
	ToExportableElement(node.Ingredient1, &res.Children[0], visited)
	ToExportableElement(node.Ingredient2, &res.Children[1], visited)
}
