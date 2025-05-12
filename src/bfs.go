package main

import (
	"fmt"
	"sync"
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

func bfs_single(root *ElementNode, elements map[string]*ElementNode, recipes []RecipeNode) {
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

			if base1.Tier >= current.Tier || base2.Tier >= current.Tier {
				continue
			}

			if ok1 && ok2 {
				if len(current.Children) < 1 {

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
		}
		//Dequeue
		q = q[1:]

	}
}

func bfs(root *ElementNode, elements map[string]*ElementNode, recipes []*RecipeNode) {
	// q := make(chan *ElementNode, 100)
	visited := make(map[string]bool)
	var mu sync.Mutex

	mu.Lock()
	visited[root.Name] = true
	root.IsVisited = true
	mu.Unlock()

	fmt.Printf("Starting BFS for %s\n", root.Name)

	// done := make(chan struct{})

	// const numberofWoker = 4
	// var wg sync.WaitGroup

	// wg.Add(1)
	// q <- root
	currentLevel := []*ElementNode{root}

	for len(currentLevel) > 0 {
		// wg.Add(1)
		var wg sync.WaitGroup
		var nextLevel []*ElementNode

		// go func() {
		for _, current := range currentLevel {
			// fmt.Println("Worker: ", i)
			// wg.Done()
			// fmt.Println(current.Name)
			wg.Add(1)
			go func(current *ElementNode) {
				defer wg.Done()

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

					mu.Lock()
					current.Children = append(current.Children, &RecipeNode{
						Result:      current.Name,
						Ingredient1: base1,
						Ingredient2: base2,
					})
					mu.Unlock()
					//BFS
					mu.Lock()
					if !visited[base1.Name] {
						// fmt.Println("Enque: ", base1.Name)
						//Enqueue
						visited[base1.Name] = true
						base1.IsVisited = true
						// wg.Add(1)
						nextLevel = append(nextLevel, base1)
						fmt.Printf("Enqueue: %s\n", base1.Name)
					}

					if !visited[base2.Name] {
						//Enqueue
						visited[base2.Name] = true
						base2.IsVisited = true
						// wg.Add(1)
						nextLevel = append(nextLevel, base2)
						fmt.Printf("Enqueue: %s\n", base2.Name)
					}
					mu.Unlock()
				}
			}(current)
		}
		wg.Wait()
		currentLevel = nextLevel
		// }()
	}

	// go func() {
	// 	wg.Wait()
	// 	close(done)
	// }()

	// <-done
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
