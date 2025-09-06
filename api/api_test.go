package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"testing"

	_ "modernc.org/sqlite"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	// Setup: create a temporary database for testing
	db, err := sql.Open("sqlite", "test_api.db")
	if err != nil {
		log.Fatal(err)
	}
	testDB = db
	defer os.Remove("test_api.db") // clean up

	// run schema
	command := exec.Command("bash", "-c", "sqlite3 test_api.db < ../schema.sql")
	output, err := command.CombinedOutput()
	if err != nil {
		log.Fatalf("Failed to run schema: %s, %s", err, string(output))
	}

	// Run tests
	exitCode := m.Run()

	// Teardown: close the database connection
	testDB.Close()

	os.Exit(exitCode)
}

// --- Helper Functions ---

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(createUser)
	handler.ServeHTTP(rr, req)
	return rr
}

func checkStatus(t *testing.T, rr *httptest.ResponseRecorder, expectedStatus int) {
	t.Helper()
	if status := rr.Code; status != expectedStatus {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, expectedStatus)
	}
}

func createUserRequest(user User) (*http.Request, error) {
	jsonUser, _ := json.Marshal(user)
	return http.NewRequest("POST", "/users", bytes.NewBuffer(jsonUser))
}

func clearUserTable() {
	testDB.Exec("DELETE FROM user")
	testDB.Exec("UPDATE sqlite_sequence SET seq = 0 WHERE name = 'user'")
}

// --- Tests ---

func TestCreateUser(t *testing.T) {
	// Initialize the API with the test database
	db = testDB
	clearUserTable()

	newUser := User{
		Username: "testuser",
		Passhash: "testpass",
		Email:    "test@example.com",
	}

	// Scenario 1: Create user successfully
	req, err := createUserRequest(newUser)
	if err != nil {
		t.Fatal(err)
	}
	rr := executeRequest(req)
	checkStatus(t, rr, http.StatusCreated)

	// Check the response body
	var createdUser User
	err = json.NewDecoder(rr.Body).Decode(&createdUser)
	if err != nil {
		t.Fatal(err)
	}
	if createdUser.Username != newUser.Username {
		t.Errorf("handler returned unexpected body: got %v want %v",
			createdUser.Username, newUser.Username)
	}

	// Scenario 2: Try to create the same user again (should fail)
	req, err = createUserRequest(newUser)
	if err != nil {
		t.Fatal(err)
	}
	rr = executeRequest(req)
	checkStatus(t, rr, http.StatusConflict)
}
