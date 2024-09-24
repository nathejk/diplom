package data

import (
	"database/sql"
	"errors"
	"time"

	"github.com/nathejk/shared-go/types"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type Models struct {
	Teams interface {
		GetStartedTeamIDs(Filters) ([]types.TeamID, Metadata, error)
		GetDiscontinuedTeamIDs(Filters) ([]types.TeamID, Metadata, error)
		GetPatruljer(Filters) ([]*Patrulje, Metadata, error)
		GetPatrulje(types.TeamID) (*Patrulje, error)
		GetPatruljeByYearAndNumber(types.YearSlug, string) (*Patrulje, error)
		GetKlan(types.TeamID) (*Klan, error)
		GetContact(types.TeamID) (*Contact, error)
		RequestedSeniorCount() int
	}
	Members interface {
		GetSpejdere(Filters) ([]*Spejder, Metadata, error)
		GetSeniore(Filters) ([]*Senior, Metadata, error)
		GetInactive(Filters) ([]*SpejderStatus, Metadata, error)
	}
	Permissions interface {
		AddForUser(int64, ...string) error
		GetAllForUser(int64) (Permissions, error)
	}
	Personnel interface {
		//Insert(*Personnel) error
		GetByPhone(types.PhoneNumber) (*Personnel, error)
		GetByID(types.UserID) (*Personnel, error)
		//Update(*Personnel) error
	}
	Tokens interface {
		New(userID int64, ttl time.Duration, scope string) (*Token, error)
		Insert(token *Token) error
		DeleteAllForUser(scope string, userID int64) error
	}
	Users interface {
		Insert(*User) error
		GetByEmail(string) (*User, error)
		Update(*User) error
		GetForToken(string, string) (*User, error)
	}
	Signup interface {
		GetByID(types.TeamID) (*Signup, error)
		ConfirmBySecret(string) (types.TeamID, error)
	}
	Checkpoint interface {
		GetLastCheckpoint(types.YearSlug) (*Checkpoint, error)
	}
	Scan interface {
		GetByTeamIDAndCheckpoint(types.TeamID, string) (*Scan, error)
	}
}

func NewModels(db *sql.DB) Models {
	return Models{
		Teams:       TeamModel{DB: db},
		Members:     MemberModel{DB: db},
		Permissions: PermissionModel{DB: db},
		Personnel:   PersonnelModel{DB: db},
		Tokens:      TokenModel{DB: db},
		Users:       UserModel{DB: db},
		Signup:      SignupModel{DB: db},
		Checkpoint:  CheckpointModel{DB: db},
		Scan:        ScanModel{DB: db},
	}
}
