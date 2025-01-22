package main

import (
	"fmt"
	"time"

	"github.com/kamuridesu/trnotes/cmd"
	e "github.com/kamuridesu/trnotes/internal/editor"
	t "github.com/kamuridesu/trnotes/internal/trilium"
)

func main() {
	args := cmd.Argparse()
	if *args.Debug {
		fmt.Println("Starting in DEBUG mode")
		cmd.Debug()
		return
	}
	config := cmd.Check(cmd.Setup(args))
	trilium := cmd.Check(t.FromConfig(config))
	var notes *[]*t.Note
	if *args.List {
		if *args.DatePrefix != "" {
			tm := cmd.Check(time.Parse("2006-01-02", *args.DatePrefix))
			notes = cmd.Check(trilium.GetAllDateNotes(tm))
		} else {
			notes = cmd.Check(trilium.GetAllDateNotes(time.Now().Local()))
		}
		for i, note := range *notes {
			fmt.Printf("%d. %s\n", i+1, note.Title)
		}
		return
	}

	if *args.Edit {
		if *args.DatePrefix != "" {
			notes = cmd.Check(trilium.SearchInDate(*args.DatePrefix, *args.Name))
		} else {
			notes = cmd.Check(trilium.SearchInTodayNotes(*args.Name))
		}
		note := (*notes)[0]
		if len(*notes) > 1 {
			note = cmd.PromptMultiNotes(notes)
		}
		content := cmd.Check(trilium.FetchNoteContent(note.Id))
		tempFile := cmd.Check(e.EditFile(*content))
		if *tempFile == *content {
			return
		}
		cmd.Check[any](nil, trilium.UpdateNote(note.Id, *tempFile))
		return
	}
	tempFile := cmd.Check(e.NewFile())
	if *tempFile == "" {
		return
	}
	cmd.Check[any](nil, trilium.SaveNote(tempFile, args.Name))
}
