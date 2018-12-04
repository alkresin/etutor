package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"encoding/xml"
	egui "github.com/alkresin/external"
)

func main() {

	if egui.Init("") != 0 {
		return
	}

	pWindow := &egui.Widget{X: 100, Y: 100, W: 800, H: 400, Title: "External"}

	egui.InitMainWindow(pWindow)

	pWindow.AddWidget(&egui.Widget{Type: "browse", Name: "brw", X: 10, Y: 10, W: 780, H: 300,
		Anchor: egui.A_TOPABS + egui.A_BOTTOMABS + egui.A_LEFTABS + egui.A_RIGHTABS})

	pWindow.Activate()

	egui.Exit()

}

func getdata() {

	type Item struct {
		Title string  `xml:"title"`
		Link string  `xml:"link"`
		Date string  `xml:"pubDate"`
		Guid string  `xml:"guid"`
	}

	type Chan struct {
		Items []Item  `xml:"item"`
	}

	type Result struct {
		Channel Chan  `xml:"channel"`
	}

	response, err := http.Get("http://feeds.cbsnews.com/CBSNewsMain")
	if err != nil {
		fmt.Println(err)
	} else {

		htmlData, _ := ioutil.ReadAll(response.Body)
		if err != nil {
			fmt.Printf("ioutil.ReadAll error: %v", err)
			return
		}

		v := Result{}
		err = xml.Unmarshal(htmlData, &v)
		if err != nil {
			fmt.Printf("unmarshal error: %v", err)
			return
		}

		response.Body.Close()

		for _, p := range v.Channel.Items {
			fmt.Println(p.Date,p.Title)
		}
	}
}

