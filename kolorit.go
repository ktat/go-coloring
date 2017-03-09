package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/ktat/go-ansistrings"
	"github.com/mitchellh/go-homedir"
	toml "github.com/pelletier/go-toml"
)

var isDebug bool

type kolorit struct {
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

type optDef struct {
	k        string
	order    int
	isBool   bool
	isString bool
	boolDef  bool
	strDef   string
	help     string
}

type colorName struct {
	s string
	l string
}

var opt []optDef
var homeDir string
var homeDirRegexp *regexp.Regexp
var resetRegexp = regexp.MustCompile("(?m)^(\\033\\[0m)?")
var colorArray []colorName
var colorMap = make(map[string]string)
var colorNames []string

func init() {
	var err error
	homeDir, err = homedir.Dir()
	if err != nil {
		errCheck(err, "error on getting HOME dir")
	} else {
		homeDir += string(os.PathSeparator)
	}
	homeDirRegexp = regexp.MustCompile(fmt.Sprintf("^%s", homeDir))

	opt = []optDef{
		optDef{k: "help", isBool: true, boolDef: false, help: "show usage"},
		optDef{k: "h", isBool: true, boolDef: false, help: "show usage"},
		optDef{k: "conf", isString: true, strDef: homeDir + ".kolorit.toml", help: "path of config file"},
		optDef{k: "use", isString: true, strDef: "", help: "use predefined setting from config file($HOME/.kolorit.toml)"},
		optDef{k: "grep", isBool: true, boolDef: false, help: "take string and ignore not matched lines with it like grep. cannot use it with -s"},
		optDef{k: "and", isBool: true, boolDef: false, help: "change grep option behavior. take string only when all regexps are matched."},
		optDef{k: "ngrep", isBool: true, boolDef: false, help: "ignore grep option"},
		optDef{k: "s", isBool: true, boolDef: false, help: "regexp option. treat given content as single line(default as multi line)"},
		optDef{k: "i", isBool: true, boolDef: false, help: "regexp option. do case insensitive pattern matching."},
		optDef{k: "R", isBool: true, boolDef: false, help: "recursively read directory."},
		optDef{k: "f", isString: true, strDef: "", help: "file pattern. read from matched file."},
		optDef{k: "e", isString: true, strDef: "", help: "erase matched string"},
		optDef{k: "B", isBool: true, boolDef: false, help: "matched string to be bold"},
		optDef{k: "nB", isBool: true, boolDef: false, help: "ignore -B option"},
		optDef{k: "I", isBool: true, boolDef: false, help: "matched string background color to be inverted"},
		optDef{k: "nI", isBool: true, boolDef: false, help: "ignore -I option"},
		optDef{k: "dot", isBool: true, boolDef: false, help: "dot includes files starts with '.'"},
		optDef{k: "vcs", isBool: true, boolDef: false, help: "vcs includes vcs files/dirs"},
		optDef{k: "ext", isBool: true, boolDef: false, help: "ext includes predefined extensions to ignore(images,movies,audios etc.)"},
		optDef{k: "force", isBool: true, boolDef: false, help: "forcely read file even if file has not utf-8 string"},
		optDef{k: "d", isBool: true, boolDef: false, help: "debug mode"},
	}

	colorArray = []colorName{
		colorName{s: "r", l: "red"},
		colorName{s: "g", l: "green"},
		colorName{s: "b", l: "blue"},
		colorName{s: "y", l: "yellow"},
		colorName{s: "p", l: "purple"},
		colorName{s: "c", l: "cyan"},
		colorName{s: "k", l: "black"},
		colorName{s: "w", l: "white"},
		colorName{s: "lr", l: "light_red"},
		colorName{s: "lg", l: "light_green"},
		colorName{s: "lb", l: "light_blue"},
		colorName{s: "ly", l: "light_yellow"},
		colorName{s: "lp", l: "light_purple"},
		colorName{s: "lc", l: "light_cyan"},
		colorName{s: "dgr", l: "dark_gray"},
		colorName{s: "lgr", l: "light gray"},
	}

	for _, v := range colorArray {
		colorNames = append(colorNames, v.s)
		colorMap[v.s] = v.l
	}
}

func usage() {
	fmt.Println(`Usage:
	
  kolorit [options] [FILES]
  kolorit [options] -f "*.go"
  kolorit [options] -R [FILES/DIRECTORIES]

Options:
`)
	// flag.PrintDefaults()
	for _, v := range opt {
		k := v.k
		if v.isBool {
			if len(k) > 1 {
				fmt.Printf("  -%s\n   \t%s\n", k, v.help)
			} else {
				fmt.Printf("  -%s\t%s\n", k, v.help)
			}
		} else {
			fmt.Printf("  -%s string\n   \t%s\n", k, v.help)
		}
	}
	fmt.Print("\nColor Options:\n\n")
	for _, k := range colorNames {
		fmt.Printf("  -%s regexp\t%s\n", k, "to be "+colorMap[k])
	}
	fmt.Print("\nBack Ground Color Options:\n  * color_name is name of color like 'black', 'red', 'light_blue' etc.\n\n")
	for _, k := range colorNames {
		fmt.Printf("  -b%s color_name\t%s\n", k, "background color of "+colorMap[k])
	}

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
	kolorit := kolorit{
		options:    make(map[string]bool),
		strOptions: make(map[string]string),
		bg:         make(map[string]int),
		files:      make([]string, 0),
	}
	kolorit.parseOptions()

	re, regexpErr := regexp.Compile(kolorit.pattern)
	errCheck(regexpErr, "wrong regexp: "+kolorit.pattern)

	reErase, eraseRegexpErr := regexp.Compile(kolorit.erasePattern)
	errCheck(eraseRegexpErr, "wrong regexp: "+kolorit.erasePattern)

	var ioerr error

	if kolorit.fromSTDIN {
		// read from STDIN
		if kolorit.asSingle {
			whole, ioerr := ioutil.ReadAll(os.Stdin)
			errCheck(ioerr, "error on reading STDIN")
			str, _, e := kolorit.coloringText(re, reErase, string(whole))
			if e != nil {
				errCheck(e)
			}
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
					colored, n, e := kolorit.coloringText(re, reErase, l)
					if e != nil {
						errCheck(e)
					}
					if kolorit.options["grep"] && (!kolorit.options["and"] || n == kolorit.numOfRegexps) && colored != string(l) {
						fmt.Println(colored)
					} else if !kolorit.options["grep"] {
						fmt.Println(colored)
					}
				}
			}
		}
	} else {
		// read from file or dir
		if len(kolorit.files) == 0 {
			errMessage("files are not given.")
		}
		if isDebug {
			log.Println("### read from file or dir in main")
			log.Println("File Name: " + kolorit.fileName)
			log.Println("Files: " + strings.Join(kolorit.files, ", "))
			log.Printf("Num of Files: %d\n", len(kolorit.files))
			log.Printf("Is Recursive: %t\n", kolorit.isRecursive)
		}

		if kolorit.asSingle {
			var whole []byte

			for i := 0; i < len(kolorit.files); i++ {
				fi, err := os.Stat(kolorit.files[i])
				if err != nil {
					log.Println(err.Error() + ":error on stat file: " + kolorit.files[i])
					continue
				}
				if fi.IsDir() {
					continue
				}
				whole, ioerr = ioutil.ReadFile(kolorit.files[i])
				if ioerr != nil {
					log.Println(ioerr.Error() + ":error on reading file: " + kolorit.files[i])
					continue
				}
				colored, _, e := kolorit.coloringText(re, reErase, string(whole))
				if e != nil {
					log.Println(e.Error() + " : " + kolorit.files[i])
					continue
				}
				kolorit.printColored(colored, i, 0)

			}
		} else {
			for i := 0; i < len(kolorit.files); i++ {
				fi, err := os.Stat(kolorit.files[i])
				if err != nil {
					if isDebug {
						log.Println(err.Error() + " : " + kolorit.files[i])
					}
					continue
				}
				if kolorit.isRecursive && fi.IsDir() {
					continue
				}
				var fp *os.File
				fp, ioerr = os.Open(kolorit.files[i])
				if ioerr != nil {
					log.Println(ioerr.Error() + " :cannot open file: " + kolorit.files[i])
					continue
				}
				reader := bufio.NewReaderSize(fp, 4096)
				lineNumber := 0
				for {
					lineNumber++
					line, _, ioerr := reader.ReadLine()
					if ioerr != nil && ioerr != io.EOF {
						log.Println(ioerr.Error() + " :error on reading file content: " + kolorit.files[i])
						break
					} else if ioerr == io.EOF {
						break
					}

					colored, n, e := kolorit.coloringText(re, reErase, string(line))
					if e != nil {
						log.Println(e.Error() + " : " + kolorit.files[i])
						break
					}
					if kolorit.options["grep"] && (!kolorit.options["and"] || n == kolorit.numOfRegexps) && colored != string(line) {
						kolorit.printColored(colored, i, lineNumber)
					} else if !kolorit.options["grep"] {
						kolorit.printColored(colored, i, lineNumber)

					}
				}
				ioerr = fp.Close()
				errCheck(ioerr, "error on closing file: "+kolorit.files[i])
			}
		}
	}
	os.Exit(0)
}

