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
}

func parseOptions () (string, string) {
        replace := make([]string, 0)
        var (
                fileName string
                c int
        )

        for {
                if c = Getopt("r:g:b:y:p:k:c:w:f:h"); c == EOF {
                        break
                }
                switch c {
                case 'r':
                        replace = append(replace, "(?P<red>" + OptArg + ")")
                case 'g':
                        replace = append(replace, "(?P<green>" + OptArg + ")")
                case 'b':
                        replace = append(replace, "(?P<blue>" + OptArg + ")")
                case 'y':
                        replace = append(replace, "(?P<yellow>" + OptArg + ")")
                case 'p':
                        replace = append(replace, "(?P<purple>" + OptArg + ")")
                case 'c':
                        replace = append(replace, "(?P<cyan>" + OptArg + ")")
                case 'k':
                        replace = append(replace, "(?P<black>" + OptArg + ")")
                case 'w':
                        replace = append(replace, "(?P<white>" + OptArg + ")")
                case 'f':
                        fileName = OptArg
                case 'h':
                        usage()
                        os.Exit(1)
                }
        }
        OptErr = 0

        if len(replace) == 0 {
                println("any of -r, -g, -b, -y, -w, -c, -k or -p are required.")
                os.Exit(1)
        }

        pattern := "(?s)" + strings.Join(replace, "|")

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

        // should improved
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
