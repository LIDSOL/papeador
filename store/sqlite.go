package store

import (
	"context"
	"database/sql"
	"log"
	"time"

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
	passhash, passsalt, err := security.HashPassword(u.Password, security.Argon2Params)
	if err != nil {
		return err
	}

	// Inserting user
	res, err := s.DB.ExecContext(ctx, "INSERT INTO user (username,passhash,passsalt,email) VALUES (?, ?, ?, ?)", u.Username, passhash, passsalt, u.Email)

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
	err := s.DB.QueryRowContext(ctx, "SELECT user_id  FROM user WHERE username=?", username).Scan(&id)
	if err != nil {
		log.Println("ERR", err)
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
	rows, err := s.DB.QueryContext(ctx, "SELECT c.contest_id, c.contest_name, c.start_date, c.end_date, u.username FROM contest c JOIN user u ON c.organizer_id = u.user_id")

	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	defer rows.Close()

	var contests []Contest

	for rows.Next() {
		var c Contest
		if err := rows.Scan(&c.ContestID, &c.ContestName, &c.StartDate, &c.EndDate, &c.OrganizerName); err != nil {
			return contests, err
		}
		contests = append(contests, c)
	}

	// if err := rows.Err(); err != nil {
	// 	return contests, nil
	// }

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
	err := s.DB.QueryRowContext(ctx, "SELECT c.contest_id, c.contest_name, c.start_date, c.end_date, c.organizer_id, u.username FROM contest c JOIN user u ON c.organizer_id = u.user_id WHERE c.contest_id=?", id).Scan(&c.ContestID, &c.ContestName, &c.StartDate, &c.EndDate, &c.OrganizerID, &c.OrganizerName)
	if err != nil {
		return Contest{}, err
	}

	return c, nil
}

func (s *SQLiteStore) GetContestProblems(ctx context.Context, id int) ([]Problem, error) {
	rows, err := s.DB.Query("SELECT p.problem_id, c.contest_id, p.problem_name from problem p JOIN contest_has_problem c ON p.problem_id = c.problem_id WHERE c.contest_id = ?", id)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var problems []Problem

	for rows.Next() {
		var p Problem
		if err := rows.Scan(&p.ProblemID, &p.ContestID, &p.ProblemName); err != nil {
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

	res, err := s.DB.ExecContext(ctx, "INSERT INTO problem (creator_id, problem_name, base_score, description) VALUES (?, ?, ?, ?)", p.CreatorID, p.ProblemName, p.BaseScore, p.Description)

	if err != nil {
		return err
	}

	if id, ierr := res.LastInsertId(); ierr == nil {
		p.ProblemID = id

		_, err := s.DB.ExecContext(ctx, "INSERT INTO contest_has_problem (contest_id, problem_id, score) VALUES (?, ?, 0)", p.ContestID, p.ProblemID)

		if err != nil {
			return err
		}

	}
	return nil
}

func (s *SQLiteStore) GetProblemByIDs(ctx context.Context, contestID, problemID int) (*Problem, error) {
	p := Problem{}
	err := s.DB.QueryRowContext(ctx, "SELECT p.problem_id, p.problem_name, p.description from problem p JOIN contest_has_problem c ON p.problem_id = c.problem_id WHERE c.contest_id = ? AND p.problem_id = ?", contestID, problemID).Scan(&p.ProblemID, &p.ProblemName, &p.Description)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}
	log.Println("HOLA")

	err = s.DB.QueryRowContext(ctx, "SELECT t.expected_out, t.given_input from test_case t JOIN problem p ON p.problem_id = t.problem_id WHERE t.problem_id = ? ORDER BY t.num_test_case ASC LIMIT 1", p.ProblemID).Scan(&p.SampleOut, &p.SampleInput)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}

	return &p, nil
}

func (s *SQLiteStore) CreateTestCase(ctx context.Context, t *TestCase) error {
	var pid int64
	err := s.DB.QueryRowContext(ctx, "SELECT problem_id FROM problem WHERE problem_id=?", *&t.ProblemID).Scan(&pid)
	if err == sql.ErrNoRows {
		return ErrNotFound
	} else if err != nil {
		return err
	}

	res, err := s.DB.ExecContext(ctx, "INSERT INTO test_case (problem_id, num_test_case, time_limit, expected_out, given_input) VALUES (?, ?, ?, ?, ?)", t.ProblemID, t.NumTestCase, t.TimeLimit, t.ExpectedOut, t.GivenInput)

	if err != nil {
		return err
	}

	if id, ierr := res.LastInsertId(); ierr == nil {
		t.TestCaseID = id
	}
	return nil
}

func (s *SQLiteStore) GetTestCases(ctx context.Context, problemID int) ([]TestCase, error) {
	rows, err := s.DB.Query("SELECT time_limit, expected_out, given_input from test_case  WHERE problem_id = ?", problemID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var testcases []TestCase

	for rows.Next() {
		var t TestCase
		if err := rows.Scan(&t.TimeLimit, &t.ExpectedOut, &t.GivenInput); err != nil {
			return testcases, err
		}
		testcases = append(testcases, t)
	}

	if err := rows.Err(); err != nil {
		return testcases, err
	}

	return testcases, nil
}

func (s *SQLiteStore) Login(ctx context.Context, u *User) error {
	var username, storedHash, salt string

	err := s.DB.QueryRowContext(ctx, "SELECT username, passhash, passsalt FROM user WHERE username = ? OR email = ?", u.Username, u.Email).Scan(&username, &storedHash, &salt)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	//Verificar password
	inputPass := []byte(u.Password)

	val, err := security.VerifyHash(inputPass, []byte(storedHash), []byte(salt), security.Argon2Params)
	if err != nil {
		return err
	}
	if !val {
		log.Println("ERR", err)
		return security.ErrInvalidCredentials
	}	

	return nil
	
}

func (s *SQLiteStore) GetProblemScore(ctx context.Context, problemID int) (int, error) {
	var score int
	err := s.DB.QueryRowContext(ctx, "SELECT base_score FROM problem WHERE problem_id = ?", problemID).Scan(&score)
	if err == sql.ErrNoRows {
		return 0, ErrNotFound
	}
	if err != nil {
		return 0, err
	}
	return score, nil
}

// Calculamos el puntaje de una subission usando la formula de Codeforces:
// max(0.3⋅x, x−⌊(120x⋅t)/(250d)⌋−50w)
// 
// x = el puntaje incial del problema
// t = tiempo en minutos cuando se resolvió el problema
// d = duración del concurso en minutos
// w = número de envíos incorrectos antes del primero aceptado
func (s *SQLiteStore) CalculateSubmissionScore(ctx context.Context, submissionID int64) (int, error) {
	var (
		baseScore       int
		submissionDate  string
		contestStart    string
		contestEnd      string
		userID          int64
		problemID       int64
		statusID        int64
	)
	
	query := `
		SELECT 
			p.base_score,
			s.date,
			c.start_date,
			c.end_date,
			s.user_id,
			s.problem_id,
			s.status_id
		FROM submission s
		JOIN problem p ON s.problem_id = p.problem_id
		JOIN contest_has_problem chp ON p.problem_id = chp.problem_id
		JOIN contest c ON chp.contest_id = c.contest_id
		WHERE s.submission_id = ?
	`
	
	err := s.DB.QueryRowContext(ctx, query, submissionID).Scan(
		&baseScore, &submissionDate, &contestStart, &contestEnd, 
		&userID, &problemID, &statusID,
	)
	if err == sql.ErrNoRows {
		return 0, ErrNotFound
	}
	if err != nil {
		return 0, err
	}
	
	contestDuration, err := calculateDurationInMinutes(contestStart, contestEnd)
	if err != nil {
		return 0, err
	}
	
	submissionTime, err := calculateDurationInMinutes(contestStart, submissionDate)
	if err != nil {
		return 0, err
	}
	
	wrongSubmissionsQuery := `
		SELECT COUNT(*) 
		FROM submission 
		WHERE user_id = ? 
		AND problem_id = ? 
		AND submission_id < ? 
		AND status_id != 1
	`
	
	var wrongSubmissions int
	err = s.DB.QueryRowContext(ctx, wrongSubmissionsQuery, userID, problemID, submissionID).Scan(&wrongSubmissions)
	if err != nil {
		return 0, err
	}
	
	// aplicamos la formula
	x := float64(baseScore)
	t := float64(submissionTime)
	d := float64(contestDuration)
	w := float64(wrongSubmissions)
	
	minScore := 0.3 * x
	timeDecay := float64(int((120 * x * t) / (250 * d)))
	penaltyScore := x - timeDecay - (50 * w)
	
	var finalScore int
	if penaltyScore > minScore {
		finalScore = int(penaltyScore)
	} else {
		finalScore = int(minScore)
	}
	
	return finalScore, nil
}


// Para el calulo uso el formato de tiempo YYYY-MM-DDTHH:mm:ssZ por default
func calculateDurationInMinutes(start, end string) (int, error) {
	const layout = "2006-01-01 12:12:12"
	
	startTime, err := parseFlexibleTime(start, layout)
	if err != nil {
		return 0, err
	}
	
	endTime, err := parseFlexibleTime(end, layout)
	if err != nil {
		return 0, err
	}
	
	duration := endTime.Sub(startTime)
	minutes := int(duration.Minutes())
	
	return minutes, nil
}


func parseFlexibleTime(timeStr, preferredLayout string) (time.Time, error) {
	t, err := time.Parse(preferredLayout, timeStr)
	if err == nil {
		return t, nil
	}
	
	t, err = time.Parse(time.RFC3339, timeStr)
	if err == nil {
		return t, nil
	}
	
	layouts := []string{
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	
	for _, layout := range layouts {
		t, err = time.Parse(layout, timeStr)
		if err == nil {
			return t, nil
		}
	}
	
	return time.Time{}, err
}