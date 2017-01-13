// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var total int64 = 0
var succeed int64 = 0
var failed int64 = 0

type PRDiff struct {
	FromHash     string `json:"fromHash"`
	ToHash       string `json:"toHash"`
	ContextLines int    `json:"contextLines"`
	Whitespace   string `json:"whitespace"`
	Diffs        []struct {
		Source struct {
			Components []string `json:"components"`
			Parent     string   `json:"parent"`
			Name       string   `json:"name"`
			Extension  string   `json:"extension"`
			ToString   string   `json:"toString"`
		} `json:"source"`
		Destination struct {
			Components []string `json:"components"`
			Parent     string   `json:"parent"`
			Name       string   `json:"name"`
			Extension  string   `json:"extension"`
			ToString   string   `json:"toString"`
		} `json:"destination"`
		Hunks []struct {
			SourceLine      int `json:"sourceLine"`
			SourceSpan      int `json:"sourceSpan"`
			DestinationLine int `json:"destinationLine"`
			DestinationSpan int `json:"destinationSpan"`
			Segments        []struct {
				Type  string `json:"type"`
				Lines []struct {
					Source      int    `json:"source"`
					Destination int    `json:"destination"`
					Line        string `json:"line"`
					Truncated   bool   `json:"truncated"`
				} `json:"lines"`
				Truncated bool `json:"truncated"`
			} `json:"segments"`
			Truncated bool `json:"truncated"`
		} `json:"hunks"`
		Truncated bool `json:"truncated"`
	} `json:"diffs"`
	Truncated bool `json:"truncated"`
}

func main() {
	prID := os.Args[1]
	url := "https://git.acronis.com/rest/api/1.0/projects/AMB/repos/mobile-backup-client-ios/pull-requests/" + prID + "/diff"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Print(err)
	}
	req.SetBasicAuth("<YOUR USERNAME>", "<YOUR PASSWORD>")
	client := &http.Client{}
	response, err := client.Do(req)
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Print(err)
	}
	// fmt.Print(string(body))
	var prDiff PRDiff
	json.Unmarshal(body, &prDiff)
	// fmt.Print(prDiff)
	//
	processDiff(prDiff)
}

func processDiff(prDiff PRDiff) {
	for i := 0; i < len(prDiff.Diffs); i++ {
		diff := prDiff.Diffs[i]
		fileName := diff.Source.Name
		if !strings.Contains(fileName, ".strings") {
			continue
		}
		hunks := diff.Hunks
		for j := 0; j < len(hunks); j++ {
			segments := hunks[j].Segments
			for k := 0; k < len(segments); k++ {
				segment := segments[k]
				if segment.Type == "REMOVED" || segment.Type == "CONTEXT" {
					continue
				}
				lines := segment.Lines
				for l := 0; l < len(lines); l++ {
					total++
					line := lines[l].Line
					fmt.Print(segment.Type + " ")
					fmt.Print(line)
					fmt.Print("\n")
					checkString(line)
				}
			}
		}
	}
	fmt.Print("Report: Total:" + strconv.FormatInt(total, 10) + " Succeed: " + strconv.FormatInt(succeed, 10) + " Failed: " + strconv.FormatInt(failed, 10))
}

func checkString(str string) {
	matched, err := regexp.Match("\";$", []byte(str))
	if err != nil {
		fmt.Print("Error checking " + str + ". Error:")
		fmt.Print(err)
	}
	if matched {
		succeed++
		fmt.Print("Checked " + str + "OK\n\n")
	} else {
		failed++
		fmt.Print("Checked " + str + "FAILED\n\n")
	}

}
