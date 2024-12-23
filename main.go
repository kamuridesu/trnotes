package main

import (
	"fmt"

	c "github.com/kamuridesu/trnotes/internal/config"
	e "github.com/kamuridesu/trnotes/internal/editor"
	t "github.com/kamuridesu/trnotes/internal/trilium"
)

func ppanic(err error) {
	if err != nil {
		panic(err)
	}
}

func setNewConfig() (*c.Config, error) {
	fmt.Println("Let' do an initial config")
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

func main() {
	config, err := setup()
	ppanic(err)
	trilium, err := t.FromConfig(config)
	ppanic(err)
	tempFile, err := e.OpenEditor()
	ppanic(err)
	err = trilium.SaveNote(tempFile)
	ppanic(err)
}
