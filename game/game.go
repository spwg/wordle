// Package game provides functionality for creating and playing Wordle games.
package game

import (
	"fmt"
	"math/rand"
	"os"
	"strings"

	"github.com/fatih/color"
)

// Wordle is a game of wordle.
type Wordle struct {
	ans     string
	guesses []string
	dict    map[string]struct{}
}

// New constructs a *Wordle.
func New(dictionaryPath string, answer string) (*Wordle, error) {
	b, err := os.ReadFile(dictionaryPath)
	if err != nil {
		return nil, fmt.Errorf("ReadFile(%q) failed: %v", dictionaryPath, err)
	}
	words := strings.Fields(string(b))
	lenFiveWords := []string{}
	for _, w := range words {
		if len(w) == 5 {
			lenFiveWords = append(lenFiveWords, w)
		}
	}
	if answer == "" {
		answer = lenFiveWords[rand.Intn(len(lenFiveWords))]
	}
	if len(answer) != 5 {
		return nil, fmt.Errorf("answer %q must have 5 characters, not %v", answer, len(answer))
	}
	answer = strings.ToLower(answer)
	dict := map[string]struct{}{}
	for _, w := range lenFiveWords {
		dict[strings.ToLower(w)] = struct{}{}
	}
	return &Wordle{ans: answer, dict: dict}, nil
}

func (w *Wordle) format(won bool) (string, error) {
	var b strings.Builder
	for i := 0; i < len(w.guesses); i++ {
		if _, err := b.WriteString(formatGuess(w.guesses[i], w.ans, won)); err != nil {
			return "", err
		}
		if i < len(w.guesses)-1 {
			if _, err := b.WriteString("\n"); err != nil {
				return "", err
			}
		}
	}
	if _, err := b.WriteString("\n"); err != nil {
		return "", err
	}
	return b.String(), nil
}

func formatGuess(guess, answer string, won bool) string {
	var s string
	var share string
	for i := 0; i < len(guess); i++ {
		var c *color.Color
		switch {
		case guess[i] == answer[i]:
			share += "🟩"
			c = color.New(color.BgGreen)
		case strings.ContainsRune(answer, rune(guess[i])):
			share += "🟨"
			c = color.New(color.BgYellow)
		default:
			share += "⬜"
			c = color.New(color.BgBlack).Add(color.BgWhite)
		}
		c = c.Add(color.FgBlack)
		s += c.Sprintf(" %s ", string(guess[i]))
		s += " "
	}
	if won {
		return share
	}
	return s
}

// GuessResult tells you the result of guessing a word.
type GuessResult struct {
	inDictionary bool
	won          bool
	guess        string
	g            *Wordle
}

func (gr *GuessResult) Format() (string, error) {
	var s string
	switch {
	case gr.won:
		s = gr.guess + " is correct!"
	case gr.inDictionary:
		s = gr.guess + " is not correct."
	default:
		s = gr.guess + " is not a word."
	}
	gf, err := gr.g.format(gr.won)
	if err != nil {
		return "", err
	}
	return s + "\n" + gf, nil
}

// Guess processes the word to check for if the word is correct and returns a
// *GuessResult that contains information about the state of the game.
func (w *Wordle) Guess(word string) (*GuessResult, error) {
	if len(word) != 5 {
		return nil, fmt.Errorf("word %q does have 5 characters", word)
	}
	gr := &GuessResult{guess: word, g: w}
	if w.isInDictionary(word) {
		gr.inDictionary = true
		w.saveGuess(word)
	}
	if w.isGuessCorrect(word) {
		gr.won = true
	}
	return gr, nil
}

// Search returns possible answers with the information that the game
// player has, not the answer directly.
func (w *Wordle) Search() ([]string, error) {
	// There are more efficient ways to search text than what this method uses.
	// Some alternatives: a trigram, constructing a regexp and filtering based
	// on that, and binary searching the sorted words (it is a dictionary after
	// all) for a prefix if the guesses reveal one.
	results := []string{}
	exact := map[int]rune{}
	contains := map[rune]int{}
	for _, g := range w.guesses {
		for i, c := range g {
			contains[c] = strings.Count(w.ans, string(c))
			if c == rune(w.ans[i]) {
				exact[i] = c
			}
		}
	}
	for w := range w.dict {
		wOk := true
		for i, c := range w {
			r, ok := exact[i]
			if ok && r != c {
				wOk = false
				break
			}
		}
		if !wOk {
			continue
		}
		for r, cnt := range contains {
			if strings.Count(w, string(r)) != cnt {
				wOk = false
				break
			}
		}
		if wOk {
			results = append(results, w)
		}
	}
	return results, nil
}

func (w *Wordle) isInDictionary(word string) bool {
	_, ok := w.dict[word]
	return ok
}

func (w *Wordle) saveGuess(guess string) {
	w.guesses = append(w.guesses, guess)
}

func (w *Wordle) isGuessCorrect(guess string) bool {
	return guess == w.ans
}
