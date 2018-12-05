// This sample shows a browse (or a table, if you want) - a most complicated widget.
// We will get data from a RSS feed of CBS to represent it in a browse.
// For to not disturb CDS every time we save a received xml file as _cbs.xml to use
// it while a next program launch.
package main

import (
	"encoding/xml"
	"fmt"
	egui "github.com/alkresin/external"
	"io/ioutil"
	"net/http"
	"os"
)

var sErr string
var pArr [][]string

const (
	CLR_LGRAY1 = 0xeeeeee
	CLR_LGRAY2 = 0xdddddd
)

func main() {

	var arr = [][]string{{"", "", "Waiting for data..."}}

	if egui.Init("log=0") != 0 {
		return
	}

	// launch a goroutine to get data from a RSS feed or from a saved file
	go getdata()

	pFont := egui.CreateFont(&egui.Font{Name: "f1", Family: "Georgia", Height: 16})
	pWindow := &egui.Widget{X: 100, Y: 100, W: -800, H: -400, Title: "RSS browse", Font: pFont}

	egui.InitMainWindow(pWindow)

	// Adding of a browse widget
	pBrw := pWindow.AddWidget(&egui.Widget{Type: "browse", Name: "brw", X: 10, Y: 10, W: 780, H: 320,
		Anchor: egui.A_TOPABS + egui.A_BOTTOMABS + egui.A_LEFTABS + egui.A_RIGHTABS})

	// Colors setting
	pBrw.SetParam("bColorSel", CLR_LGRAY2)
	pBrw.SetParam("htbColor", CLR_LGRAY2)
	pBrw.SetParam("tColorSel", 0)
	pBrw.SetParam("httColor", 0)

	// Setting of an initial info
	egui.BrwSetArray(pBrw, &arr)

	// Columns setting
	egui.BrwSetColumn(pBrw, 1, "Date", 1, 0, false, 14)
	egui.BrwSetColumn(pBrw, 2, "Time", 1, 0, false, 12)
	egui.BrwSetColumn(pBrw, 3, "Title", 0, 0, false, 0)

	pWindow.AddWidget(&egui.Widget{Type: "button", X: 350, Y: 350, W: 100, H: 32, Title: "Ok",
		Anchor: egui.A_BOTTOMABS + egui.A_LEFTABS + egui.A_RIGHTABS})
	egui.PLastWidget.SetCallBackProc("onclick", nil, "hwg_EndWindow()")

	pWindow.Activate()

	egui.Exit()

}

func getdata() {

	var htmlData []byte
	var err error
	var bGetData = false

	type Item struct {
		Title string `xml:"title"`
		Link  string `xml:"link"`
		Date  string `xml:"pubDate"`
		Guid  string `xml:"guid"`
	}

	type Chan struct {
		Items []Item `xml:"item"`
	}

	type Result struct {
		Channel Chan `xml:"channel"`
	}

	if isFileExists("_cbs.xml") {
		// If a _cbs.xml exists, we read an info from it
		htmlData, err = ioutil.ReadFile("_cbs.xml")
		if err == nil {
			bGetData = true
		}
	}

	if !bGetData {
		// If no, get it from the CBS RSS feed
		response, err := http.Get("http://feeds.cbsnews.com/CBSNewsMain")
		if err != nil {
			sErr = fmt.Sprintln(err)
			return
		} else {
			htmlData, err = ioutil.ReadAll(response.Body)
			response.Body.Close()
			if err != nil {
				return
			}
			// Save the data as _cbs.xml
			ioutil.WriteFile("_cbs.xml", htmlData, 0644)
			bGetData = true
		}
	}

	if bGetData {
		// if the data is receved, we convert it to xml structures
		v := Result{}
		err := xml.Unmarshal(htmlData, &v)
		if err != nil {
			sErr = fmt.Sprintln(err)
			return
		}

		// and create a slice with received information for browsing
		for _, p := range v.Channel.Items {
			pArr = append(pArr, []string{p.Date[5:16], p.Date[17:25], p.Title})
		}
		// Set the function, which will put data to the browse after main window activation
		egui.AddFuncToIdle(setBrowse)
	}
}

func isFileExists(sPath string) bool {
	if _, err := os.Stat(sPath); os.IsNotExist(err) {
		return false
	}
	return true
}

func setBrowse() {

	pBrw := egui.Widg("main.brw")
	egui.BrwSetArray(pBrw, &pArr)
}
