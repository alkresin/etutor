package main

import (
	"encoding/xml"
	"fmt"
	egui "github.com/alkresin/external"
	"io/ioutil"
	"os"
	"os/exec"
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
	CLR_LGRAY5   = 0x333333
	ES_MULTILINE = 4
)

type Module struct {
	Name     string `xml:"name,attr"`
	Path     string `xml:"path"`
	Code     string `xml:"code"`
	NodeName string
}

type Chapter struct {
	Name    string    `xml:"name,attr"`
	Module  []Module  `xml:"module"`
	Chapter []Chapter `xml:"chapter"`
	Text    string    `xml:"text"`
	NodeName string
}

type Tutor struct {
	Chapter []Chapter `xml:"chapter"`
}

type Book struct {
	Name     string `xml:"name,attr"`
	Path     string `xml:"path,attr"`
}

var pTutor *Tutor
var pBooks []Book
var iBookCurrent int
var pFontMain, pFontCode, pFontResult *egui.Font
var pHilight *egui.Highlight

func main() {

	pTutor = &Tutor{}

	egui.BeginPacket()

	sInit := getini()
	if getxml(pBooks[0].Path, pTutor) == false {
		return
	}

	if egui.Init(sInit) != 0 {
		return
	}

	egui.SetImagePath("images/")
	egui.CreateStyle(&(egui.Style{Name: "st1", Orient: 1, Colors: []int32{CLR_WHITE, CLR_LGRAY1}}))
	egui.CreateStyle(&(egui.Style{Name: "st2", Colors: []int32{CLR_LGRAY1}, BorderW: 3}))
	egui.CreateStyle(&(egui.Style{Name: "st3", Colors: []int32{CLR_LGRAY1},
		BorderW: 2, BorderClr: 	CLR_LGRAY2}))

	pWindow := &(egui.Widget{X: 100, Y: 100, W: 800, H: 600, Title: "Go Tutor",
		Font: pFontMain, AProps: map[string]string{"Icon": "etutor"}})
	egui.InitMainWindow(pWindow)

	pPanel := pWindow.AddWidget(&(egui.Widget{Type: "paneltop", Name: "pane", H: 40,
		AProps: map[string]string{"HStyle": "st1"}}))

	// Buttons
	pPanel.AddWidget(&(egui.Widget{Type: "ownbtn", Name: "btnm", X: 0, Y: 0, W: 40, H: 40,
		Anchor: egui.A_LEFTABS,
		AProps: map[string]string{"Image": "menu.bmp", "Transpa": "true", "HStyles": egui.ToString("st1", "st2", "st3")}}))
	egui.PLastWidget.SetCallBackProc("onclick", fmenu, "fmenu")

	pPanel.AddWidget(&(egui.Widget{Type: "ownbtn", Name: "btn+", X: 620, Y: 0, W: 30, H: 40, Title: "+",
		Anchor: egui.A_RIGHTABS,
		AProps: map[string]string{"HStyles": egui.ToString("st1", "st2", "st3")}}))
	egui.PLastWidget.SetCallBackProc("onclick", fontincr, "fontincr")
	pPanel.AddWidget(&(egui.Widget{Type: "ownbtn", Name: "btn-", X: 650, Y: 0, W: 30, H: 40, Title: "-",
		Anchor: egui.A_RIGHTABS,
		AProps: map[string]string{"HStyles": egui.ToString("st1", "st2", "st3")}}))
	egui.PLastWidget.SetCallBackProc("onclick", fontdecr, "fontdecr")

	pPanel.AddWidget(&(egui.Widget{Type: "ownbtn", Name: "btnfmt", X: 680, Y: 0, W: 60, H: 40, Title: "Fmt",
		Anchor: egui.A_RIGHTABS,
		AProps: map[string]string{"HStyles": egui.ToString("st1", "st2", "st3")}}))
	egui.PLastWidget.SetCallBackProc("onclick", ffmt, "ffmt")
	egui.PLastWidget.Enable( false )

	pPanel.AddWidget(&(egui.Widget{Type: "ownbtn", Name: "btnrun", X: 740, Y: 0, W: 60, H: 40, Title: ">",
		Anchor: egui.A_RIGHTABS,
		AProps: map[string]string{"HStyles": egui.ToString("st1", "st2", "st3")}}))
	egui.PLastWidget.SetCallBackProc("onclick", frun, "frun")
	egui.PLastWidget.Enable( false )

	// Tree
	pTree := pWindow.AddWidget(&(egui.Widget{Type: "tree",
		X: 0, Y: 40, W: 200, H: 340,
		AProps: map[string]string{"AImages": egui.ToString("folder.bmp", "folderopen.bmp")}}))
	pTree.SetCallBackProc("onsize", nil, "{|o,x,y|o:Move(,,,y-40)}")

	// First code editor
	pHilight = egui.CreateHighliter( "higo", "package import func { }", "", "//", "/* */", true )
	pEdi1 := pWindow.AddWidget(&(egui.Widget{Type: "cedit", Name: "edi1",
		X: 204, Y: 40, W: 596, H: 360, Font: pFontCode, TColor: CLR_LGRAY5}))
	pEdi1.SetCallBackProc("onsize", nil, "{|o,x,y|o:Move(,,x-o:nLeft)}")
	pEdi1.SetParam("lTabs", true)
	egui.SetHiliOpt( pEdi1, egui.HILI_KEYW, nil, CLR_BLACK, CLR_WHITE )
	egui.SetHiliOpt( pEdi1, egui.HILI_QUOTE, nil, CLR_BLUE, CLR_WHITE )
	egui.SetHiliOpt( pEdi1, egui.HILI_COMM, nil, CLR_GREEN, CLR_WHITE )

	// Results edit
	pEdi2 := pWindow.AddWidget(&(egui.Widget{Type: "cedit", Name: "edi2",
		X: 204, Y: 404, W: 596, H: 240, BColor: CLR_LGRAY0, Font: pFontResult}))
	pEdi2.SetCallBackProc("onsize", nil, "{|o,x,y|o:Move(,,x-o:nLeft,y-o:nTop)}")
	pEdi2.SetParam("lReadOnly", true)
	pEdi2.SetParam("bColorCur", CLR_LGRAY0)
	pEdi2.SetParam("nMarginL", 8)

	pSpliH := pWindow.AddWidget(&(egui.Widget{Type: "splitter", X: 204, Y: 400, W: 596, H: 4,
		AProps: map[string]string{"ALeft": egui.ToString(pEdi1), "ARight": egui.ToString(pEdi2)}}))
	pSpliH.SetCallBackProc("onsize", nil, "{|o,x,y|o:Move(,,x-o:nLeft)}")

	pWindow.AddWidget(&(egui.Widget{Type: "splitter", X: 200, Y: 40, W: 4, H: 560,
		AProps: map[string]string{"ALeft": egui.ToString(pTree), "ARight": egui.ToString(pEdi1, pSpliH, pEdi2)}}))
	egui.PLastWidget.SetCallBackProc("onsize", nil, "{|o,x,y|o:Move(,,,y-40)}")

	buildTree(pTree, pTutor.Chapter, "")

	egui.MenuContext("mm")
	{
		egui.AddMenuItem("About", fabout, "fabout")
		egui.AddMenuSeparator()
		egui.AddMenuItem("Exit", nil, "hwg_EndWindow()")
	}
	egui.EndMenu()

	egui.EndPacket()

	fldOnClick([]string{"","n0"})
	pWindow.Activate()
	egui.Exit()
}

