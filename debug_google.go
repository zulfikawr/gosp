package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func main() {
	req, _ := http.NewRequest("GET", "https://www.google.com/search?q=golang&udm=14", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)")
	req.Header.Set("Cookie", "CONSENT=YES+cb.20210328-17-p0.en+FX+410")
	
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	
	f, _ := os.Create("google_debug.html")
	io.Copy(f, resp.Body)
	fmt.Println("Saved google_debug.html")
}
