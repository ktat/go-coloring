package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ktat/kolorit"
	"github.com/mitchellh/go-homedir"
	toml "github.com/pelletier/go-toml"
)

var isDebug bool

func errCheck(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func errMessage(e string) {
	fmt.Println("Error: " + e + "\n")
	flag.Usage()
	os.Exit(1)
}

func readStdin(in chan string) {
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		var s = scanner.Text()
		in <- s
	}
	close(in)
}

func main() {
	pattern, files, fileName, dirName, erasePattern, options := parseOptions()

	re, regexpErr := regexp.Compile(pattern)
	errCheck(regexpErr)

	reErase, eraseRegexpErr := regexp.Compile(erasePattern)
	errCheck(eraseRegexpErr)

	var ioerr error

	if fileName == "-" {
		// read from STDIN
		if options["s"] {
			whole, ioerr := ioutil.ReadAll(os.Stdin)
			errCheck(ioerr)
			fmt.Println(coloringText(re, reErase, string(whole), options["B"]))
		} else {
			in := make(chan string)
			go readStdin(in)
			// read from STDIN with channel
			for {
				l, ok := <-in
				if ok == false {
					break
				} else {
					fmt.Println(coloringText(re, reErase, l, options["B"]))
				}
			}
		}
	} else {
		// read from file or dir
		if dirName != "" {
			seekDir(&files, dirName, fileName, options["R"])
		} else {
			files = append(files, fileName)
		}
		if isDebug {
			log.Println("### read from file or dir in main")
			log.Println("File Name: " + fileName)
			log.Println("Files: " + strings.Join(files, ", "))
			log.Println("Dir Name: " + dirName)
		}

		if options["s"] {
			var whole []byte

			for i := 0; i < len(files); i++ {
				whole, ioerr = ioutil.ReadFile(files[i])
				errCheck(ioerr)
				colored := coloringText(re, reErase, string(whole), options["B"])

				printColored(colored, files, i)
			}
		} else {
			for i := 0; i < len(files); i++ {
				var fp *os.File
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

					colored := coloringText(re, reErase, string(line), options["B"])
					if !options["grep"] || colored != string(line) {
						printColored(colored, files, i)
					}
				}
				ioerr = fp.Close()
				errCheck(ioerr)
			}
		}
	}
	os.Exit(0)
}

func printColored(colored string, files []string, i int) {
	if len(files) == 1 {
		fmt.Print(colored)
	} else {
		fmt.Print(addFileName(colored, files[i]))
	}
}

func seekDir(files *[]string, dirName string, fileName string, isRecursive bool) {
	if isDebug {
		log.Println("### seekDir")
		log.Println("File Name:" + fileName)
		log.Println("Dir Name:" + dirName)
	}
	fileInfo, ioerr := ioutil.ReadDir(dirName)
	errCheck(ioerr)
	for i := 0; i < len(fileInfo); i++ {
		fullName := filepath.Join(dirName, fileInfo[i].Name())
		if fileInfo[i].IsDir() == false {
			if fileName == "" || checkFileName(fullName, fileName) {
				if isDebug {
					log.Println("File Full Name: " + fullName)
				}
				*files = append(*files, fullName)
			}
		} else if isRecursive && fileInfo[i].Name()[0] != '.' {
			if isDebug {
				log.Println("Seek Dir: " + filepath.Join(dirName, fileInfo[i].Name()))
			}
			seekDir(files, filepath.Join(dirName, fileInfo[i].Name()), fileName, true)
		}
	}
}

func addFileName(content string, fileName string) string {
	var r = regexp.MustCompile("(?m)^(\\033\\[0m)?")
	return r.ReplaceAllString(content, "$1"+fileName+":") + "\n"
}

func checkFileName(targetFile string, fileName string) bool {
	pattern := fileName
	pattern = strings.Replace(pattern, ".", "\\.", -1)
	pattern = strings.Replace(pattern, "*", ".*", -1)
	matched, err := regexp.MatchString("(^|/)"+pattern+"$", targetFile)
	if isDebug {
		log.Println("### checkFileName")
		log.Println("Target File: " + targetFile)
		log.Println("File Name: " + fileName)
		log.Println("Pattern: " + pattern)
		log.Printf("Matched: %t\n", matched)
	}
	if err == nil && matched {
		return true
	}
	return false
}

type optDef struct {
	isBool   bool
	isString bool
	boolDef  bool
	strDef   string
	help     string
}

