package main

import (
	"fmt"
	"time"
)

// type job func(in, out chan interface{})
// go run .\main.go .\common.go .\signer.go
func main() {
	// inputData := []int{0, 1, 1, 2, 3, 5, 8}
	inputData := []int{0, 1}

	testResult := "NOT_SET"
	hashSignJobss := []job{
		job(func(in, out chan interface{}) {
			for _, fibNum := range inputData {
				out <- fibNum
			}
		}),
		job(SingleHash),
		job(MultiHash),
		job(CombineResults),
		job(func(in, out chan interface{}) {
			dataRaw := <-in
			data, ok := dataRaw.(string)
			if !ok {
				fmt.Println("cant convert result data to string")
			}
			testResult = data
		}),
	}
	start := time.Now()

	ExecutePipeline(hashSignJobss...)

	end := time.Since(start)

	fmt.Println(float32(end) / float32(time.Second))
	fmt.Println(testResult)
}
