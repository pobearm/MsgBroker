package api

type Valid struct {
	username string
	password string
}

func New_Valid(username string, password string) *Valid {
	return &Valid{
		username: username,
		password: password,
	}
}

func (v *Valid) ValidateUser(username string, password string) bool {
	if username == v.username && password == v.password {
		return true
	}
	return false
}
