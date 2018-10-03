package bot

import (
	"fmt"
)

func (br *BotRequest) CheckForAd(ch chan bool) {
	tokens := tokenizer(br.Message.Text)  // X features

	ch <- true
}

// 1. tokenize the sentence to array of words (unigrams) -- regexp?
// 2. normalize each word (lemmatizing vs stemming)
// 3. remove stopwords
// 4. return the list of tokens
func tokenizer(text string) []string {

}

type TextClassifierModel struct {

}

type NaiveBayesClassifier struct {
	X []string  // convert to slice
	y []string
}

type TextClassifier interface {
	fit() TextClassifierModel
	predict() map[string]float32
}

// train the NB clf model
func (clf *NaiveBayesClassifier) train() TextClassifierModel {

}

// naive bayes classification
// returns the probability
func (clf *NaiveBayesClassifier) predict() map[string]float32 {

}