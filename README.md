# Stats Aggregator (aggregator.go)

Aggregator and formatter for SauerDuels stats.

## Requirements

The Go compiler must be installed on your system. This program additionally requires the Go package *github.com/jlouis/glicko2*, which can be installed by running the following command:

```
go get github.com/jlouis/glicko2
```

## Input

This tool takes zero or more directory paths as arguments. Each of these directories must contain one or more subdirectories whose names correspond to event stages (group, finals, ...) Each of these subdirectories must contain one or more files in the [SauerDuels server](https://github.com/sauerduels/server) stats log format (specified by the `-s` server switch).

If given no arguments, the program simply rewrites the CSV and HTML files based on the data in *state.json*.

## Output

The output is as follows:

- **state.json**: Contains the entire state of player stats. This file is typically only useful for internal use by the program, and should be deleted if you wish to clear old stats and start over.
- **[mode]_[group].html**: A portion of an HTML file containing only two `<table>` elements with player stats in each *[mode]* and *[group]*, and a front matter with a `title` value for use with Jekyll.
- **[mode]_[group].csv**: A CSV file of player stats in each *[mode]* and *[group]*.

## Notes

- Directories (aka events) are always processed in the order in which they appear in the argument list.
- Forfeits are represented by games in the same stats log format, but with all stats set to 0 except the frags of the winning player, which are set to `99999`.
- You can add new events simply by running the program with the new event directory as an argument, as long as *state.json* is intact.
- Do **not** run the program on the same event directory more than once, or else stats are duplicated.

# Demo Stats Extractor (extractor.go)

Extracts stats from demos and outputs them in the [SauerDuels server](https://github.com/sauerduels/server) stats log format.

## Requirements

The Go compiler must be installed on your system.

## Input

The program takes one or more .dmo files and directory paths as arguments. If an argument is a directory, all files directly inside of it are parsed.

## Output

For each file or directory given as an argument, this program creates a single file with the results of the demo(s) in the [SauerDuels server](https://github.com/sauerduels/server) stats log format.

## Notes

- There is a bug with overtime handling. Games with overtime will appear as if intermission had occurred instead of the first overtime.

# Name Replacer (replace-names.sh)

Replaces, in-place and recursively, the names of players in files with their well-known counterparts, based on a predefined table.

## Input

One directory to be searched recursively.

## Output

In-place.

# Country Flag Inserter (insert-country-flags.sh)

Inserts country flags into HTML files for known players.

## Input

One HTML file.

## Output

In-place.

## Notes

- It is very important that you keep a backup of the file you want to run this script on. This script is destructive and may cause data corruption.
