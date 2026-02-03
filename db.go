package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Note struct {
	ID        int64
	Body      string
	CreatedAt string
}

// Turso HTTP API request/response types

type pipelineRequest struct {
	Requests []stmtRequest `json:"requests"`
}

type stmtRequest struct {
	Type string    `json:"type"`
	Stmt *stmtBody `json:"stmt,omitempty"`
}

type stmtBody struct {
	SQL  string     `json:"sql"`
	Args []stmtArg  `json:"args,omitempty"`
}

type stmtArg struct {
	Type  string `json:"type"`
	Value any    `json:"value"`
}

type pipelineResponse struct {
	Results []stmtResult `json:"results"`
}

type stmtResult struct {
	Type     string        `json:"type"`
	Response *stmtResponse `json:"response,omitempty"`
	Error    *stmtError    `json:"error,omitempty"`
}

type stmtResponse struct {
	Type string      `json:"type"`
	Result resultBody `json:"result"`
}

type resultBody struct {
	Cols         []colInfo   `json:"cols"`
	Rows         [][]colVal  `json:"rows"`
	AffectedRows int64       `json:"affected_row_count"`
	LastInsertID string      `json:"last_insert_rowid,omitempty"`
}

type colInfo struct {
	Name string `json:"name"`
}

type colVal struct {
	Type  string `json:"type"`
	Value any    `json:"value"`
}

type stmtError struct {
	Message string `json:"message"`
}

type DB struct {
	url    string
	token  string
	client *http.Client
}

func NewDB(url, token string) *DB {
	return &DB{
		url:   url,
		token: token,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (d *DB) execute(stmts []stmtBody) (*pipelineResponse, error) {
	reqs := make([]stmtRequest, len(stmts))
	for i := range stmts {
		reqs[i] = stmtRequest{Type: "execute", Stmt: &stmts[i]}
	}
	reqs = append(reqs, stmtRequest{Type: "close"})

	body, err := json.Marshal(pipelineRequest{Requests: reqs})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", d.url+"/v2/pipeline", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+d.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed (are you online?): %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("turso API error (%d): %s", resp.StatusCode, string(respBody))
	}

	var result pipelineResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	for _, r := range result.Results {
		if r.Type == "error" && r.Error != nil {
			return nil, fmt.Errorf("SQL error: %s", r.Error.Message)
		}
	}

	return &result, nil
}

func (d *DB) Init() error {
	_, err := d.execute([]stmtBody{{
		SQL: `CREATE TABLE IF NOT EXISTS notes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			body TEXT NOT NULL,
			created_at TEXT NOT NULL DEFAULT (datetime('now'))
		)`,
	}})
	return err
}

func (d *DB) Add(body string) (int64, error) {
	resp, err := d.execute([]stmtBody{{
		SQL:  "INSERT INTO notes (body) VALUES (?)",
		Args: []stmtArg{{Type: "text", Value: body}},
	}})
	if err != nil {
		return 0, err
	}

	if len(resp.Results) > 0 && resp.Results[0].Response != nil {
		id := resp.Results[0].Response.Result.LastInsertID
		var n int64
		fmt.Sscanf(id, "%d", &n)
		return n, nil
	}
	return 0, nil
}

func (d *DB) List() ([]Note, error) {
	resp, err := d.execute([]stmtBody{{
		SQL: "SELECT id, body, created_at FROM notes ORDER BY id DESC",
	}})
	if err != nil {
		return nil, err
	}
	return parseNotes(resp, 0), nil
}

func (d *DB) Get(id int64) (*Note, error) {
	resp, err := d.execute([]stmtBody{{
		SQL:  "SELECT id, body, created_at FROM notes WHERE id = ?",
		Args: []stmtArg{{Type: "integer", Value: fmt.Sprintf("%d", id)}},
	}})
	if err != nil {
		return nil, err
	}

	notes := parseNotes(resp, 0)
	if len(notes) == 0 {
		return nil, fmt.Errorf("note #%d not found", id)
	}
	return &notes[0], nil
}

func (d *DB) Latest() (*Note, error) {
	resp, err := d.execute([]stmtBody{{
		SQL: "SELECT id, body, created_at FROM notes ORDER BY id DESC LIMIT 1",
	}})
	if err != nil {
		return nil, err
	}

	notes := parseNotes(resp, 0)
	if len(notes) == 0 {
		return nil, fmt.Errorf("no notes yet")
	}
	return &notes[0], nil
}

func (d *DB) Delete(id int64) error {
	resp, err := d.execute([]stmtBody{{
		SQL:  "DELETE FROM notes WHERE id = ?",
		Args: []stmtArg{{Type: "integer", Value: fmt.Sprintf("%d", id)}},
	}})
	if err != nil {
		return err
	}

	if len(resp.Results) > 0 && resp.Results[0].Response != nil {
		if resp.Results[0].Response.Result.AffectedRows == 0 {
			return fmt.Errorf("note #%d not found", id)
		}
	}
	return nil
}

func parseNotes(resp *pipelineResponse, idx int) []Note {
	if resp == nil || idx >= len(resp.Results) {
		return nil
	}
	r := resp.Results[idx]
	if r.Response == nil {
		return nil
	}

	var notes []Note
	for _, row := range r.Response.Result.Rows {
		if len(row) < 3 {
			continue
		}
		var id int64
		fmt.Sscanf(fmt.Sprintf("%v", row[0].Value), "%d", &id)

		notes = append(notes, Note{
			ID:        id,
			Body:      fmt.Sprintf("%v", row[1].Value),
			CreatedAt: fmt.Sprintf("%v", row[2].Value),
		})
	}
	return notes
}
