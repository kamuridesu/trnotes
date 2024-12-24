package trilium

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/kamuridesu/trnotes/internal/config"
	req "github.com/kamuridesu/trnotes/internal/request"
)

type Trilium struct {
	Url   string
	Token string
}

type Note struct {
	Id          string `json:"noteId"`
	Title       string `json:"title"`
	Type        string `json:"type"`
	Mime        string `json:"mime"`
	IsProtected bool   `json:"isProtected"`
	BlobId      string `json:"blobId"`
}

func FromConfig(c *config.Config) (*Trilium, error) {
	tr, err := New(c.Url)
	if err != nil {
		return nil, err
	}
	tr.SetToken(c.Token)
	return tr, nil
}

func New(Url string) (*Trilium, error) {
	if !strings.HasPrefix(Url, "http") {
		return nil, fmt.Errorf("missing schema for URL")
	}
	_, err := http.Head(Url)
	if err != nil {
		return nil, err
	}
	return &Trilium{Url: Url}, nil
}

func (t *Trilium) SetToken(token string) {
	t.Token = token
}

func (t *Trilium) Authorize(password string) (string, error) {
	url := fmt.Sprintf("%s/etapi/auth/login", t.Url)
	name, err := config.GetComputerName()
	if err != nil {
		return "", err
	}
	reqBody := fmt.Sprintf(`{"password": "%s", "tokenName": "TRNotes on %s"}`, password, name)
	body, err := req.New("POST", url, 201).SetBody(reqBody).SetHeaders(map[string]string{
		"Content-Type":  "application/json",
		"Authorization": t.Token}).Send()
	if err != nil {
		return "", err
	}

	bodyContent := make(map[string]string)
	if err = json.Unmarshal([]byte(body), &bodyContent); err != nil {
		return "", err
	}

	token, ok := bodyContent["authToken"]
	if !ok {
		return "", fmt.Errorf("fail to fetch authToken from api, response is %v", bodyContent)
	}

	t.Token = token

	return token, nil

}

func (t *Trilium) GetCurrentDayNote() (*Note, error) {
	date := time.Now().Local().Format(time.DateOnly)
	url := fmt.Sprintf(`%s/etapi/calendar/days/%s`, t.Url, date)

	body, err := req.New("GET", url, 200).SetHeaders(map[string]string{
		"Authorization": t.Token}).Send()
	if err != nil {
		return nil, err
	}

	note := &Note{}
	if err = json.Unmarshal([]byte(body), note); err != nil {
		return nil, err
	}

	return note, nil

}

func (t *Trilium) SaveNote(content *string) error {
	parent, err := t.GetCurrentDayNote()
	if err != nil {
		return err
	}
	url := fmt.Sprintf(`%s/etapi/create-note`, t.Url)

	postBodyJson, err := json.Marshal(map[string]string{"parentNoteId": parent.Id,
		"title":   "Note",
		"type":    "text",
		"content": *content})
	if err != nil {
		return err
	}

	_, err = req.New("POST", url, 201).SetHeaders(map[string]string{
		"Content-Type":  "application/json",
		"Authorization": t.Token}).SetBody(string(postBodyJson)).Send()
	if err != nil {
		return err
	}
	return nil
}
