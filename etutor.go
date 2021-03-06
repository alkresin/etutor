package main

import (
	"encoding/xml"
	"fmt"
	egui "github.com/alkresin/external"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

const (
	CLR_BLACK    = 0
	CLR_WHITE    = 0xffffff
	CLR_BLUE     = 0xff0000
	CLR_GREEN    = 32768
	CLR_LGRAY0   = 0xeeeeee
	CLR_LGRAY1   = 0xbbbbbb
	CLR_LGRAY2   = 0x999999
	CLR_LGRAY4   = 0x666666
	CLR_LGRAY5   = 0x444444
	ES_MULTILINE = 4
)

const ITEM_TUTOR = 1001

type Module struct {
	Name     string `xml:"name,attr"`
	Path     string `xml:"path"`
	Code     string `xml:"code"`
	NodeName string
}

type Chapter struct {
	Name     string    `xml:"name,attr"`
	Module   []Module  `xml:"module"`
	Chapter  []Chapter `xml:"chapter"`
	Text     string    `xml:"text"`
	NodeName string
}

type Tutor struct {
	Chapter []Chapter `xml:"chapter"`
}

type Book struct {
	Name string `xml:"name,attr"`
	Path string `xml:"path,attr"`
}

type XHiOpt struct {
	TColor int32 `xml:"tcolor,attr"`
	BColor int32 `xml:"bcolor,attr"`
	Bold   bool  `xml:"bold,attr"`
	Italic bool  `xml:"italic,attr"`
}

type XHiOpts struct {
	Normal   XHiOpt `xml:"normal"`
	Command  XHiOpt `xml:"command"`
	Function XHiOpt `xml:"function"`
	Comment  XHiOpt `xml:"comment"`
	Quote    XHiOpt `xml:"quote"`
}

var pTutor *Tutor
var pBooks []Book
var iBookCurrent int
var sTutorErr = ""
var pFontMain, pFontCode, pFontResult *egui.Font
var pHilight *egui.Highlight
var pHiopts XHiOpts
var pResOpts XHiOpt

func main() {

	pTutor = &Tutor{}

	egui.BeginPacket()

	sInit := getini()
	sTutorErr = getxml(pBooks[0].Path, pTutor)

	if egui.Init(sInit) != 0 {
		return
	}

	egui.SetImagePath("images/")
	egui.CreateStyle(&egui.Style{Name: "st1", Orient: 1, Colors: []int32{CLR_WHITE, CLR_LGRAY1}})
	egui.CreateStyle(&egui.Style{Name: "st2", Colors: []int32{CLR_LGRAY1}, BorderW: 3})
	egui.CreateStyle(&egui.Style{Name: "st3", Colors: []int32{CLR_LGRAY1},
		BorderW: 2, BorderClr: CLR_LGRAY2})

	pWindow := &egui.Widget{X: 200, Y: 150, W: -800, H: -600, Title: "Go Tutor",
		Font: pFontMain, AProps: map[string]string{"Icon": "etutor"}}
	egui.InitMainWindow(pWindow)

	pPanel := pWindow.AddWidget(&egui.Widget{Type: "paneltop", Name: "pane", H: 40,
		AProps: map[string]string{"HStyle": "st1"}})

	// Buttons
	pPanel.AddWidget(&egui.Widget{Type: "ownbtn", Name: "btnm", X: 0, Y: 0, W: 40, H: 40,
		Anchor: egui.A_LEFTABS,
		AProps: map[string]string{"Image": "menu.bmp", "Transpa": "true", "HStyles": egui.ToString("st1", "st2", "st3")}})
	egui.PLastWidget.SetCallBackProc("onclick", fmenu, "fmenu")

	pPanel.AddWidget(&egui.Widget{Type: "ownbtn", Name: "btn+", X: 612, Y: 0, W: 30, H: 40, Title: "+",
		Anchor: egui.A_RIGHTABS,
		AProps: map[string]string{"HStyles": egui.ToString("st1", "st2", "st3")}})
	egui.PLastWidget.SetCallBackProc("onclick", fontincr, "fontincr")
	pPanel.AddWidget(&egui.Widget{Type: "ownbtn", Name: "btn-", X: 642, Y: 0, W: 30, H: 40, Title: "-",
		Anchor: egui.A_RIGHTABS,
		AProps: map[string]string{"HStyles": egui.ToString("st1", "st2", "st3")}})
	egui.PLastWidget.SetCallBackProc("onclick", fontdecr, "fontdecr")

	pPanel.AddWidget(&egui.Widget{Type: "ownbtn", Name: "btnfmt", X: 672, Y: 0, W: 60, H: 40, Title: "Fmt",
		Anchor: egui.A_RIGHTABS,
		AProps: map[string]string{"HStyles": egui.ToString("st1", "st2", "st3")}})
	egui.PLastWidget.SetCallBackProc("onclick", ffmt, "ffmt")
	egui.PLastWidget.Enable(false)

	pPanel.AddWidget(&egui.Widget{Type: "ownbtn", Name: "btnrun", X: 732, Y: 0, W: 60, H: 40, Title: "Run",
		Anchor: egui.A_RIGHTABS,
		AProps: map[string]string{"HStyles": egui.ToString("st1", "st2", "st3")}})
	egui.PLastWidget.SetCallBackProc("onclick", frun, "frun")
	egui.PLastWidget.Enable(false)

	// Tree
	pTree := pWindow.AddWidget(&egui.Widget{Type: "tree", Name: "tree",
		X: 0, Y: 40, W: 200, H: 560, Winstyle: egui.WS_VSCROLL, Anchor: egui.A_TOPABS + egui.A_BOTTOMABS,
		AProps: map[string]string{"AImages": egui.ToString("folder.bmp", "folderopen.bmp")}})
	//pTree.SetCallBackProc("onsize", nil, "{|o,x,y|o:Move(,,,y-40)}")

	// First code editor
	pEdi1 := pWindow.AddWidget(&egui.Widget{Type: "cedit", Name: "edi1",
		X: 204, Y: 40, W: 596, H: 360, Font: pFontCode, TColor: CLR_LGRAY5})
	pEdi1.SetCallBackProc("onsize", nil, "{|o,x,y|o:Move(,,x-o:nLeft)}")
	pEdi1.SetParam("lTabs", true)
	if pHiopts.Command.TColor != 0 || pHiopts.Command.BColor != 0 {
		egui.SetHiliOpt(pEdi1, egui.HILI_KEYW, nil, pHiopts.Command.TColor, pHiopts.Command.BColor)
	} else {
		egui.SetHiliOpt(pEdi1, egui.HILI_KEYW, nil, CLR_BLACK, CLR_WHITE)
	}
	if pHiopts.Function.TColor != 0 || pHiopts.Function.BColor != 0 {
		egui.SetHiliOpt(pEdi1, egui.HILI_FUNC, nil, pHiopts.Function.TColor, pHiopts.Function.BColor)
	}
	if pHiopts.Quote.TColor != 0 || pHiopts.Quote.BColor != 0 {
		egui.SetHiliOpt(pEdi1, egui.HILI_QUOTE, nil, pHiopts.Quote.TColor, pHiopts.Quote.BColor)
	} else {
		egui.SetHiliOpt(pEdi1, egui.HILI_QUOTE, nil, CLR_BLUE, CLR_WHITE)
	}
	if pHiopts.Comment.TColor != 0 || pHiopts.Comment.BColor != 0 {
		egui.SetHiliOpt(pEdi1, egui.HILI_COMM, nil, pHiopts.Comment.TColor, pHiopts.Comment.BColor)
	} else {
		egui.SetHiliOpt(pEdi1, egui.HILI_COMM, nil, CLR_GREEN, CLR_WHITE)
	}

	// Results edit
	pEdi2 := pWindow.AddWidget(&egui.Widget{Type: "cedit", Name: "edi2",
		X: 204, Y: 404, W: 596, H: 240, BColor: CLR_LGRAY0, Font: pFontResult})
	pEdi2.SetCallBackProc("onsize", nil, "{|o,x,y|o:Move(,,x-o:nLeft,y-o:nTop)}")
	pEdi2.SetParam("lReadOnly", true)
	if pResOpts.TColor != 0 || pResOpts.BColor != 0 {
		pEdi2.SetColor(pResOpts.TColor, pResOpts.BColor)
	} else {
		pEdi2.SetParam("bColorCur", CLR_LGRAY0)
	}
	pEdi2.SetParam("nMarginL", 8)

	pSpliH := pWindow.AddWidget(&egui.Widget{Type: "splitter", X: 204, Y: 400, W: 596, H: 4,
		AProps: map[string]string{"ALeft": egui.ToString(pEdi1), "ARight": egui.ToString(pEdi2)}})
	pSpliH.SetCallBackProc("onsize", nil, "{|o,x,y|o:Move(,,x-o:nLeft)}")

	pWindow.AddWidget(&egui.Widget{Type: "splitter", X: 200, Y: 40, W: 4, H: 560,
		AProps: map[string]string{"ALeft": egui.ToString(pTree), "ARight": egui.ToString(pEdi1, pSpliH, pEdi2)}})
	egui.PLastWidget.SetCallBackProc("onsize", nil, "{|o,x,y|o:Move(,,,y-40)}")

	if sTutorErr == "" {
		buildTree(pTree, pTutor.Chapter, "")
	}

	egui.MenuContext("mm")
	{
		egui.AddMenuItem("About", 0, fabout, "fabout")
		egui.AddMenuSeparator()
		for i, p := range pBooks {
			egui.AddCheckMenuItem(p.Name, ITEM_TUTOR+i, fSetTutor, "fsettutor", strconv.Itoa(i))
		}
		egui.AddMenuSeparator()
		egui.AddMenuItem("Exit", 0, nil, "hwg_EndWindow()")
	}
	egui.EndMenu()
	egui.MenuItemCheck("", "mm", ITEM_TUTOR, true)

	egui.EndPacket()

	if sTutorErr == "" {
		fldOnClick([]string{"", "n0"})
	} else {
		pEdi1.SetText(sTutorErr)
	}

	pWindow.Activate()
	egui.Exit()
}

func isFileExists(sPath string) bool {
	if _, err := os.Stat(sPath); os.IsNotExist(err) {
		return false
	}
	return true
}

func buildTree(pTree *egui.Widget, pChapter []Chapter, sPrefix string) {

	for iChap, _ := range pChapter {
		pChap := &(pChapter[iChap])
		sNodeName := sPrefix + "n" + strconv.Itoa(iChap)
		pChap.NodeName = sNodeName
		egui.InsertNode(pTree, sPrefix, sNodeName, pChap.Name, "", nil, fldOnClick, "fldOnClick")
		if len(pChap.Chapter) > 0 {
			buildTree(pTree, pChap.Chapter, sNodeName)
		}
		for iMod, _ := range pChap.Module {
			pMod := &(pChap.Module[iMod])
			sModName := sNodeName + "m" + strconv.Itoa(iMod)
			pMod.NodeName = sModName
			egui.InsertNode(pTree, sNodeName, sModName, pMod.Name, "", []string{"book.bmp"}, nodeOnClick, "nodeOnClick")
			if pMod.Path != "" {
			} else if pMod.Code != "" {
			}
		}
	}
}

func getMod(pChapter []Chapter, sModName string) string {

	for _, oChap := range pChapter {
		for _, oMod := range oChap.Module {
			if oMod.NodeName == sModName {
				if oMod.Path != "" {
					b, err := ioutil.ReadFile(oMod.Path)
					if err != nil {
						return fmt.Sprintf("error reading file: %v", err)
					} else {
						return string(b)
					}
				} else {
					return oMod.Code
				}
			}
		}
		if len(oChap.Chapter) > 0 {
			if sCode := getMod(oChap.Chapter, sModName); sCode != "" {
				return sCode
			}
		}
	}
	return ""
}

func getText(pChapter []Chapter, sChapName string) string {

	for _, oChap := range pChapter {
		if oChap.NodeName == sChapName {
			return oChap.Text
		}
		if len(oChap.Chapter) > 0 {
			if sText := getText(oChap.Chapter, sChapName); sText != "" {
				return sText
			}
		}
	}
	return ""
}

func getini() string {

	type XFont struct {
		Family string `xml:"family,attr"`
		Height int    `xml:"height,attr"`
	}

	type XHili struct {
		Keywords  string `xml:"keywords"`
		Functions string `xml:"functions"`
		SLcomm    string `xml:"single_line_comment"`
		MLcomm    string `xml:"multi_line_comment"`
	}

	type Ini struct {
		Guiserver  string  `xml:"guiserver"`
		Ipaddr     string  `xml:"ip"`
		Port       int     `xml:"port"`
		Log        int     `xml:"log"`
		FontMain   XFont   `xml:"fontmain"`
		FontCode   XFont   `xml:"fontcode"`
		FontResult XFont   `xml:"fontresult"`
		Books      []Book  `xml:"book"`
		Hili       XHili   `xml:"hilighter"`
		Hiliopt    XHiOpts `xml:"hiliopt"`
		Results    XHiOpt  `xml:"results"`
	}

	var pIni = &Ini{}
	var sInit = ""

	// Check, is a current directory the same, where etutor's files are placed.
	// If no, try to change it to that one where executable is.
	sCurrDir, _ := os.Getwd()
	if !isFileExists(sCurrDir + "/etutor.ini") {
		ex, _ := os.Executable()
		sCurrDir := filepath.Dir(ex)
		if isFileExists(sCurrDir + "/etutor.ini") {
			os.Chdir(sCurrDir)
		}
	}

	getxml("etutor.ini", pIni)

	if pIni.Guiserver != "" {
		sInit = "guiserver=" + pIni.Guiserver + "\n"
	}
	if pIni.Port != 0 {
		sInit += fmt.Sprintf("port=%d\n", pIni.Port)
	}
	if pIni.Log == 1 {
		sInit += "log=1\n"
	} else if pIni.Log == 2 {
		sInit += "log=2\n"
	}

	if pIni.FontMain.Family != "" {
		pFontMain = egui.CreateFont(&egui.Font{Name: "fm",
			Family: pIni.FontMain.Family, Height: pIni.FontMain.Height})
	} else {
		pFontMain = egui.CreateFont(&egui.Font{Name: "fm", Family: "Courier New", Height: -19})
	}
	if pIni.FontCode.Family != "" {
		pFontCode = egui.CreateFont(&egui.Font{Name: "fc",
			Family: pIni.FontCode.Family, Height: pIni.FontCode.Height})
	} else {
		pFontCode = pFontMain
	}
	if pIni.FontResult.Family != "" {
		pFontResult = egui.CreateFont(&egui.Font{Name: "fc",
			Family: pIni.FontResult.Family, Height: pIni.FontResult.Height})
	} else {
		pFontResult = pFontMain
	}

	pBooks = pIni.Books
	if len(pBooks) == 0 {
		pBooks = append(pBooks, Book{Name: "Main tutorial", Path: "etutor.xml"})
	}

	if pIni.Hili.Keywords != "" {
		pHilight = egui.CreateHighliter("higo", pIni.Hili.Keywords, "",
			pIni.Hili.SLcomm, pIni.Hili.MLcomm, true)
	} else {
		pHilight = egui.CreateHighliter("higo", "package import func { }", "", "//", "/* */", true)
	}

	pHiopts = pIni.Hiliopt
	pResOpts = pIni.Results

	return sInit
}

func getxml(sPath string, pXml interface{}) string {

	data, err := ioutil.ReadFile(sPath)
	if err != nil {
		return fmt.Sprintf("error reading file: %v", err)
	}

	err = xml.Unmarshal([]byte(data), pXml)
	if err != nil {
		return fmt.Sprintf("unmarshal error: %v", err)
	}
	return ""
}

func fldOnClick(p []string) string {

	pEdi1 := egui.Widg("main.edi1")
	sText := getText(pTutor.Chapter, p[1])
	egui.BeginPacket()
	pEdi1.SetColor(CLR_BLACK, CLR_WHITE)
	pEdi1.SetParam("bColorCur", CLR_WHITE)
	egui.EvalProc("Widg('main.edi1'):SetWrap(.T.)")
	egui.SetHighliter(pEdi1, nil)
	pEdi1.SetText(sText)
	egui.Widg("main.pane.btnfmt").Enable(false)
	egui.Widg("main.pane.btnrun").Enable(false)
	egui.EndPacket()
	return ""
}

func nodeOnClick(p []string) string {

	pEdi1 := egui.Widg("main.edi1")
	sCode := getMod(pTutor.Chapter, p[1])
	egui.BeginPacket()
	if pHiopts.Normal.TColor != 0 || pHiopts.Normal.BColor != 0 {
		pEdi1.SetColor(pHiopts.Normal.TColor, pHiopts.Normal.BColor)
		pEdi1.SetParam("bColorCur", pHiopts.Normal.BColor)
	} else {
		pEdi1.SetColor(CLR_LGRAY5, CLR_WHITE)
		pEdi1.SetParam("bColorCur", CLR_WHITE)
	}
	egui.EvalProc("Widg('main.edi1'):SetWrap(.F.)")
	egui.SetHighliter(pEdi1, pHilight)
	pEdi1.SetText(sCode)
	egui.Widg("main.pane.btnfmt").Enable(true)
	egui.Widg("main.pane.btnrun").Enable(true)
	egui.EndPacket()
	return ""
}

func frun([]string) string {

	pEdi2 := egui.Widg("main.edi2")
	pEdi2.SetText("Wait...")

	sCode := egui.Widg("main.edi1").GetText()

	if len(sCode) < 32 {
		pEdi2.SetText("Nothing to run.")
		return ""
	}

	f, err := os.OpenFile("tmp.go", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		pEdi2.SetText(fmt.Sprintf("%v", err))
		return ""
	}
	f.WriteString(sCode)
	f.Close()

	cmd := exec.Command("go", "run", "tmp.go")
	result, err := cmd.CombinedOutput()
	if err != nil {
		pEdi2.SetText(string(result) + fmt.Sprintf("%v", err))
		return ""
	}
	pEdi2.SetText(string(result))

	return ""
}

func ffmt([]string) string {

	pEdi1 := egui.Widg("main.edi1")
	pEdi2 := egui.Widg("main.edi2")
	sCode := pEdi1.GetText()

	if len(sCode) < 32 {
		return ""
	}

	ioutil.WriteFile("tmp.go", []byte(sCode), 0600)

	cmd := exec.Command("gofmt", "-w", "tmp.go")
	result, err := cmd.Output()
	if err != nil {
		pEdi2.SetText(fmt.Sprintf("%v", err))
		return ""
	}
	pEdi2.SetText(string(result))

	b, err := ioutil.ReadFile("tmp.go")
	pEdi1.SetText(string(b))

	return ""
}

func fontincr([]string) string {
	height := pFontCode.Height
	if height > 0 {
		height += 2
	} else {
		height -= 2
	}
	pFontCode = egui.CreateFont(&egui.Font{Family: pFontCode.Family, Height: height})
	egui.Widg("main.edi1").SetFont(pFontCode)
	return ""
}

func fontdecr([]string) string {
	height := pFontCode.Height
	if height > 0 {
		height -= 2
	} else {
		height += 2
	}
	pFontCode = egui.CreateFont(&egui.Font{Family: pFontCode.Family, Height: height})
	egui.Widg("main.edi1").SetFont(pFontCode)
	return ""
}

func fmenu([]string) string {

	egui.ShowMenuContext("mm", egui.PLastWindow)
	return ""
}

func fabout([]string) string {

	sVer := "Golang Tutorial\r\nVersion 1.1\r\n(C) Alexander S.Kresin\r\n\r\n" + egui.GetVersion(2)
	egui.MsgInfo(sVer, "About", nil, "", "")
	return ""
}

func fSetTutor(p []string) string {

	i, _ := strconv.Atoi(p[0])
	if i != iBookCurrent {
		egui.MenuItemCheck("", "mm", ITEM_TUTOR+iBookCurrent, false)
		egui.MenuItemCheck("", "mm", ITEM_TUTOR+i, true)
		iBookCurrent = i
		egui.EvalProc("Widg('main.tree'):Clean()")
		pTutor = &Tutor{}
		sTutorErr = getxml(pBooks[i].Path, pTutor)
		if sTutorErr == "" {
			egui.BeginPacket()
			buildTree(egui.Widg("main.tree"), pTutor.Chapter, "")
			egui.EndPacket()
			fldOnClick([]string{"", "n0"})
		} else {
			egui.Widg("main.edi1").SetText(sTutorErr)
		}
	}
	return ""
}
