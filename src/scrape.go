package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
)

type Element struct {
	Name    string      `json:"name"`
	Tier    int         `json:"tier"`
	Recipes [][2]string `json:"recipes"`
}

func scrape() []Element {
	c := colly.NewCollector()
	var recipeMap []Element
	var currentTier int

	c.OnHTML("tr, h3", func(e *colly.HTMLElement) {
		if e.Name == "h3" {
			// Save the current <h3> tier title
			tier := strings.TrimSpace(e.DOM.Find("span.mw-headline").Text())
			if len(tier) == 0 {
				return
			}
			if tier == "Starting elements" {
				fmt.Println("Starting elements")
				currentTier = 0
			} else {
				currentTier = int(tier[5] - '0')
				if tier[6] != ' ' {
					currentTier = currentTier*10 + int(tier[6]-'0')
				}
				fmt.Println("Current tier:", currentTier)
			}
			return
		}

		tds := e.DOM.Find("td")
		if tds.Length() < 2 {
			return
		}

		var elmt Element
		if tds.Eq(0).Find("a[title]").First().AttrOr("title", "") == "Elements (Little Alchemy 1)" {
			return
		}

		elmt.Name = tds.Eq(0).Find("a[title]").First().AttrOr("title", "")
		if elmt.Name == "Time" {
			fmt.Println("Skipping Time element")
			return
		}

		elmt.Tier = currentTier
		tds.Eq(1).Find("li").Each(func(i int, li *goquery.Selection) {
			var ingredients [2]string
			li.Find("a[title]").Each(func(j int, a *goquery.Selection) {
				if a.AttrOr("title", "") == "Time" {
					fmt.Println("Skipping Time recipe for element:", elmt.Name)
					return
				}
				ingredients[j] = a.AttrOr("title", "")
			})
			if ingredients[0] != "" && ingredients[1] != "" {
				elmt.Recipes = append(elmt.Recipes, ingredients)
			}
		})
		recipeMap = append(recipeMap, elmt)
	})

	c.OnScraped(func(r *colly.Response) {
	})

	c.Visit("https://little-alchemy.fandom.com/wiki/Elements_(Little_Alchemy_2)")

	return recipeMap
}

func convertToJson(recipes []Element) ([]byte, error) {
	jsonBytes, err := json.MarshalIndent(recipes, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling to JSON:", err)
		return nil, err
	}
	if jsonBytes == nil {
		fmt.Println("jsonBytes is nil")
		return nil, fmt.Errorf("jsonBytes is nil")
	}
	return jsonBytes, nil
}

// func main() {
// 	recipes := scrape()
// 	convertToJson(recipes)
// }
