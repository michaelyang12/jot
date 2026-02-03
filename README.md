# jot

Quick sticky notes from the terminal, synced to [Turso](https://turso.tech).

## Install

```
go install github.com/michaelyang12/jot@latest
```

## Setup

Create a Turso database and set your env vars:

```bash
turso db create jot
export JOT_URL="$(turso db show jot --url)"
export JOT_TOKEN="$(turso db tokens create jot)"
```

Add the exports to your shell profile (`~/.bashrc`, `~/.zshrc`, etc.) to persist them.

The table is created automatically on first use.

## Usage

```
jot <text>       add a note
jot ls           list all notes
jot peek <id>    view a note
jot rm <id>      delete a note
jot pop          view + delete the latest note
```

### Examples

```
$ jot train api for image classifier
noted (#1)

$ jot look into turso branching
noted (#2)

$ jot ls
  2  look into turso branching   just now
  1  train api for image classifier   2 mins ago

$ jot peek 1
#1  2 mins ago

train api for image classifier

$ jot pop
#2  just now

look into turso branching

(removed)
```
