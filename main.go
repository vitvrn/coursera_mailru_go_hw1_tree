package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	//	"strings"
)

//TODO [filepath] func Walk(root string, walkFn WalkFunc) error
//TODO use strings module ???

//func panicOnErr(err error) {
//	if err != nil {
//		panic(err)
//	}
//}

func prefix(lvl int, areLast map[int]bool) (result string) {
	const ( //move to global scope?
		last0_currLvl = "├───" //TODO ??? use map[bool,bool]string or other map
		last0_prevLvl = "│\t"
		last1_currLvl = "└───"
		last1_prevLvl = "\t"
	)
	for l := 1; l < lvl; l++ { //drawing one of prevLvl
		if !areLast[l] {
			result += last0_prevLvl
		} else {
			result += last1_prevLvl
		}
	}
	if !areLast[lvl] {
		result += last0_currLvl
	} else {
		result += last1_currLvl
	}
	return
}

func fileInfo(out io.Writer, f *os.File, lvl int, areLast map[int]bool) (err error) {
	if lvl == 0 {
		return //err ? "корневой элемент не выводим"
	}
	finfo, err := f.Stat()
	if err != nil {
		return
	}
	var sizeStr string
	if finfo.Size() == 0 {
		sizeStr = "empty"
	} else {
		sizeStr = fmt.Sprintf("%db", finfo.Size())
	}
	_, err = out.Write([]byte(fmt.Sprintf("%s%s (%s)\n", prefix(lvl, areLast), filepath.Base(f.Name()), sizeStr)))
	return
}

func dirInfo(out io.Writer, f *os.File, lvl int, areLast map[int]bool) (err error) {
	if lvl == 0 {
		return //err ? "корневой элемент не выводим"
	}
	_, err = out.Write([]byte(fmt.Sprintf("%s%s\n", prefix(lvl, areLast), filepath.Base(f.Name()))))
	return
}

//func dirTree(out *os.File, path string, printFiles bool) error {
func dirTree(out io.Writer, path string, printFiles bool) (err error) {
	// уровень вложенности каталогов относительно точки входа, определяет:
	//   1) префикс (отступы, символы графики);
	//   2) lvl==0 -> fileInfo вообще не печатаем!
	var lvl int                                                               //= 0                                                      //= false
	var areLast map[int]bool = make(map[int]bool)                             //остались ли еще файлы для обработки на [уровнях_вложенности]
	var dirTreeR func(path string, lvl int, areLast map[int]bool) (err error) //by-ptr *areLast - не обязательно ???

	//==========================================================================
	dirTreeR = func(path string, lvl int, areLast map[int]bool) (err error) {
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		stat, err := f.Stat()
		if err != nil {
			return err
		}

		if stat.IsDir() {
			err = dirInfo(out, f, lvl, areLast)
			if err != nil {
				return err
			}
			// read-dir
			var fnames sort.StringSlice
			//fnames, err = f.Readdirnames(0)
			finfos, err := f.Readdir(0)
			if err != nil {
				return err
			}
			for _, finfo := range finfos { //building []string of names
				if printFiles || finfo.IsDir() { //dirs are to add anyway, if printFiles then add all
					fnames = append(fnames, finfo.Name())
				}
			}
			// sort-output
			fnames.Sort()
			// recurse(dirTreeR)
			for i, name := range fnames {
				if isLastName := i == len(fnames)-1; isLastName {
					areLast[lvl+1] = true //= isLastName //??? areLast[lvl]
					//fmt.Println("+++", name, areLast, i, len(fnames))
				}
				err = dirTreeR(path+string(os.PathSeparator)+name, lvl+1, areLast) //if err!=nil { TODO? }
			}
			//update areLast for higher (than current dir's) nesting levels:
			for key, _ := range areLast {
				if key > lvl {
					areLast[key] = false //OR delete?
				}
			}

		} else /* if printFiles*/ { //"if printFiles" not needed - files already are filtered (and lvl=0 won't be printed anyway)
			err = fileInfo(out, f, lvl, areLast) //if err!=nil { nothing to do }
		}

		return err
	}
	//==========================================================================
	dirTreeR(path, lvl, areLast)

	return err
}

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
