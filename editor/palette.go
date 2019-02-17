package editor

import (
	"fmt"
	"math"

	"github.com/akiyosi/gonvim/fuzzy"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/svg"
	"github.com/therecipe/qt/widgets"
)

// Palette is the popup for fuzzy finder, cmdline etc
type Palette struct {
	// mu               sync.Mutex
	// procCount        int
	ws               *Workspace
	hidden           bool
	widget           *widgets.QWidget
	patternText      string
	resultItems      []*PaletteResultItem
	resultWidget     *widgets.QWidget
	resultMainWidget *widgets.QWidget
	itemHeight       int
	width            int
	cursor           *widgets.QWidget
	cursorX          int
	resultType       string
	itemTypes        []string
	max              int
	showTotal        int
	pattern          *widgets.QLabel
	patternPadding   int
	patternWidget    *widgets.QWidget
	scrollBar        *widgets.QWidget
	scrollBarPos     int
	scrollCol        *widgets.QWidget
}

// PaletteResultItem is the result item
type PaletteResultItem struct {
	p          *Palette
	hidden     bool
	icon       *svg.QSvgWidget
	iconType   string
	iconHidden bool
	base       *widgets.QLabel
	baseText   string
	widget     *widgets.QWidget
	selected   bool
}

func initPalette() *Palette {
	width := 600
	mainLayout := widgets.NewQVBoxLayout()
	mainLayout.SetContentsMargins(0, 0, 0, 0)
	mainLayout.SetSpacing(0)
	mainLayout.SetSizeConstraint(widgets.QLayout__SetMinAndMaxSize)
	widget := widgets.NewQWidget(nil, 0)
	widget.SetLayout(mainLayout)
	widget.SetContentsMargins(1, 1, 1, 1)
	// widget.SetFixedWidth(width)
	widget.SetObjectName("palette")
	shadow := widgets.NewQGraphicsDropShadowEffect(nil)
	shadow.SetBlurRadius(35)
	shadow.SetColor(gui.NewQColor3(0, 0, 0, 200))
	shadow.SetOffset3(-2, 8)
	widget.SetGraphicsEffect(shadow)

	resultMainLayout := widgets.NewQHBoxLayout()
	resultMainLayout.SetContentsMargins(0, 0, 0, 0)
	resultMainLayout.SetSpacing(0)
	resultMainLayout.SetSizeConstraint(widgets.QLayout__SetMinAndMaxSize)

	padding := 8
	resultLayout := widgets.NewQVBoxLayout()
	resultLayout.SetContentsMargins(0, 0, 0, 0)
	resultLayout.SetSpacing(0)
	resultLayout.SetSizeConstraint(widgets.QLayout__SetMinAndMaxSize)
	resultWidget := widgets.NewQWidget(nil, 0)
	resultWidget.SetLayout(resultLayout)
	resultWidget.SetContentsMargins(0, 0, 0, 0)

	scrollCol := widgets.NewQWidget(nil, 0)
	scrollCol.SetContentsMargins(0, 0, 0, 0)
	scrollCol.SetFixedWidth(5)
	scrollBar := widgets.NewQWidget(scrollCol, 0)
	scrollBar.SetFixedWidth(5)

	resultMainWidget := widgets.NewQWidget(nil, 0)
	resultMainWidget.SetContentsMargins(0, 0, 0, 0)
	resultMainLayout.AddWidget(resultWidget, 0, 0)
	resultMainLayout.AddWidget(scrollCol, 0, 0)
	resultMainWidget.SetLayout(resultMainLayout)

	pattern := widgets.NewQLabel(nil, 0)
	pattern.SetContentsMargins(padding, padding, padding, padding)
	pattern.SetFixedWidth(width - padding*2)
	pattern.SetSizePolicy2(widgets.QSizePolicy__Preferred, widgets.QSizePolicy__Maximum)
	patternLayout := widgets.NewQVBoxLayout()
	patternLayout.AddWidget(pattern, 0, 0)
	patternLayout.SetContentsMargins(0, 0, 0, 0)
	patternLayout.SetSpacing(0)
	patternLayout.SetSizeConstraint(widgets.QLayout__SetMinAndMaxSize)
	patternWidget := widgets.NewQWidget(nil, 0)
	patternWidget.SetLayout(patternLayout)
	patternWidget.SetContentsMargins(padding, padding, padding, padding)

	cursor := widgets.NewQWidget(nil, 0)
	cursor.SetParent(pattern)
	cursor.SetFixedSize2(1, pattern.SizeHint().Height()-padding*2)
	cursor.Move2(padding, padding)

	mainLayout.AddWidget(patternWidget, 0, 0)
	mainLayout.AddWidget(resultMainWidget, 0, 0)

	palette := &Palette{
		width:            width,
		widget:           widget,
		resultWidget:     resultWidget,
		resultMainWidget: resultMainWidget,
		pattern:          pattern,
		patternPadding:   padding,
		patternWidget:    patternWidget,
		scrollCol:        scrollCol,
		scrollBar:        scrollBar,
		cursor:           cursor,
	}

	resultItems := []*PaletteResultItem{}
	max := 30
	for i := 0; i < max; i++ {
		itemWidget := widgets.NewQWidget(nil, 0)
		itemWidget.SetContentsMargins(0, 0, 0, 0)
		itemLayout := newVFlowLayout(padding, padding*2, 0, 0, 9999)
		itemLayout.SetSizeConstraint(widgets.QLayout__SetMinAndMaxSize)
		itemWidget.SetLayout(itemLayout)
		itemWidget.SetStyleSheet("background-color: rgba(0, 0, 0, 0);")
		resultLayout.AddWidget(itemWidget, 0, 0)
		icon := svg.NewQSvgWidget(nil)
		icon.SetFixedWidth(editor.iconSize - 1)
		icon.SetFixedHeight(editor.iconSize - 1)
		icon.SetContentsMargins(0, 0, 0, 0)
		icon.SetStyleSheet("background-color: rgba(0, 0, 0, 0);")
		base := widgets.NewQLabel(nil, 0)
		base.SetText("base")
		base.SetContentsMargins(0, padding, 0, padding)
		base.SetStyleSheet("background-color: rgba(0, 0, 0, 0); white-space: pre-wrap;")
		// base.SetSizePolicy2(widgets.QSizePolicy__Preferred, widgets.QSizePolicy__Maximum)
		itemLayout.AddWidget(icon)
		itemLayout.AddWidget(base)
		resultItem := &PaletteResultItem{
			p:      palette,
			widget: itemWidget,
			icon:   icon,
			base:   base,
		}
		resultItems = append(resultItems, resultItem)
	}
	palette.max = max
	palette.resultItems = resultItems
	return palette
}

