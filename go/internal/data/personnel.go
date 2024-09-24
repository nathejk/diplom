package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/nathejk/shared-go/types"
	"nathejk.dk/internal/validator"
)

var (
	ErrDuplicatePhone = errors.New("duplicate phone")
)

type Personnel struct {
	UserID    types.UserID       `json:"id"`
	CreatedAt string             `json:"created_at"`
	Name      string             `json:"name"`
	Email     types.EmailAddress `json:"email"`
	//	Password  password  `json:"-"`
	Phone      types.PhoneNumber `json:"phone"`
	Korps      types.CorpsSlug   `json:"corps"`
	MedlemNr   string            `json:"medlemNr"`
	Version    int               `json:"-"`
	Department string            `json:"department"`
	Pincode    string            `json:"pincode"`
	Diet       string            `json:"diet"`
}

/*
	func (u *User) ID() int64 {
		return u.UserID
	}

	func (u *User) IsAnonymous() bool {
		return u == AnonymousUser
	}

	func (u *User) IsActivated() bool {
		return u.Activated
	}

	type password struct {
		plaintext *string
		hash      []byte
	}

// The Set() method calculates the bcrypt hash of a plaintext password, and stores both
// the hash and the plaintext versions in the struct.

	func (p *password) Set(plaintextPassword string) error {
		hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
		if err != nil {
			return err
		}
		p.plaintext = &plaintextPassword
		p.hash = hash
		return nil
	}

// The Matches() method checks whether the provided plaintext password matches the
// hashed password stored in the struct.

	func (p *password) Matches(plaintextPassword string) (bool, error) {
		err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
		if err != nil {
			switch {
			case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
				return false, nil
			default:
				return false, err
			}
		}
		return true, nil
	}

	func ValidatePasswordPlaintext(v validator.Validator, password string) {
		v.Check(password != "", "password", "must be provided")
		v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
		v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
	}
*/
func (user *Personnel) Validate(v validator.Validator) {
	v.Check(user.Name != "", "name", "must be provided")
	v.Check(len(user.Name) <= 500, "name", "must not be more than 500 bytes long")
	//v.Check(user.Email != "", "email", "must be provided")
	//v.CheckEmail(string(user.Email), "email", "must be a valid email address")
	/*
	   	if user.Password.plaintext != nil {
	   		ValidatePasswordPlaintext(v, *user.Password.plaintext)
	   	}

	   // If the password hash is ever nil, this will be due to a logic error in the codebase

	   	if user.Password.hash == nil {
	   		panic("missing password hash for user")
	   	}
	*/
}

// Create a PersonnelModel struct which wraps the connection pool.
type PersonnelModel struct {
	DB *sql.DB
}

func (m PersonnelModel) Insert(user *Personnel) error {
	query := `
INSERT INTO users (name, email, password_hash, activated) VALUES ($1, $2, $3, $4)
RETURNING id, created_at, version`

	args := []any{user.UserID, user.Phone, user.Pincode}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.UserID, &user.CreatedAt, &user.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		default:
			return err
		}
	}
	return nil
}

func (m PersonnelModel) GetByPhone(phone types.PhoneNumber) (*Personnel, error) {
	query := `
		SELECT userId, createdAt, name, email, phone, corps, medlemNr, department, pincode, diet  FROM personnel
		WHERE phone = ?`
	var user Personnel
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, phone.Normalize()).Scan(
		&user.UserID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Phone,
		&user.Korps,
		&user.MedlemNr,
		&user.Department,
		&user.Pincode,
		&user.Diet,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &user, nil
}

func (m PersonnelModel) GetByID(ID types.UserID) (*Personnel, error) {
	query := `
		SELECT userId, createdAt, name, email, phone, corps, medlemNr, department, pincode, diet  FROM personnel
		WHERE userId = ?`
	var user Personnel
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, ID).Scan(
		&user.UserID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Phone,
		&user.Korps,
		&user.MedlemNr,
		&user.Department,
		&user.Pincode,
		&user.Diet,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &user, nil
}

func (m PersonnelModel) Update(user *Personnel) error {
	query := ` UPDATE users
		SET name = $1, email = $2, password_hash = $3, activated = $4, version = version + 1 WHERE id = $5 AND version = $6
		RETURNING version`

	args := []any{
		user.Name,
		user.Email,
		user.Department,
		user.UserID,
		user.Version,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

/*
func (m PersonnelModel) GetForToken(tokenScope, tokenPlaintext string) (*User, error) {
	// Calculate the SHA-256 hash of the plaintext token provided by the client.
	// Remember that this returns a byte *array* with length 32, not a slice.
	tokenHash := sha256.Sum256([]byte(tokenPlaintext))

	query := `
		SELECT users.id, users.created_at, users.name, users.email, users.password_hash, users.activated, users.version FROM users
		INNER JOIN tokens
		ON users.id = tokens.user_id
		WHERE tokens.hash = $1
		AND tokens.scope = $2
		AND tokens.expiry > $3`

	args := []any{tokenHash[:], tokenScope, time.Now()}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var user User
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&user.UserID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}*/
