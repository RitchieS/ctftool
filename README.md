# CTFTOOL

[![Go Reference](https://pkg.go.dev/badge/github.com/ritchies/ctftool.svg)](https://pkg.go.dev/github.com/ritchies/ctftool)
[![Go Report Card](https://goreportcard.com/badge/github.com/ritchies/ctftool)](https://goreportcard.com/report/github.com/ritchies/ctftool)
[![](https://img.shields.io/github/workflow/status/ritchies/ctftool/Tests?longCache=tru&label=Tests&logo=github%20actions&logoColor=fff)](https://github.com/ritchies/ctftool/actions?query=workflow%3ATests)

A cli tool to check upcoming CTF's from [ctftime.org](https://ctftime.org) and download challenges from [CTFd](https://ctfd.io).

# Overview

`ctftool` is a cli tool for interacting with [ctftime](https://ctftime.org) and [CTFd](https://ctfd.io) to list upcoming CTFs, download challenges from CTFd, get the scores of your team, and more.

It can:

- Interact with [ctftime](https://ctftime.org) and [CTFd](https://ctfd.io)
- List upcoming CTFs
- List the top 10 teams on ctftime and CTFd
- Download challenges from CTFd
- Creates a writeup template for each challenge

## Concepts

The tool is designed to allow quick acces to different features from your terminal using structured commands.

It is structured to allow for easy and quick access to different features from either ctftime or ctfd.

```bash
$ ctftool --help/help # Show help
$ ctftool ctftime # Aliases: time
$ ctftool ctfd # Aliases: download, d
```

If no command is given, the default command is `ctftool ctftime` or if a config file is found it will be used and the default command will be `ctftool ctfd` to update the challenges from that CTF.

#### Current issues

- [ ] Does not handle cloudflare correctly, yet
- [ ] Does not handle catpcha's, yet
- [ ] Does not support other CTF instances yet like;
  - [ ] rCTF

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
  -h, --help                help for ctftool
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
