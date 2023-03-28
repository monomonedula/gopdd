package gopdd

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCollectPuzzlesOk(t *testing.T) {
	origindir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	defer os.Chdir(origindir)
	err = os.Chdir("../..")
	if err != nil {
		panic(err)
	}
	path := "resources/foobar.py"
	content, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	puzzles, err := Source{file: path, source: string(content)}.CollectPuzzles(false)
	assert.Equal(t, nil, err)
	assert.Equal(t, len(puzzles), 2, "Must find two puzzles")
	assert.Equal(
		t,
		Puzzle{
			Id:       "209-c992021",
			Ticket:   "209",
			Estimate: 30,
			Role:     "DEV",
			Lines:    "3-5",
			Body:     "whatever 1234. Please fix soon 1.",
			File:     "resources/foobar.py",
			Author:   "monomonedula",
			Email:    "valh@tuta.io",
			Time:     "2023-03-26T23:27:31+03:00",
		},
		puzzles[0],
	)
	assert.Equal(
		t,
		Puzzle{
			Id:       "321-b7bbd66",
			Ticket:   "321",
			Estimate: 60,
			Role:     "DEV",
			Lines:    "9-11",
			Body:     "very important issue. Please fix soon 2.",
			File:     "resources/foobar.py",
			Author:   "monomonedula",
			Email:    "valh@tuta.io",
			Time:     "2023-03-26T23:27:31+03:00",
		},
		puzzles[1],
	)
}

func TestCollectSourcesOk(t *testing.T) {
	origindir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	defer os.Chdir(origindir)
	err = os.Chdir("../..")
	if err != nil {
		panic(err)
	}
	sources, err := MakeSources(
		".",
		[]string{
			"*/must_be_excluded.py",
			"*.txt",
			"pkg/*",
			"go.mod",
			"go.sum",
			"cmd/*",
		},
		[]string{},
		true,
	)
	assert.Equal(t, nil, err)
	found := sources.fetch()
	assert.Equal(t, 1, len(found))
	cwd, _ := os.Getwd()
	assert.Equal(t, path.Join(cwd, "resources/foobar.py"), found[0].file)
}
