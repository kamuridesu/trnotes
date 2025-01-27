package cmd

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	c "github.com/kamuridesu/trnotes/internal/config"
	t "github.com/kamuridesu/trnotes/internal/trilium"
)

type Args struct {
	Name       *string
	Debug      *bool
	Edit       *bool
	DatePrefix *string
	List       *bool
	ConfigFile *string
}

func Check[T any](x T, err error) T {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return x
}

func Argparse() *Args {
	args := Args{}
	args.Debug = flag.Bool("debug", false, "Use Debug function")
	args.Edit = flag.Bool("edit", false, "Edit existing note")
	args.List = flag.Bool("list", false, "List current date notes")
	args.ConfigFile = flag.String("config", "", "Sets the config file to be used")
	flag.Parse()
	name := strings.Join(flag.Args(), " ")
	if *args.Edit && name == "" {
		Check[any](nil, fmt.Errorf("edit should be used with named notes only"))
	}
	args.Name = &name
	args.DatePrefix = CheckIfNoteDatePrefix(*args.Name)
	if *args.DatePrefix != "" {
		name = strings.Join(strings.Split(name, "/")[1:], "/")
		args.Name = &name
	}

	if *args.ConfigFile == "" {
		config := os.Getenv("TRNOTES_CONFIG")
		args.ConfigFile = &config
	}

	return &args
}

func SetNewConfig(args *Args) (*c.Config, error) {
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
	return c.New(url, token, *args.ConfigFile)
}

func Setup(args *Args) (*c.Config, error) {
	conf, err := c.GetExistingConfig(*args.ConfigFile)
	if err != nil {
		return nil, err
	}
	if conf == nil {
		conf, err = SetNewConfig(args)
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

func Debug() {
	args := &Args{}
	empty := ""
	args.ConfigFile = &empty
	config := Check(Setup(args))
	trilium := Check(t.FromConfig(config))
	note := Check(trilium.GetCurrentDayNote())
	fmt.Println(note.Title)
	if len(note.ChildNoteIds) > 0 {
		notes := Check(trilium.FetchChildrenNotes(note.ChildNoteIds))
		for _, note := range notes {
			fmt.Println(note.Title)
			if note.Title == "Note" {
				body := Check(trilium.FetchNoteContent(note.Id))
				fmt.Println(body)
				fmt.Println(note.Mime)
				fmt.Println(note.Type)
			}
		}
	}
}

func PromptMultiNotes(notes []*t.Note) *t.Note {
	fmt.Println("More than one note found!")
	for i, note := range notes {
		fmt.Printf("%d. %s\n", i+1, note.Title)
	}
	choosenStr := ""
	fmt.Print("Please, select one for the list: ")
	fmt.Scanln(&choosenStr)
	choosen := Check(strconv.Atoi(choosenStr))
	if choosen < 0 || choosen > len(notes) {
		Check[any](nil, fmt.Errorf("error: selected number is not in the list"))
	}
	return (notes)[choosen-1]
}

func CheckIfNoteDatePrefix(noteTitle string) *string {
	r := ""
	result := Check(regexp.Compile(`\d+-\d+-\d+`)).Find([]byte(noteTitle))
	if result != nil {
		r = string(result)
	}
	return &r
}
