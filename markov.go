package main

import (
	"encoding/gob"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"bytes"
	"io/ioutil"
)

var (
	END         = "@END@"
	START       = [2]string{"@START 1@", "@START 2@"}
	punctuation = map[string]bool{
		".": true,
		",": true,
		"!": true,
		"?": true,
		";": true,
		":": true,
		"&": true,
	}
	SplitRegex    = regexp.MustCompile(`([\w'-]+|[.,!?;&])`)
	MaxMessageLen int
)

func generateMarkovResponse(inputText string) string {
	seed := processText(preprocessText(inputText))
	previousItems := [2]string{}
	var response string
	if len(seed) > 1 {
		previousItems[0] = seed[0]
		previousItems[1] = seed[1]
		response = seed[0] + " " + seed[1]
	} else if len(seed) == 1 {
		previousItems[0] = START[1]
		previousItems[1] = seed[0]
		response = seed[0]
	} else {
		previousItems = START
	}
	DataDict.RLock()
	defer DataDict.RUnlock()
	if _, ok := DataDict.Map[previousItems]; !ok {
		return "Error! I don't understand that =("
	}
	for i := 0; i < MaxMessageLen; i++ {
		options, ok := DataDict.Map[previousItems]
		if !ok {
			break
		}
		nextItem := options[rand.Intn(len(options))]
		if nextItem == END {
			break
		}
		if _, isPunctuation := punctuation[nextItem]; isPunctuation {
			response = response + nextItem
		} else {
			response = response + " " + nextItem
		}
		previousItems[0] = previousItems[1]
		previousItems[1] = nextItem
	}
	return response
}

func trainMessage(msg string) {
	items := processText(preprocessText(msg)) // Split by whitespace to get individual words
	previousItems := START
	if len(items) < 1 {
		return
	}
	DataDict.Lock()
	defer DataDict.Unlock()
	for _, item := range items {
		if _, ok := DataDict.Map[previousItems]; !ok {
			DataDict.Map[previousItems] = []string{}
		}
		DataDict.Map[previousItems] = append(DataDict.Map[previousItems], item)
		previousItems[0] = previousItems[1]
		previousItems[1] = item
	}
	DataDict.Map[previousItems] = append(DataDict.Map[previousItems], END)
}

func loadDataset(path string) (DataMapType, error) {
	res := make(DataMapType)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return res, nil
	}

	fileBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	reader := bytes.NewReader(fileBytes)

	dataDecoder := gob.NewDecoder(reader)

	err = dataDecoder.Decode(&res)
	return res, err
}

func saveDataset(path string) error {
	b := new(bytes.Buffer)

	dataEncoder := gob.NewEncoder(b)

	DataDict.RLock()
	err := dataEncoder.Encode(DataDict.Map)
	DataDict.RUnlock()
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, b.Bytes(), 0644)
}

func importFile(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return err
	}

	fileBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	msgSplt := bytes.Split(fileBytes, []byte("\n"))

	for _, msg := range msgSplt {
		trainMessage(string(msg))
	}
	return nil
}

func preprocessText(text string) string {
	// TODO: Add more here to clean up punctuation, etc.
	return strings.ToLower(text)
}

func processText(text string) []string {
	return SplitRegex.FindAllString(text, -1)
}

func postprocessText(text string) string {
	return text
}
