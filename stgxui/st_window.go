package stgxui

import (
	"fmt"
	"github.com/google/gxui"
	"github.com/yofu/st/stlib"
	"path/filepath"
	"math"
	"os"
)

// Constants & Variables
// General
var (
	// aliases            map[string]*Command
	sectionaliases     map[int]string
	Inps               []string
	SearchingInps      bool
	SearchingInpsDone  chan bool
	// DoubleClickCommand = []string{"TOGGLEBOND", "EDITPLATEELEM"}
	comhistpos         int
	undopos            int
	completepos        int
	completes          []string
	prevkey            int
	clineinput         string
)
const (
	nRecentFiles = 3
	nUndo        = 10
)
var (
	gopath          = os.Getenv("GOPATH")
	home            = os.Getenv("HOME")
	releasenote     = filepath.Join(home, ".st/help/releasenote.html")
	tooldir         = filepath.Join(home, ".st/tool")
	pgpfile         = filepath.Join(home, ".st/st.pgp")
	recentfn        = filepath.Join(home, ".st/recent.dat")
	historyfn       = filepath.Join(home, ".st/history.dat")
	NOUNDO          = false
	ALTSELECTNODE   = true
)

var (
	fixRotate               = false
	fixMove                 = false
	deg10                   = 10.0 * math.Pi / 180.0
)

const (
	CanvasRotateSpeedX = 0.01
	CanvasRotateSpeedY = 0.01
	CanvasMoveSpeedX   = 0.05
	CanvasMoveSpeedY   = 0.05
	CanvasScaleSpeed   = 500
)

var MouseButtonNil = gxui.MouseButton(-1)

const (
	EPS = 1e-4
)

var (
	STLOGO = &TextBox{
		Value:    []string{"         software", "     forstructural", "   analysisthename", "  ofwhichstandsfor", "", " sigmatau  stress", "structure  steel", "andsometh  ing", " likethat"},
		Position: []float64{100.0, 100.0},
		Angle:    0.0,
		Font:     NewFont(),
		Hide:     false,
	}
)

type Window struct { // {{{
	Home string
	Cwd  string

	Frame *st.Frame

	driver  gxui.Driver
	theme   gxui.Theme
	dlg     gxui.Window
	draw    gxui.Image
	cline   gxui.TextBox
	history gxui.TextBox

	CanvasSize []int // width, height

	SelectNode []*st.Node
	SelectElem []*st.Elem

	PageTitle *TextBox
	Title     *TextBox
	Text      *TextBox
	TextBox   map[string]*TextBox

	papersize uint

	Version  string
	Modified string

	startX int
	startY int
	endX   int
	endY   int
	downkey gxui.MouseButton

	// lastcommand     *Command
	lastexcommand   string
	lastfig2command string

	Labels map[string]gxui.Label

	InpModified bool
	Changed     bool

	exmodech  chan (interface{})
	exmodeend chan (int)

	comhist     []string
	recentfiles []string
	undostack   []*st.Frame
}

// }}}

func (stw *Window) sideBar() gxui.PanelHolder {
	label := func(text string) gxui.Label {
		label := stw.theme.CreateLabel()
		label.SetText(text)
		return label
	}
	lblbx := func(text string) gxui.LinearLayout {
		label := stw.theme.CreateLabel()
		tbox := stw.theme.CreateTextBox()
		label.SetText(text)
		layout := stw.theme.CreateLinearLayout()
		layout.SetDirection(gxui.LeftToRight)
		layout.AddChild(label)
		layout.AddChild(tbox)
		return layout
	}
	holder := stw.theme.CreatePanelHolder()
	vpanel := stw.theme.CreateLinearLayout()
	vpanel.AddChild(label("VIEW"))
	vpanel.AddChild(lblbx("  GFACT"))
	vpanel.AddChild(label("  PERSPECTIVE"))
	vpanel.AddChild(label("DISTS"))
	vpanel.AddChild(lblbx("  R"))
	vpanel.AddChild(lblbx("  L"))
	vpanel.AddChild(label("ANGLE"))
	vpanel.AddChild(lblbx("  PHI"))
	vpanel.AddChild(lblbx("  THETA"))
	vpanel.AddChild(label("FOCUS"))
	vpanel.AddChild(lblbx("  X"))
	vpanel.AddChild(lblbx("  Y"))
	vpanel.AddChild(lblbx("  Z"))
	scrollvp := stw.theme.CreateScrollLayout()
	scrollvp.SetChild(vpanel)
	holder.AddPanel(scrollvp, "View")
	holder.AddPanel(label("Show"), "Show")
	holder.AddPanel(label("Property"), "Property")
	return holder
}

func (stw *Window) commandArea() gxui.LinearLayout {
	rtn := stw.theme.CreateLinearLayout()
	rtn.SetDirection(gxui.TopToBottom)
	stw.history = stw.theme.CreateTextBox()
	stw.history.SetMultiline(true)
	stw.cline = stw.theme.CreateTextBox()
	rtn.AddChild(stw.history)
	rtn.AddChild(stw.cline)
	stw.history.SetDesiredWidth(800)
	stw.cline.SetDesiredWidth(800)
	return rtn
}

