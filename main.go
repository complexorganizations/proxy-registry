package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"flag"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	// Files
	inclusionList = "assets/inclusion"
	exclusionList = "assets/exclusion"
	hostsFile     = "assets/hosts"
	historyFile   = "assets/history"
	// Waitgroups
	protocolWaitGroup sync.WaitGroup
	// The user expresses his or her opinion on what should be done.
	update bool
)

func init() {
	if len(os.Args) > 1 {
		tempUpdate := flag.Bool("update", false, "Make any necessary changes to the listings.")
		flag.Parse()
		update = *tempUpdate
	} else {
		log.Fatal("Error: No flags provided. Please use -help for more information.")
	}
}

func main() {
	if update {
		scrapeTheLists()
	}
}

func scrapeTheLists() {
	// Create a map of the proxy list.
	proxyList := []string{
		"https://raw.githubusercontent.com/KUTlime/ProxyList/main/ProxyList.txt",
		"https://raw.githubusercontent.com/ShiftyTR/Proxy-List/master/http.txt",
		"https://raw.githubusercontent.com/ShiftyTR/Proxy-List/master/https.txt",
		"https://raw.githubusercontent.com/ShiftyTR/Proxy-List/master/socks4.txt",
		"https://raw.githubusercontent.com/ShiftyTR/Proxy-List/master/socks5.txt",
		"https://raw.githubusercontent.com/TheSpeedX/PROXY-List/master/http.txt",
		"https://raw.githubusercontent.com/TheSpeedX/PROXY-List/master/socks4.txt",
		"https://raw.githubusercontent.com/TheSpeedX/PROXY-List/master/socks5.txt",
		"https://raw.githubusercontent.com/TundzhayDzhansaz/proxy-list-auto-pull-in-30min/main/proxies/http.txt",
		"https://raw.githubusercontent.com/Volodichev/proxy-list/main/http.txt",
		"https://raw.githubusercontent.com/almroot/proxylist/master/list.txt",
		"https://raw.githubusercontent.com/clarketm/proxy-list/master/proxy-list-raw.txt",
		"https://raw.githubusercontent.com/complexorganizations/proxy-registry/main/assets/history",
		"https://raw.githubusercontent.com/complexorganizations/proxy-registry/main/assets/hosts",
		"https://raw.githubusercontent.com/drakelam/Free-Proxy-List/main/proxy_all.txt",
		"https://raw.githubusercontent.com/hendrikbgr/Free-Proxy-Repo/master/proxy_list.txt",
		"https://raw.githubusercontent.com/hookzof/socks5_list/master/proxy.txt",
		"https://raw.githubusercontent.com/jetkai/proxy-list/main/online-proxies/txt/proxies-http.txt",
		"https://raw.githubusercontent.com/jetkai/proxy-list/main/online-proxies/txt/proxies-https.txt",
		"https://raw.githubusercontent.com/jetkai/proxy-list/main/online-proxies/txt/proxies-socks4.txt",
		"https://raw.githubusercontent.com/jetkai/proxy-list/main/online-proxies/txt/proxies-socks5.txt",
		"https://raw.githubusercontent.com/jetkai/proxy-list/main/online-proxies/txt/proxies.txt",
		"https://raw.githubusercontent.com/mmpx12/proxy-list/master/http.txt",
		"https://raw.githubusercontent.com/mmpx12/proxy-list/master/https.txt",
		"https://raw.githubusercontent.com/mmpx12/proxy-list/master/socks4.txt",
		"https://raw.githubusercontent.com/mmpx12/proxy-list/master/socks5.txt",
		"https://raw.githubusercontent.com/monosans/proxy-list/main/proxies/http.txt",
		"https://raw.githubusercontent.com/monosans/proxy-list/main/proxies/socks4.txt",
		"https://raw.githubusercontent.com/monosans/proxy-list/main/proxies/socks5.txt",
		"https://raw.githubusercontent.com/roosterkid/openproxylist/main/HTTPS_RAW.txt",
		"https://raw.githubusercontent.com/roosterkid/openproxylist/main/SOCKS4_RAW.txt",
		"https://raw.githubusercontent.com/roosterkid/openproxylist/main/SOCKS5_RAW.txt",
		"https://raw.githubusercontent.com/rx443/proxy-list/main/online/http.txt",
		"https://raw.githubusercontent.com/rx443/proxy-list/main/online/https.txt",
		"https://raw.githubusercontent.com/sunny9577/proxy-scraper/master/proxies.txt",
		"https://www.proxy-list.download/api/v1/get?type=http",
		"https://www.proxy-list.download/api/v1/get?type=https",
		"https://www.proxyscan.io/download?type=http",
		"https://www.proxyscan.io/download?type=https",
		"https://www.proxyscan.io/download?type=socks4",
		"https://www.proxyscan.io/download?type=socks5",
	}
	// Create a list of scraped data.
	var scrapedData []string
	// Go through the proxy list and validate the proxies.
	for _, value := range proxyList {
		var tempScrapedData []string = getDataFromURL(value)
		scrapedData = combineMultipleSlices(tempScrapedData, scrapedData)
	}
	// Remove all the empty things from slice.
	scrapedData = removeEmptyFromSlice(scrapedData)
	// Remove all the dubplicates from slice.
	scrapedData = removeDuplicatesFromSlice(scrapedData)
	// Remove all the prefix from the proxies.
	scrapedData = removePrefixFromProxy(scrapedData)
	// Remove the old file.
	removeFile(hostsFile)
	// Get all the protocol for the proxies and than save that data.
	for _, value := range scrapedData {
		protocolWaitGroup.Add(1)
		go validateEachProxyProtocolAndWriteToDisk(value)
	}
	protocolWaitGroup.Wait()
	// Cleanup the file.
	cleanupTheFiles(hostsFile)
	cleanUpTheHistoryFile()
}

