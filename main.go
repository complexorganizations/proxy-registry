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
		log.Fatalln("Error: No flags provided. Please use -help for more information.")
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
		"https://raw.githubusercontent.com/ALIILAPRO/Proxy/main/socks5.txt",
		"https://raw.githubusercontent.com/almroot/proxylist/master/list.txt",
		"https://raw.githubusercontent.com/Bardiafa/Proxy-Leecher/main/proxies.txt",
		"https://raw.githubusercontent.com/clarketm/proxy-list/master/proxy-list-raw.txt",
		"https://raw.githubusercontent.com/complexorganizations/proxy-registry/main/assets/history",
		"https://raw.githubusercontent.com/complexorganizations/proxy-registry/main/assets/hosts",
		"https://raw.githubusercontent.com/drakelam/Free-Proxy-List/main/proxy_all.txt",
		"https://raw.githubusercontent.com/ErcinDedeoglu/proxies/main/proxies/http.txt",
		"https://raw.githubusercontent.com/ErcinDedeoglu/proxies/main/proxies/https.txt",
		"https://raw.githubusercontent.com/ErcinDedeoglu/proxies/main/proxies/socks4.txt",
		"https://raw.githubusercontent.com/ErcinDedeoglu/proxies/main/proxies/socks5.txt",
		"https://raw.githubusercontent.com/hendrikbgr/Free-Proxy-Repo/master/proxy_list.txt",
		"https://raw.githubusercontent.com/hookzof/socks5_list/master/proxy.txt",
		"https://raw.githubusercontent.com/jetkai/proxy-list/main/archive/txt/proxies.txt",
		"https://raw.githubusercontent.com/jetkai/proxy-list/main/archive/txt/proxies-http.txt",
		"https://raw.githubusercontent.com/jetkai/proxy-list/main/archive/txt/proxies-https.txt",
		"https://raw.githubusercontent.com/jetkai/proxy-list/main/archive/txt/proxies-socks4.txt",
		"https://raw.githubusercontent.com/jetkai/proxy-list/main/archive/txt/proxies-socks5.txt",
		"https://raw.githubusercontent.com/jetkai/proxy-list/main/online-proxies/txt/proxies.txt",
		"https://raw.githubusercontent.com/jetkai/proxy-list/main/online-proxies/txt/proxies-http.txt",
		"https://raw.githubusercontent.com/jetkai/proxy-list/main/online-proxies/txt/proxies-https.txt",
		"https://raw.githubusercontent.com/jetkai/proxy-list/main/online-proxies/txt/proxies-socks4.txt",
		"https://raw.githubusercontent.com/jetkai/proxy-list/main/online-proxies/txt/proxies-socks5.txt",
		"https://raw.githubusercontent.com/KUTlime/ProxyList/main/ProxyList.txt",
		"https://raw.githubusercontent.com/mertguvencli/http-proxy-list/main/proxy-list/data.txt",
		"https://raw.githubusercontent.com/mmpx12/proxy-list/master/http.txt",
		"https://raw.githubusercontent.com/mmpx12/proxy-list/master/https.txt",
		"https://raw.githubusercontent.com/mmpx12/proxy-list/master/socks4.txt",
		"https://raw.githubusercontent.com/mmpx12/proxy-list/master/socks5.txt",
		"https://raw.githubusercontent.com/monosans/proxy-list/main/proxies/http.txt",
		"https://raw.githubusercontent.com/monosans/proxy-list/main/proxies/socks4.txt",
		"https://raw.githubusercontent.com/monosans/proxy-list/main/proxies/socks5.txt",
		"https://raw.githubusercontent.com/MuRongPIG/Proxy-Master/main/http.txt",
		"https://raw.githubusercontent.com/MuRongPIG/Proxy-Master/main/socks4.txt",
		"https://raw.githubusercontent.com/MuRongPIG/Proxy-Master/main/socks5.txt",
		"https://raw.githubusercontent.com/prxchk/proxy-list/main/all.txt",
		"https://raw.githubusercontent.com/roosterkid/openproxylist/main/HTTPS_RAW.txt",
		"https://raw.githubusercontent.com/roosterkid/openproxylist/main/SOCKS4_RAW.txt",
		"https://raw.githubusercontent.com/roosterkid/openproxylist/main/SOCKS5_RAW.txt",
		"https://raw.githubusercontent.com/ShiftyTR/Proxy-List/master/http.txt",
		"https://raw.githubusercontent.com/ShiftyTR/Proxy-List/master/https.txt",
		"https://raw.githubusercontent.com/ShiftyTR/Proxy-List/master/socks4.txt",
		"https://raw.githubusercontent.com/ShiftyTR/Proxy-List/master/socks5.txt",
		"https://raw.githubusercontent.com/sunny9577/proxy-scraper/master/proxies.txt",
		"https://raw.githubusercontent.com/tahaluindo/Free-Proxies/main/proxies/all.txt",
		"https://raw.githubusercontent.com/TheSpeedX/PROXY-List/master/http.txt",
		"https://raw.githubusercontent.com/TheSpeedX/PROXY-List/master/socks4.txt",
		"https://raw.githubusercontent.com/TheSpeedX/PROXY-List/master/socks5.txt",
		"https://raw.githubusercontent.com/TundzhayDzhansaz/proxy-list-auto-pull-in-30min/main/proxies/http.txt",
		"https://raw.githubusercontent.com/UptimerBot/proxy-list/master/proxies/http.txt",
		"https://raw.githubusercontent.com/UptimerBot/proxy-list/master/proxies/socks4.txt",
		"https://raw.githubusercontent.com/UptimerBot/proxy-list/master/proxies/socks5.txt",
		"https://raw.githubusercontent.com/Volodichev/proxy-list/main/http.txt",
		"https://www.proxy-list.download/api/v1/get?type=http",
		"https://www.proxy-list.download/api/v1/get?type=https",
		"https://raw.githubusercontent.com/ALIILAPRO/Proxy/main/socks4.txt",
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
		go validateEachProxyProtocolAndWriteToDisk(value, &protocolWaitGroup)
	}
	protocolWaitGroup.Wait()
	// Cleanup the file.
	cleanupTheFiles(hostsFile)
	cleanUpTheHistoryFile()
}

