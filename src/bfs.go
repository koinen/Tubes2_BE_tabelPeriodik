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

func bfs(root *ElementNode, elements map[string]*ElementNode, recipes []*RecipeNode, limitRecipe int, ch chan int) {
	// q := make(chan *ElementNode, 100)
	visited := make(map[string]bool)
	var mu sync.Mutex
	var ru sync.Mutex

	mu.Lock()
	visited[root.Name] = true
	root.IsVisited = true
	mu.Unlock()
	// var recipeCount int
	// recipeCount = 1
	var recipeMu sync.Mutex
	recipeCount := make(map[string]int)
	// done := make(chan struct{})

	// const numberofWoker = 4
	// var wg sync.WaitGroup

	// wg.Add(1)
	// q <- root
	currentLevel := []*ElementNode{root}
	// temp := []*RecipeNode{}
	count := 0

	for len(currentLevel) > 0 {
		// wg.Add(1)
		var wg sync.WaitGroup
		var nextLevel []*ElementNode

		// go func() {
		// if ch != nil {
		// 	fmt.Println("Level: ", currentLevel[0].Tier)
		// 	ch <- currentLevel[0].Tier
		// }
		for _, current := range currentLevel {
			// fmt.Println("Worker: ", i)
			// wg.Done()
			// fmt.Println(current.Name)
			// if current.Tier == 0 {

			// 	continue
			// }
			
			current.Children = []*RecipeNode{}

			wg.Add(1)
			go func(current *ElementNode) {
				defer wg.Done()

				for _, recipe := range recipes {

					if recipe.Result != current.Name {
						continue
					}
					if recipe.Ingredient1 == nil || recipe.Ingredient2 == nil {
						// fmt.Printf("Skipping recipe with nil ingredient: %+v\n", recipe)
						continue
					}
					base1, ok1 := elements[recipe.Ingredient1.Name]
					base2, ok2 := elements[recipe.Ingredient2.Name]
					if !ok1 || !ok2 {
						continue
					}
					if base1.Tier >= current.Tier || base2.Tier >= current.Tier {

						continue
					}

					mu.Lock()

					ru.Lock()

					if current == root {
						if count >= limitRecipe {
							mu.Unlock()
							ru.Unlock()
							return
						}
						// recipeMu.Lock()
						// if recipeCount > limitRecipe {
						// 	mu.Unlock()
						// 	ru.Unlock()
						// 	recipeMu.Unlock()
						// 	return
						// }
						exist := false
						for _, c := range current.Children {
							if base1.Name == c.Ingredient1.Name && base2.Name == c.Ingredient2.Name {
								exist = true
								break
							}
						}

						if exist {
							mu.Unlock()
							ru.Unlock()
							continue
						}
						count++
						// recipeCount++
						// if len(current.Children) > 0 {

						// 	recipeCount = recipeCount / len(current.Children)
						// 	recipeCount = recipeCount * (len(current.Children) + 1)
						// }
						// recipeMu.Unlock()

						current.Children = append(current.Children, &RecipeNode{
							Result:      current.Name,
							Ingredient1: elements[base1.Name],
							Ingredient2: elements[base2.Name],
						})
						recipeMu.Lock()
						recipeCount[current.Name] += 1
						recipeMu.Unlock()

					} else {
						exist := false
						for _, c := range current.Children {
							if base1.Name == c.Ingredient1.Name && base2.Name == c.Ingredient2.Name {
								exist = true
								break
							}
						}
						// recipeMu.Lock()
						// if recipeCount > limitRecipe {
						// 	mu.Unlock()
						// 	ru.Unlock()
						// 	recipeMu.Unlock()
						// 	return
						// }
						// recipeCount++
						// if len(current.Children) > 0 {

						// 	recipeCount = recipeCount / len(current.Children)
						// 	recipeCount = recipeCount * (len(current.Children) + 1)
						// }
						// recipeMu.Unlock()

						if exist {
							mu.Unlock()
							ru.Unlock()
							continue
						}
						res := RecipeNode{
							Result:      current.Name,
							Ingredient1: elements[base1.Name],
							Ingredient2: elements[base2.Name],
						}
						recipeMu.Lock()

						current_count := 1
						recipeCount[current.Name] += 1
						for _, v := range recipeCount {
							// if v >= 1 {
							current_count *= v
							// }
						}

						if current_count > limitRecipe {
							recipeCount[current.Name] -= 1

							ru.Unlock()
							mu.Unlock()
							recipeMu.Unlock()
							break
						}
						recipeMu.Unlock()
						current.Children = append(current.Children, &res)
						fmt.Printf("Appending recipe for %s, %s + %s\n", current.Name, base1.Name, base2.Name)
						if ch != nil {
							fmt.Println("Level: ", currentLevel[0].Tier)
							ch <- currentLevel[0].Tier
						}
					}
					ru.Unlock()
					mu.Unlock()
					//BFS
					mu.Lock()
					if !base1.IsVisited {
						fmt.Println("Enqueue: ", base1.Name)
						//Enqueue
						// wg.Add(1)
						// if base1.Name == "Stone" {
						// 	fmt.Println("STONE")
						// }
						nextLevel = append(nextLevel, base1)
						visited[base1.Name] = true
						base1.IsVisited = true
					}

					if !base2.IsVisited {
						//Enqueue
						// wg.Add(1)
						nextLevel = append(nextLevel, base2)
						fmt.Printf("Enqueue: %s\n", base2.Name)
						visited[base2.Name] = true
						base2.IsVisited = true
					}
					mu.Unlock()
				}
			}(current)
			
		}
		wg.Wait()

		fmt.Println("Level: ", currentLevel[0].Tier)
		currentLevel = nextLevel

		// }()
		// if len(root.Children) >= limitRecipe {
		// 	return
		// }
	}

	// go func() {
	// 	wg.Wait()
	// 	close(done)
	// }()

	// <-done

	if ch != nil {
		fmt.Println("DONE")
		close(ch)
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
