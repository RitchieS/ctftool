# CTFTOOL

A cli tool to check upcoming CTF's from [ctftime.org](https://ctftime.org) and download challenges from [CTFd](https://ctfd.io).

[![Go Reference](https://pkg.go.dev/badge/github.com/ritchies/ctftool.svg)](https://pkg.go.dev/github.com/ritchies/ctftool)
[![Go Report Card](https://goreportcard.com/badge/github.com/ritchies/ctftool)](https://goreportcard.com/report/github.com/ritchies/ctftool)
[![](https://img.shields.io/github/workflow/status/ritchies/ctftool/Tests?longCache=tru&label=Tests&logo=github%20actions&logoColor=fff)](https://github.com/ritchies/ctftool/actions?query=workflow%3ATests)

# Overview

`ctftool` is a cli tool for interacting with [ctftime](https://ctftime.org) and [CTFd](https://ctfd.io) to list upcoming CTFs, download challenges from CTFd, get the scores of your team, and more.

It provides:

- Interactive CLI interface `cobra` and `bubbletea`
- Interact with [ctftime](https://ctftime.org)
- List upcoming CTFs
- List the top 10 teams
- Adjust information for each CTF locally stored in a sqlite db
- Interact with [CTFd](https://ctfd.io)
- Download challenges from CTFd
- Creates a writeup template for each challenge
- Get the top 10 teams in a CTF

### Upcoming features

- [ ] Get the scores of your team on ctftime.org
- [ ] Get the scores of your team on ctfd
- [ ] Search teams by name
- [ ] Search teams by country
- [ ] Scoreboard by country
- [ ] Ability to announce CTFs on Discord
- [ ] Various Discord related features, like creating channels, threads and allowing users to solve challenges
- [ ] Complete interactive TUI using bubbletea

## Concepts

The tool is designed to allow quick acces to different features from your terminal using structured commands.

It is structured to allow for easy and quick access to different features from either ctftime or ctfd.

```bash
$ ctftool --help/help # Show help
$ ctftool ctftime # Aliases: time
$ ctftool ctfd # Aliases: download, d
```

If no command is given, the default command is `ctftool ctftime` or if a config file is found it will be used and the default command will be `ctftool ctfd` to update the challenges from that CTF.

#### Warning

For `ctftool ctfd` to work without supplying `--username`, `--password`, `--url`, and `--output` you can save a config file to the output directory by specifying `--save-config`. **However, this will save the credentials in plaintext to the output directory, do not use this if you don't want to store your credentials in plaintext. - YOU HAVE BEEN WARNED**

#### Current issues

- [ ] Does not handle cloudflare correctly
- [ ] Does not handle catpcha's yet

## Commands

```plain
Available Commands:
  completion  Generate shell completion script
  ctfd        Query CTFd instance
  ctftime     Query CTFTime
  help        Help about any command
  version     Print the version number

Flags:
      --config string       Config file (default is .ctftool.yaml)
      --db-path string      Path to the database file (default "ctftool.sqlite")
  -h, --help                help for ctftool
      --interactive         Interactive mode
      --log-format string   Logger output format (text|json) (default "text")
  -v, --verbose             Verbose logging
  -V, --version             Print version information
```

```bash
$ ctftool ctftime # Aliases: time
$ ctftool ctftime top # List the top 10 teams
$ ctftool ctftime team 1 # Get information about a specific team by id
$ ctftool ctftime event 1 # Get information about a specific event by id
$ ctftool ctftime event --id=1 # Get information about a specific event

$ ctftool ctfd # Aliases: download, d
$ ctftool ctfd top --url=<url> # List the top 10 teams
$ ctftool ctfd --username=<user> --password=<pass> --url=<url> --output=<output> # Download challenges from CTFd
$ ctftool ctfd --username=<user> .... --save-config # Download challenges from CTFd and save config file
```

# Installing

Installing ctftool is easy, make sure you are on a recent version of Go and then run:

```bash
go install github.com/ritchies/ctftool@latest
```

# Usage

`ctftool` is a command line program to interact with [ctftime](https://ctftime.org) and [CTFd](https://ctfd.io). It is designed to allow quick access to different features from your terminal using structured commands to quickly interact with the CTFs.

## Completion

```bash
Bash:

  $ source <(ctftool completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ ctftool completion bash > /etc/bash_completion.d/ctftool
  # macOS:
  $ ctftool completion bash > /usr/local/etc/bash_completion.d/ctftool

Zsh:

  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ ctftool completion zsh > "${fpath[1]}/_ctftool"

  # You will need to start a new shell for this setup to take effect.

fish:

  $ ctftool completion fish | source

  # To load completions for each session, execute once:
  $ ctftool completion fish > ~/.config/fish/completions/ctftool.fish

PowerShell:

  PS> ctftool completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> ctftool completion powershell > ctftool.ps1
  # and source this file from your PowerShell profile.

Usage:
  ctftool completion [bash|zsh|fish|powershell]
```

# License

ctftool is released under the MIT license. See [LICENSE](https://github.com/ritchies/ctftool/blob/master/LICENSE) for more information.
