package main

import (
	"fmt"
	"github.com/ktat/go-coloring/coloring"
	"github.com/ktat/go-pager"
	. "github.com/mattn/go-getopt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
	"path/filepath"
)

var isDebug bool
var usePager bool

func usage() {
	const v = `
usage: coloring [-f file|-[rgbycpwk] regexp|-n pattern|-R dir|-h]

        -f file_name/pattern ... read from file instead of stdin
        -R dir  ... recursively read directory
        -r regexp ... to be red
        -g regexp ... to be green
        -b regexp ... to be blue
        -y regexp ... to be yellow
        -c regexp ... to be cyan
        -p regexp ... to be purple
        -w regexp ... to be white
        -k regexp ... to be black
        -m ... regexp for multiline
        -i ... regexp is case insensitive
        -P ... use builtin pager
        -h ... help
`
	os.Stderr.Write([]byte(v))
	os.Exit(1)
}

func errCheck(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func main() {

	pattern, fileName, dirName := parseOptions()

	re, regexpErr := regexp.Compile(pattern)
	errCheck(regexpErr)

	var (
		whole []byte
		ioerr error
	)

	if dirName == "" && fileName == "" {
		// read from STDIN
		whole, ioerr = ioutil.ReadAll(os.Stdin)
		errCheck(ioerr)
		if usePager {
			for {
				var p pager.Pager
				p.Init()
				p.SetContent(coloringText(re, string(whole)))
				if p.PollEvent() == false {
					p.Close()
					break
				}
			}
		} else {
			fmt.Println(coloringText(re, string(whole)))
		}
	} else {
		var files = make([]string, 0)
		var isRecursve bool = false
		if dirName == "" {
			dirName = "."
		} else {
			isRecursve = true
		}
		if fileName == "" {
			fileName = "*.*"
		}
		seekDir(&files, dirName, fileName, isRecursve)
		if isDebug {
			fmt.Println(files)
			fmt.Println(isRecursve)
		}
		var p pager.Pager
		if usePager {
			p.Init()
			p.Files = files
		}

		for i := 0; i < len(files); i++ {
			whole, ioerr = ioutil.ReadFile(files[i])
			errCheck(ioerr)
			colored := coloringText(re, string(whole))

			if usePager {
				p.Index = i
				p.SetContent(colored)
				p.File = files[i]
				if p.PollEvent() {
					i = p.Index
				} else {
					break
				}
			} else {
				fmt.Println(colored)
			}
		}
		if usePager {
			p.Close()
		}
	}
	os.Exit(0)
}

func seekDir(files *[]string, dirName string, fileName string, isRecursive bool) {
	if isDebug {
		println(fileName)
	}
	fileInfo, ioerr := ioutil.ReadDir(dirName)
	errCheck(ioerr)
	for i := 0; i < len(fileInfo); i++ {
		fullName := filepath.Join(dirName, fileInfo[i].Name())
		if fileInfo[i].IsDir() == false {
			if fileName == "" || checkFileName(fullName, fileName) {
				*files = append(*files, fullName)
			}
		} else if isRecursive && fileInfo[i].Name()[0] != '.' {
			println("seek dir")
			seekDir(files, filepath.Join(dirName,fileInfo[i].Name()), fileName, true)
		}
	}
}

func checkFileName(targetFile string, fileName string) bool {
	pattern := fileName
	pattern = strings.Replace(pattern, ".", "\\.", -1)
	pattern = strings.Replace(pattern, "*", ".*", -1)
	matched, err := regexp.MatchString("/"+pattern+"$", targetFile)
	if isDebug {
		println(targetFile, fileName, pattern, matched)
	}
	if err == nil && matched {
		return true
	}
	return false
}

func parseOptions() (pattern string, fileName string, dirName string) {
	replace := make([]string, 0)
	var (
		c int
	)
	regexpFlg := ""
	regexpFlgs := make(map[byte]bool)
	regexpFlgs['s'] = true

	colorMap := map[byte]string{
		'r': "red",
		'g': "green",
		'b': "blue",
		'y': "yellow",
		'p': "pink",
		'c': "cyan",
		'k': "black",
		'w': "white",
	}

	options := "imdhPR:f:n:"
	colorOptions := make([]string, 0)
	colorHelp := make([]string, 0)
	for k := range colorMap {
		colorOptions = append(colorOptions, string(k))
		colorHelp = append(colorHelp, "-"+string(k))
	}

	for {
		if c = Getopt(options + strings.Join(colorOptions, ":") + ":"); c == EOF {
			break
		}

		switch c {
		case 'h':
			usage()
		case 'f':
			fileName = OptArg
		case 'R':
			dirName = OptArg
		case 'P':
			usePager = true
		case 'd':
			isDebug = true
		case 'm':
			regexpFlgs['m'] = true
			delete(regexpFlgs, 's')
		case 'i':
			regexpFlgs['i'] = true
		default:
			if color, ok := colorMap[byte(c)]; ok {
				replace = append(replace, fmt.Sprintf("(?P<%s>%s)", color, OptArg))
			} else {
				os.Exit(1)
			}
		}
	}
	OptErr = 0

	if len(replace) == 0 {
		println("any of " + strings.Join(colorHelp, ", ") + " is required.")
		os.Exit(1)
	}

	for k, v := range regexpFlgs {
		if v {
			regexpFlg += string(k)
		}
	}
	regexpFlg = "(?" + regexpFlg + ")"

	pattern = regexpFlg + strings.Join(replace, "|")
	if isDebug {
		fmt.Println("regexp: " + pattern)
		fmt.Println("dirName: " + dirName)
	}
	return
}

func coloringText(re *regexp.Regexp, lines string) string {
	colorFunc := map[string]interface{}{
		"red":    func(s coloring.String) string { return s.Red() },
		"green":  func(s coloring.String) string { return s.Green() },
		"blue":   func(s coloring.String) string { return s.Blue() },
		"yellow": func(s coloring.String) string { return s.Yellow() },
		"white":  func(s coloring.String) string { return s.White() },
		"cyan":   func(s coloring.String) string { return s.Cyan() },
		"black":  func(s coloring.String) string { return s.Black() },
		"purple": func(s coloring.String) string { return s.Magenta() },
	}

	// should be improved
	lines = re.ReplaceAllStringFunc(lines, func(s string) string {
		result := make(map[string]string)
		match := re.FindStringSubmatch(s)

		for i, name := range re.SubexpNames() {
			result[name] = match[i]
		}

		for k := range colorFunc {
			if len(result[k]) > 0 {
				var color coloring.String
				color.Str = s
				return colorFunc[k].(func(coloring.String) string)(color)
			}
		}
		// never come here
		return s
	})
	return string(lines)
}
