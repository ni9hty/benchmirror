// To benchmark Ubuntu mirrors
package main

import (
	//"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"menteslibres.net/gosexy/to"
)

func main() {
	var mirrors string
	flag.StringVar(&mirrors, "f", "", "path to file containing the mirror urls")
	mirrors = flag.Arg(0)
	timeout := flag.Int("t", 5, "timeout setting for the http calls (time in sec)")
	limit := flag.Int("l", 5000, "limit latency, discard mirrors slower then (time in ms) ")
	v := flag.Bool("v", false, "enable verbose output")
	b := flag.Bool("b", false, "start benchmark")
	flag.Parse()

	var LIMIT time.Duration
	LIMIT = time.Duration(*limit)

	var TIMEOUT time.Duration
	TIMEOUT = time.Duration(*timeout)

	var VERBOSE bool
	if *v {
		VERBOSE = true
		fmt.Println("entering verbose mode, fasten your seat belt..\n")
		time.Sleep(2 * time.Second)
	} else {
		VERBOSE = false
	}

	if *b == false {
		fmt.Println("Ubuntu mirror checker")
		flag.Usage()
		os.Exit(0)
	}

	if len(mirrors) > 0 && *b {
		mirrors_read := read_url_list(mirrors, VERBOSE)
		mirrors_checked := bench(mirrors_read, TIMEOUT, VERBOSE)
		mirrors_sorted := sort_results(mirrors_checked, mirrors_read, VERBOSE, LIMIT)
		output(mirrors_sorted, VERBOSE)
	}

	if *b && len(mirrors) == 0 {
		mirrors_read := get_mirrors(TIMEOUT, VERBOSE)
		mirrors_checked := bench(mirrors_read, TIMEOUT, VERBOSE)
		mirrors_sorted := sort_results(mirrors_checked, mirrors_read, VERBOSE, LIMIT)
		output(mirrors_sorted, VERBOSE)
	}

}

//mirrors.txt are mirrors related to "yourcountry"
func get_mirrors(tout time.Duration, v bool) []string {
	timeout := tout * time.Second

	tr := &http.Transport{}
	client := &http.Client{Transport: tr,
		Timeout: timeout,
	}
	req, _ := http.NewRequest("GET", "http://mirrors.ubuntu.com/mirrors.txt", nil)
	if v == true {
		fmt.Println("GET http://mirrors.ubuntu.com/mirrors.txt")
	}

	resp, err := client.Do(req) //do the request
	if err != nil {
		mirror_err := fmt.Sprint("Can't fetch mirrors.", err)
		fmt.Println(mirror_err)
	}
	defer resp.Body.Close()
	responseData, _ := ioutil.ReadAll(resp.Body) // fetch body

	//validate and convert to []string
	var mirrors_greped []string
	mirrors_listed := strings.Split(to.String(responseData), "\n")
	count := 0
	for i := 0; i < len(mirrors_listed); i++ {
		if strings.Contains(mirrors_listed[i], "http://") {
			count++
			mirrors_greped = append(mirrors_greped, mirrors_listed[i])
			if v == true {
				fmt.Println("grepped", count, mirrors_listed[i])
			}
		}
	}
	return mirrors_greped
}

type HttpResponse struct {
	url      string
	response string
	latency  time.Duration
}

func bench(urls []string, tout time.Duration, v bool) []*HttpResponse {

	timeout := tout * time.Second
	tr := &http.Transport{}
	client := &http.Client{Transport: tr,
		Timeout: timeout,
	}

	channel := make(chan *HttpResponse) //need for multithread request execution
	responses := []*HttpResponse{}
	if v == true {
		fmt.Println("\nBenchmarking the urls")
	}

	for _, url := range urls {

		go func(url string) {
			host_to_check := fmt.Sprint(url)
			//fmt.Println("Fetching :", host_to_check)
			start := time.Now()
			resp, err := client.Get(host_to_check) //do the request
			ms := time.Since(start)
			mserr := time.Millisecond * 9999
			if err != nil {
				channel <- &HttpResponse{url, to.String(err), mserr}
			} else {
				responseData, _ := ioutil.ReadAll(resp.Body) // fetch body
				channel <- &HttpResponse{url, to.String(responseData), ms}
			}

		}(url)
	}
	for {
		select {
		case r := <-channel:
			if v == true {
				fmt.Println("got result: ", r.latency, "- ", r.url)
			}

			responses = append(responses, r)
			if len(responses) == len(urls) {
				return responses
			}
		case <-time.After(50 * time.Millisecond):
			/*if v == true {
				fmt.Printf(".")
			}*/

		}
	}
	return responses
}

func sort_results(checked_map []*HttpResponse, list []string, v bool, l time.Duration) map[string][]string {
	sorted_map := make(map[string][]string, len(list))
	limit := l * time.Millisecond
	//fmt.Println("limit: ", limit, "l: ", l)
	for _, value := range checked_map {
		if value.latency > limit {
			//do nothing
		} else {
			sorted_map["url"] = append(sorted_map["url"], value.url)
			sorted_map["latency"] = append(sorted_map["latency"], to.String(value.latency))
		}

	}
	if v == true {
		fmt.Println("\ndone.\n")
		for j := 0; j < len(sorted_map["url"]); j++ {
			fmt.Println(sorted_map["latency"][j], "-", sorted_map["url"][j])
		}
	}
	return sorted_map
}

func read_url_list(file string, v bool) []string {
	var urls_greped []string
	path := to.String(file)
	if v == true {
		fmt.Println("Reading urls from", path, "..")
	}

	urls, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println("urls file not found. ", err)
		os.Exit(1)
	} else {
		urls_listed := strings.Split(to.String(urls), "\n")
		count := 0
		for i := 0; i < len(urls_listed); i++ {
			if strings.Contains(urls_listed[i], "http") {
				count++
				urls_greped = append(urls_greped, urls_listed[i])
			}
		}
		if v == true {
			fmt.Println(to.String(count), "url(s) found.")
		}

	}
	return urls_greped
}

func output(mirrors_benched map[string][]string, v bool) {
	if v == true {
		fmt.Println("\n fastest mirrors are:\n")
	}
	for i := 0; i < len(mirrors_benched["url"]); i++ {
		fmt.Println(mirrors_benched["url"][i])
	}
}
