package f5

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"

	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/goHttpEcho"
)

type Authentication interface {
	AuthenticateUser(ctx context.Context, user, passwordHash string) bool
	GetUserInfoFromLogin(ctx context.Context, login string) (*goHttpEcho.UserInfo, error)
}

// Authenticator Create a struct that will implement the Authentication interface
type Authenticator struct {
	// You can add fields here if needed, e.g., a database connection
	mainAdminUserLogin    string
	mainAdminPasswordHash string
	mainAdminEmail        string
	mainAdminId           int
	mainAdminExternalId   int
	jwtChecker            goHttpEcho.JwtChecker
	store                 Storage
}

// AuthenticateUser Implement the AuthenticateUser method for F5Authenticator
func (sa *Authenticator) AuthenticateUser(ctx context.Context, userLogin, passwordHash string) bool {
	l := sa.jwtChecker.GetLogger()
	l.Info("AuthenticateUser", "userLogin", userLogin)
	err := ValidateLogin(userLogin)
	if err != nil {
		l.Warn("invalid user login", "error", err)
		return false
	}
	err = ValidatePasswordHash(passwordHash)
	if err != nil {
		l.Warn("invalid password hash", "error", err)
		return false
	}
	// check if it's the env admin user
	if userLogin == sa.mainAdminUserLogin && passwordHash == sa.mainAdminPasswordHash {
		return true
	}
	// look in db
	if sa.store.Exist(ctx, userLogin) {
		return true
	}
	//MAYBE add login failure to DB ?
	l.Warn("AuthenticateUser is false user will not be authenticated", "userLogin", userLogin)
	return false
}

// GetUserInfoFromLogin Get the JWT claims from the login User
func (sa *Authenticator) GetUserInfoFromLogin(ctx context.Context, login string) (*goHttpEcho.UserInfo, error) {
	l := sa.jwtChecker.GetLogger()
	l.Info("GetUserInfoFromLogin", "login", login)
	if sa.store.Exist(ctx, login) {
		u, err := sa.store.Get(ctx, login)
		if err != nil {
			msg := fmt.Sprintf("GetUserInfoFromLogin(%s) failed doing store.Get err: %s", login, err)
			l.Warn(msg)
			return nil, errors.New(msg)
		} else {
			var isItAdmin bool
			var userDefaultGroup []int
			isItAdmin = false
			userDefaultGroup = append(userDefaultGroup, 0)
			if u.Id < 10 {
				isItAdmin = true
				userDefaultGroup = append(userDefaultGroup, 1) // this is the group id of the global_admin group
			}
			user := &goHttpEcho.UserInfo{
				UserId:     int(u.Id),
				ExternalId: int(u.Id),
				Name:       u.Name,
				Email:      u.Email,
				Login:      login,
				IsAdmin:    isItAdmin,
				Groups:     userDefaultGroup,
			}
			return user, nil
		}
	} else {
		msg := fmt.Sprintf("GetUserInfoFromLogin(%s) failed because user does not exist", login)
		l.Warn(msg)
		return nil, errors.New(msg)
	}
}

// NewF5Authenticator Function to create an instance of F5Authenticator
func NewF5Authenticator(u *goHttpEcho.UserInfo, mainAdminPassword string, jwtCheck goHttpEcho.JwtChecker, store Storage) Authentication {
	l := jwtCheck.GetLogger()
	h := sha256.New()
	h.Write([]byte(mainAdminPassword))
	mainAdminPasswordHash := fmt.Sprintf("%x", h.Sum(nil))
	l.Info("mainAdminUserLogin", "login", u.Login)
	return &Authenticator{
		mainAdminUserLogin:    u.Login,
		mainAdminPasswordHash: mainAdminPasswordHash,
		mainAdminEmail:        u.Email,
		mainAdminId:           u.UserId,
		mainAdminExternalId:   u.ExternalId,
		jwtChecker:            jwtCheck,
		store:                 store,
	}
}
