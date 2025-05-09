package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
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

	fmt.Print("Enter element tier: ")
	elementTierStr, _ := reader.ReadString('\n')
	elementTierStr = strings.TrimSpace(elementTierStr)
	elementTier, err := strconv.Atoi(elementTierStr)
	if err != nil {
		fmt.Println("Invalid tier")
		return
	}

	// Prepare maps
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
					Ingredient1: &ElementNode{Name: r[0]},
					Ingredient2: &ElementNode{Name: r[1]},
				})
			}
		}
	}

	// Build tree from user input
	root := &ElementNode{Name: elementName, Tier: elementTier, Children: []*RecipeNode{}}
	wg := &sync.WaitGroup{}

	wg.Add(1)
	DFS_Single(root, wg, elementMap, allRecipes)
	wg.Wait()

	fmt.Println("DFS completed")

	// Export tree
	exportList := ExportableElement{
		Name:       root.Name,
		Attributes: "element",
		Children:   make([]ExportableRecipe, 0, len(root.Children)),
	}
	ToExportableElement(root, &exportList)

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
