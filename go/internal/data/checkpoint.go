package data

import (
	"database/sql"
	"errors"

	"github.com/nathejk/shared-go/types"
)

type Checkpoint struct {
	Year         types.YearSlug
	GroupID      string
	GroupName    string
	Index        string
	Name         string
	OpenFromUts  int64
	OpenUntilUts int64
}

type CheckpointModel struct {
	DB *sql.DB
}

func (m CheckpointModel) GetLastCheckpoint(year types.YearSlug) (*Checkpoint, error) {
	query := `SELECT year, controlGroupId, controlGroupName, controlIndex, controlName, openFromUts, openUntilUts
		FROM controlpoint
		WHERE year = ?
		ORDER BY openFromUts DESC LIMIT 1`
	var p Checkpoint
	err := m.DB.QueryRow(query, year).Scan(
		&p.Year,
		&p.GroupID,
		&p.GroupName,
		&p.Index,
		&p.Name,
		&p.OpenFromUts,
		&p.OpenUntilUts,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &p, nil
}
