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
	ContestID     int64  `json:"contest_id"`
	ContestName   string `json:"contest_name"`
	StartDate     string `json:"start_date"`
	EndDate       string `json:"end_date"`
	OrganizerID   int64  `json:"organizer_id"`
	OrganizerName string `json:"organizer_name"`
}

type Problem struct {
	ProblemID   int64  `json:"problem_id"`
	ContestID   *int64 `json:"contest_id"`
	CreatorID   int64  `json:"creator_id"`
	ProblemName string `json:"problem_name"`
	BaseScore   string `json:"problem_name"`
	Description []byte `json:"description"`
}

// Store defines persistence operations used by the HTTP layer.
type Store interface {
	CreateUser(ctx context.Context, u *User) error
	CreateContest(ctx context.Context, c *Contest) error
	CreateProblem(ctx context.Context, p *Problem) error
	Login(ctx context.Context, u *User) error
	GetUserByID(ctx context.Context, id int) (string, error)
	GetUserID(ctx context.Context, username string) (int, error)
	GetContests(ctx context.Context) ([]Contest, error)
	GetContestByName(ctx context.Context, name string) (Contest, error)
	GetContestByID(ctx context.Context, id int) (Contest, error)
	GetContestProblems(ctx context.Context, id int) ([]Problem, error)
}