func (p *Palette) setColor() {
	fg := editor.colors.widgetFg.String()
	bg := editor.colors.widgetBg
	inputArea := editor.colors.widgetInputArea
	sbg := editor.colors.scrollBarBg
	// transparent := editor.config.Editor.Transparent / 4.0
	transparent := transparent() * transparent()
	p.cursor.SetStyleSheet(fmt.Sprintf("background-color: %s;", fg))
	//p.widget.SetStyleSheet(fmt.Sprintf(" QWidget#palette { border: 1px solid %s; } .QWidget { background-color: rgba(%d, %d, %d, %f); } * { color: %s; } ", bg, bg, fg))
	p.widget.SetStyleSheet(fmt.Sprintf(" .QWidget { background-color: rgba(%d, %d, %d, %f); } * { color: %s; } ", bg.R, bg.G, bg.B, transparent, fg))
	p.scrollBar.SetStyleSheet(fmt.Sprintf("background-color: rgba(%d, %d, %d, %f);", sbg.R, sbg.G, sbg.B, transparent))
	p.pattern.SetStyleSheet(fmt.Sprintf("background-color: rgba(%d, %d, %d, %f);", inputArea.R, inputArea.G, inputArea.B, transparent))
}

func (p *Palette) resize() {
	// if p.procCount > 0 {
	// 	return
	// }
	// go func() {
	// p.mu.Lock()
	// defer p.mu.Unlock()
	// p.procCount = 1
	// defer func() { p.procCount = 0 }()

	padding := 8
	p.width = int(math.Trunc(float64(editor.width) * 0.7))
	cursorBoundary := p.cursor.Pos().X() + 35

	if cursorBoundary > p.width {
		p.width = cursorBoundary
	}
	if p.width > editor.width {
		p.width = editor.width
		p.pattern.SetAlignment(core.Qt__AlignRight | core.Qt__AlignCenter)
		return
	} else if p.width <= editor.width {
		if p.pattern.Alignment() != core.Qt__AlignLeft {
			p.pattern.SetAlignment(core.Qt__AlignLeft)
			return
		}
	}

	p.pattern.SetFixedWidth(p.width - padding*2)
	p.widget.SetMaximumWidth(p.width)
	p.widget.SetMinimumWidth(p.width)

	x := editor.width - p.width
	if x < 0 {
		x = 0
	}
	p.widget.Move2(x/2, 10)

	itemHeight := p.resultItems[0].widget.SizeHint().Height()
	p.itemHeight = itemHeight
	p.showTotal = int(float64(p.ws.height)/float64(itemHeight)*0.5) - 1
	if p.ws.uiAttached {
		fuzzy.UpdateMax(p.ws.nvim, p.showTotal)
	}
	for i := p.showTotal; i < len(p.resultItems); i++ {
		p.resultItems[i].hide()
	}
	// }()
}

