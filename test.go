package main

import "fmt"


func main() {
	value := make(chan int)
	
	for i := 0; i < 5; i++ {
		go func() {
			fmt.Printf("Hello from goroutine %d\n", i)
			value <- i
		}()
	}

	for i := 0; i < 5; i++ {
		fmt.Printf("Hello from main goroutine\n")
		fmt.Println(<-value)
	}
}