// Send a http get request to a given url and return the data from that url.
func getDataFromURL(uri string) []string {
	response, err := http.Get(uri)
	if err != nil {
		log.Fatalln(err)
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatalln(err)
	}
	err = response.Body.Close()
	if err != nil {
		log.Fatalln(err)
	}
	// Examine the page's response code.
	if response.StatusCode != 200 {
		log.Println("Sorry, but we were unable to scrape the page you requested due to a error.", uri)
	}
	// Scraped data is read and appended to an array.
	scanner := bufio.NewScanner(bytes.NewReader(body))
	scanner.Split(bufio.ScanLines)
	var returnContent []string
	for scanner.Scan() {
		returnContent = append(returnContent, scanner.Text())
	}
	return returnContent
}

// Check if a given proxy is working and return a bool.
func validateProxy(proxy string) bool {
	proxyURL, err := url.Parse(proxy)
	if err != nil {
		return false
	}
	transport := &http.Transport{
		Proxy:           http.ProxyURL(proxyURL),
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   time.Second * 60,
	}
	requestDomainList := []string{
		"https://aws.amazon.com",
		"https://cloud.google.com",
	}
	for _, domain := range requestDomainList {
		request, err := http.NewRequest("GET", domain, nil)
		if err != nil {
			return false
		}
		response, err := client.Do(request)
		if err != nil {
			return false
		}
		if response.StatusCode != 200 {
			return false
		}
		err = response.Body.Close()
		if err != nil {
			return false
		}
	}
	return true
}

// Append and write to file
func appendAndWriteToFile(path string, content string) {
	filePath, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalln(err)
	}
	_, err = filePath.WriteString(content + "\n")
	if err != nil {
		log.Fatalln(err)
	}
	err = filePath.Close()
	if err != nil {
		log.Fatalln(err)
	}
}

