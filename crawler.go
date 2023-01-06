package main

import (
    "encoding/json"
    "fmt"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
    "io/ioutil"
    "path/filepath"
    "path"
    "strings"
)


type ArrayOfArrays [][]string

func read_json() []string{
    const fileName = "image_json_anne.json"
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
		fmt.Println(innerArray[0], innerArray[1])
        new_url := "https://web.archive.org/web/" + innerArray[2] + "if_/" + innerArray[0]
        urls = append(urls, new_url)
	}
    return urls
}

func main() {
    result := read_json()
    for i, url := range result {
        if i == 0 {
            continue
        }
        fmt.Println(url)
        fileName := path.Base(url)
        fmt.Println(fileName)
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
        // Print the parts after the "wp-content" part.
	    fmt.Println(file_path) // Output: [themes my-theme style.css]
        err := downloadFile(url, file_path, dir_path)
        if err != nil {
          	log.Fatal(err)
        }
    }
}

func downloadFile(URL, file_path string, dirPath string) error {
	//Get the response bytes from the url
	response, err := http.Get(URL)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return errors.New("Received non 200 response code")
	}

    // Create the subdirectory.
	err = os.MkdirAll(dirPath, 0755)
	if err != nil {
		fmt.Println(err)
		return err
	}

	//Create a empty file
	file, err := os.Create(file_path)
	if err != nil {
		return err
	}
	defer file.Close()

	//Write the bytes to the fiel
	_, err = io.Copy(file, response.Body)
	if err != nil {
		return err
	}

	return nil
}