package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/jessevdk/go-flags"
	fasthttp "github.com/valyala/fasthttp"
)

type arcgisResult struct {
	Query     string        `json:"query"`
	Total     int           `json:"total"`
	Start     int           `json:"start"`
	Num       int           `json:"num"`
	NextStart int           `json:"nextStart"`
	Groups    []arcgisGroup `json:"results,omitempty"`
	Users     []arcgisUser  `json:"users,omitempty"`
}

type arcgisGroup struct {
	ID       string `json:"id"`
	Title    string `json:"title,omitempty"`
	Owner    string `json:"owner"`
	Created  int64  `json:"created,omitempty"`
	Modified int64  `json:"modified,omitempty" `
}

type arcgisUser struct {
	Username   string `json:"username"`
	Fullname   string `json:"fullName"`
	MemberType string `json:"memberType"`
	Joined     int64  `json:"joined"`
}

// CLI options
var opts struct {
	Threads   int    `short:"t" long:"threads" default:"40" description:"Number of concurrent threads"`
	RestURI   string `short:"u" long:"url" required:"true" description:"ArcGIS Portal REST URL (https://maps.company.net/portal/sharing/rest/)"`
	UserAgent string `short:"a" long:"agent" default:"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2227.0 Safari/537.36"`
	Verbose   bool   `short:"v" long:"verbose" description:"Turns on verbose logging"`
	Esri      bool   `short:"e" long:"esri" description:"Will also print default ESRI (esri_*) users"`
}

func parseResults(body []byte) (*arcgisResult, error) {
	var s = new(arcgisResult)
	err := json.Unmarshal(body, &s)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERR: was not able to unmarshall", err)
	}
	return s, err
}

func main() {
	_, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(1)
	}
	var result *arcgisResult
	// Create one webclient
	client := &fasthttp.Client{
		MaxConnsPerHost:               1024,
		DisableHeaderNamesNormalizing: true,
		MaxConnWaitTimeout:            40 * time.Second,
		ReadTimeout:                   30 * time.Second,
		NoDefaultUserAgentHeader:      true,
		TLSConfig:                     &tls.Config{InsecureSkipVerify: true},
	}
	numWorkers := opts.Threads
	work := make(chan arcgisGroup)
	go func() {
		url := fmt.Sprintf("%s/community/groups?f=json&q=access:public&sortField=title&sortOrder=&num=100", opts.RestURI)
		req := fasthttp.AcquireRequest()
		resp := fasthttp.AcquireResponse()
		req.Header.SetUserAgent(opts.UserAgent)
		req.Header.Set(fasthttp.HeaderAccept, "application/json")
		req.Header.SetMethod(fasthttp.MethodGet)
	redo:
		req.SetRequestURI(url)
		err := client.Do(req, resp)
		if err != nil {
			fmt.Fprintln(os.Stderr, "ERR: errors occurred during request of ", url, "Error:", err, "continueing")
		}
		result, _ = parseResults([]byte(resp.Body())) //don't care about errors
		fasthttp.ReleaseResponse(resp)
		if result != nil {
			for i := range result.Groups {
				work <- result.Groups[i]
			}
			// json returns nextStart if there are more items, if we are at last it shows -1
			if result.NextStart != -1 {
				url = fmt.Sprintf("%s/community/groups?f=json&q=access:public&sortField=title&sortOrder=&num=100&start=%d", opts.RestURI, result.NextStart)
				goto redo //re-do the steps with updated URL
			}
		}
		fasthttp.ReleaseRequest(req)
		close(work)
	}()
	// Create a waiting group
	wg := &sync.WaitGroup{}
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go doWork(work, wg, client) //Schedule the work
	}
	wg.Wait() //Wait for it all to complete
}

func doWork(work chan arcgisGroup, wg *sync.WaitGroup, wc *fasthttp.Client) {
	defer wg.Done()
	//It is unsafe using Request object from concurrently running goroutines, even for marshaling the request.
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	req.Header.SetUserAgent(opts.UserAgent)
	req.Header.Set(fasthttp.HeaderAccept, "application/json")
	req.Header.Set(fasthttp.HeaderAcceptEncoding, "gzip, deflate")
	req.Header.SetMethod(fasthttp.MethodGet)
	resp.ImmediateHeaderFlush = true //only care about body

	for group := range work {
		if opts.Verbose {
			fmt.Sprintln("VERBOSE:", "Currently processing group:", group.ID, " - ", group.Title)
		}
		//construct URL
		userlistURL := fmt.Sprintf("%s/community/groups/%s/userList?f=json&num=100", opts.RestURI, group.ID)
	redo:
		req.SetRequestURI(userlistURL)
		err := wc.Do(req, resp)
		if err != nil {
			fmt.Fprintln(os.Stderr, "ERR: Was not able to query the userList for ", userlistURL, "Error:", err)
			fasthttp.ReleaseResponse(resp)
			continue
		}
		//get body, don't check for errs will come later
		var body []byte // encoding support
		encoding := resp.Header.Peek("Content-Encoding")
		switch string(encoding) {
		case "gzip":
			body, _ = resp.BodyGunzip()
		case "deflate":
			body, _ = resp.BodyInflate()
		default:
			body = resp.Body()
		}
		fasthttp.ReleaseResponse(resp)
		// map json to struct
		result := arcgisResult{}
		err = json.Unmarshal(body, &result)
		if err != nil {
			fmt.Fprintln(os.Stderr, "ERR: Was not able to read the parse JSON for URL: ", userlistURL, "Error:", err)
			continue
		}
		for _, user := range result.Users {
			//groupid	username	fullname	joined
			if opts.Esri || !strings.HasPrefix(user.Username, "esri_") {
				fmt.Printf("%s, %s, \"%s\", %d \n", group.ID, user.Username, user.Fullname, user.Joined) //print output to STDOUT
			}
		} // json returns nextStart if there are more items, if we are at last it shows -1
		if result.NextStart != -1 {
			userlistURL = fmt.Sprintf("%s/community/groups/%s/userList?f=json&num=100&start=%d", opts.RestURI, group.ID, result.NextStart)
			goto redo //re-do the steps with updated URL
		}
	}
	fasthttp.ReleaseRequest(req)
}
