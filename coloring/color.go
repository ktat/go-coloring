package coloring

import (
	"fmt"
)

type String struct {
	Str string
}

func (s String) String() string {
	return s.Str
}

func (s String) Black() string {
	return fmt.Sprintf("\033[30m%s\033[0m", s)
}

func (s String) Red() string {
	return fmt.Sprintf("\033[31m%s\033[0m", s)
}

func (s String) Green() string {
	return fmt.Sprintf("\033[32m%s\033[0m", s)
}

func (s String) Yellow() string {
	return fmt.Sprintf("\033[33m%s\033[0m", s)
}

func (s String) Blue() string {
	return fmt.Sprintf("\033[34m%s\033[0m", s)
}

func (s String) Magenta() string {
	return fmt.Sprintf("\033[35m%s\033[0m", s)
}

func (s String) Cyan() string {
	return fmt.Sprintf("\033[36m%s\033[0m", s)
}

func (s String) White() string {
	return fmt.Sprintf("\033[37m%s\033[0m", s)
}

func (s String) Bold() string {
	return fmt.Sprintf("\033[1m%s\033[0m", s)
}
