package main

import (
	"fmt"
)

// type ElementNode struct {
// 	IsVisited bool
// 	Name      string
// 	Tier      int
// 	Recipes   []*RecipeNode
// }

// type RecipeNode struct {
// 	Result      string
// 	Ingredient1 *ElementNode
// 	Ingredient2 *ElementNode
// }

// type ExportableElement struct {
// 	Name    string             `json:"name"`
// 	Tier    int                `json:"tier"`
// 	Recipes []ExportableRecipe `json:"recipes"`
// }

// type ExportableRecipe struct {
// 	Ingredient1 string `json:"ingredient1"`
// 	Ingredient2 string `json:"ingredient2"`
// }

func NewQueue[T any]() Queue[T] {
	return Queue[T]{}
}

type Queue[T any] []T

func (q *Queue[T]) Len() int {
	return len(*q)
}

func (q *Queue[T]) Enqueue(value T) {
	*q = append(*q, value)
}

func (q *Queue[T]) Dequeue() T {
	temp := *q

	if q.Len() >= 1 {
		res := temp[0]
		*q = temp[1:]
		return res
	} else {
		fmt.Print("Queue kosong")
		var empty T
		return empty
	}
}

// func bfs(root *ElementNode, recipes []RecipeNode, elements map[string]*ElementNode) {
// 	q := []*ElementNode{root}
// 	visited := make(map[string]bool)
// 	visited[root.Name] = true

// 	for len(q) > 0 {
// 		current := q[0]

// 		for _, recipe := range recipes {
// 			if recipe.Result != current.Name {
// 				continue
// 			}

// 			base1, ok1 := elements[recipe.Ingredient1.Name]
// 			base2, ok2 := elements[recipe.Ingredient2.Name]

// 			if ok1 && ok2 {
// 				current.Children = append(current.Children, &RecipeNode{
// 					Result: current.Name,
// 				})
// 			}
// 		}
// 	}

// }
