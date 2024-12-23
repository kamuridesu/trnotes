package main

import (
	c "github.com/kamuridesu/trnotes/internal/config"
	e "github.com/kamuridesu/trnotes/internal/editor"
	t "github.com/kamuridesu/trnotes/internal/trilium"
)

func ppanic(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	config, err := c.New("example.yaml")
	ppanic(err)
	trilium, err := t.New(config)
	ppanic(err)
	tempFile, err := e.OpenEditor()
	ppanic(err)
	err = trilium.SaveNote(tempFile)
	ppanic(err)

}
