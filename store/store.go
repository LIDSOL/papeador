package store

import (
	"context"
	"errors"
)

var (
	ErrAlreadyExists = errors.New("already exists")
	ErrNotFound      = errors.New("not found")
)

type User struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	JWT      string `json:"jwt"`
}

type Contest struct {
	ContestID   int64  `json:"contest_id"`
	ContestName string `json:"contest_name"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
}

type Problem struct {
	ProblemID   int64  `json:"problem_id"`
	ContestID   *int64 `json:"contest_id"`
	CreatorID   int64  `json:"creator_id"`
	ProblemName string `json:"problem_name"`
	Description string `json:"description"`
}

// Store defines persistence operations used by the HTTP layer.
type Store interface {
	CreateUser(ctx context.Context, u *User) error
	CreateContest(ctx context.Context, c *Contest) error
	CreateProblem(ctx context.Context, p *Problem) error
	Login(ctx context.Context, u *User) error
}
