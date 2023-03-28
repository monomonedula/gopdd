package gopdd

import (
	"encoding/json"
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
	b.Logger.Info("My version is ...", buildInfo.Main.Version)
	b.Logger.Infof("Go version is %s", runtime.Version())
	b.Logger.Info("Reading from root dir %s", dir)

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
			b.Logger.Info("Puzzle %s %s/%s at %s", p.Id, p.Estimate, p.Role, p.File)
			puzzles = append(puzzles, p)
		}
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
