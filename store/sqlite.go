package store

import (
    "context"
    "database/sql"
)

type SQLiteStore struct{
    DB *sql.DB
}

func NewSQLiteStore(db *sql.DB) Store {
    return &SQLiteStore{DB: db}
}

func (s *SQLiteStore) CreateUser(ctx context.Context, u *User) error {
    // Verify unique username/email
    var username string
    err := s.DB.QueryRowContext(ctx, "SELECT username FROM user WHERE username=? OR email=?", u.Username, u.Email).Scan(&username)
    if err == nil {
        return ErrAlreadyExists
    } else if err != sql.ErrNoRows {
        return err
    }

    res, err := s.DB.ExecContext(ctx, "INSERT INTO user (username,passhash,email) VALUES (?, ?, ?)", u.Username, u.Passhash, u.Email)
    if err != nil {
        return err
    }
    if id, ierr := res.LastInsertId(); ierr == nil {
        u.UserID = id
    }
    return nil
}

func (s *SQLiteStore) CreateContest(ctx context.Context, c *Contest) error {
    var name string
    err := s.DB.QueryRowContext(ctx, "SELECT contest_name FROM contest WHERE contest_name=?", c.ContestName).Scan(&name)
    if err == nil {
        return ErrAlreadyExists
    } else if err != sql.ErrNoRows {
        return err
    }

    res, err := s.DB.ExecContext(ctx, "INSERT INTO contest (contest_name) VALUES (?)", c.ContestName)
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
