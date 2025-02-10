package api

import (
	"errors"
	"net/http"
	"time"

	z "github.com/Oudwins/zog"
	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/joaopsramos/fincon/internal/config"
	"github.com/joaopsramos/fincon/internal/domain"
	"github.com/joaopsramos/fincon/internal/util"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	tokenExpiresIn = time.Hour * 24 * 7
	jwtContextKey  = "token"
)

var userCreateSchema = z.Struct(z.Schema{
	"email":    z.String().Trim().Max(160).Email().Required(),
	"password": z.String().Trim().Min(8).Max(72).Required(),
})

var userLoginSchema = z.Struct(z.Schema{
	"email":    z.String().Trim().Required(),
	"password": z.String().Trim().Required(),
})

func (a *Api) CreateUser(c *fiber.Ctx) error {
	var params struct {
		Email    string
		Password string
	}

	if errs := util.ParseZodSchema(userCreateSchema, c.Body(), &params); errs != nil {
		return a.HandleZodError(c, errs)
	}

	if _, err := a.userRepo.GetByEmail(params.Email); err == nil {
		return c.Status(http.StatusConflict).JSON(util.M{"error": "email already in use"})
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		panic(err)
	}

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(params.Password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}

	user := &domain.User{Email: params.Email, HashPassword: string(hashPassword)}
	err = a.userRepo.Create(user)
	if err != nil {
		panic(err)
	}

	return c.Status(http.StatusCreated).JSON(util.M{"user": user, "token": a.generateToken(user.ID)})
}

func (a *Api) UserLogin(c *fiber.Ctx) error {
	var params struct {
		Email    string
		Password string
	}

	if errs := util.ParseZodSchema(userLoginSchema, c.Body(), &params); errs != nil {
		return a.HandleZodError(c, errs)
	}

	user, err := a.userRepo.GetByEmail(params.Email)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return c.Status(http.StatusUnprocessableEntity).JSON(util.M{"error": "invalid email or password"})
	} else if err != nil {
		panic(err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.HashPassword), []byte(params.Password)); err != nil {
		return c.Status(http.StatusUnprocessableEntity).JSON(util.M{"error": "invalid email or password"})
	}

	return c.Status(http.StatusCreated).JSON(util.M{"token": a.generateToken(user.ID)})
}

func (a *Api) ValidateTokenMiddleware() fiber.Handler {
	return jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: config.SecretKey()},
		ContextKey: jwtContextKey,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			if err.Error() == jwtware.ErrJWTMissingOrMalformed.Error() {
				return c.Status(fiber.StatusBadRequest).JSON(util.M{"error": jwtware.ErrJWTMissingOrMalformed.Error()})
			}
			return c.Status(fiber.StatusUnauthorized).JSON(util.M{"error": "invalid or expired JWT"})
		},
	})
}

func (a *Api) PutUserIDMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Locals(jwtContextKey).(*jwt.Token)
		id, err := token.Claims.GetSubject()
		if err != nil {
			panic(err)
		}

		c.Locals("user_id", id)

		return c.Next()
	}
}

func (a *Api) generateToken(userID uuid.UUID) string {
	return domain.CreateToken(userID, tokenExpiresIn)
}
