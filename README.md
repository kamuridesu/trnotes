# trnotes

CLI that uses your $EDITOR to write notes and save it on Trilium

## Install
1. Clone the repository: `git clone https://github.com/kamuridesu/trnotes.git`
2. Run `go build -ldflags='-s -w -extldflags "-static"' -o trnotes`

## Running
Just run ./trnotes and follow the instruction, the config file is saved to ~/.config/trnotes on linux and %APPDATA%\trnotes on Windows.

If no $EDITOR variable is set, it'll default to nano(1) on Linux and notepad.exe on Windows

## Notes
The notes are saved to your Journal using the date as base to get today's note. The notes are saved with the title "Note" if no argument is passed, otherwise the argument will be the name of the note.
