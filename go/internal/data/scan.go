package data

import (
	"database/sql"
	"errors"

	"github.com/nathejk/shared-go/types"
)

type Scan struct {
	ID           types.ScanID
	Year         types.YearSlug
	QrID         types.QrID
	TeamID       string
	TeamNumber   string
	ScannerID    string
	ScannerPhone string
	Uts          int64
	Coordinate   types.Coordinate
}

type ScanModel struct {
	DB *sql.DB
}

func (m ScanModel) GetByTeamIDAndCheckpoint(teamID types.TeamID, checkgroupID string) (*Scan, error) {
	query := `SELECT id, year, qrId, teamId, teamNumber, scannerId, scannerPhone, uts, latitude, longitude
		FROM scan s right JOIN controlgroup_user u ON s.scannerId = u.userId
		WHERE teamId = ? AND controlGroupId = ? AND uts >= startUts AND uts <= enduts;`
	var p Scan
	err := m.DB.QueryRow(query, teamID, checkgroupID).Scan(
		&p.ID,
		&p.Year,
		&p.QrID,
		&p.TeamID,
		&p.TeamNumber,
		&p.ScannerID,
		&p.ScannerPhone,
		&p.Uts,
		&p.Coordinate.Latitude,
		&p.Coordinate.Longitude,
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
