package goHttpEcho

import (
	"crypto/sha256"
	"fmt"
)

type Authentication interface {
	AuthenticateUser(user, passwordHash string) bool
	GetUserInfoFromLogin(login string) (*UserInfo, error)
}

// SimpleAdminAuthenticator Create a struct that will implement the Authentication interface
type SimpleAdminAuthenticator struct {
	// You can add fields here if needed, e.g., a database connection
	mainAdminUserLogin    string
	mainAdminPasswordHash string
	mainAdminEmail        string
	mainAdminId           int
	jwtChecker            JwtChecker
}

// AuthenticateUser Implement the AuthenticateUser method for SimpleAdminAuthenticator
func (sa *SimpleAdminAuthenticator) AuthenticateUser(userLogin, passwordHash string) bool {
	l := sa.jwtChecker.GetLogger()
	l.Info("mainAdminUserLogin: %s", userLogin)
	//l.Info("mainAdminPasswordHash: %s", passwordHash)
	if userLogin == sa.mainAdminUserLogin && passwordHash == sa.mainAdminPasswordHash {
		return true
	}
	sa.jwtChecker.GetLogger().Info("User %s was not authenticated", userLogin)
	return false
}

// GetUserInfoFromLogin Get the JWT claims from the login User
func (sa *SimpleAdminAuthenticator) GetUserInfoFromLogin(login string) (*UserInfo, error) {
	user := &UserInfo{
		UserId:  sa.mainAdminId,
		Name:    fmt.Sprintf("SimpleAdminAuthenticator_%s", sa.mainAdminUserLogin),
		Email:   sa.mainAdminEmail,
		Login:   login,
		IsAdmin: true,
	}
	return user, nil
}

// NewSimpleAdminAuthenticator Function to create an instance of SimpleAdminAuthenticator
func NewSimpleAdminAuthenticator(mainAdminUser, mainAdminPassword, mainAdminEmail string, mainAdminId int, jwtCheck JwtChecker) Authentication {
	l := jwtCheck.GetLogger()
	h := sha256.New()
	h.Write([]byte(mainAdminPassword))
	mainAdminPasswordHash := fmt.Sprintf("%x", h.Sum(nil))
	l.Info("mainAdminUserLogin: %s", mainAdminUser)
	//l.Info("mainAdminUserPassword: %s", mainAdminPassword)
	//l.Info("mainAdminPasswordHash: %s", mainAdminPasswordHash)
	return &SimpleAdminAuthenticator{
		mainAdminUserLogin:    mainAdminUser,
		mainAdminPasswordHash: mainAdminPasswordHash,
		mainAdminEmail:        mainAdminEmail,
		mainAdminId:           mainAdminId,
		jwtChecker:            jwtCheck,
	}
}
