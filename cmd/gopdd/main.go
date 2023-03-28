package main

import (
	"fmt"
	"log"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/monomonedula/gopdd/pkg/gopdd"
)

func main() {
	app := &cli.App{
		Name:  "GoPdd",
		Usage: "Todo puzzle collector",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "source",
				Value:   ".",
				Usage:   "Source directory to parse ('.' by default)",
				Aliases: []string{"s"},
			},
			&cli.StringFlag{
				Name:    "file",
				Usage:   "File to save JSON output into",
				Aliases: []string{"f"},
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Usage:   "Enable verbose mode (a lot of logging)",
				Aliases: []string{"v"},
			},
			&cli.BoolFlag{
				Name:  "skip-gitignore",
				Usage: "Don't look into .gitignore for excludes",
			},
			&cli.BoolFlag{
				Name:  "skip-errors",
				Usage: "Suppress error as warning and skip badly formatted puzzles",
			},
			&cli.StringSliceFlag{
				Name:    "rule",
				Aliases: []string{"r"},
				Usage: "Rule to apply (can be used many times). Possible values: " +
					"'max-estimate:<int>', 'min-estimate:<int>', 'min-words:<int>', 'available-roles:<ROLENAME>,<ROLANME>...'",
			},
			&cli.StringSliceFlag{
				Name:    "include",
				Aliases: []string{"n"},
				Usage:   "Glob pattern to include, e.g. \"**/*.jpg\"",
			},
			&cli.StringSliceFlag{
				Name:    "exclude",
				Aliases: []string{"e"},
				Usage:   "Glob pattern to exclude, e.g. \"**/*.jpg\"",
			},
		},
		Action: Run,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func Run(ctx *cli.Context) error {
	dir := ctx.String("source")
	file := ctx.String("file")
	verbose := ctx.Bool("verbose")
	skipGitignore := ctx.Bool("skip-gitignore")
	skipErrors := ctx.Bool("skip-errors")
	rules := ctx.StringSlice("rule")
	include := ctx.StringSlice("include")
	exclude := ctx.StringSlice("exclude")
	logger := logrus.New()
	if verbose {
		logger.Level = logrus.InfoLevel
	} else {
		logger.Level = logrus.WarnLevel
	}
	output := gopdd.Base{
		Dir:           dir,
		Exclude:       exclude,
		Include:       include,
		SkipGitignore: skipGitignore,
		Rules:         gopdd.RulesOf(rules),
		Logger:        logger,
	}.JsonPuzzles(skipErrors)
	if file == "" {
		fmt.Println(string(output))
		return nil
	}
	err := os.WriteFile(file, output, 0644)
	return err
}
