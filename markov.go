package main

import (
	"encoding/gob"
	"log"
	"math/rand"
	"os"
	"regexp"
	"strings"
)

const End = "@END@"

var (
	Start = [2]string{"@START_1@", "@START_2@"}
	punctuation = map[string]bool{
		".": true,
		",": true,
		"!": true,
		"?": true,
		";": true,
		":": true,
		"&": true,
	}
	SplitRegex = regexp.MustCompile(`([\w'-]+|[.,!?;&])`)
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
		previousItems[0] = Start[1]
		previousItems[1] = seed[0]
		response = seed[0]
	} else {
		previousItems = Start
	}
	if _, ok := DataDict[previousItems]; !ok {
		return "Error! I don't understand that =("
	}
	for {
		options, ok := DataDict[previousItems]
		if !ok {
			return response
		}
		nextItem := options[rand.Intn(len(options))]
		if nextItem == End {
			return response
		}
		if _, isPunctuation := punctuation[nextItem]; isPunctuation {
			response = response + nextItem
		} else {
			response = response + " " + nextItem
		}
	}
}

func trainMessage(msg string) {
	items := processText(preprocessText(msg)) // Split by whitespace to get individual words
	previousItems := Start
	if len(items) < 1 {
		return
	}
	for _, item := range items {
		if _, ok := DataDict[previousItems]; !ok {
			DataDict[previousItems] = []string{}
		}
		DataDict[previousItems] = append(DataDict[previousItems], item)
		previousItems[0] = previousItems[1]
		previousItems[1] = item
	}
	DataDict[previousItems] = append(DataDict[previousItems], End)
}

func loadDataset(path string) (DataMapType, error) {
	res := make(DataMapType)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return res, nil

	}

	dataFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer log.Println(dataFile.Close())

	dataDecoder := gob.NewDecoder(dataFile)

	err = dataDecoder.Decode(&res)
	return res, err
}

func saveDataset(path string) error {
	dataFile, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer log.Println(dataFile.Close())

	dataEncoder := gob.NewEncoder(dataFile)
	err = dataEncoder.Encode(DataDict)
	if err != nil {
		return err
	}

	return nil
}

// TODO
func importFile(fp string) {
}

// TODO: Add more here to clean up punctuation, etc.
func preprocessText(text string) string {
	return strings.ToLower(text)
}

func processText(text string) []string {
	return SplitRegex.FindAllString(text, -1)
}

func postprocessText(text string) string {
	return text
}
