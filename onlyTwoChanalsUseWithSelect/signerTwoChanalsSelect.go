package main

import (
	"fmt"
	"slices"
	"strconv"
	"sync"
)

// $env:CGO_ENABLED = "1"

// сюда писать код
// - Написание функции ExecutePipeline которая обеспечивает нам конвейерную обработку функций-воркеров, которые что-то делают.
// - Написание нескольких функций, которые считают нам какую-то условную хеш-сумму от входных данных

// На все расчеты у нас 3 сек.
func ExecutePipeline(jobs ...job) {
	fmt.Println("ExecutePipeline")
	in := make(chan interface{}, MaxInputDataLen)
	out := make(chan interface{}, MaxInputDataLen)
	for _, job := range jobs {
		fmt.Println("injob")
		job(in, out)
	LOOP:
		for {
			select {
			case i := <-out:
				in <- i
			default:
				break LOOP
			}
		}
		fmt.Println("outloop")
	}
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
LOOP:
	for {
		select {
		case data := <-in:
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
		default:
			break LOOP
		}
	}
	wg.Wait()
}

// считает конкатенацию в порядке расчета значений crc32(th+data)), где th=0..5
func MultiHash(in, out chan interface{}) {
	//fmt.Println("MultiHash")
	wg := new(sync.WaitGroup)
LOOP:
	for {
		select {
		case data := <-in:
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
							// if i < 5 {
							// 	res += "_"
							// }
						}
						out <- res
						fmt.Println("MultiHash RES", res)
						break
					}
				}
			}(out, crc32Out, wg)
		default:
			break LOOP
		}
	}
	wg.Wait()
}

// получает все результаты, сортирует, объединяет отсортированный
// результат через _ (символ подчеркивания) в одну строку
func CombineResults(in, out chan interface{}) {
	// fmt.Println("CombineResults")
	ress := make([]string, 0, MaxInputDataLen)
LOOP:
	for {
		select {
		case data := <-in:
			dataStr, ok := data.(string)
			if !ok {
				panic("Пришёл не тот тип")
			}
			ress = append(ress, dataStr)
		default:
			break LOOP
		}
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
}
