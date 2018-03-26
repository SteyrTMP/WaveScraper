package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"golang.org/x/net/html"
	"strings"
	"regexp"
	"os"
)

func main() {
	// Compile the expression once, usually at init time.
	// Use raw strings to avoid having to quote the backslashes.
	// 	var validID = regexp.MustCompile(`^[a-z]+\[[0-9]+\]$`)

	getMp3Files("http://www.thanatosrealms.com/war2/horde-sounds")

}

func getMp3Files(url string) {
	rootUrlRegexp := regexp.MustCompile(`/[^/]*$`)
	rootUrl := rootUrlRegexp.ReplaceAllString(url, "")
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	tokenized := html.NewTokenizer(res.Body)
	for {
	tokenType := tokenized.Next()

	switch {
	case tokenType == html.ErrorToken:
		fmt.Println("End of file")
		return

	case tokenType == html.StartTagToken:
		currentToken := tokenized.Token()
		if (currentToken.Data == "a") {
			fmt.Println("Link Found")
				for _, attribute := range currentToken.Attr {
					if attribute.Key == "href" {
						fmt.Println("URL=",attribute.Val)
						match, _ := regexp.MatchString("wav$", attribute.Val)
						if match {
							fmt.Println("It's an MP3!")
							urlToJoin := [2]string{rootUrl, attribute.Val}
							var sliceJoin []string = urlToJoin[0:2]
							mp3Url := strings.Join(sliceJoin, "/")
							fmt.Println("Getting MP3 file at ", mp3Url)
							err := DownloadFile(attribute.Val, mp3Url)
							if err != nil {
								log.Fatal(err)
							}
						}

					}
				}
		}
	}
	}

}

func DownloadFile(filepath string, url string) error {

    // Create the file
   	rootDirRegexp := regexp.MustCompile(`/[^/]*$`)
	rootDir := rootDirRegexp.ReplaceAllString(filepath, "")
	os.MkdirAll(rootDir, os.ModePerm)
    out, err := os.Create(filepath)
    if err != nil {
        return err
    }
    defer out.Close()

    // Get the data
    resp, err := http.Get(url)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    // Write the body to file
    _, err = io.Copy(out, resp.Body)
    if err != nil {
        return err
    }

    return nil
}