package main

import (
	"encoding/json"
	"strings"
	"fmt"
	"io"
	"net/http"
	"sync"
	"path/filepath"
	"os"
	"io/ioutil"
	"log"
	"time"
)

const numWorkers = 10
type ArrayOfArrays [][]string

func readJsonFile() []string{
    const fileName = "urls.json"
    data, err := ioutil.ReadFile(fileName)
    // if we os.Open returns an error then handle it
    if err != nil {
        fmt.Println(err)
    }
    fmt.Println("Successfully Opened" + fileName)
    // defer the closing of our jsonFile so that we can parse it later on
    var aoa ArrayOfArrays
    if err := json.Unmarshal(data, &aoa); err != nil {
		log.Fatal(err)
	}

    // Iterate over the outer array.
    urls := []string{}
	for _, innerArray := range aoa {
        new_url := "https://web.archive.org/web/" + innerArray[2] + "if_/" + innerArray[0]
        urls = append(urls, new_url)
	}
    return urls
}

func checkFileExists(filename string) (bool, error) {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false, err
	} else if err != nil {
		fmt.Println("Error checking if file exists:", err)
		return false, err
	} else {
		return true, nil
	}
}

func download(url string, wg *sync.WaitGroup, bufPool *sync.Pool) error {
	parts := strings.Split(url, "/")
	// Find the index of the "wp-content" part.
	var wpContentIndex int
	for i, part := range parts {
		if part == "wp-content" {
			wpContentIndex = i
			break
		}
	}
	dir_path := filepath.Join(parts[wpContentIndex+1:len(parts)-1]...)
	file_path := filepath.Join(parts[wpContentIndex+1:]...)

	// Check if file exists
	exists, err := checkFileExists(file_path)
	if exists {
		s := fmt.Sprintf("File Exists: %s", file_path)
		fmt.Println(s)
		return nil
	}


	// Decrement the wait group counter when the goroutine completes
	defer wg.Done()

	// Get a buffer from the pool
	buf := bufPool.Get().([]byte)

	// Defer returning the buffer to the pool
	defer bufPool.Put(buf)

	// Create a HTTP client with a transport that limits the number of idle connections per host
	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: numWorkers,
		},
	}
	// Fetch the URL
	resp, err := client.Get(url)
	if err != nil {
		s := fmt.Sprintf("Error Retry %s, err: ", url, err)
		fmt.Println(s)
		time.Sleep(time.Minute)
		download(url, wg, bufPool)
		return err
	}
	defer resp.Body.Close()

	s := fmt.Sprintf("Downloading %s from %s", file_path, url)
	fmt.Println(s)

	// Create the subdirectory.
	err = os.MkdirAll(dir_path, 0755)
	if err != nil {
		fmt.Println(err)
	}

	//Create a empty file
	file, err := os.Create(file_path)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	// fmt.Println("Downloading", url)
	// Save the image data to a file
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	return nil
}

func main() {
	var wg sync.WaitGroup

	// Create a pool to cache and reuse image data buffers
	bufPool := &sync.Pool{
		New: func() interface{} {
			// Allocate a new buffer with a size of 1 MB
			return make([]byte, 1<<20)
		},
	}

	// Add the URLs to the channel
	result := readJsonFile()
    for i, url := range result {
        if i == 0 {
            continue
        }
		// Add the URL to the wait group
		wg.Add(1)
		if i % 10 == 0 {
			ticker := time.Tick(time.Second)
			<-ticker
		}
		err_download := download(url, &wg, bufPool)
		if err_download != nil {
			ticker := time.Tick(time.Minute)
			<-ticker
		}
    }

	wg.Wait()
}
