// Package kolorit returns ASCII colored string
package kolorit

import (
	"fmt"
	"regexp"
)

const (
	bLACK   = "\033[30m"
	rED     = "\033[31m"
	gREEN   = "\033[32m"
	yELLOW  = "\033[33m"
	bLUE    = "\033[34m"
	mAGENTA = "\033[35m"
	cYAN    = "\033[36m"
	wHITE   = "\033[37m"
	bOLD    = "\033[1m"
	rESET   = "\033[0m"
)

// String is struct which contains string to be colored
type String struct {
	Str      string
	withBold bool
}

// WithBold tell string should be bold
func (s *String) WithBold(t bool) {
	s.withBold = t
}

func (s String) resetWithLineBreak(color string) string {
	var r = regexp.MustCompile("(?s)([\r\n]+)")
	return r.ReplaceAllString(s.Str, rESET+"$1"+color)
}

// String returns string you set
func (s String) String() string {
	return s.Str
}

func (s String) sprintf(c string) string {
	if s.withBold {
		c += bOLD
	}
	return fmt.Sprintf(c+"%s"+rESET, s.resetWithLineBreak(c))
}

// Black returns black colored string
func (s String) Black() string {
	return s.sprintf(bLACK)
}

// Red returns red colored string
func (s String) Red() string {
	return s.sprintf(rED)
}

// Green returns green colored string
func (s String) Green() string {
	return s.sprintf(gREEN)
}

// Yellow returns yellow colored string
func (s String) Yellow() string {
	return s.sprintf(yELLOW)
}

// Blue returns blue colored string
func (s String) Blue() string {
	return s.sprintf(bLUE)
}

// Magenta returns blue colored string
func (s String) Magenta() string {
	return s.sprintf(mAGENTA)
}

// Cyan returns cyan colored string
func (s String) Cyan() string {
	return s.sprintf(cYAN)
}

// White returns white colored string
func (s String) White() string {
	return s.sprintf(wHITE)
}

// Bold returns bold string
func (s String) Bold() string {
	return fmt.Sprintf(bOLD+"%s"+rESET, s.resetWithLineBreak(bOLD))
}
