package auth

import (
	"fmt"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"project/foundation/web"
	"project/internal/commands"
	"project/internal/repository/postgres/user"
)

// build is the git version of this hard_skill. It is set using build flags in the makefile.
var build = "develop"

type Controller struct {
	user User
}

func NewController(user User) *Controller {
	return &Controller{user: user}
}

func (uc Controller) SignIn(c *web.Context) error {
	var data user.SignInRequest

	err := c.BindFunc(&data, "Login", "Password")
	if err != nil {
		return c.RespondError(err)
	}

	detail, err := uc.user.GetByLogin(c.Ctx, data.Login)
	if err != nil {
		return c.RespondError(err)
	}

	if detail.Password == nil {
		return c.RespondError(&web.Error{
			Err:    errors.New("area not found"),
			Status: http.StatusNotFound,
		})
	}

	//if *detail.Password != data.Password {
	//	return c.RespondError(&web.Error{
	//		Err:    errors.New("incorrect password"),
	//		Status: http.StatusBadRequest,
	//	})
	//}

	if err = bcrypt.CompareHashAndPassword([]byte(*detail.Password), []byte(data.Password)); err != nil {
		return c.RespondError(web.NewRequestError(errors.New(fmt.Sprintf("incorrect password. error: %v", err)), http.StatusBadRequest))
	}

	//var cfg struct {
	//	conf.Version
	//	Args conf.Args
	//	DB   struct {
	//		User       string `conf:"default:postgres"`
	//		Password   string `conf:"default:1"`
	//		Host       string `conf:"default:0.0.0.0"`
	//		Name       string `conf:"default:onda_b2b"`
	//		DisableTLS bool   `conf:"default:true"`
	//	}
	//}
	//cfg.Version.SVN = build

	token, err := commands.GenToken(user.AuthClaims{
		ID:   detail.ID,
		Role: *detail.Role,
	}, "./private.pem")

	if err != nil {
		return c.RespondError(err)
	}

	return c.Respond(map[string]interface{}{
		"status": true,
		"data": map[string]string{
			"token": token,
		},
		"error": nil,
	}, http.StatusOK)
}
