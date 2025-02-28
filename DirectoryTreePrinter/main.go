package main

import (
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
)

const (
	enter         = "\n"
	tab           = "\t"
	separator     = "│\t" // \t
	lastBranch    = "└───"
	anotherBranch = "├───"
)

func walker(out io.Writer, startDir string, currentLevel int,
	levels []bool, printFiles bool) error {
	file, fo_err := os.Open(startDir)
	if fo_err != nil {
		fmt.Println("Ошибка открытия ", fo_err)
		return fo_err
	}
	defer file.Close()

	infos, err := file.Readdir(0) // file.Readdirnames(0)
	if err != nil && err != io.EOF {
		fmt.Println("Ошибка чтения ", err)
		return err
	}

	slices.SortFunc(infos, func(a, b os.FileInfo) int {
		return strings.Compare(a.Name(), b.Name())
	})

	currentLevel++
	levels = append(levels, false)

	filteredInfos := make([]os.FileInfo, 0)
	for _, inf := range infos {
		if inf.IsDir() || printFiles {
			filteredInfos = append(filteredInfos, inf)
		}
	}

	for i, inf := range filteredInfos {
		for level, closed := range levels {
			if closed {
				fmt.Fprint(out, tab)
			} else {
				if level == currentLevel {
					if i == len(filteredInfos)-1 {
						fmt.Fprint(out, lastBranch)
						levels[currentLevel] = true
					} else {
						fmt.Fprint(out, anotherBranch)
					}
				} else {
					fmt.Fprint(out, separator)
				}

			}
		}
		name := inf.Name()
		isFile := !inf.IsDir()
		if isFile {
			size := inf.Size()
			if size > 0 {
				fmt.Fprintf(out, "%s (%db)\n", name, size)
			} else {
				fmt.Fprintf(out, "%s (empty)\n", name)
			}
		} else {
			fmt.Fprint(out, name+enter) // +" "+fmt.Sprint(levels)+" "
			err := walker(out, startDir+string(os.PathSeparator)+name, currentLevel, levels, printFiles)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// map[name]info
func dirTree(out io.Writer, startDir string, printFiles bool) error {

	levels := make([]bool, 0, 6)
	err := walker(out, startDir, -1, levels, printFiles)
	if err != nil {
		return err
	} else {
		return nil
	}
}

// go run main.go .\testdata\ -f
func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}
