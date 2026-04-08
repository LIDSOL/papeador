package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"lidsol.org/papeador/store"

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
	// run schema
	schemaBytes, err := os.ReadFile("../schema.sql")
	if err != nil {
		log.Fatalf("Failed to read schema.sql: %v", err)
	}
	if _, err := db.Exec(string(schemaBytes)); err != nil {
		log.Fatalf("Failed to execute schema: %v", err)
	}

	// Run tests
	exitCode := m.Run()

	// Teardown: close the database connection
	testDB.Close()
	// remove the temporary DB file
	if err := os.Remove("test_api.db"); err != nil {
		log.Printf("Warning: failed to remove test DB: %v", err)
	}

	os.Exit(exitCode)
}

// --- Helper Functions ---

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	// create an ApiContext that uses the test sqlite DB
	apiCtx := ApiContext{Store: store.NewSQLiteStore(testDB)}
	handler := http.HandlerFunc(apiCtx.createUser)
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

func createUserRequest(user store.User) (*http.Request, error) {
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
	// apiCtx is constructed inside executeRequest per-call using testDB
	clearUserTable()
	newUser := store.User{
		Username: "testuser",
		Password: "testpass",
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
	var createdUser store.User
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