func (kolorit *kolorit) printColored(colored string, i int, ln int) {
	if len(kolorit.files) == 0 {
		fmt.Println(colored)
	} else if len(kolorit.files) == 1 {
		if ln == 0 {
			fmt.Println(colored)
		} else {
			fmt.Println(addLineNum(colored, ln))
		}
	} else {
		fmt.Print(addFileName(colored, kolorit.files[i], ln))
	}
}

func (kolorit *kolorit) seekDir(files *[]string, dirName string) {
	if isDebug {
		log.Println("### seekDir")
		log.Println("File Name:" + kolorit.fileName)
		log.Println("Dir Name:" + dirName)
	}
	fileInfo, ioerr := ioutil.ReadDir(dirName)
	if ioerr != nil {
		log.Println(ioerr.Error() + " :error on reading dir: " + dirName)
	} else {
		for i := 0; i < len(fileInfo); i++ {
			fullName := filepath.Join(dirName, fileInfo[i].Name())
			if fileInfo[i].IsDir() == false {
				if !kolorit.isIgnoreFile(fileInfo[i].Name()) && (kolorit.fileName == "" || kolorit.checkFileName(fullName)) {
					if isDebug {
						log.Println("File Full Name: " + fullName)
					}
					*files = append(*files, fullName)
				}
			} else if kolorit.isRecursive && !kolorit.isIgnoreDirs(fileInfo[i].Name()) {
				if isDebug {
					log.Println("Seek Dir: " + filepath.Join(dirName, fileInfo[i].Name()))
				}
				kolorit.seekDir(files, filepath.Join(dirName, fileInfo[i].Name()))
			}
		}
	}
}

