// Package kolorit returns ASCII colored/decorated string
package kolorit

import (
	"errors"
	"fmt"
	"regexp"
)

// constant value of colors
const (
	BLACK        = 30
	RED          = 31
	GREEN        = 32
	YELLOW       = 33
	BLUE         = 34
	MAGENTA      = 35
	CYAN         = 36
	LightGRAY    = 37
	DarkGRAY     = 90
	LightRED     = 91
	LightGREEN   = 92
	LightYELLOW  = 93
	LightBLUE    = 94
	LightMAGENTA = 95
	LightCYAN    = 96
	WHITE        = 97
	bOLD         = "\033[1m"
	uNDERLINE    = "\033[4m"
	bLINK        = "\033[5m"
	iNVERTED     = "\033[7m"
	rESET        = "\033[0m"
)

var name2color = map[string]int{
	"black":         BLACK,
	"red":           RED,
	"green":         GREEN,
	"yellow":        YELLOW,
	"blue":          BLUE,
	"magenta":       MAGENTA,
	"cyan":          CYAN,
	"light_gray":    LightGRAY,
	"dark_gray":     DarkGRAY,
	"light_red":     LightRED,
	"light_green":   LightGREEN,
	"light_yellow":  LightYELLOW,
	"light_blue":    LightBLUE,
	"light_magenta": LightMAGENTA,
	"light_cyan":    LightCYAN,
	"white":         WHITE,
}

// ANSIString is struct which contains string with ANSI escaping setting
type ANSIString struct {
	Str   string
	color struct {
		color int
		isSet bool
	}
	colorN struct {
		color int
		isSet bool
	}
	bgColor struct {
		color int
		isSet bool
	}
	bgColorN struct {
		color int
		isSet bool
	}
	withBold      bool
	withBlink     bool
	withInverted  bool
	withUnderline bool
}

// ColorNumFromName return number of colors
func ColorNumFromName(n string) (cn int, e error) {
	v, ok := name2color[n]
	if ok {
		cn = v
	} else {
		e = errors.New("unknown color name: " + n)
	}
	return cn, e
}

func (s ANSIString) String() string {
	color := ""
	if s.color.isSet {
		color += fmt.Sprintf("\033[%dm", s.color.color)
	}
	if s.colorN.isSet {
		c := s.colorN.color
		if (c >= 30 && c <= 37) || (c >= 90 && c <= 97) {
			color += fmt.Sprintf("\033[%dm", s.colorN.color)
		} else {
			color += fmt.Sprintf("\033[38;5;%dm", s.colorN.color)

		}
	}
	if s.bgColorN.isSet {
		color += fmt.Sprintf("\033[48;5;%dm", s.bgColorN.color)
	}
	if s.bgColor.isSet {
		color += fmt.Sprintf("\033[%dm", s.bgColor.color+10)
	}
	if s.withBold {
		color += bOLD
	}
	if s.withUnderline {
		color += uNDERLINE
	}
	if s.withBlink {
		color += bLINK
	}
	if s.withInverted {
		color += iNVERTED
	}
	if color == "" {
		return s.Str
	}
	return fmt.Sprintf(color+"%s"+rESET, s.resetWithLineBreak(s.Str, color))
}

func (s *ANSIString) resetWithLineBreak(str string, color string) string {
	var r = regexp.MustCompile("(?s)([\r\n]+)")
	return r.ReplaceAllString(str, rESET+"$1"+color)
}

// BgColor set background color of string
func (s *ANSIString) BgColor(c int) {
	s.bgColor.color = c
	s.bgColor.isSet = true
	s.bgColorN.isSet = false
}

// BgColorN set background color of string
func (s *ANSIString) BgColorN(c int) {
	s.bgColorN.color = c
	s.bgColor.isSet = false
	s.bgColorN.isSet = true
}

// UnsetColor unset color of string
func (s *ANSIString) UnsetColor() {
	s.color.isSet = false
	s.colorN.isSet = false
}

// UnsetBgColor unset background color of string
func (s *ANSIString) UnsetBgColor() {
	s.bgColor.isSet = false
	s.bgColorN.isSet = false
}

// Black returns black colored string
func (s *ANSIString) Black() string {
	s.colorN.isSet = false
	s.color.isSet = true
	s.color.color = BLACK
	return s.String()
}

