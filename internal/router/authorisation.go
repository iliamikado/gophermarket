package router

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v4"
)

const SecretKey = "secret key"

type userLoginKey struct {}

type Claims struct {
    jwt.RegisteredClaims
    UserLogin string
}

func authMiddleware(next http.Handler) http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("JWT")
		fmt.Println(err)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		token := cookie.Value
		userLogin, err := getUserLogin(token)
		fmt.Println(err)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		fmt.Println(userLogin)
		r = r.WithContext(context.WithValue(r.Context(), userLoginKey{}, userLogin))
		next.ServeHTTP(w, r)
	})
}

func getUserLogin(tokenString string) (string, error) {
    claims := &Claims{}
    token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, errors.New("unexpected signing method")
        }
        return []byte(SecretKey), nil
    })

	if err != nil {
        return "", err
    }

    if !token.Valid {
        return "", errors.New("token is not valid")
    }

    return claims.UserLogin, nil
}

func buildJWTString(userLogin string) string {
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims {
        RegisteredClaims: jwt.RegisteredClaims{},
        UserLogin: userLogin,
    })

    tokenString, _ := token.SignedString([]byte(SecretKey))
    return tokenString
}