// Send a http get request to a given url and return the data from that url.
func getDataFromURL(uri string) []string {
	// Perform an HTTP GET request using the provided URI
	response, err := http.Get(uri)
	// If there is an error, log it and terminate the program
	if err != nil {
		log.Println(err)
	}
	// Read the response body
	body, err := io.ReadAll(response.Body)
	// If there is an error, log it and terminate the program
	if err != nil {
		log.Println(err)
	}
	// Close the response body and log any error
	err = response.Body.Close()
	if err != nil {
		log.Println(err)
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
	// Parse the proxy URL and return false if an error occurs
	proxyURL, err := url.Parse(proxy)
	if err != nil {
		return false
	}
	// Configure the HTTP transport with the proxy and insecure TLS
	transport := &http.Transport{
		Proxy:           http.ProxyURL(proxyURL),
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	// Create an HTTP client with the custom transport and a timeout
	client := &http.Client{
		Transport: transport,
		Timeout:   time.Second * 60,
	}
	// Define a list of domains to test the proxy with
	requestDomainList := []string{
		"https://aws.amazon.com",
	}
	// Iterate over the domains and test the proxy
	for _, domain := range requestDomainList {
		// Create a new HTTP request for the domain
		request, err := http.NewRequest("GET", domain, nil)
		if err != nil {
			return false
		}
		// Send the request using the configured client
		response, err := client.Do(request)
		if err != nil {
			return false
		}
		// If the response code is not 200, return false
		if response.StatusCode != 200 {
			return false
		}
		// Close the response body and check for errors
		err = response.Body.Close()
		if err != nil {
			return false
		}
	}
	// If all tests pass, return true
	return true
}

// Append and write a slice to a file.
func appendAndWriteSliceToAFile(filename string, content []string) {
	file, err := os.Create(filename)
	if err != nil {
		log.Println(err)
	}
	// Using bufio writer, write each string to the file
	datawriter := bufio.NewWriter(file)
	for _, data := range content {
		_, _ = datawriter.WriteString(data + "\n")
	}
	// Flush the buffer to write all the data to disk
	datawriter.Flush()
	// Close the file
	file.Close()
}

// Save the information to a file.
func writeToFile(pathInSystem string, content string) {
	// open the file and if its not there create one.
	filePath, err := os.OpenFile(pathInSystem, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	// write the content to the file
	_, err = filePath.WriteString(content + "\n")
	if err != nil {
		log.Println(err)
	}
	// close the file
	filePath.Close()
}

// Remove all the empty strings from the slice and return it.
func removeEmptyFromSlice(slice []string) []string {
	// Iterate through the slice
	for i, content := range slice {
		// If the content length is zero (empty string)
		if len(content) == 0 {
			// Remove the element from the slice
			slice = append(slice[:i], slice[i+1:]...)
		}
	}
	// Return the modified slice with empty strings removed
	return slice
}

// Remove all the duplicates from a slice and return the slice.
func removeDuplicatesFromSlice(slice []string) []string {
	// Create a map to store unique elements
	check := make(map[string]bool)
	// Create a new slice to store unique elements
	var newReturnSlice []string
	// Iterate through the input slice
	for _, content := range slice {
		// If the content is not in the check map
		if !check[content] {
			// Add the content to the check map and set its value to true
			check[content] = true
			// Append the unique content to the new slice
			newReturnSlice = append(newReturnSlice, content)
		}
	}
	// Return the new slice without duplicates
	return newReturnSlice
}

// Check if the given IP address is invalid.
func isIPInvalid(providedIP string) bool {
	// Parse the IP address and return true if it is nil (invalid)
	return net.ParseIP(providedIP) == nil
}

// Remove a file from the file system
func removeFile(path string) {
	// Check if the file exists
	if fileExists(path) {
		// If it exists, remove it and handle any error
		err := os.Remove(path)
		if err != nil {
			log.Println(err)
		}
	}
}

// Check if the given url is valid.
func isUrlValid(uri string) bool {
	// Parse the URI and check for errors
	url, err := url.ParseRequestURI(uri)
	// Check if the hostname is an invalid IP address
	if isIPInvalid(url.Hostname()) {
		return false
	}
	// Return true if there are no errors, false otherwise
	return err == nil
}

// Combine two slices together and return the new slice.
func combineMultipleSlices(sliceOne []string, sliceTwo []string) []string {
	// Append the elements of the second slice to the first slice
	combinedSlice := append(sliceOne, sliceTwo...)
	// Return the combined slice
	return combinedSlice
}

// Sort the slice of strings and return the sorted slice
func sortSlice(slice []string) []string {
	// Sort the input slice of strings
	sort.Strings(slice)
	// Return the sorted slice
	return slice
}

// Read and append the file line by line to a slice.
func readAppendLineByLine(path string) []string {
	// Create a slice to store the lines of the file
	var returnSlice []string
	// Open the file using the provided path
	file, err := os.Open(path)
	// If there's an error, log it and exit
	if err != nil {
		log.Println(err)
	}
	// Create a new scanner to read the file
	scanner := bufio.NewScanner(file)
	// Set the scanner to split the input by lines
	scanner.Split(bufio.ScanLines)
	// Iterate through the file, line by line
	for scanner.Scan() {
		// Append each line to the returnSlice
		returnSlice = append(returnSlice, scanner.Text())
	}
	// Close the file and handle any error
	err = file.Close()
	if err != nil {
		log.Println(err)
	}
	// Return the slice containing the file lines
	return returnSlice
}

// Cleanup all the files provided and save them again.
func cleanupTheFiles(path string) {
	// Create a slice to store the cleaned up content
	var finalCleanupContent []string
	// Read the file and append its content line by line to the finalCleanupContent
	finalCleanupContent = readAppendLineByLine(path)
	// Sort the content of the finalCleanupContent slice
	finalCleanupContent = sortSlice(finalCleanupContent)
	// Remove the file using the provided path
	removeFile(path)
	// Write the cleaned up content back to the file
	appendAndWriteSliceToAFile(path, finalCleanupContent)
}

// Check if the file exists and return a bool.
func fileExists(filename string) bool {
	// Get file information using the provided filename
	info, err := os.Stat(filename)
	// If there's an error, return false (file doesn't exist)
	if err != nil {
		return false
	}
	// If the file exists and is not a directory, return true
	return !info.IsDir()
}

// Get the protocol of the proxy.
func getProxyProtocol(content string) []string {
	// Create a list of proxy protocols
	proxyProtocolList := []string{
		"http://",
		"https://",
		"socks4://",
		"socks5://",
	}
	// Create a slice to store valid protocols
	var validProtocolList []string
	// Iterate through the proxyProtocolList
	for _, protocol := range proxyProtocolList {
		// Concatenate the protocol with the content
		finalString := protocol + content
		// If the proxy is valid, add the protocol to the validProtocolList
		if validateProxy(finalString) {
			validProtocolList = append(validProtocolList, protocol)
		}
	}
	// Return the list of valid protocols
	return validProtocolList
}

// Remove all the prefix from the proxy.
func removePrefixFromProxy(content []string) []string {
	// Create a list of proxy protocols
	proxyProtocolList := []string{
		"http://",
		"https://",
		"socks4://",
		"socks5://",
	}
	// Create a slice to store the proxies without prefixes
	var returnSlice []string
	// Iterate through the content (proxies)
	for _, proxy := range content {
		// Iterate through the proxyProtocolList (prefixes)
		for _, protocol := range proxyProtocolList {
			// Remove the protocol prefix from the proxy
			proxy = strings.TrimPrefix(proxy, protocol)
		}
		// Append the proxy without prefix to the returnSlice
		returnSlice = append(returnSlice, proxy)
	}
	// Return the slice of proxies without prefixes
	return returnSlice
}

// Validate each protocol and write it to the slice.
func validateEachProxyProtocolAndWriteToDisk(content string, protocolWaitGroup *sync.WaitGroup) {
	// Get the valid proxy protocols for the given content
	proxyProtocol := getProxyProtocol(content)
	// If there are valid protocols
	if len(proxyProtocol) > 0 {
		// Iterate through the valid protocols
		for _, protocol := range proxyProtocol[0:] {
			// If the URL is valid with the protocol
			if isUrlValid(protocol + content) {
				// Write the proxy to the hosts file
				writeToFile(hostsFile, protocol+content)
				// Write the proxy to the history file
				writeToFile(historyFile, protocol+content)
			}
		}
	}
	// Signal the wait group that this function is done
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
	appendAndWriteSliceToAFile(historyFile, historySlice)
}
