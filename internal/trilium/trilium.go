package trilium

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"time"

	c "github.com/kamuridesu/trnotes/internal/config"
)

type Trilium struct {
	Config *c.Config
}

type Note struct {
	Id          string `json:"noteId"`
	Title       string `json:"title"`
	Type        string `json:"type"`
	Mime        string `json:"mime"`
	IsProtected bool   `json:"isProtected"`
	BlobId      string `json:"blobId"`
}

func New(config *c.Config) (*Trilium, error) {
	_, err := http.Head(config.Url)
	if err != nil {
		return nil, err
	}
	return &Trilium{Config: config}, nil
}

func (t *Trilium) authorize(req *http.Request) {
	token := base64.RawStdEncoding.EncodeToString([]byte(t.Config.Username + ":" + t.Config.Token))
	req.Header.Add("Authorization", "Bearer "+token)
}

func (t *Trilium) GetCurrentDayNote() (*Note, error) {
	url := fmt.Sprintf(`%s/calendar/days/`, t.Config.Url, time.Date())
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	t.authorize(req)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}


}

func (t *Trilium) SaveNote(content *string) error {
	return nil
}
