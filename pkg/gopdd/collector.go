package gopdd

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"gopkg.in/alessio/shellescape.v1"
)

type Source struct {
	source     string
	file       string
	skipErrors bool
}

func (s Source) CollectPuzzles() ([]Puzzle, error) {
	puzzles := []Puzzle{}
	lines := strings.Split(strings.ReplaceAll(s.source, "\r\n", "\n"), "\n")
	for i, line := range lines {
		todo, err := FindTodo(line)
		if err != nil && !s.skipErrors {
			return puzzles, collectorError(err, s.file, i)
		} else if todo != nil {
			puzzle, err := s.PuzzleOf(*todo, lines[i+1:], i)
			if err != nil && !s.skipErrors {
				return puzzles, collectorError(err, s.file, i)
			}
			puzzles = append(puzzles, puzzle)
		}
	}
	return puzzles, nil
}

func collectorError(err error, path string, idx int) error {
	return fmt.Errorf("Error at %s:%d: %w", path, idx+1, err)
}

type TodoLine struct {
	line   string
	prefix string
	marker string
	title  string
}

func FindTodo(line string) (*TodoLine, error) {
	if !strings.Contains(strings.ToLower(line), "todo") {
		return nil, nil
	}
	formatErr := badFormat(line)
	if formatErr != "" {
		return nil, errors.New(formatErr)
	}
	re := regexp.MustCompile(`(.*(?:^|\s))(?:\x40todo|TODO:|TODO)\s+#([\w\-.:/]+)\s+(.+)`)
	matches := re.FindStringSubmatch(line)
	return &TodoLine{
		line:   matches[0],
		prefix: matches[1],
		marker: matches[2],
		title:  matches[3],
	}, nil
}

func badFormat(line string) string {
	if regexp.MustCompile(`[^\s]\x40todo`).MatchString(line) {
		return getNoLeadingSpaceError("@todo")
	}
	if regexp.MustCompile(`\x40todo(\s*[^\s#]|[^\s]*#)`).MatchString(line) {
		return getNoPuzzleMarkerError("@todo")
	}
	if regexp.MustCompile(`\x40todo\s+#\s`).MatchString(line) {
		return getNoSpaceAfterHashError("@todo")
	}

	if regexp.MustCompile(`[^\s]TODO:?`).MatchString(line) {
		return getNoLeadingSpaceError("TODO")
	}
	if regexp.MustCompile(`TODO:?(\s*[^\s#]|[^\s]*#)`).MatchString(line) {
		return getNoPuzzleMarkerError("TODO")
	}
	if regexp.MustCompile(`TODO:?\s+#\s`).MatchString(line) {
		return getNoSpaceAfterHashError("TODO")
	}
	return ""
}

func getNoLeadingSpaceError(todo string) string {
	return todo + " must have a leading space to become " +
		"a puzzle, as this page explains: https://github.com/cqfn/pdd#how-to-format"
}

func getNoSpaceAfterHashError(todo string) string {
	return todo + " found, but there is an unexpected space " +
		"after the hash sign, it should not be there, " +
		"see https://github.com/cqfn/pdd#how-to-format"
}

func getNoPuzzleMarkerError(todo string) string {
	return todo + " found, but puzzle can't be parsed, " +
		fmt.Sprintf("most probably because %s is not followed by a puzzle marker, ", todo) +
		"as this page explains: https://github.com/cqfn/pdd#how-to-format"
}

type Puzzle struct {
	id    string
	lines string
	body  string
	file  string
	git   GitInfo
}

func (s Source) PuzzleOf(td TodoLine, following []string, idx int) (Puzzle, error) {
	tail, err := TailOf(following, td.prefix, todoOffset(td.line))
	if err != nil {
		return Puzzle{}, err
	}
	body := strings.TrimSpace(regexp.MustCompile(`\s+`).ReplaceAllString(td.title+strings.Join(tail, " "), " "))
	body = strings.TrimSpace(strings.TrimSuffix(body, "*/-->"))
	marker, err := MarkerOf(td.marker)
	return Puzzle{
		id:    PuzzleId(marker, body),
		lines: fmt.Sprintf("%d-%d", idx, idx+len(tail)+1),
		body:  body,
		file:  s.file,
		git:   s.GetGitInfo(idx + 1),
	}, nil

}

func PuzzleId(marker Marker, body string) string {
	return fmt.Sprintf("%s-%s", marker.ticket, getMD5Hash(body)[0:7])
}

func getMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func todoOffset(line string) int {
	return len(line) - len(strings.TrimLeftFunc(line, unicode.IsSpace))
}

func TailOf(rest []string, prefix string, offset int) ([]string, error) {
	if prefix == "" {
		prefix = strings.Repeat(" ", offset)
	}

	indented := isIndented(rest)
	tail := []string{}
	for _, line := range rest {
		ended, err := tailEnded(line, prefix, indented)
		if ended || err != nil {
			return tail, err
		}
		tail = append(tail, strings.TrimLeftFunc(line[len(prefix):], unicode.IsSpace))
	}
	return tail, nil
}

func tailEnded(line string, prefix string, indented bool) (bool, error) {
	match, err := FindTodo(line)
	if err != nil {
		return true, err
	}
	if match != nil {
		return true, nil
	}
	start := prefix
	if len(line) <= len(prefix) {
		start = strings.TrimRightFunc(prefix, unicode.IsSpace)
	}
	if indented && !strings.HasPrefix(line[len(prefix):], start) {
		return true, nil
	}
	return false, nil
}

func isIndented(rest []string) bool {
	var first string
	if len(rest) > 0 {
		first = rest[0]
	}
	return strings.HasPrefix(first, " ")
}

type Marker struct {
	ticket   string
	estimate int
	role     string
}

func MarkerOf(text string) (Marker, error) {
	match := regexp.MustCompile(`([\w\-.]+)(?::(\d+)(?:(m|h)[a-z]*)?)?(?:/([A-Z]+))?`).FindStringSubmatch(text)
	if match == nil {
		return Marker{}, errors.New(fmt.Sprintf("Invalid puzzle marker \"%s\", most probably formatted"+
			" against the rules explained here: https://github.com/cqfn/pdd#how-to-format", text))
	}
	role := "DEV"
	if match[4] != "" {
		role = match[4]
	}
	return Marker{
		ticket:   match[1],
		estimate: minutesOf(match[2], match[3]),
		role:     role,
	}, nil

}

func minutesOf(num string, units string) int {
	if num == "" {
		return 0
	}
	minutes, err := strconv.Atoi(num)
	if err != nil {
		panic(err)
	}
	if units == "" || strings.HasPrefix(units, "h") {
		minutes *= 60
	}
	return minutes
}

type GitInfo struct {
	author string
	email  string
	time   string
}

func (s Source) GetGitInfo(pos int) GitInfo {
	info := GitInfo{}
	if !IsInsideWorkTree(s.file) {
		return info
	}
	for _, line := range GetBlame(s.file, pos) {
		if regexp.MustCompile(`^author `).MatchString(line) {
			info.author = regexp.MustCompile(`^author `).ReplaceAllString(line, "")
		} else if regexp.MustCompile(`^author-mail [^@]+@[^.]+\..+`).MatchString(line) {
			info.email = regexp.MustCompile(`^author-mail <(.+)>$`).ReplaceAllString(line, "$1")
		} else if regexp.MustCompile(`^author-time ([0-9]+)$`).MatchString(line) {
			ts, err := strconv.ParseInt(
				regexp.MustCompile(`^author-time ([0-9]+)$`).ReplaceAllString(line, "$1"),
				10, 64)
			if err != nil {
				panic(err)
			}
			info.time = time.Unix(ts, 0).Format(time.RFC3339)
		}
	}

	return info
}

func IsInsideWorkTree(file string) bool {
	dir := shellescape.Quote(path.Dir(file))
	output, err := exec.Command(
		"bash",
		"-c",
		fmt.Sprintf("cd %s && git rev-parse --is-inside-work-tree 2>/dev/null", dir),
	).Output()
	if err != nil {
		panic(err)
	}
	return strings.TrimSpace(string(output)) == "true"
}

func GetBlame(file string, linenum int) []string {
	dir := shellescape.Quote(path.Dir(file))
	name := shellescape.Quote(path.Base(file))
	println(fmt.Sprintf("cd %s && git blame -L %d,%d --porcelain %s", dir, linenum, linenum, name))
	output, err := exec.Command(
		"bash",
		"-c",
		fmt.Sprintf("cd %s && git blame -L %d,%d --porcelain %s", dir, linenum, linenum, name),
		// fmt.Sprintf("cd %s && git blame -L %d,%d --porcelain %s", dir, linenum, linenum, name),
	).Output()
	println(err.Error())
	if err != nil {
		panic(err)
	}
	return strings.Split(string(output), "\n")
}
