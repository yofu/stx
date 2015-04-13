package stgxui

import (
	"github.com/google/gxui"
)

const (
	defaultfontface = "IPA明朝"
	defaultfontsize = 12
)

var (
	defaultfontcolor = gxui.White
)

type TextBox struct {
	Value    []string
	Position []int
	Angle    float64
	Font     *Font
	Hide     bool
}

type Font struct {
	Face  string
	Size  int
	Color gxui.Color
}

func NewTextBox() *TextBox {
	rtn := new(TextBox)
	rtn.Value = make([]string, 0)
	rtn.Position = []int{0, 0}
	rtn.Font = NewFont()
	rtn.Hide = true
	return rtn
}

func NewFont() *Font {
	rtn := new(Font)
	rtn.Face = defaultfontface
	rtn.Size = defaultfontsize
	rtn.Color = defaultfontcolor
	return rtn
}
