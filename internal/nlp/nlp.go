// SPDX-License-Identifier: GPL-2.0
package nlp

import (
	"fmt"
)

func Vectorize(ch chan bool, text string) {
	tokens := tokenizer(text) // X features
	if tokens != nil {
		fmt.Printf("good")
	}

	ch <- true
}

// 1. tokenize the sentence to array of words (unigrams) -- regexp?
// 2. normalize each word (lemmatizing vs stemming)
// 3. remove stopwords
// 4. return the list of tokens
func tokenizer(text string) []string {
	var tokens []string
	return tokens
}

type TextClassifierModel struct {
}

type NaiveBayesClassifier struct {
	X []string // convert to slice
	y []string
}

type TextClassifier interface {
	fit() *TextClassifierModel
	predict() map[string]float32
}

// train the NB clf model
func (clf *NaiveBayesClassifier) train() *TextClassifierModel {
	var model TextClassifierModel
	return &model
}

// naive bayes classification
// returns the probability
func (clf *NaiveBayesClassifier) predict() map[string]float32 {
	var result map[string]float32

	return result
}
