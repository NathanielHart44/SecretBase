package db

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

// ConnectToDB establishes a connection to the database and returns the *sql.DB object.
func ConnectToDB() (*sql.DB, error) {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	// Fetch the database URL and auth token from environment variables
	dbURL := os.Getenv("TURSO_DATABASE_URL")
	authToken := os.Getenv("TURSO_AUTH_TOKEN")

	if dbURL == "" || authToken == "" {
		fmt.Fprintln(os.Stderr, "Environment variables TURSO_DATABASE_URL or TURSO_AUTH_TOKEN are not set.")
		os.Exit(1)
	}

	// Construct the database connection URL
	url := fmt.Sprintf("%s?authToken=%s", dbURL, authToken)

	// Attempt to connect to the database
	db, err := sql.Open("libsql", url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open db %s: %s\n", url, err)
		os.Exit(1)
	}
	// defer db.Close()

	// Ping the database to ensure the connection is successful
	if err := db.Ping(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to the database: %s\n", err)
		os.Exit(1)
	}

	return db, nil
}

// CreateUser inserts a new user into the database
func CreateUser(db *sql.DB, email, password string, admin bool) error {
	query := `INSERT INTO users (email, password, admin) VALUES (?, ?, ?)`
	_, err := db.Exec(query, email, password, admin)
	if err != nil {
		return fmt.Errorf("failed to create user: %v", err)
	}
	fmt.Println("User created successfully")
	return nil
}

// CreateProject inserts a new project into the database and creates associated environments
func CreateProject(db *sql.DB, name string) error {
	var existingID int
	err := db.QueryRow("SELECT id FROM projects WHERE name = ?", name).Scan(&existingID)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("error checking for existing project: %v", err)
	}

	if existingID != 0 {
		return fmt.Errorf("a project with the name '%s' already exists", name)
	}

	// Insert the project into the projects table
	query := `INSERT INTO projects (name, active) VALUES (?, 1)`
	res, err := db.Exec(query, name)
	if err != nil {
		return fmt.Errorf("failed to create project: %v", err)
	}

	projectID, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to retrieve project ID: %v", err)
	}

	// Insert the environments associated with this project
	envTypes := []string{"development", "staging", "production"}
	for _, envType := range envTypes {
		envQuery := `INSERT INTO environments (project_id, environment_type) VALUES (?, ?)`
		_, err := db.Exec(envQuery, projectID, envType)
		if err != nil {
			return fmt.Errorf("failed to create environment (%s): %v", envType, err)
		}
	}

	fmt.Println("Project and associated environments created successfully")
	return nil
}

// ProjectExists checks if a project with the given name already exists
func ProjectExists(db *sql.DB, name string) (bool, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM projects WHERE name = ?", name).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("error checking if project exists: %v", err)
	}
	return count > 0, nil
}

// SecretExists checks if a secret with the given key, project, and environment already exists
func SecretExists(db *sql.DB, key, projectName, environmentType string) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM secrets s
		INNER JOIN environment_secrets es ON s.id = es.secret_id
		INNER JOIN environments e ON es.environment_id = e.id
		INNER JOIN projects p ON e.project_id = p.id
		WHERE s.key = ? AND p.name = ? AND e.environment_type = ?`

	var count int
	err := db.QueryRow(query, key, projectName, environmentType).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("error checking if secret exists: %v", err)
	}

	return count > 0, nil
}

// CreateSecret inserts a new secret into the database
func CreateSecret(db *sql.DB, key, value, location, projectName, environmentType string) error {
	// Find the environment ID
	var environmentID int
	err := db.QueryRow(`
		SELECT e.id
		FROM environments e
		INNER JOIN projects p ON e.project_id = p.id
		WHERE p.name = ? AND e.environment_type = ?`,
		projectName, environmentType).Scan(&environmentID)
	if err != nil {
		return fmt.Errorf("error finding environment ID: %v", err)
	}

	// Insert the secret into the secrets table
	secretQuery := `INSERT INTO secrets (key, value, location, creator_id) VALUES (?, ?, ?, ?)`
	res, err := db.Exec(secretQuery, key, value, location, 1) // Assuming creator_id = 1 for simplicity
	if err != nil {
		return fmt.Errorf("error creating secret: %v", err)
	}

	secretID, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("error getting last insert ID: %v", err)
	}

	// Link the secret to the environment
	linkQuery := `INSERT INTO environment_secrets (environment_id, secret_id) VALUES (?, ?)`
	_, err = db.Exec(linkQuery, environmentID, secretID)
	if err != nil {
		return fmt.Errorf("error linking secret to environment: %v", err)
	}

	return nil
}

// UpdateSecret updates an existing secret in the database
func UpdateSecret(db *sql.DB, key, value, location, projectName, environmentType string) error {
	query := `
		UPDATE secrets
		SET value = ?, location = ?
		WHERE id = (
			SELECT s.id
			FROM secrets s
			INNER JOIN environment_secrets es ON s.id = es.secret_id
			INNER JOIN environments e ON es.environment_id = e.id
			INNER JOIN projects p ON e.project_id = p.id
			WHERE s.key = ? AND p.name = ? AND e.environment_type = ?)`

	_, err := db.Exec(query, value, location, key, projectName, environmentType)
	if err != nil {
		return fmt.Errorf("error updating secret: %v", err)
	}

	fmt.Println("Secret updated successfully")
	return nil
}

// GetAllSecretsKeys returns all keys for a given project and environment
func GetAllSecretsKeys(db *sql.DB, projectName, environmentType string) ([]string, error) {
	query := `
		SELECT s.key
		FROM secrets s
		INNER JOIN environment_secrets es ON s.id = es.secret_id
		INNER JOIN environments e ON es.environment_id = e.id
		INNER JOIN projects p ON e.project_id = p.id
		WHERE p.name = ? AND e.environment_type = ?`

	rows, err := db.Query(query, projectName, environmentType)
	if err != nil {
		return nil, fmt.Errorf("error fetching keys from database: %v", err)
	}
	defer rows.Close()

	var keys []string
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}
		keys = append(keys, key)
	}

	return keys, nil
}

// DeleteSecret deletes a secret from the database
func DeleteSecret(db *sql.DB, key, projectName, environmentType string) error {
	query := `
		DELETE FROM secrets
		WHERE id = (
			SELECT s.id
			FROM secrets s
			INNER JOIN environment_secrets es ON s.id = es.secret_id
			INNER JOIN environments e ON es.environment_id = e.id
			INNER JOIN projects p ON e.project_id = p.id
			WHERE s.key = ? AND p.name = ? AND e.environment_type = ?)`

	_, err := db.Exec(query, key, projectName, environmentType)
	if err != nil {
		return fmt.Errorf("error deleting secret: %v", err)
	}

	return nil
}

// GetSecrets returns all secrets for a given project and environment
func GetSecrets(db *sql.DB, projectName, environmentType string) ([]Secret, error) {
	query := `
		SELECT s.key, s.value, s.location
		FROM secrets s
		INNER JOIN environment_secrets es ON s.id = es.secret_id
		INNER JOIN environments e ON es.environment_id = e.id
		INNER JOIN projects p ON e.project_id = p.id
		WHERE p.name = ? AND e.environment_type = ?`

	rows, err := db.Query(query, projectName, environmentType)
	if err != nil {
		return nil, fmt.Errorf("error fetching secrets: %v", err)
	}
	defer rows.Close()

	var secrets []Secret
	for rows.Next() {
		var secret Secret
		if err := rows.Scan(&secret.Key, &secret.Value, &secret.Location); err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}
		secrets = append(secrets, secret)
	}

	return secrets, nil
}
