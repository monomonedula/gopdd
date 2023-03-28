package gopdd

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/exp/slices"
)

type Rule interface {
	ApplyTo([]Puzzle) []string
}

type MinStimateRule struct {
	value int
}

func (r MinStimateRule) ApplyTo(puzzles []Puzzle) []string {
	var errors []string
	for _, p := range puzzles {
		if p.Estimate < r.value {
			errors = append(
				errors,
				fmt.Sprintf(
					"estimate on %s:%s is too low: %d. min estimate allowed: %d",
					p.File,
					p.Lines,
					p.Estimate,
					r.value,
				),
			)
		}
	}
	return errors
}

type MaxEstimateRule struct {
	value int
}

func (r MaxEstimateRule) ApplyTo(puzzles []Puzzle) []string {
	var errors []string
	for _, p := range puzzles {
		if p.Estimate > r.value {
			errors = append(
				errors,
				fmt.Sprintf(
					"estimate on %s:%s is too high: %d. max estimate allowed: %d",
					p.File,
					p.Lines,
					p.Estimate,
					r.value,
				),
			)
		}
	}
	return errors
}

type MinWordsRule struct {
	value int
}

func (r MinWordsRule) ApplyTo(puzzles []Puzzle) []string {
	var errors []string
	for _, p := range puzzles {
		wordCount := len(strings.Split(p.Body, " "))
		if wordCount > r.value {
			errors = append(
				errors,
				fmt.Sprintf(
					"puzzle on %s:%s has a very short description of just %d words"+
						" while a minimum of just %d is required.",
					p.File,
					p.Lines,
					p.Estimate,
					r.value,
				),
			)
		}
	}
	return errors
}

type MaxDuplicatesRule struct {
	value int
}

func (r MaxDuplicatesRule) ApplyTo(puzzles []Puzzle) []string {
	var errors []string
	m := map[string][]Puzzle{}
	for _, p := range puzzles {
		m[p.Body] = append(m[p.Body], p)

	}
	for _, v := range m {
		if len(v) > r.value {
			var dupes []string
			for _, puzzle := range v {
				dupes = append(dupes, puzzle.File+":"+puzzle.Lines)
			}
			errors = append(
				errors,
				fmt.Sprintf(
					"there are %d duplicate(s) of the same puzzle: %s, while maximum of %d duplicates is allowed",
					len(v),
					strings.Join(dupes, ", "),
					r.value,
				),
			)
		}
	}
	return errors
}

type AvailableRolesRule struct {
	roles []string
}

func (r AvailableRolesRule) ApplyTo(puzzles []Puzzle) []string {
	var errors []string
	for _, p := range puzzles {
		if !slices.Contains(r.roles, p.Role) {
			err := fmt.Sprintf("puzzle %s:%s", p.File, p.Lines)
			if p.Role == "" {
				err += " doesn't definy any role, while one of there roles is required: " +
					strings.Join(r.roles, ", ")
			} else {
				err += fmt.Sprintf(" defines role %s while one of there roles is required: %s", p.Role, strings.Join(r.roles, ", "))
			}
			errors = append(errors, err)
		}
	}
	return errors
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

func RulesOf(rulesstrings []string) []Rule {
	var rules []Rule
	for _, rulestr := range rulesstrings {
		if strings.HasPrefix(rulestr, "min-estimate:") {
			value := strings.Split(rulestr, ":")[1]
			parsed, err := strconv.Atoi(value)
			if err != nil {
				panic(err)
			}
			rules = append(rules, MinStimateRule{parsed})
		} else if strings.HasPrefix(rulestr, "max-estimate:") {
			value := strings.Split(rulestr, ":")[1]
			parsed, err := strconv.Atoi(value)
			if err != nil {
				panic(err)
			}
			rules = append(rules, MaxEstimateRule{parsed})
		} else if strings.HasPrefix(rulestr, "min-words:") {
			value := strings.Split(rulestr, ":")[1]
			parsed, err := strconv.Atoi(value)
			if err != nil {
				panic(err)
			}
			rules = append(rules, MinWordsRule{parsed})
		} else if strings.HasPrefix(rulestr, "available-roles:") {
			value := strings.Split(rulestr, ":")[1]
			roles := strings.Split(value, ",")
			rules = append(rules, AvailableRolesRule{roles})
		}
	}
	rules = append(rules, MaxDuplicatesRule{1})
	return rules
}
