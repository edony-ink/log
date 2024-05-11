package log

import "github.com/fatih/color"

type ColorFormat struct {
	Foreground color.Attribute
	Background color.Attribute
	Actions    color.Attribute
}

func SetupColor() {
}