func buildTree(pTree *egui.Widget, pChapter []Chapter, sPrefix string) {

	for iChap, _ := range pChapter {
		pChap := &(pChapter[iChap])
		sNodeName := sPrefix + "n" + strconv.Itoa(iChap)
		pChap.NodeName = sNodeName
		egui.InsertNode(pTree, sPrefix, sNodeName, pChap.Name, "", nil, fldOnClick, "fldOnClick")
		for iMod, _ := range pChap.Module {
			pMod := &(pChap.Module[iMod])
			sModName := sNodeName + "m" + strconv.Itoa(iMod)
			pMod.NodeName = sModName
			egui.InsertNode(pTree, sNodeName, sModName, pMod.Name, "", []string{"book.bmp"}, nodeOnClick, "nodeOnClick")
			if pMod.Path != "" {
			} else if pMod.Code != "" {
			}
		}
		if len(pChap.Chapter) > 0 {
			buildTree(pTree, pChap.Chapter, sNodeName)
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
		Family     string `xml:"family,attr"`
		Height     int `xml:"height,attr"`
	}

	type Ini struct {
		Guiserver  string `xml:"guiserver"`
		Ipaddr     string `xml:"ip"`
		Port       int    `xml:"port"`
		Log        int    `xml:"log"`
		FontMain   XFont  `xml:"fontmain"`
		FontCode   XFont  `xml:"fontcode"`
		FontResult XFont  `xml:"fontresult"`
		Books      []Book `xml:"book"`
	}

	var pIni = &Ini{}
	var sInit = ""

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
		pFontMain = egui.CreateFont(&(egui.Font{Name: "fm",
			Family: pIni.FontMain.Family, Height: pIni.FontMain.Height }))
	} else {
		pFontMain = egui.CreateFont(&(egui.Font{Name: "fm", Family: "Courier New", Height: -19}))
	}
	if pIni.FontCode.Family != "" {
		pFontCode = egui.CreateFont(&(egui.Font{Name: "fc",
			Family: pIni.FontCode.Family, Height: pIni.FontCode.Height }))
	} else {
		pFontCode = pFontMain
	}
	if pIni.FontResult.Family != "" {
		pFontResult = egui.CreateFont(&(egui.Font{Name: "fc",
			Family: pIni.FontResult.Family, Height: pIni.FontResult.Height }))
	} else {
		pFontResult = pFontMain
	}

	pBooks = pIni.Books
	if len(pBooks) == 0 {
		pBooks = append( pBooks, Book{ Name: "Main tutorial", Path: "etutor.xml" } )
	}

	return sInit
}

