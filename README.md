# trnotes

CLI that uses your `$EDITOR` to write notes and save it on Trilium

## Install
1. Clone the repository: `git clone https://github.com/kamuridesu/trnotes.git`
2. Run `go build -ldflags='-s -w -extldflags "-static"' -o trnotes`

## Running
Just run `./trnotes` and follow the instruction, the config file is saved to `~/.config/trnotes` on linux and `%APPDATA%\trnotes` on Windows but you can use `-config` or set the environment variable `TRNOTES_CONFIG` to set a custom config location .

If no $EDITOR variable is set, it'll default to nano(1) on Linux and notepad.exe on Windows

## Notes
The notes are saved to your Journal using the date as base to get today's note. The notes are saved with the title "Note" if no argument is passed, otherwise the argument will be the name of the note.

You can use `-edit` to edit an existing note, like `./trnotes -edit Note`. It should be used with a name and by default searches in today notes. If there's more than one note with the same name, you will be prompted to choose which note to edit.

If you want to edit a note on a certain date, you can pass the date with the title, like: `./trnotes -edit 2024-12-31/Note`.

If you want to list all notes, you can use the `-list` option. Using a date like `2024-12-31` will list all the notes from that date.

## Examples
1. `trnote` will create a new note named Note
2. `trnote My Note` will create a new note named My Note
3. `trnote -edit My Note` will search and edit a note called My Note in Today
4. `trnote -edit 2024-12-31/My Note` will search and edit a note called My Note in the date 2024-12-31
5. `trnote -list` will list all notes in Today
6. `trnote -list 2024-12-31` will list all notes in the date 2024-12-31
