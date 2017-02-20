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

	"github.com/ktat/go-ansistrings"
	"github.com/mitchellh/go-homedir"
	toml "github.com/pelletier/go-toml"
)

var isDebug bool

type kolor struct {
	strOptions   map[string]string
	options      map[string]bool
	bg           map[string]int
	pattern      string
	erasePattern string
	numOfRegexps int
	files        []string
	fileName     string
	isRecursive  bool
	fromSTDIN    bool
	asSingle     bool
}

func usage() {
	fmt.Println(`Usage:
	
  kolorit [options] [FILES]
  kolorit [options] -f "*.go"
  kolorit [options] -R [FILES/DIRECTORIES]

Options:
`)
	flag.PrintDefaults()
	os.Exit(1)
}

func errCheck(e error, m ...string) {
	if e != nil {
		if len(m) > 0 && m[0] != "" {
			fmt.Printf("%s\nmessage: ", m[0])
			fmt.Println(e)
		} else {
			fmt.Println(e)
		}
		os.Exit(1)
	}
}

func errMessage(e string) {
	fmt.Println("Error: " + e + "\n")
	usage()
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
	kolor := kolor{
		options:    make(map[string]bool),
		strOptions: make(map[string]string),
		bg:         make(map[string]int),
		files:      make([]string, 0),
	}
	kolor.parseOptions()

	re, regexpErr := regexp.Compile(kolor.pattern)
	errCheck(regexpErr, "wrong regexp: "+kolor.pattern)

	reErase, eraseRegexpErr := regexp.Compile(kolor.erasePattern)
	errCheck(eraseRegexpErr, "wrong regexp: "+kolor.erasePattern)

	var ioerr error

	if kolor.fromSTDIN {
		// read from STDIN
		if kolor.asSingle {
			whole, ioerr := ioutil.ReadAll(os.Stdin)
			errCheck(ioerr, "error on reading STDIN")
			str, _ := kolor.coloringText(re, reErase, string(whole))
			fmt.Println(str)
		} else {
			in := make(chan string)
			go readStdin(in)
			// read from STDIN with channel
			for {
				l, ok := <-in
				if ok == false {
					break
				} else {
					colored, n := kolor.coloringText(re, reErase, l)
					if kolor.options["grep"] && (kolor.options["or"] || n == kolor.numOfRegexps) && colored != string(l) {
						fmt.Println(colored)
					} else if !kolor.options["grep"] {
						fmt.Println(colored)
					}
				}
			}
		}
	} else {
		// read from file or dir
		if len(kolor.files) == 0 {
			errMessage("files are not given.")
		}
		if isDebug {
			log.Println("### read from file or dir in main")
			log.Println("File Name: " + kolor.fileName)
			log.Println("Files: " + strings.Join(kolor.files, ", "))
			log.Printf("Num of Files: %d\n", len(kolor.files))
			log.Printf("Is Recursive: %t\n", kolor.isRecursive)
		}

		if kolor.asSingle {
			var whole []byte

			for i := 0; i < len(kolor.files); i++ {
				fi, err := os.Stat(kolor.files[i])
				if err != nil {
					errCheck(err, "error on stat file: "+kolor.files[i])
				}
				if kolor.isRecursive && fi.IsDir() {
					continue
				}
				whole, ioerr = ioutil.ReadFile(kolor.files[i])
				errCheck(ioerr, "error on reading file: "+kolor.files[i])
				colored, _ := kolor.coloringText(re, reErase, string(whole))
				kolor.printColored(colored, i)
			}
		} else {
			for i := 0; i < len(kolor.files); i++ {
				fi, err := os.Stat(kolor.files[i])
				if err != nil {
					if isDebug {
						log.Println(kolor.files[i])
					}
					errCheck(err)
				}
				if kolor.isRecursive && fi.IsDir() {
					continue
				}
				var fp *os.File
				fp, ioerr = os.Open(kolor.files[i])
				errCheck(ioerr)
				reader := bufio.NewReaderSize(fp, 4096)
				for {
					line, _, ioerr := reader.ReadLine()
					if ioerr != io.EOF {
						errCheck(ioerr, "error on reading file: "+kolor.files[i])
					} else if ioerr == io.EOF {
						break
					}

					colored, n := kolor.coloringText(re, reErase, string(line))
					if kolor.options["grep"] && (kolor.options["or"] || n == kolor.numOfRegexps) && colored != string(line) {
						kolor.printColored(colored, i)
					} else if !kolor.options["grep"] {
						kolor.printColored(colored, i)
					}
				}
				ioerr = fp.Close()
				errCheck(ioerr, "error on closing file: "+kolor.files[i])
			}
		}
	}
	os.Exit(0)
}

