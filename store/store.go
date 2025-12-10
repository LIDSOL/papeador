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
	ProblemID      int64  `json:"problem_id"`
	ContestID      *int64 `json:"contest_id"`
	CreatorID      int64  `json:"creator_id"`
	ProblemName    string `json:"problem_name"`
	BaseScore      string `json:"base_score"`
	Description    []byte `json:"description"`
	SampleInput    []byte `json:"sample_input"`
	SampleOut      []byte `json:"sample_out"`
	SampleInputStr string `json:"sample_input_str"`
	SampleOutStr   string `json:"sample_out_str"`
}

type TestCase struct {
	TestCaseID  int64  `json:"test_case_id"`
	ProblemID   int64  `json:"problem_id"`
	NumTestCase int64  `json:"num_test_case"`
	TimeLimit   int64  `json:"time_limit"`
	ExpectedOut []byte `json:"expected_out"`
	GivenInput  []byte `json:"given_input"`
}
type UserScore struct {
	Rank   int `json:"rank"`
	UserID int `json:"user_id"`
	Score  int `json:"score"`
}
type ScoringInput struct {
	Status        string
	ExecutionTime float64
}

// Store defines persistence operations used by the HTTP layer.
type Store interface {
	CreateUser(ctx context.Context, u *User) error
	CreateContest(ctx context.Context, c *Contest) error
	CreateProblem(ctx context.Context, p *Problem) error
	CreateTestCase(ctx context.Context, t *TestCase) error
	Login(ctx context.Context, u *User) error
	GetUserByID(ctx context.Context, id int) (string, error)
	GetUserID(ctx context.Context, username string) (int, error)
	GetContests(ctx context.Context) ([]Contest, error)
	GetContestByName(ctx context.Context, name string) (Contest, error)
	GetContestByID(ctx context.Context, id int) (Contest, error)
	GetContestProblems(ctx context.Context, id int) ([]Problem, error)
	GetProblemByIDs(ctx context.Context, contestID, problemID int) (*Problem, error)
	GetTestCases(ctx context.Context, problemID int) ([]TestCase, error)
	UserScore(ctx context.Context, userID int, score int) error
	GetGlobalRanking(ctx context.Context) ([]UserScore, error)
}
