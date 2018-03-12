// Converts google transcripts into subtitle srt files

package main

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
)

var isTime = regexp.MustCompile("^[0-9]{2,3}:[0-9]{2}$")

func main() {
	printIntroText()
	files, err := ioutil.ReadDir(".")
	if err != nil {
		fmt.Println("Error reading files in current directory")
		fmt.Println(err)
		return
	}
	fmt.Println("Converting...")
	totalConverted := 0
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		name := f.Name()
		// name must end in .txt
		if len(name) < 5 {
			continue
		}
		if name[len(name)-4:len(name)] == ".txt" {
			srtContent := convertTxtToSrt(name)
			if len(srtContent) == 0 {
				continue
			}
			srtFilename := name[0:len(name)-4] + ".srt"
			saveSrt(srtContent, srtFilename)
			totalConverted = totalConverted + 1
		}
	}
	if totalConverted == 0 {
		fmt.Println("No files converted. Are there .txt files in this directory?")
	}
}

func printIntroText() {
	fmt.Println("youtube_transcript_to_srt v0.1.0")
	fmt.Println("")
	fmt.Println("About:")
	fmt.Println("Creates .srt files from any .txt files in the current")
	fmt.Println("directory.")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("1. Open a youtube video in your browser.")
	fmt.Println("2. Press the Menu button (...) then Open Transcript.")
	fmt.Println("3. Copy the youtube transcript text.")
	fmt.Println("4. Paste the text into a .txt file in this directory.")
	fmt.Println("5. Run youtube_transcript_to_srt")
	fmt.Println("6. The subtitles will be in this directory renamed with .srt")
	fmt.Println("")
}

func convertTxtToSrt(filename string) string {
	contentBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("Error reading file: " + filename)
		fmt.Println(err)
		return ""
	}
	// initialise srt content
	srt := "\n"
	// file alternates each line with time / text
	content := string(contentBytes)
	lines := strings.Split(content, "\n")
	count := 1
	previousSeconds := 0
	previousLine := ""
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// ignore blank lines
		if len(line) == 0 {
			continue
		}
		// check if the line is a time
		// TODO edge case where subtitle text is just a time
		if isTime.Match([]byte(line)) {
			bits := strings.Split(line, ":")
			if len(bits) != 2 {
				fmt.Println("Incorrect timing line:")
				fmt.Println(line)
				fmt.Println("Ignoring file")
				return ""
			}
			// get current seconds elapsed
			minutes, err := strconv.Atoi(bits[0])
			if err != nil {
				fmt.Println("Error parsing minutes in line", line)
				fmt.Println(err)
				return ""
			}
			seconds, err := strconv.Atoi(bits[1])
			if err != nil {
				fmt.Println("Error parsing seconds in line", line)
				fmt.Println(err)
				return ""
			}
			currentSeconds := minutes*60 + seconds
			// create next line in srt file
			start := secondsToSrtTime(previousSeconds)
			end := secondsToSrtTime(currentSeconds)
			if start == end {
				continue
			}
			newSrtLines := fmt.Sprintf("%v\n%v --> %v\n%v\n\n", count, start, end, previousLine)
			srt = srt + newSrtLines
			// keep track for next line of subtitles
			previousSeconds = currentSeconds
			previousLine = ""
			count = count + 1
		} else {
			previousLine = strings.TrimSpace(previousLine + "\n" + line)
		}
	}
	// add last line
	if previousLine != "" {
		start := secondsToSrtTime(previousSeconds)
		end := secondsToSrtTime(previousSeconds + 3)
		newSrtLines := fmt.Sprintf("%v\n%v --> %v\n%v\n\n", count, start, end, previousLine)
		srt = srt + newSrtLines
	}
	return srt
}

func secondsToSrtTime(t int) string {
	h := t / 3600
	m := (t - (h * 3600)) / 60
	mPad := zeroLeftPad(m, 2)
	s := t - (h * 3600) - (m * 60)
	sPad := zeroLeftPad(s, 2)
	return fmt.Sprintf("%v:%v:%v,000", h, mPad, sPad)
}

func saveSrt(content, filename string) {
	err := ioutil.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		fmt.Println("Error writing srt file", filename)
		fmt.Println(err)
		return
	}
	fmt.Println("Created", filename)
}

func zeroLeftPad(i, finalLength int) string {
	s := strconv.Itoa(i)
	for len(s) < finalLength {
		s = "0" + s
	}
	return s
}
