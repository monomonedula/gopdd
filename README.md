# GoPdd 

### PDD puzzle collector



A port of [the original PDD written in Ruby](https://github.com/cqfn/pdd/), but with JSON output and written in Go, so it compiles into a single pretty executable.

Expected formatting and most of the functionality is the same, so feel free to use documentation from the original repo for reference on how to write puzzles.

```
NAME:
   GoPdd - Todo puzzle collector

USAGE:
   GoPdd [global options] command [command options] [arguments...]

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --source value, -s value                                 Source directory to parse (default: ".")
   --file value, -f value                                   File to save JSON output into
   --verbose, -v                                            Enable verbose mode (a lot of logging) (default: false)
   --skip-gitignore                                         Don't look into .gitignore for excludes (default: false)
   --skip-errors                                            Suppress error as warning and skip badly formatted puzzles (default: false)
   --rule value, -r value [ --rule value, -r value ]        Rule to apply (can be used many times). Possible values: 'max-estimate:<int>', 'min-estimate:<int>', 'min-words:<int>', 'available-roles:<ROLENAME>,<ROLANME>...'
   --include value, -n value [ --include value, -n value ]  Glob pattern to include, e.g. "**/*.jpg"
   --exclude value, -e value [ --exclude value, -e value ]  Glob pattern to exclude, e.g. "**/*.jpg"
   --help, -h                                               show help
```


Example output:
```json
[
  {
    "id": "209-c992021",
    "ticket": "209",
    "estimate": 30,
    "role": "DEV",
    "lines": "3-5",
    "body": "whatever 1234. Please fix soon 1.",
    "file": "resources/foobar.py",
    "author": "monomonedula",
    "email": "email@xxx.xyz",
    "time": "2023-03-26T23:27:31+03:00"
  },
  {
    "id": "321-b7bbd66",
    "ticket": "321",
    "estimate": 60,
    "role": "DEV",
    "lines": "9-11",
    "body": "very important issue. Please fix soon 2.",
    "file": "resources/foobar.py",
    "author": "monomonedula",
    "email": "email@xxx.xyz",
    "time": "2023-03-26T23:27:31+03:00"
  }
]
```

Installation:
```
go install -v github.com/monomonedula/gopdd/cmd/gopdd@latest
```

Tested on MacOS and Linux.


