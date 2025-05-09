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

// func NewQueue[T any]() Queue[T] {
// 	return Queue[T]{}
// }

// type Queue[T any] []T

// func (q *Queue[T]) Len() int {
// 	return len(*q)
// }

// func (q *Queue[T]) Enqueue(value T) {
// 	*q = append(*q, value)
// }

// func (q *Queue[T]) Dequeue() T {
// 	temp := *q

// 	if q.Len() >= 1 {
// 		res := temp[0]
// 		*q = temp[1:]
// 		return res
// 	} else {
// 		fmt.Print("Queue kosong")
// 		var empty T
// 		return empty
// 	}
// }

func bfs(root *ElementNode, elements map[string]*ElementNode, recipes []RecipeNode) {
	q := []*ElementNode{root}
	visited := make(map[string]bool)
	visited[root.Name] = true
	root.IsVisited = true
	for len(q) > 0 {
		current := q[0]
		// fmt.Println(current.Name)

		for _, recipe := range recipes {
			base1, ok1 := elements[recipe.Ingredient1.Name]
			base2, ok2 := elements[recipe.Ingredient2.Name]

			if !ok1 || !ok2 {
				continue
			}

			if recipe.Result != current.Name {
				continue
			}

			if base1.Tier > current.Tier || base2.Tier > current.Tier {
				continue
			}

			if ok1 && ok2 {

				current.Children = append(current.Children, &RecipeNode{
					Result:      current.Name,
					Ingredient1: base1,
					Ingredient2: base2,
				})
				//BFS

				if !visited[base1.Name] {
					// fmt.Println("Enque: ", base1.Name)
					//Enqueue
					q = append(q, base1)
					visited[base1.Name] = true
					base1.IsVisited = true
				}

				if !visited[base2.Name] {
					//Enqueue
					q = append(q, base2)
					visited[base2.Name] = true
					base2.IsVisited = true
				}
			}
		}
		//Dequeue
		q = q[1:]

	}
}

func (n ElementNode) display() {
	r := n.Children
	fmt.Println("Name: ", n.Name)
	for _, i := range r {
		fmt.Println("===========================")
		fmt.Println(i.Ingredient1.Name)
		fmt.Println(i.Ingredient2.Name)
	}
	fmt.Println("END")
}
