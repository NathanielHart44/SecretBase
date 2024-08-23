package db

type EnvironmentType string

const (
	Development EnvironmentType = "development"
	Staging     EnvironmentType = "staging"
	Production  EnvironmentType = "production"
)

type User struct {
	ID       int
	Email    string
	Password string
	Admin    bool
}

type Secret struct {
	ID       int
	Key      string
	Creator  User
	Value    string
	Location string
}

type Environment struct {
	ID      int
	Secrets map[string]Secret // A map of Secrets (key-value pairs)
	Type    EnvironmentType
}

type Project struct {
	ID          int
	Name        string
	Development Environment
	Staging     Environment
	Production  Environment
	Active      bool // Indicates if the project has been sunset or not
}
