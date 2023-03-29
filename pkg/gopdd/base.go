package gopdd

import (
	"encoding/json"
	"errors"
	"os"
	"runtime"
	"runtime/debug"

	"github.com/sirupsen/logrus"
)

type Base struct {
	Dir           string
	Exclude       []string
	Include       []string
	Rules         []Rule
	SkipGitignore bool
	Logger        *logrus.Logger
}

func (b Base) Puzzles(skipErrors bool) []Puzzle {
	var dir string
	if b.Dir != "" {
		dir = b.Dir
	} else {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			panic(err)
		}
	}
	buildInfo, _ := debug.ReadBuildInfo()
	b.Logger.Infof("My version is %s", buildInfo.Main.Version)
	b.Logger.Infof("Go version is %s", runtime.Version())
	b.Logger.Infof("Reading from root dir %s", dir)

	sources, err := MakeSources(dir, b.Exclude, b.Include, !b.SkipGitignore)
	if err != nil {
		panic(err)
	}
	var puzzles []Puzzle
	for _, file := range sources.fetch() {
		collected, err := file.CollectPuzzles(skipErrors)
		if err != nil {
			panic(err)
		}
		for _, p := range collected {
			b.Logger.Infof("Puzzle %s %d/%s at %s", p.Id, p.Estimate, p.Role, p.File)
			puzzles = append(puzzles, p)
		}
	}
	err = b.ApplyRules(puzzles)
	if err != nil {
		panic(err)
	}
	return puzzles
}

func (b Base) JsonPuzzles(skipErrors bool) []byte {
	out, err := json.Marshal(b.Puzzles(skipErrors))
	if err != nil {
		panic(err)
	}
	return out
}

func (b Base) ApplyRules(puzzles []Puzzle) error {
	total := 0
	for _, rule := range b.Rules {
		errors := rule.ApplyTo(puzzles)
		total += len(errors)
		for _, err := range errors {
			b.Logger.Error(err)
		}
	}
	if total == 0 {
		return nil
	}
	b.Logger.Errorf("Got %d errors. See logs above", total)
	return errors.New("puzzles do not comply with the rules")
}
