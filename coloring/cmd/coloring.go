package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/ktat/go-coloring/coloring"
	"github.com/ktat/go-pager"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var isDebug bool
var usePager bool

func errCheck(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func readStdin(in chan string, usePager bool) {
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		var s = scanner.Text()
		in <- s
	}

	if !usePager {
		close(in)
	}
}

func main() {

	pattern, files, fileName, dirName, erasePattern, usePager, asGrep := parseOptions()

	re, regexpErr := regexp.Compile(pattern)
	errCheck(regexpErr)

	reErase, eraseRegexpErr := regexp.Compile(erasePattern)
	errCheck(eraseRegexpErr)
	var (
		ioerr error
	)
	if dirName == "" && fileName == "" && len(files) == 0 {
		fmt.Println("-R or -f is requreid")
		os.Exit(1)
	} else if fileName == "-" {
		// read from STDIN with channel
		stat, _ := os.Stdin.Stat()

		if usePager {
			if (stat.Mode() & os.ModeCharDevice) != 0 {
				fmt.Fprintln(os.Stderr, "pager doesn't work without pipe.")
				os.Exit(1)
			}

			in := make(chan string)
			go readStdin(in, usePager)

			var p pager.Pager
			p.Init()

			pollEnd := make(chan int)
			go func(in chan string, pe chan int) {
				if p.PollEvent() == false {
					p.Close()
					if !usePager {
						close(in)
					}
					pe <- 1
					return
				}
			}(in, pollEnd)

			go func() {
				for {
					l, ok := <-in
					if ok == false {
						break
					} else {
						p.AddContent(coloringText(re, reErase, l+"\n"))
						p.Draw()
					}

				}
			}()

			<-pollEnd

			close(pollEnd)
			defer p.Close()
		} else {
			in := make(chan string)
			go readStdin(in, usePager)

			for {
				l, ok := <-in
				if ok == false {
					break
				} else {
					fmt.Println(coloringText(re, reErase, l))
				}
			}
		}
	} else {
		// read from file or dir
		var isRecursive bool = true
		if len(files) == 0 {
			if fileName == "" {
				fileName = "*.*"
			}
			dirName = "."
		}
		if dirName != "" {
			seekDir(&files, dirName, fileName, isRecursive)
		}
		if isDebug {
			log.Println("filename: " + fileName)
			log.Println("files:")
			log.Println(files)
			log.Println("dirName: " + dirName)
		}
		var p pager.Pager
		if usePager {
			p.Init()
			p.Files = files
		}

		var fp *os.File
		for i := 0; i < len(files); i++ {
			fp, ioerr = os.Open(files[i])
			errCheck(ioerr)
			reader := bufio.NewReaderSize(fp, 4096)
			for {
				line, _, ioerr := reader.ReadLine()
				if ioerr != io.EOF {
					errCheck(ioerr)
				} else if ioerr == io.EOF {
					break
				}

				colored := coloringText(re, reErase, string(line))
				if !asGrep || colored != string(line) {
					if usePager {
						p.AddContent(colored + "\n")
					} else {
						fmt.Println(colored)
					}
				}
			}
			if usePager {
				p.Index = i
				p.File = files[i]
				if p.PollEvent() {
					i = p.Index
				} else {
					break
				}
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
		log.Println(fileName)
	}
	fileInfo, ioerr := ioutil.ReadDir(dirName)
	errCheck(ioerr)
	for i := 0; i < len(fileInfo); i++ {
		fullName := filepath.Join(dirName, fileInfo[i].Name())
		if fileInfo[i].IsDir() == false {
			if fileName == "" || checkFileName(fullName, fileName) {
				if isDebug {
					log.Println(fullName)
				}
				*files = append(*files, fullName)
			}
		} else if isRecursive && fileInfo[i].Name()[0] != '.' {
			if isDebug {
				log.Println("seek dir")
			}
			seekDir(files, filepath.Join(dirName, fileInfo[i].Name()), fileName, true)
		}
	}
}

func checkFileName(targetFile string, fileName string) bool {
	pattern := fileName
	pattern = strings.Replace(pattern, ".", "\\.", -1)
	pattern = strings.Replace(pattern, "*", ".*", -1)
	matched, err := regexp.MatchString("(^|/)"+pattern+"$", targetFile)
	if isDebug {
		log.Println(targetFile, fileName, pattern, matched)
	}
	if err == nil && matched {
		return true
	}
	return false
}

func parseOptions() (pattern string, files []string, fileName string, dirName string, erasePattern string, usePager bool, asGrep bool) {
	replace := make([]string, 0)

	regexpFlg := ""
	regexpFlgs := make(map[byte]*bool)

	colorMap := map[byte]string{
		'r': "red",
		'g': "green",
		'b': "blue",
		'y': "yellow",
		'p': "purple",
		'c': "cyan",
		'k': "black",
		'w': "white",
	}
	colorOptions := make([]string, 0)
	colorHelp := make([]string, 0)

	regexps := make(map[byte]*string)
	for k := range colorMap {
		regexps[k] = flag.String(string(k), "", "to be "+colorMap[k])
	}
	help := flag.Bool("h", false, "show usage")
	regexpFlgs['m'] = flag.Bool("m", false, "regexp for multilines")
	regexpFlgs['i'] = flag.Bool("i", false, "regexp for case insensitive")
	usePagerP := flag.Bool("P", false, "use bultin pager")
	isDebugP := flag.Bool("d", false, "debug mode")
	asGrepP := flag.Bool("grep", false, "take string and ignore not matched lines with it like grep")
	dirNameP := flag.String("R", "", "file_name/pattern/-(stdin) ... read from file. read stdin if '-' is given")
	fileNameP := flag.String("f", "", "recursively read directory")
	erasePatternP := flag.String("e", "", "erase matched string")

	flag.Parse()

	isDebug = *isDebugP
	usePager = *usePagerP
	asGrep = *asGrepP
	dirName = *dirNameP
	fileName = *fileNameP
	erasePattern = *erasePatternP

	if *help {
		flag.Usage()
	}
	if *regexpFlgs['m'] {
		delete(regexpFlgs, 's')
	}

	for k := range colorMap {
		if *regexps[k] != "" {
			replace = append(replace, fmt.Sprintf("(?P<%s>%s)", colorMap[k], *regexps[k]))
			colorOptions = append(colorOptions, string(k))
			colorHelp = append(colorHelp, "-"+string(k))
		}
	}

	if !*regexpFlgs['m'] {
		var tx = true
		regexpFlgs['s'] = &tx
	}

	for n := 0; n < flag.NArg(); n++ {
		if isDebug {
			log.Println("add file: " + flag.Arg(n))
		}
		files = append(files, flag.Arg(n))
	}
	if len(replace) == 0 {
		println(123)
		println("any of " + strings.Join(colorHelp, ", ") + " AND -R or -f is required.")
		os.Exit(1)
	}

	for k, v := range regexpFlgs {
		if *v {
			regexpFlg += string(k)
		}
	}
	regexpFlg = "(?" + regexpFlg + ")"

	pattern = regexpFlg + strings.Join(replace, "|")
	if isDebug {
		log.Println("regexp: " + pattern)
		log.Println("dirName: " + dirName)
	}
	return
}

func coloringText(re *regexp.Regexp, reErase *regexp.Regexp, lines string) string {
	lines = reErase.ReplaceAllString(lines, "")
	colorFunc := map[string]func(s coloring.String) string{
		"red":    func(s coloring.String) string { return s.Red() },
		"green":  func(s coloring.String) string { return s.Green() },
		"blue":   func(s coloring.String) string { return s.Blue() },
		"yellow": func(s coloring.String) string { return s.Yellow() },
		"white":  func(s coloring.String) string { return s.White() },
		"cyan":   func(s coloring.String) string { return s.Cyan() },
		"black":  func(s coloring.String) string { return s.Black() },
		"purple": func(s coloring.String) string { return s.Magenta() },
	}

	lines = re.ReplaceAllStringFunc(lines, func(s string) string {
		result := make(map[string][]int)
		match := re.FindAllStringSubmatchIndex(s, -1)
		lastName := ""
		for i, name := range re.SubexpNames() {
			if i < 1 || match[0][i*2] == -1 {
				continue
			}
			if lastName != "" && name == "" {
				result[lastName] = append(result[lastName], match[0][i*2], match[0][i*2+1])
			} else {
				result[name] = append(result[name], match[0][i*2], match[0][i*2+1])
				lastName = name
			}
		}

		for k := range colorFunc {
			newStr := ""
			if len(result[k]) > 2 { // if parenthese exists in regexp, ignore first match which matches whole string
				result[k] = result[k][2:]
			}
			for i := len(result[k]) - 1; i >= 0; i -= 2 {
				if result[k][i] > 0 {
					var matched_index []int
					matched_index = append(matched_index, result[k][i-1], result[k][i])
					var color coloring.String

					if matched_index[1] > 0 {
						color.Str = s[matched_index[0]:matched_index[1]]
					}
					if matched_index[0] > 0 {
						newStr = s[0:matched_index[0]]
					}
					newStr += colorFunc[k](color)
					if matched_index[1] > 0 && matched_index[1] < len(s) {
						newStr += s[matched_index[1]:len(s)]
					}
					s = newStr
				}
			}
		}
		return s
	})
	return string(lines)
}