func (kolorit *kolorit) isIgnoreFile(file string) (ignore bool) {
	ignore, _ = regexp.MatchString("^(\\.#.+|.+~|#.*#)$", file)
	if ignore {
		return
	}
	if !kolorit.options["vcs"] {
		switch file {
		case "=RELEASE-ID":
			ignore = true
		case "=meta-update":
			ignore = true
		case "=update":
			ignore = true
		case ".gitignore":
			ignore = true
		case ".gitmodules":
			ignore = true
		case ".gitattributes":
			ignore = true
		case ".cvsignore":
			ignore = true
		case ".bzr":
			ignore = true
		case ".bzrignore":
			ignore = true
		case ".bzrtags":
			ignore = true
		case ".hg":
			ignore = true
		case ".hgignore":
			ignore = true
		case ".hgrags":
			ignore = true
		case "_darcs":
			ignore = true
		}
	}
	if !kolorit.options["ext"] {
		// https://en.wikipedia.org/wiki/Image_file_formats#Raster_formats
		ignore, _ = regexp.MatchString("(?i)\\.(jpe?g|png|gif|bmp|raw2?|tiff?|p[pgbn]|hei[fc]|bpg|webp|ico|psd|xcf|svg|swf|pdf|ai|cgm|gbr)", file)
		if ignore {
			return
		}
		// https://en.wikipedia.org/wiki/Video_file_format
		ignore, _ = regexp.MatchString("(?i)\\.(webm|flv|vob|ogv|ogg|drc|gifv|mng|avi|mov|qt|wmv|yuv|rm|rmvb|asf|amv|mp4|m4[pv]|mp[g2v]|mpeg?|svi|3g[2p]|mxf|roq|nsv|f[l4]v|f4[pab])$", file)
		if ignore {
			return
		}
		// https://en.wikipedia.org/wiki/Audio_file_format
		ignore, _ = regexp.MatchString("(?i)\\.(3gp|aa[cx]?|act|aiff|amr|ape|au|awb|dct|dss|dvf|flac?|gsm|iklax|m4[abp]|mmf|mp[3c]|msv|m?og[ga]|opus|r[am]|raw|sln|tta|vox|wav|wma|wv|webm)$", file)
		if ignore {
			return
		}

		//https://en.wikipedia.org/wiki/List_of_archive_formats
		ignore, _ = regexp.MatchString("(?i)\\.(ar?|cpio|shar|lbr|iso|mar|tar|bz2|gz|lz(?:ma|o)?|rz|sfark|sz|xz|z|s?7z|ace|afa|alz|apk|arc|arj|b[1ah]|ca[br]|cfs|cpt|dar|dd|dgc|dmg|ear|gca|ha|hki|ice|jar|kgb|lz[ha]|pak|partimg|pag|pea|pim|pit|qda|rar|rk|sda|sea|sen|sfx|shk|si|sitx|sqx|uc\\d?|uca|uha|war|wim|xar|xp3|yz1|zipx?|zoo|zpaq|zz)$", file)
		if ignore {
			return
		}
	}
	return
}