func getxml(sPath string, pXml interface{}) bool {

	data, err := ioutil.ReadFile(sPath)
	if err != nil {
		fmt.Printf("error reading file: %v", err)
		return false
	}

	err = xml.Unmarshal([]byte(data), pXml)
	if err != nil {
		fmt.Printf("unmarshal error: %v", err)
		return false
	}
	return true
}

func fldOnClick(p []string) string {

	pEdi1 := egui.Widg("main.edi1")
	sText := getText(pTutor.Chapter, p[1])
	egui.BeginPacket()
	egui.EvalProc( "Widg('main.edi1'):SetWrap(.T.)" )
	egui.SetHighliter( pEdi1, nil )
	pEdi1.SetText(sText)
	egui.Widg("main.pane.btnfmt").Enable( false )
	egui.Widg("main.pane.btnrun").Enable( false )
	egui.EndPacket()
	return ""
}

func nodeOnClick(p []string) string {

	pEdi1 := egui.Widg("main.edi1")
	sCode := getMod(pTutor.Chapter, p[1])
	egui.BeginPacket()
	egui.EvalProc( "Widg('main.edi1'):SetWrap(.F.)" )
	egui.SetHighliter( pEdi1, pHilight )
	pEdi1.SetText(sCode)
	egui.Widg("main.pane.btnfmt").Enable( true )
	egui.Widg("main.pane.btnrun").Enable( true )
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
		pEdi2.SetText(fmt.Sprintf("%v",err))
		return ""
	}
	f.WriteString(sCode)
	f.Close()

	cmd := exec.Command("go", "run", "tmp.go")
	result, err := cmd.CombinedOutput()
	if err != nil {
		pEdi2.SetText(string(result) + fmt.Sprintf("%v",err))
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
		pEdi2.SetText(fmt.Sprintf("%v",err))
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
	pFontCode = egui.CreateFont(&(egui.Font{Family: pFontCode.Family, Height: height}))
	egui.Widg("main.edi1").SetFont( pFontCode )
	return ""
}

func fontdecr([]string) string {
	height := pFontCode.Height
	if height > 0 {
		height -= 2
	} else {
		height += 2
	}
	pFontCode = egui.CreateFont(&(egui.Font{Family: pFontCode.Family, Height: height}))
	egui.Widg("main.edi1").SetFont( pFontCode )
	return ""
}

func fmenu([]string) string {

	egui.ShowMenuContext("mm", egui.PLastWindow)
	return ""
}

func fabout([]string) string {

	sVer := "Golang Tutorial\r\nVersion 1.0\r\n(C) Alexander S.Kresin\r\n\r\n" + egui.GetVersion(2)
	egui.MsgInfo( sVer, "About", "", nil, "" )
	return ""
}
