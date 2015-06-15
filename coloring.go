package main

import (
        "fmt"
        "os"
        "io/ioutil"
        "regexp"
        "strings"
        "log"
        . "github.com/mattn/go-getopt"
        "github.com/fuzzy/gocolor"
)

func usage () {
        fmt.Println(`usage: coloring [-f file|-[rgbycpwk] regexp|-h]

        -f file_name ... read from file instead of stdin
        -r regexp ... to be red
        -g regexp ... to be green
        -b regexp ... to be blue
        -y regexp ... to be yellow
        -c regexp ... to be cyan
        -p regexp ... to be purple
        -w regexp ... to be white
        -k regexp ... to be black`)
        os.Exit(1)
}

func main () {

        pattern, fileName := parseOptions()

        var (
                whole []byte
                ioerr error
        )

        if len(fileName) == 0 {
                whole,ioerr = ioutil.ReadAll(os.Stdin)
        } else {
                whole,ioerr = ioutil.ReadFile(fileName)
        }

        if ioerr != nil {
                log.Fatal(ioerr)
        }

        re,regexpErr := regexp.Compile(pattern)

        if regexpErr != nil {
                log.Fatal(regexpErr)
        }

        fmt.Println(colorling(re, string(whole)))
        os.Exit(0)
}

func parseOptions () (string, string) {
        replace := make([]string, 0)
        var (
                fileName string
                c int
                isDebug bool
        )
        regexpFlg := "(?s)"

        colorMap := map[string]string {
                "r" : "red",
                "g" : "green",
                "b" : "blue",
                "y" : "yellow",
                "p" : "pink",
                "c" : "cyan",
                "k" : "black",
                "w" : "white",
        }

        options := "mdhf:"
        colorOptions := make([]string, 0)
        colorHelp    := make([]string, 0)
        for k := range colorMap {
                colorOptions = append(colorOptions, k)
                colorHelp    = append(colorHelp, "-" + k)
        }

        for {
                if c = Getopt(options + strings.Join(colorOptions, ":") + ":"); c == EOF {
                        break
                }

                switch c {
                case 'f':
                        fileName = OptArg
                case 'h':
                        usage()
                case 'd':
                        isDebug = true
                case 'm':
                        regexpFlg = "(?m)"
                default:
                        if color, ok := colorMap[string(c)]; ok {
                                replace = append(replace, fmt.Sprintf("(?P<%s>%s)",  color, OptArg))
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


        pattern := regexpFlg + strings.Join(replace, "|")
        if isDebug {
                fmt.Println("regexp: " + pattern)
        }

        return pattern, fileName
}

func colorling (re *regexp.Regexp, lines string) string {
        colorFunc := map [string]interface{} {
                "red"   :  func (s string) string {return string(gocolor.String(s).Red())},
                "green" :  func (s string) string {return string(gocolor.String(s).Green())},
                "blue"  :  func (s string) string {return string(gocolor.String(s).Blue())},
                "yellow":  func (s string) string {return string(gocolor.String(s).Yellow())},
                "white" :  func (s string) string {return string(gocolor.String(s).White())},
                "cyan"  :  func (s string) string {return string(gocolor.String(s).Cyan())},
                "black" :  func (s string) string {return string(gocolor.String(s).Black())},
                "purple":  func (s string) string {return string(gocolor.String(s).Purple())},
        }

        // should be improved
        lines = re.ReplaceAllStringFunc(lines, func (s string) string {
                result := make(map[string]string)
                match  := re.FindStringSubmatch(s)

                for i, name := range re.SubexpNames() {
                        result[name] = match[i]
                }

                for k := range colorFunc {
                        if len(result[k]) > 0 {
                                return colorFunc[k].(func (string) string)(s)
                        }
                }
                // never come here
                return s;
        })
        return string(lines)
}
