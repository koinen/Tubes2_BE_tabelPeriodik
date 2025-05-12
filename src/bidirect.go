package main

import (
	"fmt"
	"sync"
	"time"
)

func Bidirect_Right_DFS(
	root *ElementNode,
	wg *sync.WaitGroup,
	elementMap map[string]*ElementNode,
	depthChan chan int,
	doneChan chan struct{},
) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		DFS_Multiple(root, wg, elementMap, nil)
		fmt.Println("[DFS Right] Done")
		close(doneChan) // Notify BFS
	}()
}

func Bidirect_Right_BFS(
	root *ElementNode,
	wg *sync.WaitGroup,
	elementMap map[string]*ElementNode,
	allRecipes []*RecipeNode,
	depthChan chan int,
	doneChan chan struct{},
) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		bfs(root, elementMap, allRecipes)
		fmt.Println("[BFS Right] Done")
		close(doneChan) // Notify BFS
	}()
}

func Bidirect_Left_BFS(
	basic []*ElementNode,
	target *ElementNode,
	allElement map[string]*ElementNode,
	allRecipes []*RecipeNode,
	doneChan <-chan struct{},
) {
	fmt.Println("[BFS] Bidirect_Left_BFS started")

	discovered := make(map[string]*ElementNode)
	tierElements := make(map[int][]*ElementNode)
	for _, el := range basic {
		discovered[el.Name] = el
		el.IsVisited = true
		tierElements[el.Tier] = append(tierElements[el.Tier], el)
	}

	ingredient := make(chan *ElementNode, 100)
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Worker
	worker := func(id int, recipes []*RecipeNode) {
		defer wg.Done()
		fmt.Printf("[Worker %d] Started with %d recipes\n", id, len(recipes))
		for _, recipe := range recipes {
			select {
			case <-doneChan:
				fmt.Printf("[Worker %d] Received doneChan, exiting early\n", id)
				return
			default:
			}
			mu.Lock()
			in1 := recipe.Ingredient1
			in2 := recipe.Ingredient2
			result := allElement[recipe.Result]
			fmt.Printf("[Worker %d] Checking recipe: %s + %s -> %s\n", id, in1.Name, in2.Name, result.Name)

			if result.IsVisited {
				fmt.Printf("[Worker %d] Skipped %s (already visited)\n", id, result.Name)
				mu.Unlock()
				continue
			}
			if !in1.IsVisited || !in2.IsVisited {
				fmt.Printf("[Worker %d] Skipped %s (missing ingredients)\n", id, result.Name)
				mu.Unlock()
				continue
			}
			if result.Tier >= target.Tier {
				fmt.Printf("[Worker %d] Skipped %s (tier too high)\n", id, result.Name)
				mu.Unlock()
				continue
			}
			if result.Tier <= in1.Tier || result.Tier <= in2.Tier {
				fmt.Printf("[Worker %d] Skipped %s (tier not increasing)\n", id, result.Name)
				mu.Unlock()
				continue
			}

			result.IsVisited = true
			discovered[result.Name] = result
			tierElements[result.Tier] = append(tierElements[result.Tier], result)
			fmt.Printf("[Worker %d] Discovered new element: %s (tier %d)\n", id, result.Name, result.Tier)
			mu.Unlock()
			ingredient <- result
		}
		fmt.Printf("[Worker %d] Finished\n", id)
	}

	for currentTier := 1; currentTier < target.Tier; currentTier++ {
		fmt.Printf("[BFS] Processing tier %d -> %d\n", currentTier, currentTier+1)
		select {
		case <-doneChan:
			fmt.Println("[BFS] Cancelled by DFS (doneChan closed)")
			return
		default:
		}

		// Filter recipes that produce elements to next tier
		nextTier := currentTier + 1
		candidates := make([]*RecipeNode, 0)
		for _, recipe := range allRecipes {
			result := allElement[recipe.Result]
			if result.Tier == nextTier {
				candidates = append(candidates, recipe)
			}
		}
		fmt.Printf("[BFS] Tier %d: %d recipe candidates found\n", nextTier, len(candidates))

		// Start workers
		numWorkers := 4
		chunkSize := (len(candidates) + numWorkers - 1) / numWorkers
		for i := range numWorkers {
			start := i * chunkSize
			end := min(start + chunkSize, len(candidates))
			if start >= end {
				continue
			}
			wg.Add(1)
			go worker(i, candidates[start:end])
		}

		wg.Wait()

		// Drain ingredients
		drained := false
		count := 0
		for !drained {
			select {
			case <-doneChan:
				fmt.Println("[BFS] Received doneChan signal during draining. Exiting early.")
				return

			case newEl := <-ingredient:
				fmt.Printf("[BFS] -> New element added: %s (tier %d)\n", newEl.Name, newEl.Tier)
				count++
				if newEl.Tier == target.Tier {
					fmt.Printf("[BFS] Target tier %d reached with element %s\n", target.Tier, newEl.Name)
					return
				}

			case <-time.After(500 * time.Millisecond):
				drained = true
			}
		}
		fmt.Printf("[BFS] Tier %d complete, %d new elements discovered\n", currentTier, count)
	}

	fmt.Println("[BFS Left] Finished")
}

func Bidirect_Left_DFS(
	basic []*ElementNode,
	target *ElementNode,
	allElement map[string]*ElementNode,
	allRecipes []*RecipeNode,
	doneChan <-chan struct{},
) {
	stack := make([]*ElementNode, 0)
	var mu sync.Mutex

	// basic elements
	for _, el := range basic {
		stack = append(stack, el)
		el.IsVisited = true
	}

	// DFS Loop
	for len(stack) > 0 {
		select {
		case <-doneChan:
			fmt.Println("[DFS] Received doneChan signal, exiting early.")
			return
		default:
			currentElement := stack[len(stack)-1]
			stack = stack[:len(stack)-1] // Pop

			fmt.Printf("[DFS] Processing element: %s (tier %d)\n", currentElement.Name, currentElement.Tier)

			if currentElement.Tier == target.Tier {
				fmt.Printf("[DFS] Target tier %d reached with element %s\n", target.Tier, currentElement.Name)
				return
			}

			for _, recipe := range allRecipes {
				if currentElement.Name != recipe.Ingredient1.Name && currentElement.Name != recipe.Ingredient2.Name {
					continue
				}

				newElement := allElement[recipe.Result]
				if newElement == nil || newElement.IsVisited|| newElement.Tier >= target.Tier {
					continue
				}

				// push
				mu.Lock()
				newElement.IsVisited = true
				mu.Unlock()

				stack = append(stack, newElement)
				fmt.Printf("[DFS] Discovered new element: %s (tier %d)\n", newElement.Name, newElement.Tier)
			}
		}
	}

	fmt.Println("[DFS Left] No more elements to process. Exiting DFS.")
}