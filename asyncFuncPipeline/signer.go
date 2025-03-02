package main

import (
	"fmt"
	"slices"
	"strconv"
	"sync"
)

// go run .\main.go .\common.go .\signer.go

// сюда писать код
// - Написание функции ExecutePipeline которая обеспечивает нам конвейерную обработку функций-воркеров, которые что-то делают.
// - Написание нескольких функций, которые считают нам какую-то условную хеш-сумму от входных данных

// На все расчеты у нас 3 сек.
func ExecutePipeline(jobs ...job) {
	fmt.Println("ExecutePipeline")
	chCount := len(jobs) + 1
	chanals := make([]chan interface{}, chCount)
	for i := range chCount {
		chanals[i] = make(chan interface{}, 3)
	}
	for i, j := range jobs[:len(jobs)-1] {
		go func(j job, in, out chan interface{}) {
			j(in, out)
			for {
				if len(out) == 0 {
					close(out)
					break
				}
			}
		}(j, chanals[i], chanals[i+1])
	}
	jobs[len(jobs)-1](chanals[chCount-2], chanals[chCount-1])
	fmt.Println("ExecutePipeline OUT")

}

// может одновременно вызываться только 1 раз, считается 10 мс
func md5Worker(data string, quotaCh chan struct{}) string {
	quotaCh <- struct{}{}
	res := DataSignerMd5(data)
	<-quotaCh
	return res
}

type orderedResult struct {
	result  string
	ordinal int
}

// cчитается 1 сек
func crc32Worker(data string, out chan orderedResult, ordinal int) {
	res := DataSignerCrc32(data)
	ores := orderedResult{res, ordinal}
	out <- ores
}

// считает значение crc32(data)+"~"+crc32(md5(data))
func SingleHash(in, out chan interface{}) {
	fmt.Println("SingleHash")
	wg := new(sync.WaitGroup)
	quotaCh := make(chan struct{}, 1)
	for data := range in {
		var dataStr string
		switch data := data.(type) {
		case string:
			dataStr = data
		case int:
			dataStr = strconv.Itoa(data)
		default:
			fmt.Printf("%T\n", data)
			panic(fmt.Sprintf("Пришёл не тот тип %T", data))
		}
		crc32Out := make(chan orderedResult, 2)
		go crc32Worker(dataStr, crc32Out, 0)
		go crc32Worker(md5Worker(dataStr, quotaCh), crc32Out, 1)
		wg.Add(1)
		go func(out chan<- interface{}, crc32Out <-chan orderedResult, wg *sync.WaitGroup) {
			defer wg.Done()
			for {
				if len(crc32Out) == 2 {
					order := make([]string, 2)
					for range 2 {
						ans := <-crc32Out
						order[ans.ordinal] = ans.result
					}
					res := order[0] + "~" + order[1]
					out <- res
					fmt.Println("SingleHash RES", res)
					break
				}
			}
		}(out, crc32Out, wg)
	}
	wg.Wait()
	fmt.Println("SingleHash OUT")
}

// считает конкатенацию в порядке расчета значений crc32(th+data)), где th=0..5
func MultiHash(in, out chan interface{}) {
	fmt.Println("MultiHash")
	wg := new(sync.WaitGroup)
	for data := range in {
		dataStr, ok := data.(string)
		if !ok {
			panic("Пришёл не тот тип")
		}
		crc32Out := make(chan orderedResult, 6)
		for th := range 6 {
			go crc32Worker(strconv.Itoa(th)+dataStr, crc32Out, th)
		}
		wg.Add(1)
		go func(out chan<- interface{}, crc32Out <-chan orderedResult, wg *sync.WaitGroup) {
			defer wg.Done()
			for {
				if len(crc32Out) == 6 {
					var res string
					order := make([]string, 6)
					for range 6 {
						ans := <-crc32Out
						order[ans.ordinal] = ans.result
					}
					for i := range 6 {
						res += order[i]
					}
					out <- res
					fmt.Println("MultiHash RES", res)
					break
				}
			}
		}(out, crc32Out, wg)
	}
	wg.Wait()
	fmt.Println("MultiHash OUT")
}

// получает все результаты, сортирует, объединяет отсортированный
// результат через _ (символ подчеркивания) в одну строку
func CombineResults(in, out chan interface{}) {
	fmt.Println("CombineResults")
	ress := make([]string, 0, MaxInputDataLen)
	for data := range in {
		dataStr, ok := data.(string)
		if !ok {
			panic("Пришёл не тот тип")
		}
		ress = append(ress, dataStr)
	}

	slices.Sort(ress)
	var res string
	for i, str := range ress {
		res += str
		if i < len(ress)-1 {
			res += "_"
		}
	}
	out <- res
	fmt.Println("CombineResults OUT")
}