// Red returns red colored string
func (s *ANSIString) Red() string {
	s.colorN.isSet = false
	s.color.isSet = true
	s.color.color = RED
	return s.String()
}

// Green returns green colored string
func (s *ANSIString) Green() string {
	s.colorN.isSet = false
	s.color.isSet = true
	s.color.color = GREEN
	return s.String()
}

// Yellow returns yellow colored string
func (s *ANSIString) Yellow() string {
	s.colorN.isSet = false
	s.color.isSet = true
	s.color.color = YELLOW
	return s.String()
}

// Blue returns blue colored string
func (s *ANSIString) Blue() string {
	s.colorN.isSet = false
	s.color.isSet = true
	s.color.color = BLUE
	return s.String()
}

// Magenta returns blue colored string
func (s *ANSIString) Magenta() string {
	s.colorN.isSet = false
	s.color.isSet = true
	s.color.color = MAGENTA
	return s.String()
}

// Cyan returns cyan colored string
func (s *ANSIString) Cyan() string {
	s.colorN.isSet = false
	s.color.isSet = true
	s.color.color = CYAN
	return s.String()
}

// White returns white colored string
func (s *ANSIString) White() string {
	s.colorN.isSet = false
	s.color.isSet = true
	s.color.color = WHITE
	return s.String()
}

// LightRed returns light red colored string
func (s *ANSIString) LightRed() string {
	s.colorN.isSet = false
	s.color.isSet = true
	s.color.color = LightRED
	return s.String()
}

// LightGreen returns light green colored string
func (s *ANSIString) LightGreen() string {
	s.colorN.isSet = false
	s.color.isSet = true
	s.color.color = LightGREEN
	return s.String()
}

// LightYellow returns light yellow colored string
func (s *ANSIString) LightYellow() string {
	s.colorN.isSet = false
	s.color.isSet = true
	s.color.color = LightYELLOW
	return s.String()
}

// LightBlue returns light blue colored string
func (s *ANSIString) LightBlue() string {
	s.colorN.isSet = false
	s.color.isSet = true
	s.color.color = LightBLUE
	return s.String()
}

// LightMagenta returns light magenta colored string
func (s *ANSIString) LightMagenta() string {
	s.colorN.isSet = false
	s.color.isSet = true
	s.color.color = LightMAGENTA
	return s.String()
}

// LightCyan returns light cyan colored string
func (s *ANSIString) LightCyan() string {
	s.colorN.isSet = false
	s.color.isSet = true
	s.color.color = LightCYAN
	return s.String()
}

// LightGray returns light gray colored string
func (s *ANSIString) LightGray() string {
	s.colorN.isSet = false
	s.color.isSet = true
	s.color.color = LightGRAY
	return s.String()
}

// DarkGray returns dark gray colored string
func (s *ANSIString) DarkGray() string {
	s.colorN.isSet = false
	s.color.isSet = true
	s.color.color = DarkGRAY
	return s.String()
}

// ColorN returns colored string(argument range is from 0 to 256)
func (s *ANSIString) ColorN(c int) string {
	if c < 0 || c > 256 {
		panic("invait argument: valid range is from 0 to 256")
	}
	s.color.isSet = false
	s.colorN.color = c
	s.colorN.isSet = true
	return s.String()
}

// Bold returns bold string
func (s *ANSIString) Bold(t ...bool) string {
	if len(t) == 0 {
		s.withBold = true
	} else {
		s.withBold = t[0]
	}
	return s.String()
}

// Blink returns bold string
func (s *ANSIString) Blink(t ...bool) string {
	if len(t) == 0 {
		s.withBlink = true
	} else {
		s.withBlink = t[0]
	}
	return s.String()
}

// Inverted returns inverted string
func (s *ANSIString) Inverted(t ...bool) string {
	if len(t) == 0 {
		s.withInverted = true
	} else {
		s.withInverted = t[0]
	}

	return s.String()
}

// UnderLine returns underlined string
func (s *ANSIString) UnderLine(t ...bool) string {
	if len(t) == 0 {
		s.withUnderline = true
	} else {
		s.withUnderline = t[0]
	}
	return s.String()
}

// Reset resets ANSI setting
func (s *ANSIString) Reset() {
	str := s.Str
	*s = ANSIString{Str: str}
}
