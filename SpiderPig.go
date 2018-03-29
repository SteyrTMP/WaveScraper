package main

import (
	"fmt"
	"io"
	"net/url"
	"net/http"
	"log"
	"golang.org/x/net/html"
	"regexp"
	"os"
)

func main(){
	s,_ := url.Parse("http://www.thanatosrealms.com")
	seedUrl := *s
	var newUrls, oldUrls []url.URL
	newUrls = append(newUrls, seedUrl)

	for len(newUrls) > 0 {
		var currentUrl url.URL
		currentUrl, newUrls = newUrls[0], newUrls[1:]	//Pop the first entry in the list of URL's to visit
		fmt.Println("Scraping page: ", currentUrl.String())
		pageLinks := ScrapeLinks(currentUrl)	//Get all the URL's on target page
		fmt.Println("Found ",len(pageLinks), " links, processing...")
		for in,itm := range pageLinks {
			fmt.Println(itm.String())
			if itm.Hostname() != currentUrl.Hostname(){	//Validate inside domain
				fmt.Println("Link ",itm," is outside of the domain")
				continue
			}
			if findUrl(itm, oldUrls) {	//validate not processed
				fmt.Println("Link ",in," is already seen")
				continue
			}
			if findUrl(itm, newUrls) {	//validate not already queued
				fmt.Println("Link ",in," is already queued")
				continue
			}
			if !((itm.Scheme == "http") || (itm.Scheme == "https")) {	//validate not some javascript bullshit
				fmt.Println("Link ",itm," is not an http link")
				continue
			}
			res, _ := http.Head(itm.String())	//probe for header
			
			ctype := res.Header.Get("Content-Type")
			isHtml, _ := regexp.MatchString("html", ctype)
			if isHtml {
				ctype = "text/html"
			}
			
			switch ctype {
			case "":
				fmt.Println("Link ",itm," does not have a content type tag. Adding to queue anyway.")
				newUrls = append(newUrls,itm)
			case "audio/x-wav":	
				fmt.Println("Link ",itm," is an audio file, downloading...")
				DownloadFile(itm.EscapedPath()[1:],itm)	//download if wav
				fmt.Println("Download complete")
			case "text/html":
				fmt.Println("Link ",in," is a new html file, adding ",itm.String()," to queue")
				newUrls = append(newUrls, itm) //add to queue if valid html
			default:
				fmt.Println("Link ",itm," is a file of type ",res.Header.Get("Content-Type"),", ignoring")
			}
		}
		fmt.Println("done.\n")
		oldUrls = append(oldUrls, currentUrl)
		/*fmt.Println("URL's to explore: ")
		for in,itm := range newUrls {
			fmt.Println(in,": ",itm.String())
		}
		*/

	}

}

func findUrl(item url.URL, list []url.URL) bool {
	for _,itm := range list {
		if item == itm {
			return true
		}
	}
	return false
}

func ScrapeLinks(targetUrl url.URL) []url.URL {		//Gets all the links on a page, with correct hostnames, returning as a slice of url objects
	var output []url.URL
	res, err := http.Get(targetUrl.String())
	if err != nil {
		log.Fatal(err)
	}

	tokenized := html.NewTokenizer(res.Body)

	for {
	tokenType := tokenized.Next()

	switch {
	case tokenType == html.ErrorToken:
		fmt.Println("End of file")
		return output

	case tokenType == html.StartTagToken:
		currentToken := tokenized.Token()
		if (currentToken.Data == "a") {
			//fmt.Println("Link Found")
				for _, attribute := range currentToken.Attr {
					if attribute.Key == "href" {
						fmt.Println("URL=",attribute.Val)
						removeHashThingy := regexp.MustCompile(`#[^#]*$`)
						urlWithoutHashThingy := removeHashThingy.ReplaceAllString(attribute.Val, "")
						u, err := targetUrl.Parse(urlWithoutHashThingy)

						if err != nil{
							log.Fatal(err)
						}
						if u.Hostname() == ""{
							u.Host = targetUrl.Hostname()
						}
						if u.Scheme == ""{
							u.Scheme = "http"
						}
						output = append(output, *u)
						}

					}
				}
		}
	}

	return output

}

func DownloadFile(filepath string, targetUrl url.URL) error {

    // Create the file
	fmt.Println("Download function called on filepath ",filepath," with url ",targetUrl)
   	rootDirRegexp := regexp.MustCompile(`/[^/]*$`)
	rootDir := rootDirRegexp.ReplaceAllString(filepath, "")
	fmt.Println("making root dir ",rootDir)
	err := os.MkdirAll(rootDir, os.ModePerm)
	if err != nil {
		fmt.Println("error")
    	fmt.Println(err)
        return err
    }
	fmt.Println("making file ",filepath)
    out, err := os.Create(filepath)
    if err != nil {
    	fmt.Println("error")
    	fmt.Println(err)
        return err
    }
    defer out.Close()

    // Get the data
    fmt.Println("Getting data from ",targetUrl.String())
    resp, err := http.Get(targetUrl.String())
    if err != nil {
    	fmt.Println("error")
    	fmt.Println(err)
        return err
    }
    defer resp.Body.Close()

    // Write the body to file
    fmt.Println("Writing data to file...")
    _, err = io.Copy(out, resp.Body)
    if err != nil {
    	fmt.Println("error")
    	fmt.Println(err)
        return err
    }

    return nil
}
