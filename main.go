package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const usage = `jot — quick sticky notes

usage:
  jot <text>       add a note
  jot ls           list all notes
  jot peek <id>    view a note
  jot rm <id>      delete a note
  jot pop          view + delete the latest note`

func main() {
	args := os.Args[1:]

	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" || args[0] == "help" {
		fmt.Println(usage)
		os.Exit(0)
	}

	cfg, err := LoadConfig()
	if err != nil {
		fatal(err)
	}

	db := NewDB(cfg.URL, cfg.Token)
	if err := db.Init(); err != nil {
		fatal(err)
	}

	cmd := args[0]
	switch cmd {
	case "ls":
		cmdList(db)
	case "peek":
		if len(args) < 2 {
			fatal(fmt.Errorf("usage: jot peek <id>"))
		}
		cmdPeek(db, args[1])
	case "rm":
		if len(args) < 2 {
			fatal(fmt.Errorf("usage: jot rm <id>"))
		}
		cmdRm(db, args[1])
	case "pop":
		cmdPop(db)
	default:
		// Everything else is treated as a note body
		cmdAdd(db, strings.Join(args, " "))
	}
}

func cmdAdd(db *DB, body string) {
	id, err := db.Add(body)
	if err != nil {
		fatal(err)
	}
	fmt.Printf("noted (#%d)\n", id)
}

func cmdList(db *DB) {
	notes, err := db.List()
	if err != nil {
		fatal(err)
	}
	DisplayList(notes)
}

func cmdPeek(db *DB, idStr string) {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		fatal(fmt.Errorf("invalid note id: %s", idStr))
	}
	note, err := db.Get(id)
	if err != nil {
		fatal(err)
	}
	DisplayNote(note)
}

func cmdRm(db *DB, idStr string) {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		fatal(fmt.Errorf("invalid note id: %s", idStr))
	}
	if err := db.Delete(id); err != nil {
		fatal(err)
	}
	fmt.Printf("removed #%d\n", id)
}

func cmdPop(db *DB) {
	note, err := db.Latest()
	if err != nil {
		fatal(err)
	}
	DisplayNote(note)
	fmt.Println()
	if err := db.Delete(note.ID); err != nil {
		fatal(err)
	}
	fmt.Println(dim("(removed)"))
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "jot: %v\n", err)
	os.Exit(1)
}
