package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

type ElementNode struct {
	IsVisited bool
	Name      string
	Left      bool
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
	Attributes map[string]string  `json:"attributes"`
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
			fmt.Printf("DFS_Multiple: %s\n", current.Name)
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

	if current.Tier == 0 {
		return
	}

	count := atomic.AddInt32(&numberVisit, 1)
	fmt.Printf("Visiting node Multi (%d): %s Tier: %d RecipeLeft: %d\n", count, current.Name, current.Tier, atomic.LoadInt32(&recipeLeft))

	// fmt.Printf("Recipe len: %d", len(current.Children))
	ALLrecipes := make([]*RecipeNode, len(current.Children))
	copy(ALLrecipes, current.Children)
	current.Children = []*RecipeNode{}

	// fmt.Printf("Recipe len: %d", len(ALLrecipes))

	fistAdd := true
	for _, recipe := range ALLrecipes {
		if recipe.Result != current.Name {
			continue
		}
		// fmt.Printf("Processing recipe for %s\n", current.Name)

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

	fmt.Printf("DFS_Single: %s\n", current.Name)

	visitMu.Lock()
	if current.IsVisited {
		visitMu.Unlock()
		return
	}
	current.IsVisited = true
	visitMu.Unlock()

	count := atomic.AddInt32(&numberVisit, 1)
	fmt.Printf("Visiting node Single (%d): %s Tier: %d\n", count, current.Name, current.Tier)

	ALLrecipes := make([]*RecipeNode, len(current.Children))
	copy(ALLrecipes, current.Children)
	current.Children = []*RecipeNode{}

	for _, recipe := range ALLrecipes {
		if recipe.Result != current.Name {
			continue
		}

		// fmt.Printf("Processing recipe for %s\n", current.Name)

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
		res.Attributes = visited[node].Attributes
		if node.Left {
			res.Attributes["Side"] = "Left"
		} else {
			res.Attributes["Side"] = "Right"
		}
		res.Children = visited[node].Children
		return
	}

	visited[node] = res

	res.Name = node.Name
	res.Attributes = make(map[string]string)
	if node.Left {
		res.Attributes["Side"] = "Left"
	} else {
		res.Attributes["Side"] = "Right"
	}

	if node.Left {
		fmt.Printf("Left: %s\n", node.Name)
	}
	res.Attributes["Type"] = "element"

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

func ToExportableElement3(node *ElementNode, visited map[*ElementNode]*ExportableElement)*ExportableElement {
    if node == nil || !node.IsVisited {
        return nil
    }

    // Already visited? Return the pointer directly to avoid recursion
    if cached, ok := visited[node]; ok {
        return cached
    }

    exported := &ExportableElement{
        Name:       node.Name,
        Attributes: make(map[string]string),
        Children:   []ExportableRecipe{},
    }
	exported.Attributes["Type"] = "element"
	if node.Left {
		exported.Attributes["Side"] = "Left"
	} else {
		exported.Attributes["Side"] = "Right"
	}
    // Put early in map to avoid recursive cycles
    visited[node] = exported

    if node.Tier != 0 {
        for _, recipeNode := range node.Children {
            exported.Children = append(exported.Children, *ToExportableRecipe3(recipeNode, visited))
        }
    }

    return exported
}

func ToExportableRecipe3(recipe *RecipeNode, visited map[*ElementNode]*ExportableElement) *ExportableRecipe {
    if recipe == nil {
        return nil
    }

    return &ExportableRecipe{
        Attributes: "recipe",
        Children: []ExportableElement{
            *ToExportableElement3(recipe.Ingredient1, visited),
            *ToExportableElement3(recipe.Ingredient2, visited),
        },
    }
}
