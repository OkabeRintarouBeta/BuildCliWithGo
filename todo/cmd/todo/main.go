package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/okaberintaroubeta/todo"
)

var todoFileName = ".todo.json"

func main() {
	if os.Getenv("TODO_FILENAME") != "" {
		todoFileName = os.Getenv("TODO_FILENAME")
	}

	add := flag.Bool("add", false, "Add task to the ToDo list")
	list := flag.Bool("list", false, "List all tasks")
	complete := flag.Int("complete", 0, "Item to be completed")
	verbose := flag.Bool("v", false, "Show verbose input including date/time")
	delete := flag.Int("del", 0, "Item to be deleted")
	hideCompleted := flag.Bool("hide-completed", false, "Hide completed items from being displayed")
	numTasks := flag.Int("num", 1, "Number of lines to read")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "%s tool. Developed for The Promatic Bookshelf\n", os.Args[0])
		fmt.Fprintf(flag.CommandLine.Output(), "Copyright 2020\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Usage information:")
		flag.PrintDefaults()
	}

	flag.Parse()

	l := &todo.List{}

	if err := l.Get(todoFileName); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	switch {
	case *list:
		if !*verbose && !*hideCompleted {
			fmt.Print(l)
		} else {
			fmt.Print(l.PrintFlexible(*verbose, *hideCompleted))
		}
	case *complete > 0:
		if err := l.Complete(*complete); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := l.Save(todoFileName); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case *delete > 0:
		if err := l.Delete(*delete); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := l.Save(todoFileName); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case *add:
		tasks, err := getTask(os.Stdin, *numTasks, flag.Args()...)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		for _, t := range tasks {
			l.Add(t)
		}
		if err := l.Save(todoFileName); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	default:
		fmt.Fprintln(os.Stderr, "Invalid option")
		os.Exit(1)
	}
}

func getTask(r io.Reader, numLines int, args ...string) ([]string, error) {
	tasks := []string{}
	if len(args) > 0 {
		tasks = append(tasks, strings.Join(args, " "))
		return tasks, nil
	}
	s := bufio.NewScanner(r)

	for s.Scan() && numLines > 0 {
		if err := s.Err(); err != nil {
			return nil, err
		}
		if len(s.Text()) == 0 {
			return nil, fmt.Errorf("task cannot be blank")
		}
		tasks = append(tasks, s.Text())
		numLines--
		if numLines == 0 {
			return tasks, nil
		}
	}

	return tasks, nil
}
