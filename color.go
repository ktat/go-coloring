package kolorit

import (
	"fmt"
	"regexp"
)

const (
	BLACK   = "\033[30m"
	RED     = "\033[31m"
	GREEN   = "\033[32m"
	YELLOW  = "\033[33m"
	BLUE    = "\033[34m"
	MAGENTA = "\033[35m"
	CYAN    = "\033[36m"
	WHITE   = "\033[37m"
	BOLD    = "\033[1m"
	RESET   = "\033[0m"
)

type String struct {
	Str string
}

func (s String) resetWithLineBreak(color string) string {
	var r = regexp.MustCompile("(?s)([\r\n]+)")
	return r.ReplaceAllString(s.Str, RESET+"$1"+color)
}

func (s String) String() string {
	return s.Str
}

func (s String) Black() string {
	return fmt.Sprintf(BLACK+"%s"+RESET, s.resetWithLineBreak(BLACK))
}

func (s String) Red() string {
	return fmt.Sprintf(RED+"%s"+RESET, s.resetWithLineBreak(RED))
}

func (s String) Green() string {
	return fmt.Sprintf(GREEN+"%s"+RESET, s.resetWithLineBreak(GREEN))
}

func (s String) Yellow() string {
	return fmt.Sprintf(YELLOW+"%s"+RESET, s.resetWithLineBreak(YELLOW))
}

func (s String) Blue() string {
	return fmt.Sprintf(BLUE+"%s"+RESET, s.resetWithLineBreak(BLUE))
}

func (s String) Magenta() string {
	return fmt.Sprintf(MAGENTA+"%s"+RESET, s.resetWithLineBreak(MAGENTA))
}

func (s String) Cyan() string {
	return fmt.Sprintf(CYAN+"%s"+RESET, s.resetWithLineBreak(CYAN))
}

func (s String) White() string {
	return fmt.Sprintf(WHITE+"%s"+RESET, s.resetWithLineBreak(WHITE))
}

func (s String) Bold() string {
	return fmt.Sprintf(BOLD+"%s"+RESET, s.resetWithLineBreak(BOLD))
}
