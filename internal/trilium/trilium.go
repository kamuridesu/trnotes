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
	Id           string   `json:"noteId"`
	Title        string   `json:"title"`
	Type         string   `json:"type"`
	Mime         string   `json:"mime"`
	IsProtected  bool     `json:"isProtected"`
	BlobId       string   `json:"blobId"`
	ChildNoteIds []string `json:"childNoteIds"`
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
		return "", fmt.Errorf("failed to fetch authToken from api, response is %v", bodyContent)
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

func (t *Trilium) SaveNote(content, name *string) error {
	parent, err := t.GetCurrentDayNote()
	if err != nil {
		return err
	}
	url := fmt.Sprintf(`%s/etapi/create-note`, t.Url)
	title := "Note"
	if name != nil && *name != "" {
		title = *name
	}

	postBodyJson, err := json.Marshal(map[string]string{
		"parentNoteId": parent.Id,
		"title":        title,
		"type":         "code",
		"mime":         "text/x-markdown",
		"content":      *content})
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

func (t *Trilium) FetchNote(id string) (*Note, error) {
	url := fmt.Sprintf("%s/etapi/notes/%s", t.Url, id)
	body, err := req.New("GET", url, 200).SetHeaders(map[string]string{
		"Authorization": t.Token,
	}).Send()
	if err != nil {
		return nil, err
	}
	note := &Note{}
	if err := json.Unmarshal([]byte(body), note); err != nil {
		return nil, err
	}
	return note, nil
}

func (t *Trilium) FetchChildrenNotes(ids []string) (*[]*Note, error) {
	notes := make([]*Note, 0)

	for _, id := range ids {
		note, err := t.FetchNote(id)
		if err != nil {
			return nil, err
		}
		notes = append(notes, note)
	}
	return &notes, nil
}

func (t *Trilium) FetchNoteContent(id string) (*string, error) {
	body, err := req.New("GET", fmt.Sprintf("%s/etapi/notes/%s/content", t.Url, id), 200).SetHeaders(map[string]string{
		"Authorization": t.Token,
	}).Send()
	if err != nil {
		return nil, err
	}
	return &body, nil
}

func (t *Trilium) UpdateNote(id, body string) error {
	url := fmt.Sprintf("%s/etapi/notes/%s/content", t.Url, id)
	_, err := req.New("PUT", url, 204).SetHeaders(map[string]string{
		"Authorization": t.Token,
		"Content-Type":  "text/plain",
	}).SetBody(body).Send()
	return err
}

func (t *Trilium) GetAllTodayNotes() (*[]*Note, error) {
	note, err := t.GetCurrentDayNote()
	if err != nil {
		return nil, err
	}
	notes, err := t.FetchChildrenNotes(note.ChildNoteIds)
	if err != nil {
		return nil, err
	}
	return notes, nil
}

func (t *Trilium) SearchInTodayNotes(title string) (*[]*Note, error) {
	notes, err := t.GetAllTodayNotes()
	if err != nil {
		return nil, err
	}
	matches := make([]*Note, 0)
	for _, note := range *notes {
		if note.Title == title {
			matches = append(matches, note)
		}
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("note '%s' not found", title)
	}
	return &matches, nil
}