func (p *Palette) show() {
	if !p.hidden {
		return
	}
	p.hidden = false
	p.widget.Raise()
	p.widget.SetWindowOpacity(1.0)
	p.widget.Show()
	p.resize()
}

func (p *Palette) hide() {
	if p.hidden {
		return
	}
	p.hidden = true
	p.widget.Hide()
}

func (p *Palette) setPattern(text string) {
	p.patternText = text
	p.pattern.SetText(text)
}

func (p *Palette) cursorMove(x int) {
	//p.cursorX = int(p.ws.font.defaultFontMetrics.Width(string(p.patternText[:x])))
	font := gui.NewQFontMetricsF(gui.NewQFont2(editor.config.Editor.FontFamily, editor.config.Editor.FontSize, 1, false))
	p.cursorX = int(font.HorizontalAdvance(string(p.patternText[:x]), -1))
	p.cursor.Move2(p.cursorX+p.patternPadding, p.patternPadding)
}

func (p *Palette) showSelected(selected int) {
	if p.resultType == "file_line" {
		n := 0
		for i := 0; i <= selected; i++ {
			for n++; n < len(p.itemTypes) && p.itemTypes[n] == "file"; n++ {
			}
		}
		selected = n
	}
	for i, resultItem := range p.resultItems {
		resultItem.setSelected(selected == i)
	}
}

func (f *PaletteResultItem) update() {
	c := editor.colors.selectedBg
	// transparent := editor.config.Editor.Transparent
	transparent := transparent()
	if f.selected {
		f.widget.SetStyleSheet(fmt.Sprintf(".QWidget {background-color: rgba(%d, %d, %d, %f);}", c.R, c.G, c.B, transparent))
	} else {
		f.widget.SetStyleSheet("")
	}
	f.p.widget.Hide()
	f.p.widget.Show()

}

func (f *PaletteResultItem) setSelected(selected bool) {
	if f.selected == selected {
		return
	}
	f.selected = selected
	f.update()
}

func (f *PaletteResultItem) show() {
	// if f.hidden {
	f.hidden = false
	f.widget.Show()
	// }
}

func (f *PaletteResultItem) hide() {
	if !f.hidden {
		f.hidden = true
		f.widget.Hide()
	}
}

func (f *PaletteResultItem) setItem(text string, itemType string, match []int) {
	iconType := ""
	path := false
	if itemType == "dir" {
		iconType = "folder"
		path = true
	} else if itemType == "file" {
		iconType = getFileType(text)
		path = true
	} else if itemType == "file_line" {
		iconType = "empty"
	}
	if iconType != "" {
		if iconType != f.iconType {
			f.iconType = iconType
			f.updateIcon()
		}
		f.showIcon()
	} else {
		f.hideIcon()
	}

	formattedText := formatText(text, match, path)
	if formattedText != f.baseText {
		f.baseText = formattedText
		f.base.SetText(f.baseText)
	}
}

func (f *PaletteResultItem) updateIcon() {
	svgContent := editor.getSvg(f.iconType, nil)
	f.icon.Load2(core.NewQByteArray2(svgContent, len(svgContent)))
}

func (f *PaletteResultItem) showIcon() {
	if f.iconHidden {
		f.iconHidden = false
		f.icon.Show()
	}
}

func (f *PaletteResultItem) hideIcon() {
	if !f.iconHidden {
		f.iconHidden = true
		f.icon.Hide()
	}
}
