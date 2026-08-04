package main

import (
	"github.com/asapgiri/golib/session"
	"errors"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

func Authenticate(a *session.Auth) {
    // TODO: Check if user can be mocked out to existing and be used for unauthenticated login...
    if a.Username != "" {
        user := Staff{}
        user.FindByName(a.Username)
        if "" == user.Nick {
            *a = session.Auth{}
            return
        }

        a.Id = user.Id.Hex()
        a.Username = user.Nick
        a.Email = user.Email
        a.Error = ""
    }
}

func (user *Staff) Register(password_clear_a string, password_clear_b string) error {
    old_user := Staff{}

    if old_user.FindByName(user.Nick) == nil {
        return errors.New("Nick already exists!")
    }

    if len(user.Nick) < 3 {
        return errors.New("Username must be a minimum of 3 characters long!")
    }
    if len(password_clear_a) < 5 {
        return errors.New("Password must be a minimum of 5 characters long!")
    }

    if password_clear_a != password_clear_b {
        return errors.New("Passwords doesnt match!")
    }

    pwh, _ := bcrypt.GenerateFromPassword([]byte(password_clear_a), 0)
    user.Id = primitive.NewObjectID()
    user.Passwd = string(pwh)

    user.Add()

    return nil
}

func (user *Staff) Login(nick string, password_clear string) error {
    if nil != user.FindByName(nick) {
        return errors.New("Bad username!")
    }

    if nil != bcrypt.CompareHashAndPassword([]byte(user.Passwd), []byte(password_clear)) {
        return errors.New("Bad password!")
    }

    return nil
}
