package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	z "github.com/Oudwins/zog"
	"github.com/go-chi/jwtauth/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/joaopsramos/fincon/internal/config"

	errs "github.com/joaopsramos/fincon/internal/error"
	"github.com/joaopsramos/fincon/internal/service"
	"github.com/joaopsramos/fincon/internal/util"
)

type UserIDKeyType string

var (
	tokenExpiresIn               = time.Hour * 24 * 7
	UserIDKey      UserIDKeyType = "user_id"
)

var userCreateSchema = z.Struct(z.Schema{
	"email":    z.String().Trim().Max(160).Email(z.Message("must be valid")).Required(),
	"password": z.String().Trim().Min(8, z.Message("must contain at least 8 characters")).Max(72, z.Message("must have at most 72 characters")).Required(),
	"salary":   z.Float().GT(0, z.Message("must be greater than 0")).Required(),
})

var userLoginSchema = z.Struct(z.Schema{
	"email":    z.String().Trim().Required(),
	"password": z.String().Trim().Required(),
})

func (a *Api) CreateUser(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Email    string  `json:"email"`
		Password string  `json:"password"`
		Salary   float64 `json:"salary"`
	}

	if errs := util.ParseZodSchema(userCreateSchema, r.Body, &params); errs != nil {
		fmt.Println(params, errs)
		a.HandleZodError(w, errs)
		return
	}

	if _, err := a.userService.GetByEmail(r.Context(), params.Email); err == nil {
		a.sendError(w, http.StatusConflict, "email already in use")
		return
	} else if !errors.Is(err, errs.ErrNotFound{}) {
		a.HandleError(w, err)
		return
	}

	dto := service.CreateUserDTO{
		Email:           params.Email,
		Password:        params.Password,
		CreateSalaryDTO: service.CreateSalaryDTO{Amount: params.Salary},
	}

	user, salary, err := a.userService.Create(r.Context(), dto)
	if err != nil {
		a.HandleError(w, err)
		return
	}

	a.sendJSON(w, http.StatusCreated, util.M{
		"user":   user.ToDTO(),
		"salary": salary.ToDTO(),
		"token":  a.GenerateToken(user.ID, tokenExpiresIn),
	})
}

func (a *Api) UserLogin(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if errs := util.ParseZodSchema(userLoginSchema, r.Body, &params); errs != nil {
		a.HandleZodError(w, errs)
		return
	}

	user, err := a.userService.GetByEmailAndPassword(r.Context(), params.Email, params.Password)
	if errors.Is(err, errs.ErrInvalidCredentials) {
		a.invalidCredentials(w)
		return
	} else if err != nil {
		panic(err)
	}

	a.sendJSON(w, http.StatusCreated, util.M{"token": a.GenerateToken(user.ID, tokenExpiresIn)})
}

func (a *Api) GenerateToken(userID uuid.UUID, expiresIn time.Duration) string {
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{"sub": userID, "exp": time.Now().UTC().Add(expiresIn).Unix()},
	)

	tokenString, err := token.SignedString(config.SecretKey())
	if err != nil {
		panic(err)
	}

	return tokenString
}

func (a *Api) invalidCredentials(w http.ResponseWriter) {
	a.sendError(w, http.StatusUnauthorized, "invalid email or password")
}

func (a *Api) PutUserIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, claims, err := jwtauth.FromContext(r.Context())
		if err != nil {
			panic(err)
		}

		sub, ok := claims["sub"].(string)
		if !ok {
			panic("failed to get subject from token")
		}

		userID, err := uuid.Parse(sub)
		if err != nil {
			panic(err)
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
