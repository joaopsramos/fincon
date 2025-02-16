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
	errs "github.com/joaopsramos/fincon/internal/error"
	"github.com/joaopsramos/fincon/internal/service"
	"github.com/joaopsramos/fincon/internal/util"
)

var (
	tokenExpiresIn = time.Hour * 24 * 7
	jwtContextKey  = "token"
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

func (a *Api) CreateUser(c *fiber.Ctx) error {
	var params struct {
		Email    string  `json:"email"`
		Password string  `json:"password"`
		Salary   float64 `json:"salary"`
	}

	if errs := util.ParseZodSchema(userCreateSchema, c.Body(), &params); errs != nil {
		return a.HandleZodError(c, errs)
	}

	if _, err := a.userService.GetByEmail(params.Email); err == nil {
		return c.Status(http.StatusConflict).JSON(util.M{"error": "email already in use"})
	} else if !errors.Is(err, errs.ErrNotFound{}) {
		return a.HandleError(c, err)
	}

	dto := service.CreateUserDTO{
		Email:           params.Email,
		Password:        params.Password,
		CreateSalaryDTO: service.CreateSalaryDTO{Amount: params.Salary},
	}

	user, salary, err := a.userService.Create(dto)
	if err != nil {
		return a.HandleError(c, err)
	}

	return c.Status(http.StatusCreated).JSON(util.M{
		"user":   user.ToDTO(),
		"salary": salary.ToDTO(),
		"token":  a.generateToken(user.ID),
	})
}

func (a *Api) UserLogin(c *fiber.Ctx) error {
	var params struct {
		Email    string
		Password string
	}

	if errs := util.ParseZodSchema(userLoginSchema, c.Body(), &params); errs != nil {
		return a.HandleZodError(c, errs)
	}

	user, err := a.userService.GetByEmailAndPassword(params.Email, params.Password)
	if errors.Is(err, errs.ErrInvalidCredentials) {
		return a.invalidCredentials(c)
	} else if err != nil {
		panic(err)
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
	return domain.CreateAccessToken(userID, tokenExpiresIn)
}

func (a *Api) invalidCredentials(c *fiber.Ctx) error {
	return c.Status(http.StatusUnauthorized).JSON(util.M{"error": "invalid email or password"})
}
