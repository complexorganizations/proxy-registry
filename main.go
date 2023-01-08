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
	// Waitgroups
	validateWaitGroup sync.WaitGroup
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
		"https://raw.githubusercontent.com/TheSpeedX/PROXY-List/master/http.txt",
		"https://raw.githubusercontent.com/clarketm/proxy-list/master/proxy-list-raw.txt",
		"https://raw.githubusercontent.com/ShiftyTR/Proxy-List/master/http.txt",
		"https://raw.githubusercontent.com/monosans/proxy-list/main/proxies/http.txt",
		"https://raw.githubusercontent.com/jetkai/proxy-list/main/online-proxies/txt/proxies-http.txt",
		"https://www.proxyscan.io/download?type=http",
		"https://raw.githubusercontent.com/Volodichev/proxy-list/main/http.txt",
		"https://raw.githubusercontent.com/mmpx12/proxy-list/master/http.txt",
		"https://raw.githubusercontent.com/hendrikbgr/Free-Proxy-Repo/master/proxy_list.txt",
		"https://raw.githubusercontent.com/almroot/proxylist/master/list.txt",
		"https://raw.githubusercontent.com/sunny9577/proxy-scraper/master/proxies.txt",
		"https://raw.githubusercontent.com/rx443/proxy-list/main/online/http.txt",
		"https://www.proxy-list.download/api/v1/get?type=http",
		"https://raw.githubusercontent.com/drakelam/Free-Proxy-List/main/proxy_all.txt",
		"https://raw.githubusercontent.com/jetkai/proxy-list/main/online-proxies/txt/proxies.txt",
		"https://raw.githubusercontent.com/TundzhayDzhansaz/proxy-list-auto-pull-in-30min/main/proxies/http.txt",
		"https://raw.githubusercontent.com/ShiftyTR/Proxy-List/master/https.txt",
		"https://raw.githubusercontent.com/jetkai/proxy-list/main/online-proxies/txt/proxies-https.txt",
		"https://www.proxyscan.io/download?type=https",
		"https://raw.githubusercontent.com/mmpx12/proxy-list/master/https.txt",
		"https://raw.githubusercontent.com/roosterkid/openproxylist/main/HTTPS_RAW.txt",
		"https://www.proxy-list.download/api/v1/get?type=https",
		"https://raw.githubusercontent.com/rx443/proxy-list/main/online/https.txt",
		"https://raw.githubusercontent.com/ShiftyTR/Proxy-List/master/socks4.txt",
		"https://raw.githubusercontent.com/jetkai/proxy-list/main/online-proxies/txt/proxies-socks4.txt",
		"https://www.proxyscan.io/download?type=socks4",
		"https://raw.githubusercontent.com/TheSpeedX/PROXY-List/master/socks4.txt",
		"https://raw.githubusercontent.com/monosans/proxy-list/main/proxies/socks4.txt",
		"https://raw.githubusercontent.com/roosterkid/openproxylist/main/SOCKS4_RAW.txt",
		"https://raw.githubusercontent.com/mmpx12/proxy-list/master/socks4.txt",
		"https://raw.githubusercontent.com/TheSpeedX/PROXY-List/master/socks5.txt",
		"https://raw.githubusercontent.com/ShiftyTR/Proxy-List/master/socks5.txt",
		"https://raw.githubusercontent.com/monosans/proxy-list/main/proxies/socks5.txt",
		"https://raw.githubusercontent.com/jetkai/proxy-list/main/online-proxies/txt/proxies-socks5.txt",
		"https://www.proxyscan.io/download?type=socks5",
		"https://raw.githubusercontent.com/hookzof/socks5_list/master/proxy.txt",
		"https://raw.githubusercontent.com/roosterkid/openproxylist/main/SOCKS5_RAW.txt",
		"https://raw.githubusercontent.com/mmpx12/proxy-list/master/socks5.txt",
		"https://raw.githubusercontent.com/KUTlime/ProxyList/main/ProxyList.txt",
		"https://raw.githubusercontent.com/complexorganizations/proxy-registry/main/assets/hosts",
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
	// Validate each proxy protocol.
	scrapedData = validateEachProxyProtocol(scrapedData)
	// Remove the old file.
	removeFile(hostsFile)
	// Go through the scraped data and validate the proxies.
	for _, value := range scrapedData {
		validateWaitGroup.Add(1)
		go validateAndSaveData(value)
	}
	validateWaitGroup.Wait()
	cleanupTheFiles(hostsFile)
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

func validateAndSaveData(content string) {
	if validateProxy(content) && validateProxyProtocol(content) && !checkIPIsInPrivateOrLocalRange(content) {
		appendAndWriteToFile(hostsFile, content)
	}
	validateWaitGroup.Done()
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

// Check if the given IP address is valid.
func isIPValid(providedIP string) bool {
	return net.ParseIP(providedIP) != nil
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
	_, err := url.ParseRequestURI(uri)
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

// Check if the IP is in local or private range.
func checkIPIsInPrivateOrLocalRange(content string) bool {
	uri, err := url.Parse(content)
	if err != nil {
		log.Fatalln(err)
	}
	host, _, err := net.SplitHostPort(uri.Host)
	if err != nil {
		log.Fatalln(err)
	}
	validIP := net.ParseIP(host)
	if validIP.IsLoopback() {
		return true
	}
	if validIP.IsMulticast() {
		return true
	}
	if validIP.IsPrivate() {
		return true
	}
	return false
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

// Validate proxy type is an approved type
func validateProxyProtocol(content string) bool {
	uri, err := url.Parse(content)
	if err != nil {
		log.Fatalln(err)
	}
	switch uri.Scheme {
	case "http", "https", "socks4", "socks5":
		return true
	}
	return false
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
func getProxyProtocol(content string) string {
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
	return validProtocolList[0]
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

// Validate each protocol and return the valid ones.
func validateEachProxyProtocol(content []string) []string {
	proxyProtocolList := []string{
		"http://",
		"https://",
		"socks4://",
		"socks5://",
	}
	var returnSlice []string
	for _, proxy := range content {
		for _, protocol := range proxyProtocolList {
			finalString := protocol + proxy
			if validateProxy(finalString) {
				returnSlice = append(returnSlice, finalString)
			}
		}
	}
	return returnSlice
}