func NewWindow(driver gxui.Driver, theme gxui.Theme, homedir string) *Window {
	stw := new(Window)

	stw.Home = homedir
	stw.Cwd = homedir
	stw.SelectNode = make([]*st.Node, 0)
	stw.SelectElem = make([]*st.Elem, 0)

	stw.driver = driver
	stw.theme = theme
	stw.CanvasSize = []int{1000, 1000}

	sidedraw := theme.CreateSplitterLayout()
	sidedraw.SetOrientation(gxui.Horizontal)

	side := stw.sideBar()

	stw.draw = theme.CreateImage()
	stw.draw.OnMouseUp(func (ev gxui.MouseEvent) {
		stw.downkey = MouseButtonNil
		stw.Redraw()
		switch ev.Button {
		case gxui.MouseButtonLeft:
			fmt.Println("UP: LEFT", ev.Point.X, ev.Point.Y)
		case gxui.MouseButtonMiddle:
			fmt.Println("UP: MIDDLE", ev.Point.X, ev.Point.Y)
		}
	})
	stw.draw.OnMouseDown(func (ev gxui.MouseEvent) {
		stw.downkey = ev.Button
		switch ev.Button {
		case gxui.MouseButtonLeft:
			fmt.Println("DOWN: LEFT", ev.Point.X, ev.Point.Y)
		case gxui.MouseButtonMiddle:
			fmt.Println("DOWN: MIDDLE", ev.Point.X, ev.Point.Y)
		}
		stw.startX = ev.Point.X
		stw.startY = ev.Point.Y
	})
	stw.draw.OnDoubleClick(func (ev gxui.MouseEvent) {
		switch ev.Button {
		case gxui.MouseButtonLeft:
			fmt.Println("DOUBLE: LEFT", ev.Point.X, ev.Point.Y)
		case gxui.MouseButtonMiddle:
			stw.Frame.SetFocus(nil)
			stw.RedrawNode()
			stw.ShowCenter()
		}
	})
	stw.draw.OnMouseMove(func (ev gxui.MouseEvent) {
		if stw.Frame != nil {
			switch stw.downkey {
			default:
				return
			case gxui.MouseButtonLeft:
				return
			case gxui.MouseButtonMiddle:
				stw.MoveOrRotate(ev)
				stw.RedrawNode()
			}
		}
	})
	stw.draw.OnMouseScroll(func (ev gxui.MouseEvent) {
		if stw.Frame != nil {
			val := math.Pow(2.0, float64(ev.ScrollY)/CanvasScaleSpeed)
			stw.Frame.View.Center[0] += (val - 1.0) * (stw.Frame.View.Center[0] - float64(ev.Point.X))
			stw.Frame.View.Center[1] += (val - 1.0) * (stw.Frame.View.Center[1] - float64(ev.Point.Y))
			if stw.Frame.View.Perspective {
				stw.Frame.View.Dists[1] *= val
				if stw.Frame.View.Dists[1] < 0.0 {
					stw.Frame.View.Dists[1] = 0.0
				}
			} else {
				stw.Frame.View.Gfact *= val
				if stw.Frame.View.Gfact < 0.0 {
					stw.Frame.View.Gfact = 0.0
				}
			}
			stw.Redraw()
		}
	})

	sidedraw.AddChild(side)
	sidedraw.AddChild(stw.draw)
	sidedraw.SetChildWeight(side, 0.2)
	sidedraw.SetChildWeight(stw.draw, 0.8)

	command := stw.commandArea()

	vsp := theme.CreateSplitterLayout()
	vsp.SetOrientation(gxui.Vertical)
	vsp.AddChild(sidedraw)
	vsp.AddChild(command)
	vsp.SetChildWeight(sidedraw, 0.9)
	vsp.SetChildWeight(command, 0.1)

	stw.dlg = theme.CreateWindow(1200, 900, "stx")
	stw.dlg.AddChild(vsp)
	stw.dlg.OnClose(driver.Terminate)

	stw.SetCanvasSize()

	stw.OpenFile("hiroba05.inp")
	stw.Frame.Show.NodeCaption |= st.NC_NUM
	stw.Frame.Show.ElemCaption |= st.EC_NUM
	stw.Frame.Show.ElemCaption |= st.EC_SECT

	stw.Frame.View.Center[0] = float64(stw.CanvasSize[0])*0.5
	stw.Frame.View.Center[1] = float64(stw.CanvasSize[1])*0.5

	canvas := stw.DrawFrame()
	stw.draw.SetCanvas(canvas)

	return stw
}

func (stw *Window) Snapshot() {
	stw.Changed = true
	if NOUNDO {
		return
	}
	tmp := make([]*st.Frame, nUndo)
	tmp[0] = stw.Frame.Snapshot()
	for i := 0; i < nUndo-1-undopos; i++ {
		tmp[i+1] = stw.undostack[i+undopos]
	}
	stw.undostack = tmp
	undopos = 0
}

