package main

import (
	"fmt"
	// "os"
	"context"
	"log"

	"github.com/chromedp/chromedp"
)

func main() {
	fmt.Println("Welcome SVG Auto!")
	
	// fmt.Println(getFiles())
	accessIcoMoon()
}

func accessIcoMoon() {
	url := "https://icomoon.io/app/#/select"
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	err := chromedp.Run(ctx,
			chromedp.Navigate(url),
			chromedp.WaitVisible("body"),
	)

	if err != nil {
		log.Fatal(err)
	}
}

// func getFiles() []string {
// 	var args []string
// 	
// 	for _, arg := range os.Args[1:] {
// 		args = append(args, arg)
// 	}
// 	return args
// }

// this function should access icomoon.io

// func uploadFiles() {
// 	// this function should updload files to icommon.io
// }

// func downloadFiles() {
// 	// this function should download files from iconmoon.io
// }
