package trilium

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/kamuridesu/trnotes/internal/config"
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

func (t *Trilium) authorize(req *http.Request) {
	req.Header.Add("Authorization", t.Token)
}

func (t *Trilium) Authorize(password string) (string, error) {
	url := fmt.Sprintf("%s/etapi/auth/login", t.Url)
	reqBody := fmt.Sprintf(`{"password": "%s"}`, password)
	req, err := http.NewRequest("POST", url, strings.NewReader(reqBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	if res.StatusCode != 201 {
		return "", fmt.Errorf("error creating token, status is %d and response is %s", res.StatusCode, body)
	}

	bodyContent := make(map[string]string)
	if err = json.Unmarshal(body, &bodyContent); err != nil {
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
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	t.authorize(req)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("error getting note for date %s, status is %d", date, res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	note := &Note{}
	if err = json.Unmarshal(body, note); err != nil {
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

	postBody := make(map[string]string)
	postBody["parentNoteId"] = parent.Id
	postBody["title"] = "Note"
	postBody["type"] = "text"
	postBody["content"] = *content

	postBodyJson, err := json.Marshal(postBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, strings.NewReader(string(postBodyJson)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	t.authorize(req)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	body, err := io.ReadAll(res.Body)
	if res.StatusCode != 201 {
		return fmt.Errorf("error saving note, status is %d, status is %s", res.StatusCode, body)
	}
	return nil

}
