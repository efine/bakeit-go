/*
Copyright © 2023 Edwin Fine <me@github.edfine.io>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/ini.v1"
)

var title string
var lang string
var duration int
var maxViews int
var openBrowser bool

var rootCmd = &cobra.Command{
	Use:   "bakeit",
	Short: `BakeIt is a command line utility to Pastery(https://www.pastery.net)`,
	Long: `BakeIt is a command line utility to Pastery(https://www.pastery.net), the best
	pastebin in the world. BakeIt aims to be simple to use and unobtrusive.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfgPath, err := apiKeyPath()
		cobra.CheckErr(err)

		apiKey, err := readApiKey(cfgPath)
		cobra.CheckErr(err)

		fmt.Println("Found api_key in", cfgPath)
		handleUpload(apiKey, args)
	},
}

type ResponseData struct {
	URL string `json:"url"`
}

func handleUpload(apiKey string, args []string) {
	// Read the entire file content
	var content []byte
	var err error
	if len(args) == 0 {
		content, err = readFromStdin()
		if err != nil {
			fmt.Println("Error reading from stdin:", err)
			return
		}
	} else {
		content, err = ioutil.ReadFile(args[0])
		if err != nil {
			fmt.Println("Error reading file:", err)
			return
		}
	}

	// Convert the file content to a byte array
	byteArray := []byte(content)
	URL := "https://www.pastery.net/api/paste/"

	// Set the query parameters
	params := url.Values{}
	params.Set("api_key", apiKey)
	params.Add("title", title)
	params.Add("language", lang)
	params.Add("duration", fmt.Sprint(duration))
	params.Add("max_views", fmt.Sprint(maxViews))

	queryString := params.Encode()

	fmt.Println("URL+QPs:", URL+"?"+queryString)

	req, err := http.NewRequest("POST", URL+"?"+queryString, bytes.NewBuffer(byteArray))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	// Set the appropriate headers, if needed
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Content-Length", fmt.Sprint(len(byteArray)))
	req.Header.Set("User-Agent", "Mozilla/5.0 (Go) bakeit library")

	// Create an HTTP client and send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}
	// Create a variable to hold the parsed JSON data
	var data ResponseData

	// Unmarshal the response body into the data structure
	err = json.Unmarshal(body, &data)
	if err != nil {
		fmt.Println("Error unmarshaling JSON:", err)
		return
	}

	fmt.Println("Paste URL:", data.URL)

	// Make sure to close the response body at the end
	defer resp.Body.Close()
}

func readFromStdin() ([]byte, error) {
	// Read the entire input from stdin
	input, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		fmt.Println("Error reading input from stdin:", err)
		return []byte{}, err
	}
	return input, nil
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&title, "title", "", "The title of the paste")
	rootCmd.PersistentFlags().StringVar(&lang, "lang", "", "The language highlighter to use")
	rootCmd.PersistentFlags().IntVar(&duration, "duration", 60, "The duration the paste should live for")
	rootCmd.PersistentFlags().IntVar(&maxViews, "max-views", 0, "How many times the paste can be viewed before it expires")
	rootCmd.PersistentFlags().BoolVar(&openBrowser, "open-browser", false, "Automatically open a browser window when done")
}

func apiKeyPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	ps := string(os.PathSeparator)
	return homeDir + ps + ".config" + ps + "bakeit.cfg", nil

}

// readApiKey reads the api_key from the config file in $HOME/.config/bakeit.cfg.
// It is expected to be in the [pastery] section and named "api_key".
func readApiKey(cfgPath string) (string, error) {
	iniData, err := ini.Load(cfgPath)
	if err != nil {
		return "", err
	}

	sec := iniData.Section("pastery")
	apiKey := sec.Key("api_key").String()
	if apiKey == "" {
		return "", errors.New("missing api_key")
	}
	return apiKey, nil
}