func parseOptions() (pattern string, files []string, fileName string, dirName string, erasePattern string, options map[string]bool) {
	replace := make([]string, 0)
	regexpFlg := ""
	options = make(map[string]bool)
	regexpFlgs := make(map[byte]bool)
	strOptions := make(map[string]string)
	colorOptions := make([]string, 0)
	colorHelp := make([]string, 0)
	boolParsedOpt := make(map[string]*bool)
	strParsedOpt := make(map[string]*string)
	regexps := make(map[byte]*string)
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

	homedir, err := homedir.Dir()
	if err != nil {
		errCheck(err)
	} else {
		homedir += string(os.PathSeparator)
	}

	for k := range colorMap {
		regexps[k] = flag.String(string(k), "", "regexp to be "+colorMap[k])
	}

	opt := map[string]optDef{
		"help": optDef{isBool: true, boolDef: false, help: "show usage"},
		"h":    optDef{isBool: true, boolDef: false, help: "show usage"},
		"d":    optDef{isBool: true, boolDef: false, help: "debug mode"},
		"grep": optDef{isBool: true, boolDef: false, help: "take string and ignore not matched lines with it like grep"},
		"s":    optDef{isBool: true, boolDef: false, help: "regexp option. tread given content as single line(default as multi line)"},
		"i":    optDef{isBool: true, boolDef: false, help: "regexp option. do case insensitive pattern matching."},
		"R":    optDef{isString: true, strDef: "", help: "recursively read given directory. using this option withaout -f, -f is set as '*.*'"},
		"f":    optDef{isString: true, strDef: "-", help: "file_name/pattern ... read from file. read stdin if not give."},
		"e":    optDef{isString: true, strDef: "", help: "erase matched string"},
		"B":    optDef{isBool: true, boolDef: false, help: "matched string to be bold"},
		"use":  optDef{isString: true, strDef: "", help: "use predefined setting from config file($HOME/.kolorit.toml)"},
		"conf": optDef{isString: true, strDef: homedir + ".kolorit.toml", help: "path of config file"},
	}

	for k, v := range opt {
		if v.isBool {
			boolParsedOpt[k] = flag.Bool(k, v.boolDef, v.help)
		} else if v.isString {
			strParsedOpt[k] = flag.String(k, v.strDef, v.help)
		}
	}

	flag.Parse()

	// parse options
	for k, v := range boolParsedOpt {
		options[k] = *v
	}
	for k, v := range strParsedOpt {
		strOptions[k] = *v
	}

	isDebug = options["d"]

	// print usage and exit
	if options["help"] || options["h"] {
		flag.Usage()
		os.Exit(1)
	}

	if strOptions["use"] != "" {
		parseConfig(strOptions["conf"], strOptions["use"], colorMap, &regexps, &options)
	}

	if strOptions["R"] != "" {
		options["R"] = true
	}

	fileName = strOptions["f"]
	erasePattern = strOptions["e"]
	dirName = strOptions["R"]

	if isDebug {
		log.Println("### parseOptions")
	}

	// rest args after options are regareded as files
	for n := 0; n < flag.NArg(); n++ {
		if isDebug {
			log.Println("Add File: " + flag.Arg(n))
		}
		files = append(files, flag.Arg(n))
	}

	if len(files) != 0 && dirName != "" && fileName == "-" {
		fileName = "*.*"
	}

	// buld regexp flags
	for _, k := range []byte{'s', 'i'} {
		regexpFlgs[k] = options[string(k)]
	}
	if !regexpFlgs['s'] {
		regexpFlgs['m'] = true
	}
	for k, v := range regexpFlgs {
		if v {
			regexpFlg += string(k)
		}
	}
	regexpFlg = "(?" + regexpFlg + ")"

	// build regexps
	for k := range colorMap {
		if *regexps[k] != "" {
			replace = append(replace, fmt.Sprintf("(?P<%s>%s)", colorMap[k], *regexps[k]))
			colorOptions = append(colorOptions, string(k))
		}
		colorHelp = append(colorHelp, "-"+string(k))
	}

	if len(replace) == 0 {
		errMessage("any of " + strings.Join(colorHelp, ", ") + " AND -R, -f or file names as rest of args is required.\n")
	}

	// assemble regexps
	pattern = regexpFlg + strings.Join(replace, "|")
	if isDebug {
		log.Println("regexp: " + pattern)
	}
	return
}

func parseConfig(configFile string, use string, colorMap map[byte]string, regexps *map[byte]*string, options *map[string]bool) {
	_, err := os.Stat(configFile)
	if err != nil {
		errMessage(configFile + " doesn't exist.")
	}

	config, err := toml.LoadFile(configFile)
	if err != nil {
		errCheck(err)
	}
	opt := config.Get(use)
	switch opt.(type) {
	case nil:
		errMessage("'" + use + "' is not defined in " + configFile)
	}

	for k := range colorMap {
		isRegexp := opt.(*toml.TomlTree).Get(string(k))
		switch isRegexp.(type) {
		case nil:
			continue
		default:
			regexpStr := isRegexp.(string)
			if regexpStr != "" {
				(*regexps)[k] = &regexpStr
			}
		}
	}
	boolOpts := make([]string, 0)
	boolOpts = append(boolOpts, "B", "m", "i", "s")
	for _, k := range boolOpts {
		boolOpt := opt.(*toml.TomlTree).Get(k)
		switch boolOpt.(type) {
		case nil:
			continue
		default:
			(*options)[k] = boolOpt.(bool)
		}
	}
}

func coloringText(re *regexp.Regexp, reErase *regexp.Regexp, lines string, withBold bool) string {
	lines = reErase.ReplaceAllString(lines, "")
	colorFunc := map[string]func(s kolorit.String) string{
		"red":    func(s kolorit.String) string { return s.Red() },
		"green":  func(s kolorit.String) string { return s.Green() },
		"blue":   func(s kolorit.String) string { return s.Blue() },
		"yellow": func(s kolorit.String) string { return s.Yellow() },
		"white":  func(s kolorit.String) string { return s.White() },
		"cyan":   func(s kolorit.String) string { return s.Cyan() },
		"black":  func(s kolorit.String) string { return s.Black() },
		"purple": func(s kolorit.String) string { return s.Magenta() },
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
					var matchedIndex []int
					matchedIndex = append(matchedIndex, result[k][i-1], result[k][i])
					var color kolorit.String
					color.WithBold(withBold)

					if matchedIndex[1] > 0 {
						color.Str = s[matchedIndex[0]:matchedIndex[1]]
					}
					if matchedIndex[0] > 0 {
						newStr = s[0:matchedIndex[0]]
					}
					newStr += colorFunc[k](color)
					if matchedIndex[1] > 0 && matchedIndex[1] < len(s) {
						newStr += s[matchedIndex[1]:len(s)]
					}
					s = newStr
				}
			}
		}
		return s
	})
	return string(lines)
}
