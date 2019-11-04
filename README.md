# Stats Aggregator 2

Aggregator for SauerDuels stats.

```
python3 aggregate.py
```

## Requirements

Python 3 must be installed on your system. This program additionally requires several additional Python packages which can be installed using pip with the following commands:

```
pip3 install pandas
pip3 install pyyaml
```

## Input

This tool requires the *logs* directoy to be present in the working directory. This directory must contain several directories names after SauerDuels events (*sd01*, *sd02*, ...), each of which must contain one or more files in the [SauerDuels server](https://github.com/sauerduels/sauer-server) stats log csv format (specified by the `-s` server switch).

Forfeits in server logs are represented by games in the same stats log format, but with all stats set to 0 except the frags of the winning player, which are set to any value greater than or equal to `1000`. These can be placed in their own csv file.

It also requires the file *events.yml* to be present in the working directory. It is a yaml file exported by the challonge crawler.

## Output

The output is as follows:

- **output/[ffa|insta|effic]/stats.yml**: Mode rankings in yaml format.
- **output/[ffa|insta|effic]/weapon_stats.yml**: Mode weapon stats in yaml format.
- **output/total/stats.yml**: Cumulative rankings in yaml format.

The contents of the *output* directory can be placed directly in the *_data* subdirectory of the [SauerDuels website](https://github.com/sauerduels/website) repo.

# Challonge Crawler

Given the url of a domain on challonge, this script crawls the index page and the pages of every completed event, extracts player names organized by event and group, and dumps them into a yaml file.

```
python3 crawl_challonge.py
```

## Requirements

Python 3 must be installed on your system. This program additionally requires an additional Python packages which can be installed using pip with the following command:

```
pip3 install pyyaml
pip3 install BeautifulSoup4
```

## Input

In *crawl_challonge.py* change `base_url` to the desired challonge url of the format `https://{domain}.challonge.com/`.

## Output

*events.yml*.

# Demo Stats Extractor

Extracts stats from demos and dumps them in the [SauerDuels server](https://github.com/sauerduels/sauer-server) stats log format.

```
go run extractor.go DEMO_1|DIRECTORY_1 [DEMO_2|DIRECTORY_2]
```

## Requirements

The Go compiler must be installed on your system.

## Input

The program takes one or more .dmo files and directory paths as arguments. If an argument is a directory, all files directly inside of it are parsed.

## Output

For each file or directory given as an argument, this program creates a single file with the results of the demo(s) in the [SauerDuels server](https://github.com/sauerduels/server) stats log format.

## Notes

- There is a bug with overtime handling. Games with overtime will appear as if intermission had occurred instead of the first overtime.

# Name Replacer

Replaces, in-place and recursively, the names of players in files with their recognized counterparts, based on a predefined map.

```
replace-names.sh DIRECTORY
```

## Input

One directory to be searched recursively.

## Output

In-place.

## Notes

- It is very important that you keep a backup of the file(s) on which you wish to run this script. It is destructive and may lead to data corruption.
