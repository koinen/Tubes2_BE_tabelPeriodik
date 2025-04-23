package main

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
)

func findRecipe(element string) [][2]string {
	c := colly.NewCollector()
	found := false
	recipe := [][2]string{}

	c.OnHTML("tr", func(e *colly.HTMLElement) {
		if found {
			return
		}

		tds := e.DOM.Find("td")
		if tds.Length() < 2 {
			return
		}

		if element == tds.Eq(0).Find("a[title]").First().AttrOr("title", "") {
			tds.Eq(1).Find("li").Each(func(i int, li *goquery.Selection) {
				var ingredients []string
				li.Find("a[title]").Each(func(j int, a *goquery.Selection) {
					ingredients = append(ingredients, a.AttrOr("title", ""))
				})
				if len(ingredients) == 2 {
					recipe = append(recipe, [2]string{ingredients[0], ingredients[1]})
				}
			})
			found = true
		}
	})

	c.Visit("https://little-alchemy.fandom.com/wiki/Elements_(Little_Alchemy_2)")
	return recipe
}