// Remove all the empty strings from the slice and return it.
func removeEmptyFromSlice(slice []string) []string {
	for i, content := range slice {
		if len(content) == 0 {
			slice = append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

// Remove all the duplicates from a slice and return the slice.
func removeDuplicatesFromSlice(slice []string) []string {
	check := make(map[string]bool)
	var newReturnSlice []string
	for _, content := range slice {
		if !check[content] {
			check[content] = true
			newReturnSlice = append(newReturnSlice, content)
		}
	}
	return newReturnSlice
}

// Check if the given IP address is invalid.
func isIPInvalid(providedIP string) bool {
	return net.ParseIP(providedIP) == nil
}

// Remove a file from the file system
func removeFile(path string) {
	if fileExists(path) {
		err := os.Remove(path)
		if err != nil {
			log.Fatalln(err)
		}
	}
}

// Check if the given url is valid.
func isUrlValid(uri string) bool {
	url, err := url.ParseRequestURI(uri)
	if isIPInvalid(url.Hostname()) {
		return false
	}
	return err == nil
}

// Combine two slices together and return the new slice.
func combineMultipleSlices(sliceOne []string, sliceTwo []string) []string {
	var combinedSlice []string
	combinedSlice = append(sliceOne, sliceTwo...)
	return combinedSlice
}

// Sort the slice of strings and return the sorted slice
func sortSlice(slice []string) []string {
	sort.Strings(slice)
	return slice
}

// Read and append the file line by line to a slice.
func readAppendLineByLine(path string) []string {
	var returnSlice []string
	file, err := os.Open(path)
	if err != nil {
		log.Fatalln(err)
	}
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		returnSlice = append(returnSlice, scanner.Text())
	}
	err = file.Close()
	if err != nil {
		log.Fatalln(err)
	}
	return returnSlice
}

// Cleanup all the files provided and save them again.
func cleanupTheFiles(path string) {
	var finalCleanupContent []string
	finalCleanupContent = readAppendLineByLine(path)
	finalCleanupContent = sortSlice(finalCleanupContent)
	removeFile(path)
	for _, content := range finalCleanupContent {
		appendAndWriteToFile(path, content)
	}
}

// Check if the file exists and return a bool.
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// Get the protocol of the proxy.
func getProxyProtocol(content string) []string {
	proxyProtocolList := []string{
		"http://",
		"https://",
		"socks4://",
		"socks5://",
	}
	var validProtocolList []string
	for _, protocol := range proxyProtocolList {
		finalString := protocol + content
		if validateProxy(finalString) {
			validProtocolList = append(validProtocolList, protocol)
		}
	}
	return validProtocolList
}

// Remove all the prefix from the proxy.
func removePrefixFromProxy(content []string) []string {
	proxyProtocolList := []string{
		"http://",
		"https://",
		"socks4://",
		"socks5://",
	}
	var returnSlice []string
	for _, proxy := range content {
		for _, protocol := range proxyProtocolList {
			proxy = strings.TrimPrefix(proxy, protocol)
		}
		returnSlice = append(returnSlice, proxy)
	}
	return returnSlice
}

// Validate each protocol and write it to the file.
func validateEachProxyProtocolAndWriteToDisk(content string) {
	proxyProtocol := getProxyProtocol(content)
	if len(proxyProtocol) > 0 {
		for _, protocol := range proxyProtocol[0:] {
			if isUrlValid(protocol + content) {
				// Write the proxy to the hosts file.
				appendAndWriteToFile(hostsFile, protocol+content)
				// Write the proxy to the history file.
				appendAndWriteToFile(historyFile, protocol+content)
			}
		}
	}
	protocolWaitGroup.Done()
}

// Clean up the history file.
func cleanUpTheHistoryFile() {
	// Remove the history file at the 1st of every year.
	if time.Now().Month().String() == "January" {
		if time.Now().Day() == 1 {
			removeFile(historyFile)
		}
	}
	// Read the history file and append it to a slice.
	historySlice := readAppendLineByLine(historyFile)
	// Remove all the duplicates from the slice.
	historySlice = removeDuplicatesFromSlice(historySlice)
	// Remove all the empty strings from the slice.
	historySlice = removeEmptyFromSlice(historySlice)
	// Remove the history file.
	removeFile(historyFile)
	// Sort the slice.
	historySlice = sortSlice(historySlice)
	// Write the slice to the history file.
	for _, content := range historySlice {
		appendAndWriteToFile(historyFile, content)
	}
}
