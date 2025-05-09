package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"bufio"
	"strings"
)

func main() {
	// convertToJson(scrape())
	// serve()
	// Read JSON
	data, err := os.ReadFile("data/recipes.json")
	if err != nil {
		panic(err)
	}

	var rawElements []Element
	if err := json.Unmarshal(data, &rawElements); err != nil {
		panic(err)
	}

	// Prompt user
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter element name: ")
	elementName, _ := reader.ReadString('\n')
	elementName = strings.TrimSpace(elementName)

	// fmt.Print("Enter element tier: ")
	// elementTierStr, _ := reader.ReadString('\n')
	// elementTierStr = strings.TrimSpace(elementTierStr)
	// elementTier, err := strconv.Atoi(elementTierStr)
	// if err != nil {
	// 	fmt.Println("Invalid tier")
	// 	return
	// }

	// Prepare maps
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
	// root := &ElementNode{Name: elementName, Tier: elementTier, Children: []*RecipeNode{}}
	root := elementMap[elementName]
	wg := &sync.WaitGroup{}

	wg.Add(1)
	DFS_Multiple(root, wg, elementMap)
	wg.Wait()

	fmt.Println("DFS completed")

	// Export tree
	exportList := ExportableElement{
		Name:       root.Name,
		Attributes: "element",
		Children:   make([]ExportableRecipe, 0, len(root.Children)),
	}
	visitedExport := make(map[*ElementNode]bool)
	ToExportableElement(root, &exportList, visitedExport)

	// Write to file
	jsonOut, err := json.MarshalIndent(exportList, "", "  ")
	if err != nil {
		panic(err)
	}

	if err := os.WriteFile("data/output.json", jsonOut, 0644); err != nil {
		panic(err)
	}

	fmt.Println("Exported to data/output.json")
}