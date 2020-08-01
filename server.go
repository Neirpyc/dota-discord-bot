package main

import (
	"context"
	"github.com/chromedp/chromedp"
	"log"
	"net"
	"net/http"
	"strconv"
)

func myListenAndServer(s *http.Server, err chan<- error) {
	err <- s.ListenAndServe()
}

func serveDirectory(dir string, port chan<- string, stop <-chan bool) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	portStr := strconv.Itoa(listener.Addr().(*net.TCPAddr).Port)
	_ = listener.Close()
	server := &http.Server{Addr: "localhost:" + portStr, Handler: http.FileServer(http.Dir(dir))}
	errChan := make(chan error, 1)
	go myListenAndServer(server, errChan)
	port <- portStr
	select {
	case <-stop:
		_ = server.Close()
		L.Println("Closed server " + portStr)
	case err := <-errChan:
		L.Println(err)
	}
}

func screenshotFile(file string, selectors ...string) [][]byte {
	//todo reuse same context
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	stop := make(chan bool, 0)
	port := make(chan string, 0)
	go serveDirectory("assets/", port, stop)
	portStr := <-port
	buf := make([][]byte, len(selectors))
	for i, selector := range selectors {
		err := chromedp.Run(ctx, chromedp.EmulateViewport(4096, 2160),
			elementScreenshot("http://localhost:"+portStr+"/"+file, selector, &buf[i]))
		if err != nil {
			log.Println(err)
			return nil
		}
	}
	stop <- true
	return buf
}

func elementScreenshot(urlstr, sel string, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		chromedp.WaitVisible(sel, chromedp.ByID),
		chromedp.Screenshot(sel, res, chromedp.NodeVisible, chromedp.ByID),
	}
}
