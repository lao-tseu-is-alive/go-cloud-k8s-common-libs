package goHttpEcho

import (
	"context"
	"crypto/sha256"
	"fmt"
)

type Authentication interface {
	AuthenticateUser(ctx context.Context, user, passwordHash string) bool
	GetUserInfoFromLogin(ctx context.Context, login string) (*UserInfo, error)
}

// SimpleAdminAuthenticator Create a struct that will implement the Authentication interface
type SimpleAdminAuthenticator struct {
	// You can add fields here if needed, e.g., a database connection
	mainAdminUserLogin    string
	mainAdminPasswordHash string
	mainAdminEmail        string
	mainAdminId           int
	mainAdminExternalId   int
	jwtChecker            JwtChecker
}

// AuthenticateUser Implement the AuthenticateUser method for SimpleAdminAuthenticator
func (sa *SimpleAdminAuthenticator) AuthenticateUser(ctx context.Context, userLogin, passwordHash string) bool {
	l := sa.jwtChecker.GetLogger()
	l.Info("AuthenticateUser called", "userLogin", userLogin)
	//l.Info("mainAdminPasswordHash: %s", passwordHash)
	if userLogin == sa.mainAdminUserLogin && passwordHash == sa.mainAdminPasswordHash {
		return true
	}
	sa.jwtChecker.GetLogger().Info("User was not authenticated", "userLogin", userLogin)
	return false
}

// GetUserInfoFromLogin Get the JWT claims from the login User
func (sa *SimpleAdminAuthenticator) GetUserInfoFromLogin(ctx context.Context, login string) (*UserInfo, error) {
	user := &UserInfo{
		UserId:     sa.mainAdminId,
		ExternalId: sa.mainAdminExternalId,
		Name:       fmt.Sprintf("SimpleAdminAuthenticator_%s", sa.mainAdminUserLogin),
		Email:      sa.mainAdminEmail,
		Login:      login,
		IsAdmin:    true,
		Groups:     []int{1}, // this is the group id of the global_admin group
	}
	return user, nil
}

// NewSimpleAdminAuthenticator Function to create an instance of SimpleAdminAuthenticator
func NewSimpleAdminAuthenticator(u *UserInfo, mainAdminPassword string, jwtCheck JwtChecker) Authentication {
	l := jwtCheck.GetLogger()
	h := sha256.New()
	h.Write([]byte(mainAdminPassword))
	mainAdminPasswordHash := fmt.Sprintf("%x", h.Sum(nil))
	l.Info("NewSimpleAdminAuthenticator created", "mainAdminUserLogin", u.Login)
	//l.Info("mainAdminUserPassword: %s", mainAdminPassword)
	//l.Info("mainAdminPasswordHash: %s", mainAdminPasswordHash)
	return &SimpleAdminAuthenticator{
		mainAdminUserLogin:    u.Login,
		mainAdminPasswordHash: mainAdminPasswordHash,
		mainAdminEmail:        u.Email,
		mainAdminId:           u.UserId,
		mainAdminExternalId:   u.ExternalId,
		jwtChecker:            jwtCheck,
	}
}