func (kolorit *kolorit) isIgnoreDirs(dir string) (ignore bool) {
	if !kolorit.options["dot"] && len(dir) > 1 && dir[0] == '.' {
		return true
	}
	if !kolorit.options["vcs"] {
		switch dir {
		case "CVS":
			ignore = true
		case ".svn":
			ignore = true
		case ".git":
			ignore = true
		case "RCS":
			ignore = true
		case "SCCS":
			ignore = true
		case ".arch-ids":
			ignore = true
		case "{arch}":
			ignore = true
		}
	}
	return
}

func addFileName(content string, fn string, ln int) string {

	fn = homeDirRegexp.ReplaceAllString(fn, "~/")

	a := ansistrings.NewANSIStrings()
	prefix := ""
	if ln == 0 {
		prefix = a.Str(fn).Magenta().Str(":").Cyan().String()
	} else {
		prefix = a.Str(fn).Magenta().Str(":").Cyan().Str(strconv.Itoa(ln)).Yellow().Str(":").Cyan().String()
	}
	return resetRegexp.ReplaceAllString(content, prefix+"$1") + "\n"
}

func addLineNum(content string, ln int) string {
	a := ansistrings.NewANSIStrings()
	prefix := a.Str(strconv.Itoa(ln)).Yellow().Str(":").Cyan().String()
	return resetRegexp.ReplaceAllString(content, prefix+"$1")
}