func (kolor *kolor) printColored(colored string, i int) {
	if len(kolor.files) == 1 {
		fmt.Println(colored)
	} else {
		fmt.Print(addFileName(colored, kolor.files[i]))
	}
}

func (kolor *kolor) seekDir(files *[]string, dirName string) {
	if isDebug {
		log.Println("### seekDir")
		log.Println("File Name:" + kolor.fileName)
		log.Println("Dir Name:" + dirName)
	}
	fileInfo, ioerr := ioutil.ReadDir(dirName)
	errCheck(ioerr, "error on reading dir: "+dirName)
	for i := 0; i < len(fileInfo); i++ {
		fullName := filepath.Join(dirName, fileInfo[i].Name())
		if fileInfo[i].IsDir() == false {
			if kolor.fileName == "" || kolor.checkFileName(fullName) {
				if isDebug {
					log.Println("File Full Name: " + fullName)
				}
				*files = append(*files, fullName)
			}
		} else if kolor.isRecursive && fileInfo[i].Name()[0] != '.' {
			if isDebug {
				log.Println("Seek Dir: " + filepath.Join(dirName, fileInfo[i].Name()))
			}
			kolor.seekDir(files, filepath.Join(dirName, fileInfo[i].Name()))
		}
	}
}

func addFileName(content string, fn string) string {
	var r = regexp.MustCompile("(?m)^(\\033\\[0m)?")
	return r.ReplaceAllString(content, fn+":"+"$1") + "\n"
}

