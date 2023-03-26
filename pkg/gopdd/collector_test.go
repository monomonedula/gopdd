package gopdd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCollectPuzzles(t *testing.T) {
	path := "../../resources/foobar.py"
	content, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	puzzles, err := Source{file: path, source: string(content)}.CollectPuzzles()

	assert.Equal(t, len(puzzles), 2, "Must find two puzzles")
}
