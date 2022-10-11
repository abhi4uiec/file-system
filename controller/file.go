package controller

import (
	"bufio"
	"encoding/json"
	"errors"
	constants "filesystem/constants"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
)

type requestBody struct {
	RemoteFileServerURL string `json:"remote_file_server_url"`
	LookupCharacter     string `json:"lookup_character"`
}

// key=position of the character, value = list of files having character at the same position
var positionMap = map[int]string{}
var lookupCharacter string

// Handles rest endpoint "/files"
func FetchFilesBasedOnCriteria(w http.ResponseWriter, r *http.Request) {

	var request requestBody
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Fatalln("There was an error decoding the request body into the struct")
	}
	remoteFileServerURL := request.RemoteFileServerURL
	lookupCharacter = request.LookupCharacter

	listOfFiles := allFileNamesFromRemoteServer(remoteFileServerURL)

	// read files concurrently
	err = readAndDownloadMultipleFiles(listOfFiles[:len(listOfFiles)-1], "readFile")
	if err != nil {
		log.Fatal(err)
	}

	var filesToDownload []string
	// sort the map based on key, this will give us the character at the earliest position
	if len(positionMap) > 1 {
		keys := make([]int, 0, len(positionMap))
		for k := range positionMap {
			keys = append(keys, k)
		}
		sort.Ints(keys)
		filesToDownload = strings.Split(positionMap[keys[0]], constants.EMPTY_STR)
	} else {
		for k := range positionMap {
			filesToDownload = strings.Split(positionMap[k], constants.EMPTY_STR)
			break
		}
	}

	sort.Strings(filesToDownload)

	for i := 0; i < len(filesToDownload); i++ {
		filesToDownload[i] = remoteFileServerURL + constants.FORWARD_SLASH + filesToDownload[i]
	}

	// actually download the matched files
	err = readAndDownloadMultipleFiles(filesToDownload, "downloadFile")
	if err != nil {
		log.Fatal(err)
	}

	w.Header().Set("Content-Type", "application/json")

	// encoding our podDetails array into a JSON string and then writing as part of our response.
	json.NewEncoder(w).Encode(filesToDownload)

	positionMap = map[int]string{}
}

// Gets all file names which are present on the remote server running on port 8080
func allFileNamesFromRemoteServer(remoteFileServerURL string) []string {
	resp, err := http.Get(remoteFileServerURL)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	var listOfFiles []string
	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		bodyString := string(bodyBytes)
		listOfFiles = strings.Split(bodyString, constants.ANCHOR_TAG)
		for i := 0; i < len(listOfFiles)-1; i++ {
			lastIndex := strings.LastIndex(listOfFiles[i], constants.GREATER_THAN)
			listOfFiles[i] = remoteFileServerURL + constants.FORWARD_SLASH + listOfFiles[i][lastIndex+1:]
		}
	}
	return listOfFiles
}

// Reads and downloads multiple files concurrently
func readAndDownloadMultipleFiles(urls []string, funcName string) error {
	done := make(chan bool, len(urls))
	errch := make(chan error, len(urls))
	for _, URL := range urls {
		go func(URL string, funcName string) {
			var err error
			if funcName == "readFile" {
				err = readFile(URL)
			} else {
				err = downloadFile(URL)
			}
			if err != nil {
				errch <- err
				done <- false
				return
			}
			done <- true
			errch <- nil
		}(URL, funcName)
	}
	var errStr string
	for i := 0; i < len(urls); i++ {
		<-done // blocks the flow until value is received
		if err := <-errch; err != nil {
			errStr = errStr + " " + err.Error()
		}
	}
	var err error
	if errStr != "" {
		err = errors.New(errStr)
	}
	return err
}

func readFile(URL string) error {

	filepath := URL[strings.LastIndex(URL, constants.FORWARD_SLASH)+1:]

	response, err := http.Get(URL)
	body := response.Body
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return errors.New(response.Status)
	}

	r := bufio.NewReader(body)
	bufio.NewScanner(body)
	buf := make([]byte, 0, 4*1024) // Reading in chunks
	for {
		n, err := r.Read(buf[:cap(buf)])

		buf = buf[:n]
		tempString := string(buf)
		index := strings.Index(tempString, lookupCharacter)

		if index != -1 {
			if _, ok := positionMap[index]; ok {
				positionMap[index] = positionMap[index] + " " + filepath
			} else {
				positionMap[index] = filepath
			}
			break
		}

		if n == 0 {
			if err == nil {
				continue
			}
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}

		if err != nil && err != io.EOF {
			log.Fatal(err)
		}
	}
	return nil
}

func downloadFile(URL string) error {

	filepath := URL[strings.LastIndex(URL, constants.FORWARD_SLASH)+1:]
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	response, err := http.Get(URL)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return errors.New(response.Status)
	}
	_, err = io.Copy(out, response.Body) // Writes to the file until EOF is reached
	if err != nil {
		return err
	}
	return nil
}
