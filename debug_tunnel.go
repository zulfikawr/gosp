package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"net/url"
)

func main() {
	target := "https://www.google.com/search?q=golang&gbv=1"
	tunnelURL := fmt.Sprintf("https://translate.google.com/translate?sl=en&tl=en&u=%s", url.QueryEscape(target))
	
	req, _ := http.NewRequest("GET", tunnelURL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36")
	
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	
	f, _ := os.Create("google_tunnel.html")
	io.Copy(f, resp.Body)
	fmt.Println("Saved google_tunnel.html")
}