func (kolorit *kolorit) checkFileName(targetFile string) bool {
	kolorit.pattern = kolorit.fileName
	kolorit.pattern = strings.Replace(kolorit.pattern, ".", "\\.", -1)
	kolorit.pattern = strings.Replace(kolorit.pattern, "*", ".*", -1)
	matched, err := regexp.MatchString("(^|/)"+kolorit.pattern+"$", targetFile)
	if isDebug {
		log.Println("### checkFileName")
		log.Println("Target File: " + targetFile)
		log.Println("File Name: " + kolorit.fileName)
		log.Println("Pattern: " + kolorit.pattern)
		log.Printf("Matched: %t\n", matched)
	}
	if err == nil && matched {
		return true
	}
	return false
}

func (kolorit *kolorit) parseOptions() {
	var err error
	replace := make([]string, 0)
	regexpFlg := ""
	regexpFlgs := make(map[byte]bool)
	colorOptions := make([]string, 0)
	colorHelp := make([]string, 0)
	boolParsedOpt := make(map[string]*bool)
	strParsedOpt := make(map[string]*string)
	regexps := make(map[string]*string)
	bgOptions := make(map[string]*string)

	for k := range colorMap {
		regexps[k] = flag.String(k, "", "regexp to be "+colorMap[k])
		bgOptions["b"+k] = flag.String("b"+k, "", "background color of "+colorMap[k])
	}

	for _, v := range opt {
		if v.isBool {
			boolParsedOpt[v.k] = flag.Bool(v.k, v.boolDef, v.help)
		} else if v.isString {
			strParsedOpt[v.k] = flag.String(v.k, v.strDef, v.help)
		}
	}

	flag.Parse()

	// parse options
	for k, v := range boolParsedOpt {
		if v != nil {
			kolorit.options[k] = *v
		}
	}
	for k, v := range strParsedOpt {
		if v != nil {
			kolorit.strOptions[k] = *v
		}
	}

	isDebug = kolorit.options["d"]

	kolorit.isRecursive = kolorit.options["R"]

	// print usage and exit
	if kolorit.options["help"] || kolorit.options["h"] {
		usage()
	}

	// options from config file
	kolorit.parseConfig(kolorit.strOptions["conf"], kolorit.strOptions["use"], colorMap, &regexps)

	kolorit.erasePattern = kolorit.strOptions["e"]
	kolorit.asSingle = kolorit.options["s"]

	// rest args after options are regareded as files
	for n := 0; n < flag.NArg(); n++ {
		if isDebug {
			log.Println("Add File: " + flag.Arg(n))
		}
		kolorit.files = append(kolorit.files, flag.Arg(n))
	}

	if kolorit.strOptions["f"] != "" && kolorit.strOptions["f"] != "-" {
		kolorit.fileName = kolorit.strOptions["f"]
	} else if len(kolorit.files) == 0 && !kolorit.isRecursive {
		kolorit.fromSTDIN = true
	}

	// collect target files
	if !kolorit.fromSTDIN {
		if len(kolorit.files) == 0 && kolorit.fileName != "" {
			kolorit.seekDir(&kolorit.files, ".")
		} else if kolorit.isRecursive {
			for _, f := range kolorit.files {
				fi, err := os.Stat(f)
				errCheck(err, "error on stat file: "+f)
				if fi.IsDir() {
					kolorit.seekDir(&kolorit.files, f)
				}
			}
		}
		if len(kolorit.files) == 0 {
			errMessage("files are not given/found")
		}
	}

	// build regexp flags
	for _, k := range []byte{'s', 'i'} {
		regexpFlgs[k] = kolorit.options[string(k)]
	}
	regexpFlgs['m'] = !regexpFlgs['s']

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
			kolorit.numOfRegexps++
		}
		colorHelp = append(colorHelp, "-"+string(k))
		v, ok := bgOptions["b"+k]
		if ok && *v != "" {
			kolorit.bg[k], err = ansistrings.ColorNumFromName(*v)
			if err != nil {
				errCheck(err, "unknown color name: "+*v)
			}
		}
	}

	if len(replace) == 0 {
		errMessage("any of " + strings.Join(colorHelp, ", ") + " AND -R, -f or file names as rest of args is required.\n")
	}

	// assemble regexps
	kolorit.pattern = regexpFlg + strings.Join(replace, "|")
	if isDebug {
		log.Println("regexp: " + kolorit.pattern)
	}
}

