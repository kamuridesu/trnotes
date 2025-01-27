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
	URL   string
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

type authRequest struct {
	Password  string `json:"password"`
	TokenName string `json:"tokenName"`
}

type authResponse struct {
	AuthToken string `json:"authToken"`
}

type createNoteRequest struct {
	ParentNoteID string `json:"parentNoteId"`
	Title        string `json:"title"`
	Type         string `json:"type"`
	Mime         string `json:"mime"`
	Content      string `json:"content"`
}

func FromConfig(c *config.Config) (*Trilium, error) {
	tr, err := New(c.Url)
	if err != nil {
		return nil, fmt.Errorf("failed to create trilium client: %w", err)
	}
	tr.Token = c.Token
	return tr, nil
}

func New(url string) (*Trilium, error) {
	if !strings.HasPrefix(url, "http") {
		return nil, fmt.Errorf("invalid URL format, missing schema: %s", url)
	}

	if _, err := http.Head(url); err != nil {
		return nil, fmt.Errorf("failed to connect to trilium URL: %w", err)
	}

	return &Trilium{URL: url}, nil
}

func (t *Trilium) Authorize(password string) (string, error) {
	name, err := config.GetComputerName()
	if err != nil {
		return "", fmt.Errorf("failed to get computer name: %w", err)
	}

	reqBody := authRequest{
		Password:  password,
		TokenName: fmt.Sprintf("TRNotes on %s", name),
	}

	var resp authResponse
	if err := t.doJSONRequest(http.MethodPost, "auth/login", http.StatusCreated, reqBody, &resp); err != nil {
		return "", fmt.Errorf("authorization failed: %w", err)
	}

	t.Token = resp.AuthToken
	return resp.AuthToken, nil
}

func (t *Trilium) GetSelectedDayNote(date time.Time) (*Note, error) {
	dateStr := date.Format(time.DateOnly)
	endpoint := fmt.Sprintf("calendar/days/%s", dateStr)
	var note Note
	if err := t.doJSONRequest(http.MethodGet, endpoint, http.StatusOK, nil, &note); err != nil {
		return nil, fmt.Errorf("failed to get selected day note: %w", err)
	}
	return &note, nil
}

func (t *Trilium) GetCurrentDayNote() (*Note, error) {
	return t.GetSelectedDayNote(time.Now().Local())
}

func (t *Trilium) SaveNote(content, name string) error {
	parent, err := t.GetCurrentDayNote()
	if err != nil {
		return fmt.Errorf("failed to get current day note: %w", err)
	}

	title := "Note"
	if name != "" {
		title = name
	}

	reqBody := createNoteRequest{
		ParentNoteID: parent.Id,
		Title:        title,
		Type:         "code",
		Mime:         "text/x-markdown",
		Content:      content,
	}

	return t.doJSONRequest(http.MethodPost, "create-note", http.StatusCreated, reqBody, nil)
}

func (t *Trilium) FetchNote(id string) (*Note, error) {
	endpoint := fmt.Sprintf("notes/%s", id)
	var note Note
	if err := t.doJSONRequest(http.MethodGet, endpoint, http.StatusOK, nil, &note); err != nil {
		return nil, fmt.Errorf("failed to fetch note: %w", err)
	}
	return &note, nil
}

func (t *Trilium) FetchChildrenNotes(ids []string) ([]*Note, error) {
	notes := make([]*Note, 0, len(ids))

	for _, id := range ids {
		note, err := t.FetchNote(id)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch child note %s: %w", id, err)
		}
		notes = append(notes, note)
	}
	return notes, nil
}

func (t *Trilium) FetchNoteContent(id string) (string, error) {
	endpoint := fmt.Sprintf("notes/%s/content", id)
	content, err := t.doGetRequest(endpoint, http.StatusOK)
	if err != nil {
		return "", fmt.Errorf("failed to fetch note content: %w", err)
	}
	return content, nil
}

func (t *Trilium) UpdateNote(id, content string) error {
	endpoint := fmt.Sprintf("notes/%s/content", id)
	return t.doTextRequest(http.MethodPut, endpoint, http.StatusNoContent, content)
}

func (t *Trilium) GetAllDateNotes(date time.Time) ([]*Note, error) {
	note, err := t.GetSelectedDayNote(date)
	if err != nil {
		return nil, fmt.Errorf("failed to get date note: %w", err)
	}
	return t.FetchChildrenNotes(note.ChildNoteIds)
}

func (t *Trilium) SearchInTodayNotes(title string) ([]*Note, error) {
	notes, err := t.GetAllDateNotes(time.Now().Local())
	if err != nil {
		return nil, err
	}
	return t.findNote(title, notes)
}

func (t *Trilium) SearchInDate(date, title string) ([]*Note, error) {
	d, err := time.Parse(time.DateOnly, date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}
	notes, err := t.GetAllDateNotes(d)
	if err != nil {
		return nil, err
	}
	return t.findNote(title, notes)
}

func (t *Trilium) findNote(title string, notes []*Note) ([]*Note, error) {
	var matches []*Note
	for _, note := range notes {
		if note.Title == title {
			matches = append(matches, note)
		}
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("note '%s' not found", title)
	}
	return matches, nil
}

func (t *Trilium) buildURL(endpoint string) string {
	return fmt.Sprintf("%s/etapi/%s", t.URL, endpoint)
}

func (t *Trilium) doJSONRequest(method, endpoint string, expectedStatus int, requestBody interface{}, response interface{}) error {
	req := req.New(method, t.buildURL(endpoint), expectedStatus)
	if t.Token != "" {
		req.SetHeader("Authorization", t.Token)
	}

	if requestBody != nil {
		jsonBody, err := json.Marshal(requestBody)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		req.SetBody(string(jsonBody))
		req.SetHeader("Content-Type", "application/json")
	}

	body, err := req.Send()
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	if response != nil {
		if err := json.Unmarshal([]byte(body), response); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}
	return nil
}

func (t *Trilium) doTextRequest(method, endpoint string, expectedStatus int, body string) error {
	req := req.New(method, t.buildURL(endpoint), expectedStatus)
	req.SetHeader("Authorization", t.Token)
	req.SetHeader("Content-Type", "text/plain")
	req.SetBody(body)

	_, err := req.Send()
	return err
}

func (t *Trilium) doGetRequest(endpoint string, expectedStatus int) (string, error) {
	req := req.New(http.MethodGet, t.buildURL(endpoint), expectedStatus)
	req.SetHeader("Authorization", t.Token)
	return req.Send()
}
