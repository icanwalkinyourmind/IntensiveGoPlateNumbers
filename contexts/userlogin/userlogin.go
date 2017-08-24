// Package userip provides functions for extracting a user IP address from a
// request and associating it with a Context.
package userlogin

import (
	"fmt"
	"net/http"
	"golang.org/x/net/context"
	"../../models/user"
	"crypto/rand"
  "encoding/base64"
)

func GenerateRandomBytes(n int) ([]byte, error) {
    b := make([]byte, n)
    _, err := rand.Read(b)
    if err != nil {
        return nil, err
    }
    return b, nil
}

func GenerateRandomString(s int) (string, error) {
    b, err := GenerateRandomBytes(s)
    return base64.URLEncoding.EncodeToString(b), err
}

func LoginUser(req *http.Request) (string, error) {
	password := req.FormValue("password")
	fmt.Println(password)
  u := user.User{Username: req.FormValue("username")}
	u.Get()
	if password == u.Password {
		return GenerateRandomString(32)
	} else {
		return "", nil
	}
}


type key int

const UserLoginKey key = 0

// NewContext returns a new Context carrying userLogin.
func NewContext(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, UserLoginKey, token)
}