func (kolor *kolor) checkFileName(targetFile string) bool {
	kolor.pattern = kolor.fileName
	kolor.pattern = strings.Replace(kolor.pattern, ".", "\\.", -1)
	kolor.pattern = strings.Replace(kolor.pattern, "*", ".*", -1)
	matched, err := regexp.MatchString("(^|/)"+kolor.pattern+"$", targetFile)
	if isDebug {
		log.Println("### checkFileName")
		log.Println("Target File: " + targetFile)
		log.Println("File Name: " + kolor.fileName)
		log.Println("Pattern: " + kolor.pattern)
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

func (kolor *kolor) parseOptions() {
	replace := make([]string, 0)
	regexpFlg := ""
	regexpFlgs := make(map[byte]bool)
	colorOptions := make([]string, 0)
	colorHelp := make([]string, 0)
	boolParsedOpt := make(map[string]*bool)
	strParsedOpt := make(map[string]*string)
	regexps := make(map[string]*string)
	bgOptions := make(map[string]*string)
	colorMap := map[string]string{
		"r":   "red",
		"g":   "green",
		"b":   "blue",
		"y":   "yellow",
		"p":   "purple",
		"c":   "cyan",
		"k":   "black",
		"w":   "white",
		"lr":  "light_red",
		"lg":  "light_green",
		"lb":  "light_blue",
		"ly":  "light_yellow",
		"lp":  "light_purple",
		"lc":  "light_cyan",
		"dgr": "dark_gray",
		"lgr": "light gray",
	}

	homedir, err := homedir.Dir()
	if err != nil {
		log.Println(11)
		errCheck(err, "error on getting HOME dir")
	} else {
		homedir += string(os.PathSeparator)
	}

	for k := range colorMap {
		regexps[k] = flag.String(k, "", "regexp to be "+colorMap[k])
	}
	for k := range colorMap {
		bgOptions["b"+k] = flag.String("b"+k, "", "background color of "+colorMap[k])
	}

	opt := map[string]optDef{
		"help": optDef{isBool: true, boolDef: false, help: "show usage"},
		"h":    optDef{isBool: true, boolDef: false, help: "show usage"},
		"d":    optDef{isBool: true, boolDef: false, help: "debug mode"},
		"grep": optDef{isBool: true, boolDef: false, help: "take string and ignore not matched lines with it like grep. cannot use it with -s"},
		"s":    optDef{isBool: true, boolDef: false, help: "regexp option. tread given content as single line(default as multi line)"},
		"i":    optDef{isBool: true, boolDef: false, help: "regexp option. do case insensitive pattern matching."},
		"R":    optDef{isBool: true, boolDef: false, help: "recursively read directory."},
		"f":    optDef{isString: true, strDef: "", help: "file pattern. read from matched file."},
		"e":    optDef{isString: true, strDef: "", help: "erase matched string"},
		"B":    optDef{isBool: true, boolDef: false, help: "matched string to be bold"},
		"I":    optDef{isBool: true, boolDef: false, help: "matched string background color to be inverted"},
		"use":  optDef{isString: true, strDef: "", help: "use predefined setting from config file($HOME/.kolorit.toml)"},
		"conf": optDef{isString: true, strDef: homedir + ".kolorit.toml", help: "path of config file"},
		"or":   optDef{isBool: true, boolDef: false, help: "change grep option behavior. take string if any regexp is match."},
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
		if v != nil {
			kolor.options[k] = *v
		}
	}
	for k, v := range strParsedOpt {
		if v != nil {
			kolor.strOptions[k] = *v
		}
	}

	isDebug = kolor.options["d"]

	kolor.isRecursive = kolor.options["R"]

	// print usage and exit
	if kolor.options["help"] || kolor.options["h"] {
		usage()
	}

	// options from config file
	if kolor.strOptions["use"] != "" {
		kolor.parseConfig(kolor.strOptions["conf"], kolor.strOptions["use"], colorMap, &regexps)
	}

	kolor.erasePattern = kolor.strOptions["e"]
	kolor.asSingle = kolor.options["s"]

	if isDebug {
		log.Println("### parseOptions")
	}

	// rest args after options are regareded as files
	for n := 0; n < flag.NArg(); n++ {
		if isDebug {
			log.Println("Add File: " + flag.Arg(n))
		}
		kolor.files = append(kolor.files, flag.Arg(n))
	}

	if kolor.strOptions["f"] != "" && kolor.strOptions["f"] != "-" {
		kolor.fileName = kolor.strOptions["f"]
	} else if len(kolor.files) == 0 && !kolor.isRecursive {
		kolor.fromSTDIN = true
	}

	// collect target files
	if !kolor.fromSTDIN {
		if len(kolor.files) == 0 && kolor.fileName != "" {
			kolor.seekDir(&kolor.files, ".")
		} else if kolor.isRecursive {
			for _, f := range kolor.files {
				fi, err := os.Stat(f)
				errCheck(err, "error on stat file: "+f)
				if fi.IsDir() {
					kolor.seekDir(&kolor.files, f)
				}
			}
		}
		if len(kolor.files) == 0 {
			errMessage("files are not given/found")
		}
	}

	// build regexp flags
	for _, k := range []byte{'s', 'i'} {
		regexpFlgs[k] = kolor.options[string(k)]
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
			kolor.numOfRegexps++
		}
		colorHelp = append(colorHelp, "-"+string(k))
		v, ok := bgOptions["b"+k]
		if ok && *v != "" {
			kolor.bg[k], err = ansistrings.ColorNumFromName(*v)
			if err != nil {
				errCheck(err, "unknown color name: "+*v)
			}
		}
	}

	if len(replace) == 0 {
		errMessage("any of " + strings.Join(colorHelp, ", ") + " AND -R, -f or file names as rest of args is required.\n")
	}

	// assemble regexps
	kolor.pattern = regexpFlg + strings.Join(replace, "|")
	if isDebug {
		log.Println("regexp: " + kolor.pattern)
	}
}

func (kolor *kolor) parseConfig(configFile string, use string, colorMap map[string]string, regexps *map[string]*string) {
	_, err := os.Stat(configFile)
	if err != nil {
		errCheck(err, "cannot find/read config file:"+configFile)
	}

	config, err := toml.LoadFile(configFile)
	if err != nil {
		errCheck(err, "cannot parse config file: "+configFile)
	}
	opt := config.Get(use)
	switch opt.(type) {
	case nil:
		errMessage("'" + use + "' is not defined in " + configFile)
	}

	optionKeys := make([]string, 0)
	for k := range colorMap {
		optionKeys = append(optionKeys, k)
	}
	optionKeys = append(optionKeys, "e")

	for _, k := range optionKeys {
		isRegexp := opt.(*toml.TomlTree).Get(string(k))
		switch isRegexp.(type) {
		case nil:
			continue
		default:
			regexpStr := isRegexp.(string)
			if k == "e" {
				kolor.strOptions[string(k)] = regexpStr
			} else {
				if *(*regexps)[k] == "" && regexpStr != "" {
					(*regexps)[k] = &regexpStr
				}

			}
		}
	}
	boolOpts := []string{"B", "m", "i", "s", "I"}
	for _, k := range boolOpts {
		boolOpt := opt.(*toml.TomlTree).Get(k)
		switch boolOpt.(type) {
		case nil:
			continue
		default:
			kolor.options[k] = boolOpt.(bool)
		}
	}
}

func (kolor *kolor) coloringText(re *regexp.Regexp, reErase *regexp.Regexp, lines string) (string, int) {
	lines = reErase.ReplaceAllString(lines, "")
	colorFunc := map[string]func(s ansistrings.ANSIString) string{
		"red":          func(s ansistrings.ANSIString) string { s.Red(); return s.String() },
		"green":        func(s ansistrings.ANSIString) string { s.Green(); return s.String() },
		"blue":         func(s ansistrings.ANSIString) string { s.Blue(); return s.String() },
		"yellow":       func(s ansistrings.ANSIString) string { s.Yellow(); return s.String() },
		"white":        func(s ansistrings.ANSIString) string { s.White(); return s.String() },
		"cyan":         func(s ansistrings.ANSIString) string { s.Cyan(); return s.String() },
		"black":        func(s ansistrings.ANSIString) string { s.Black(); return s.String() },
		"purple":       func(s ansistrings.ANSIString) string { s.Magenta(); return s.String() },
		"light_purple": func(s ansistrings.ANSIString) string { s.LightMagenta(); return s.String() },
		"light_red":    func(s ansistrings.ANSIString) string { s.LightRed(); return s.String() },
		"light_green":  func(s ansistrings.ANSIString) string { s.LightGreen(); return s.String() },
		"light_blue":   func(s ansistrings.ANSIString) string { s.LightBlue(); return s.String() },
		"light_yellow": func(s ansistrings.ANSIString) string { s.LightYellow(); return s.String() },
		"light_cyan":   func(s ansistrings.ANSIString) string { s.LightCyan(); return s.String() },
		"dark_gray":    func(s ansistrings.ANSIString) string { s.DarkGray(); return s.String() },
		"light_gray":   func(s ansistrings.ANSIString) string { s.LightGray(); return s.String() },
	}

	machedKind := 0
	machedName := make(map[string]int)
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
				machedName[lastName]++
				if machedName[lastName] == 1 {
					machedKind++
				}
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
					var color ansistrings.ANSIString
					if kolor.options["B"] {
						color.Bold()
					}
					if kolor.options["I"] {
						color.Inverted()
					}
					if kolor.options["U"] {
						color.UnderLine()
					}
					v, ok := kolor.bg[string(k[0])]
					if ok {
						color.BgColor(v)
					}

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
	return string(lines), machedKind
}
