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
	if err := security.IsValidUsername(u.Username); err != nil {
		return err
	}

	// Check for a valid email
	if email, err := security.ValidateEmail(u.Email); err == nil {
		u.Email = email // to_lower and trimmed
	} else {
		return err
	}

	// Check for a secure password
	if err := security.IsValidPassword(u.Password); err != nil {
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
	passhash, err := security.HashPassword(u.Password, security.Argon2Params)
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

func (s *SQLiteStore) GetUserByID(ctx context.Context, id int) (string, error) {
	username := ""
	err := s.DB.QueryRowContext(ctx, "SELECT username,  FROM user WHERE id=?", id).Scan(&username)
	if err != nil {
		return "", err
	}

	return username, nil
}

func (s *SQLiteStore) GetUserID(ctx context.Context, username string) (int, error) {
	id := 0
	err := s.DB.QueryRowContext(ctx, "SELECT id,  FROM user WHERE username=?", username).Scan(&id)
	if err != nil {
		return -1, err
	}

	return id, nil
}

func (s *SQLiteStore) CreateContest(ctx context.Context, c *Contest) error {
	var name string
	err := s.DB.QueryRowContext(ctx, "SELECT contest_name FROM contest WHERE contest_name=?", c.ContestName).Scan(&name)
	if err == nil {
		return ErrAlreadyExists
	} else if err != sql.ErrNoRows {
		return err
	}

	res, err := s.DB.ExecContext(ctx, "INSERT INTO contest (contest_name, start_date, end_date, organizer_id) VALUES (?, ?, ?, ?)", c.ContestName, c.StartDate, c.EndDate, c.OrganizerID)
	if err != nil {
		return err
	}
	if id, ierr := res.LastInsertId(); ierr == nil {
		c.ContestID = id
	}
	return nil
}

func (s *SQLiteStore) GetContests(ctx context.Context) ([]Contest, error) {
	rows, err := s.DB.Query("SELECT c.contest_id, c.contest_name, c.start_date, c.end_date, u.username from contest JOIN user ON c.organizer_id = u.user_id")

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contests []Contest

	for rows.Next() {
		var c Contest
		if err := rows.Scan(c.ContestID, c.ContestName, c.StartDate, c.EndDate, c.OrganizerName); err != nil {
			return contests, err
		}
		contests = append(contests, c)
	}

	if err := rows.Err(); err != nil {
		return contests, err
	}

	return contests, nil
}

func (s *SQLiteStore) GetContestByName(ctx context.Context, name string) (Contest, error) {
	var c Contest
	err := s.DB.QueryRowContext(ctx, "SELECT contest_id, contest_name, start_date, end_date, organizer_id,  FROM contest WHERE contest_name=?", name).Scan(&c.ContestID, &c.ContestName, &c.StartDate, &c.EndDate, &c.OrganizerID)
	if err != nil {
		return Contest{}, err
	}

	return c, nil
}

func (s *SQLiteStore) GetContestByID(ctx context.Context, id int) (Contest, error) {
	var c Contest
	err := s.DB.QueryRowContext(ctx, "SELECT c.contest_id, c.contest_name, c.start_date, c.end_date, c.organizer_id, u.organzer_name FROM contest JOIN user ON c.organizer_id = u.user_id WHERE c.contest_id=?", id).Scan(&c.ContestID, &c.ContestName, &c.StartDate, &c.EndDate, &c.OrganizerID, &c.OrganizerName)
	if err != nil {
		return Contest{}, err
	}

	return c, nil
}

func (s *SQLiteStore) GetContestProblems(ctx context.Context, id int) ([]Problem, error) {
	rows, err := s.DB.Query("SELECT problem_id, contest_id, problem_name from problem WHERE contest_id = ?", id)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var problems []Problem

	for rows.Next() {
		var p Problem
		if err := rows.Scan(p.ProblemID, p.ContestID, p.ProblemName); err != nil {
			return problems, err
		}
		problems = append(problems, p)
	}

	if err := rows.Err(); err != nil {
		return problems, err
	}

	return problems, nil
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

func (s *SQLiteStore) Login(ctx context.Context, u *User) error {
	var username, password string

	err := s.DB.QueryRowContext(ctx, "SELECT username, password FROM user WHERE username = ? OR email = ?", u.Username, u.Email).Scan(&username, &password)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	//Verificar password
	hash := []byte(u.Password)

	val, err := security.VerifyHash(password,hash, security.Argon2Params)
	if err != nil {
		return err
	}
	if !val {
		return security.ErrInvalidCredentials
	}	

	return nil
	
}