func (stw *Window) Bbox() (xmin, xmax, ymin, ymax float64) {
	if stw.Frame == nil || len(stw.Frame.Nodes) == 0 {
		return 0.0, 0.0, 0.0, 0.0
	}
	var mins, maxs [2]float64
	first := true
	for _, j := range stw.Frame.Nodes {
		if j.IsHidden(stw.Frame.Show) {
			continue
		}
		if first {
			for k := 0; k < 2; k++ {
				mins[k] = j.Pcoord[k]
				maxs[k] = j.Pcoord[k]
			}
			first = false
		} else {
			for k := 0; k < 2; k++ {
				if j.Pcoord[k] < mins[k] {
					mins[k] = j.Pcoord[k]
				} else if maxs[k] < j.Pcoord[k] {
					maxs[k] = j.Pcoord[k]
				}
			}
		}
	}
	return mins[0], maxs[0], mins[1], maxs[1]
}

func (stw *Window) SetCanvasSize() {
	size := stw.draw.Size()
	stw.CanvasSize[0] = size.W
	stw.CanvasSize[1] = size.H
}

func (stw *Window) ShowAtCanvasCenter() {
	for _, n := range stw.Frame.Nodes {
		stw.Frame.View.ProjectNode(n)
	}
	xmin, xmax, ymin, ymax := stw.Bbox()
	if xmax == xmin && ymax == ymin {
		return
	}
	stw.SetCanvasSize()
	scale := math.Min(float64(stw.CanvasSize[0])/(xmax-xmin), float64(stw.CanvasSize[1])/(ymax-ymin)) * 0.9
	if stw.Frame.View.Perspective {
		stw.Frame.View.Dists[1] *= scale
	} else {
		stw.Frame.View.Gfact *= scale
	}
	stw.Frame.View.Center[0] = float64(stw.CanvasSize[0])*0.5 + scale*(stw.Frame.View.Center[0]-0.5*(xmax+xmin))
	stw.Frame.View.Center[1] = float64(stw.CanvasSize[1])*0.5 + scale*(stw.Frame.View.Center[1]-0.5*(ymax+ymin))
}

func (stw *Window) ShowCenter() {
	stw.ShowAtCanvasCenter()
	stw.Redraw()
}

func (stw *Window) MoveOrRotate(ev gxui.MouseEvent) {
	if !fixMove && (ev.Modifier.Shift() || fixRotate) {
		stw.Frame.View.Center[0] += float64(ev.Point.X-stw.startX) * CanvasMoveSpeedX
		stw.Frame.View.Center[1] += float64(ev.Point.Y-stw.startY) * CanvasMoveSpeedY
	} else if !fixRotate {
		stw.Frame.View.Angle[0] += float64(ev.Point.Y-stw.startY) * CanvasRotateSpeedY
		stw.Frame.View.Angle[1] -= float64(ev.Point.X-stw.startX) * CanvasRotateSpeedX
	}
}

func (stw *Window) OpenFile(filename string) error {
	var err error
	var s *st.Show
	fn := st.ToUtf8string(filename)
	frame := st.NewFrame()
	if stw.Frame != nil {
		s = stw.Frame.Show
	}
	stw.SetCanvasSize()
	frame.View.Center[0] = float64(stw.CanvasSize[0]) * 0.5
	frame.View.Center[1] = float64(stw.CanvasSize[1]) * 0.5
	switch filepath.Ext(fn) {
	case ".inp":
		err = frame.ReadInp(fn, []float64{0.0, 0.0, 0.0}, 0.0, false)
		if err != nil {
			return err
		}
		stw.Frame = frame
	case ".dxf":
		err = frame.ReadDxf(fn, []float64{0.0, 0.0, 0.0}, EPS)
		if err != nil {
			return err
		}
		stw.Frame = frame
		frame.SetFocus(nil)
		stw.DrawFrameNode()
		stw.ShowCenter()
	}
	if s != nil {
		stw.Frame.Show = s
		for snum := range stw.Frame.Sects {
			if _, ok := stw.Frame.Show.Sect[snum]; !ok {
				stw.Frame.Show.Sect[snum] = true
			}
		}
	}
	openstr := fmt.Sprintf("OPEN: %s", fn)
	stw.History(openstr)
	stw.dlg.SetTitle(stw.Frame.Name)
	stw.Frame.Home = stw.Home
	// stw.LinkTextValue()
	stw.Cwd = filepath.Dir(fn)
	// stw.AddRecently(fn)
	// stw.Snapshot()
	stw.Changed = false
	// stw.HideLogo()
	return nil
}

func (stw *Window) History(str string) {
	if str == "" {
		return
	}
	stw.history.SetText(str)
}

func (stw *Window) RedrawNode() {
	stw.draw.Canvas().Release()
	canvas := stw.DrawFrameNode()
	stw.draw.SetCanvas(canvas)
}

func (stw *Window) Redraw() {
	stw.draw.Canvas().Release()
	canvas := stw.DrawFrame()
	stw.draw.SetCanvas(canvas)
}
