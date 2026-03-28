package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func generator(n int) chan int {
	randoms := make(chan int)
	go func() {
		for i := 0; i < n; i++ {
			randoms <- rand.Intn(100)
		}
		close(randoms)
	}()

	return randoms
}

func filter(nums chan int, filterFunc func(num int) bool) chan int {
	result := make(chan int)
	wg := &sync.WaitGroup{}
	wg.Add(3)
	for i := 0; i < 3; i++ {
		go func() {
			defer wg.Done()
			for num := range nums {
				time.Sleep(400 * time.Millisecond)
				if filterFunc(num) {
					result <- num
				}
			}
		}()
	}
	go func() {
		wg.Wait()
		close(result)
	}()

	return result
}

func square(nums chan int) chan int {
	result := make(chan int)
	go func() {
		for n := range nums {
			result <- n * n
		}
		close(result)
	}()

	return result
}

func main() {
	nums := generator(100)
	even := filter(nums, func(n int) bool { return n%2 == 0 })
	squares := square(even)

	timeout := time.After(time.Second)
	for {
		select {
		case num, ok := <-squares:
			if !ok {
				return
			}
			fmt.Println(num)
		case <-timeout:
			fmt.Println("timeout")
			return
		}
	}
}