package main

import (
	"github.com/yofu/stx/stgxui"
	"github.com/google/gxui/drivers/gl"
	"github.com/google/gxui"
	"github.com/google/gxui/themes/dark"
	"io/ioutil"
)

const (
	version  = "0.1.0"
	modified = "LAST CHANGE:13-Apr-2015."
	HOME     = "C:/D/CDOCS/Hogan/Debug"
	HOGAN    = "C:/D/CDOCS/Hogan/Debug"
)

func appMain(driver gxui.Driver) {
	theme := dark.CreateTheme(driver)
	f, err := ioutil.ReadFile("yumindb.ttf")
	if err == nil {
		font, err := driver.CreateFont(f, 11)
		if err == nil {
			font.LoadGlyphs(32, 126)
			theme.SetDefaultFont(font)
		}
	}
	stgxui.NewWindow(driver, theme, HOME)
}

func main() {
	gl.StartDriver(appMain)
}
