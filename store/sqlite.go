package store

import (
	"context"
	"database/sql"
	// "log"

	"lidsol.org/papeador/security"
)

type SQLiteStore struct {
	DB *sql.DB
}

func NewSQLiteStore(db *sql.DB) Store {
	return &SQLiteStore{DB: db}
}

func (s *SQLiteStore) CreateUser(ctx context.Context, u *User) error {
	// Check for valid username
	if err:=security.IsValidUsername(u.Username); err != nil {
		return err
	}

	// Check for a valid email
	if email, err := security.ValidateEmail(u.Email); err == nil {
		u.Email = email // to_lower and trimmed
	} else {
		return err
	}

	// Check for a secure password
	if err:=security.IsValidPassword(u.Password); err != nil {
		return err
	}

	// Verify duplicated Username/email
	var duplicateUsername string
	err := s.DB.QueryRowContext(ctx, "SELECT username FROM user WHERE username=? OR email=?", u.Username, u.Email).Scan(&duplicateUsername)
	if err == nil {
		return ErrAlreadyExists
	} else if err != sql.ErrNoRows {
		return err
	}


	// Password hashing
	passhash, err := security.HashPassword(u.Password, security.Argon2Params); 
	if err != nil {
		return err
	}

	// Inserting user
	res, err := s.DB.ExecContext(ctx, "INSERT INTO user (username,passhash,email) VALUES (?, ?, ?)", u.Username, passhash, u.Email)

	if err != nil {
		return err
	}

	token, err := security.GenerateJWT(u.Username)
	if err != nil {
		return err
	}

	if id, ierr := res.LastInsertId(); ierr == nil {
		u.UserID = id
		u.JWT = token
	}
	return nil
}

func (s *SQLiteStore) getUserByID(ctx context.Context, id int) (string, error) {
	username := ""
	err := s.DB.QueryRowContext(ctx, "SELECT username,  FROM user WHERE id=?", id).Scan(&username)
	if err != nil {
		return "", err
	}

	return username, nil
}

func (s *SQLiteStore) CreateContest(ctx context.Context, c *Contest) error {
	var name string
	err := s.DB.QueryRowContext(ctx, "SELECT contest_name FROM contest WHERE contest_name=?", c.ContestName).Scan(&name)
	if err == nil {
		return ErrAlreadyExists
	} else if err != sql.ErrNoRows {
		return err
	}

	res, err := s.DB.ExecContext(ctx, "INSERT INTO contest (contest_name) VALUES (?p)", c.ContestName)
	if err != nil {
		return err
	}
	if id, ierr := res.LastInsertId(); ierr == nil {
		c.ContestID = id
	}
	return nil
}

func (s *SQLiteStore) CreateProblem(ctx context.Context, p *Problem) error {
	if p.ContestID != nil {
		var cid int64
		err := s.DB.QueryRowContext(ctx, "SELECT contest_id FROM contest WHERE contest_id=?", *p.ContestID).Scan(&cid)
		if err == sql.ErrNoRows {
			return ErrNotFound
		} else if err != nil {
			return err
		}
	}

	res, err := s.DB.ExecContext(ctx, "INSERT INTO problem (contest_id, creator_id, problem_name, description) VALUES (?, ?, ?, ?)", p.ContestID, p.CreatorID, p.ProblemName, p.Description)
	if err != nil {
		return err
	}
	if id, ierr := res.LastInsertId(); ierr == nil {
		p.ProblemID = id
	}
	return nil
}
