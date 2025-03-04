package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"
)

//const filePath string = "./data/users.txt"

// вам надо написать более быструю оптимальную этой функции
func FastSearch(out io.Writer) {

	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	r := regexp.MustCompile("@")
	seenBrowsers := make(map[string]bool)
	uniqueBrowsers := 0
	foundUsers := ""

	var userPool = sync.Pool{
		New: func() any {
			return make(map[string]any)
		},
	}

	checkExists := func(browser string, seenBrowsers map[string]bool, uniqueBrowsers *int) {
		if !seenBrowsers[browser] {
			// log.Printf("SLOW New browser: %s, first seen: %s", browser, user["name"])
			seenBrowsers[browser] = true
			*uniqueBrowsers++
		}
	}

	for i := 0; scanner.Scan(); i++ {
		line := scanner.Bytes()
		user := userPool.Get().(map[string]any)
		if err := scanner.Err(); err != nil {
			fmt.Println("Ошибка при сканировании:", err)
			return
		}
		// fmt.Printf("%v %v\n", err, line)
		err := json.Unmarshal(line, &user)
		if err != nil {
			panic(err)
		}

		isAndroidUser := false
		isMSIEUser := false

		browsers, ok := user["browsers"].([]any)
		if !ok {
			log.Println("cant cast browsers")
			continue
		}

		for _, browser := range browsers {
			browser, ok := browser.(string)
			if !ok {
				log.Println("cant cast browsers")
				continue
			}
			if strings.Contains(browser, "Android") {
				isAndroidUser = true
				checkExists(browser, seenBrowsers, &uniqueBrowsers)
			}
			if strings.Contains(browser, "MSIE") {
				isMSIEUser = true
				checkExists(browser, seenBrowsers, &uniqueBrowsers)
			}
		}
		if !(isAndroidUser && isMSIEUser) {
			continue
		}

		// log.Println("Android and MSIE user:", user["name"], user["email"])
		email := r.ReplaceAllString(user["email"].(string), " [at] ")
		foundUsers += fmt.Sprintf("[%d] %s <%s>\n", i, user["name"], email)
	}

	fmt.Fprintln(out, "found users:\n"+foundUsers)
	fmt.Fprintln(out, "Total unique browsers", len(seenBrowsers))
}
