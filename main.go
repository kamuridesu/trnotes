package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	c "github.com/kamuridesu/trnotes/internal/config"
	e "github.com/kamuridesu/trnotes/internal/editor"
	t "github.com/kamuridesu/trnotes/internal/trilium"
)

type Args struct {
	Name  *string
	Debug *bool
	Edit  *bool
}

func check[T any](x T, err error) T {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return x
}

func argparse() *Args {
	args := Args{}
	args.Debug = flag.Bool("debug", false, "Use Debug function")
	args.Edit = flag.Bool("edit", false, "Edit existing note")
	flag.Parse()
	name := strings.Join(flag.Args(), " ")
	if *args.Edit && name == "" {
		check[any](nil, fmt.Errorf("edit should be used with named notes only"))
	}
	args.Name = &name
	return &args
}

func setNewConfig() (*c.Config, error) {
	fmt.Println("Let's do an initial config")
	url := ""
	password := ""
	fmt.Print("Please, insert the URL for your Trilium Server: ")
	fmt.Scanln(&url)
	tr, err := t.New(url)
	if err != nil {
		return nil, err
	}
	fmt.Print("Now enter your password to generate the access token: ")
	fmt.Scanln(&password)
	token, err := tr.Authorize(password)
	if err != nil {
		return nil, err
	}
	return c.New(url, token)
}

func setup() (*c.Config, error) {
	conf, err := c.GetExistingConfig()
	if err != nil {
		return nil, err
	}
	if conf == nil {
		conf, err = setNewConfig()
		if err != nil {
			return nil, err
		}
		err = conf.Save()
		if err != nil {
			return nil, err
		}
	}
	return conf, nil
}

func debug() {
	config := check(setup())
	trilium := check(t.FromConfig(config))
	note := check(trilium.GetCurrentDayNote())
	fmt.Println(note.Title)
	if len(note.ChildNoteIds) > 0 {
		notes := check(trilium.FetchChildrenNotes(note.ChildNoteIds))
		for _, note := range *notes {
			fmt.Println(note.Title)
			if note.Title == "Note" {
				body := check(trilium.FetchNoteContent(note.Id))
				fmt.Println(*body)
				fmt.Println(note.Mime)
				fmt.Println(note.Type)
			}
		}
	}
}

func promptMultiNotes(notes *[]*t.Note) *t.Note {
	fmt.Println("More than one note found!")
	for i, note := range *notes {
		fmt.Printf("%d. %s\n", i+1, note.Title)
	}
	choosenStr := ""
	fmt.Print("Please, select one for the list: ")
	fmt.Scanln(&choosenStr)
	choosen := check(strconv.Atoi(choosenStr))
	if choosen < 0 || choosen > len(*notes) {
		check[any](nil, fmt.Errorf("error: selected number is not in the list"))
	}
	return (*notes)[choosen-1]
}

func main() {
	args := argparse()
	if *args.Debug {
		fmt.Println("Starting in DEBUG mode")
		debug()
		return
	}
	config := check(setup())
	trilium := check(t.FromConfig(config))
	if *args.Edit {
		notes := check(trilium.SearchInTodayNotes(*args.Name))
		note := (*notes)[0]
		if len(*notes) > 1 {
			note = promptMultiNotes(notes)
		}
		content := check(trilium.FetchNoteContent(note.Id))
		tempFile := check(e.EditFile(*content))
		check[any](nil, trilium.UpdateNote(note.Id, *tempFile))
		return
	}
	tempFile := check(e.NewFile())
	if *tempFile == "" {
		return
	}
	check[any](nil, trilium.SaveNote(tempFile, args.Name))
}