func (kolorit *kolorit) parseConfig(configFile string, use string, colorMap map[string]string, regexps *map[string]*string) {
	if use == "default" {
		errMessage("cannot pass 'default' as 'use' argument")
	}

	_, err := os.Stat(configFile)
	if err != nil {
		errCheck(err, "cannot find/read config file:"+configFile)
	}

	config, err := toml.LoadFile(configFile)
	if err != nil {
		if use != "" {
			errMessage("cannot parse config file: " + configFile)
		} else {
			return
		}
	}
	opt := config.Get(use)
	switch opt.(type) {
	case nil:
		errCheck(nil, "'"+use+"' is not defined in "+configFile)
	}

	defaultOpt := config.Get("default")
	switch defaultOpt.(type) {
	case nil:
		defaultOpt = &toml.TomlTree{}
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
			isRegexp = defaultOpt.(*toml.TomlTree).Get(string(k))
		}
		switch isRegexp.(type) {
		case nil:
			continue
		default:
			regexpStr := isRegexp.(string)
			if k == "e" {
				kolorit.strOptions[string(k)] = regexpStr
			} else {
				if *(*regexps)[k] == "" && regexpStr != "" {
					(*regexps)[k] = &regexpStr
				}

			}
		}
	}
	boolOpts := []string{"B", "m", "i", "s", "I", "grep", "ngrep", "nI", "nB"}
	for _, k := range boolOpts {
		boolOpt := opt.(*toml.TomlTree).Get(k)
		switch boolOpt.(type) {
		case nil:
			boolOpt = defaultOpt.(*toml.TomlTree).Get(k)
		}
		switch boolOpt.(type) {
		case nil:
			continue
		default:
			kolorit.options[k] = boolOpt.(bool)
		}
	}
	nArry := []string{"grep", "I", "B"}
	for _, k := range nArry {
		if kolorit.options["n"+k] {
			kolorit.options[k] = false
		}
	}
}

func (kolorit *kolorit) coloringText(re *regexp.Regexp, reErase *regexp.Regexp, lines string) (string, int, error) {
	for i := 0; i < len(lines); i++ {
		if utf8.ValidString(lines) == false && !kolorit.options["force"] {
			return "", 0, errors.New("binary string or not utf-8 character is given")
		}
	}

	lines = reErase.ReplaceAllString(lines, "")
	colorFunc := make(map[string]func(s ansistrings.ANSIString) string)
	for _, colorName := range colorNames {
		name := colorMap[colorName]
		n, _ := ansistrings.ColorNumFromName(name)
		colorFunc[name] = func(s ansistrings.ANSIString) string {
			s.Color(n)
			return s.String()
		}
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

		for _, colorName := range colorNames {
			k := colorMap[colorName]
			newStr := ""
			if len(result[k]) > 2 { // if parenthese exists in regexp, ignore first match which matches whole string
				result[k] = result[k][2:]
			}
			for i := len(result[k]) - 1; i >= 0; i -= 2 {
				if result[k][i] > 0 {
					var matchedIndex []int
					matchedIndex = append(matchedIndex, result[k][i-1], result[k][i])
					var color ansistrings.ANSIString
					if kolorit.options["B"] {
						color.Bold()
					}
					if kolorit.options["I"] {
						color.Inverted()
					}
					if kolorit.options["U"] {
						color.UnderLine()
					}
					v, ok := kolorit.bg[colorName]
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
	return string(lines), machedKind, nil
}
