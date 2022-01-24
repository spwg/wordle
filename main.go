package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"time"
	"wordle/game"

	"github.com/google/shlex"
	"golang.org/x/term"
)

var (
	dictPath = flag.String("dictionary", "/usr/share/dict/words", "Absolute file path of a dictionary file.")
)

func main() {
	rand.Seed(time.Now().Unix())
	ctx := context.Background()
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()
	if err := run(ctx, *dictPath); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, dictPath string) error {
	g, err := game.New(dictPath, "")
	if err != nil {
		return err
	}
	rw := struct {
		io.Reader
		io.Writer
	}{os.Stdin, os.Stdout}
	old, err := term.MakeRaw(0)
	if err != nil {
		return err
	}
	defer term.Restore(0, old)
	t := term.NewTerminal(rw, "> ")
	gcmd := &guesscmd{g}
	scmd := &searchcmd{g}
	t.Write([]byte("Guess a word.\n"))
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		line, err := t.ReadLine()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		args, err := shlex.Split(line)
		if err != nil {
			return err
		}
		if len(args) == 0 {
			continue
		}
		var termOutput string
		switch args[0] {
		case "search":
			// This is safe because guesses have to be exactly five characters,
			// so it's not possible that "search" can a valid word in the game
			// itself.
			r, err := scmd.run(ctx)
			if err != nil {
				return err
			}
			termOutput = r
		default:
			r, err := gcmd.run(ctx, args)
			if err != nil {
				return err
			}
			termOutput = r
		}
		if _, err := t.Write([]byte(termOutput)); err != nil {
			return err
		}
	}
}

type guesscmd struct {
	g *game.Wordle
}

func (cmd *guesscmd) run(ctx context.Context, args []string) (string, error) {
	if len(args) != 1 {
		return "usage: <word>\n", nil
	}
	if len(args[0]) != 5 {
		return fmt.Sprintf("%q does not have 5 characters\n", args[0]), nil
	}
	gr, err := cmd.g.Guess(args[0])
	if err != nil {
		return "", err
	}
	grf, err := gr.Format()
	if err != nil {
		return "", err
	}
	return grf, nil
}

type searchcmd struct {
	g *game.Wordle
}

func (cmd *searchcmd) run(ctx context.Context) (string, error) {
	words, err := cmd.g.Search()
	if err != nil {
		return "", err
	}
	var s string
	for _, w := range words {
		s += w + "\n"
	}
	return s, nil
}
