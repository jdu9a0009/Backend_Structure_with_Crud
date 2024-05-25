package router

import (
	"github.com/redis/go-redis/v9"
	"project/foundation/web"
	"project/internal/auth"

	"project/internal/middleware"
	"project/internal/pkg/repository/postgresql"

	"project/internal/repository/postgres/user"

	auth_controller "project/internal/controller/http/v1/auth"
	user_controller "project/internal/controller/http/v1/user"
)

type Router struct {
	*web.App
	postgresDB         *postgresql.Database
	redisDB            *redis.Client
	port               string
	auth               *auth.Auth
	fileServerBasePath string
}

func NewRouter(
	app *web.App,
	postgresDB *postgresql.Database,
	redisDB *redis.Client,
	port string,
	auth *auth.Auth,
	fileServerBasePath string,
) *Router {
	return &Router{
		app,
		postgresDB,
		redisDB,
		port,
		auth,
		fileServerBasePath,
	}
}

func (r Router) Init() error {

	// repositories:
	// - postgresql
	userPostgres := user.NewRepository(r.postgresDB)

	// controller
	userController := user_controller.NewController(userPostgres)

	authController := auth_controller.NewController(userPostgres)

	// #auth
	r.Post("/api/v1/sign-in", authController.SignIn)

	// #user
	r.Get("/api/v1/user/list", userController.GetUserList, middleware.Authenticate(r.auth, auth.RoleAdmin))
	r.Get("/api/v1/user/:id", userController.GetUserDetailById, middleware.Authenticate(r.auth, auth.RoleAdmin))
	r.Post("/api/v1/user/create", userController.CreateUser, middleware.Authenticate(r.auth, auth.RoleAdmin))
	r.Put("/api/v1/user/:id", userController.UpdateUserAll, middleware.Authenticate(r.auth, auth.RoleAdmin))
	r.Patch("/api/v1/user/:id", userController.UpdateUserColumns, middleware.Authenticate(r.auth, auth.RoleAdmin))
	r.Delete("/api/v1/user/:id", userController.DeleteUser, middleware.Authenticate(r.auth, auth.RoleAdmin))

	return r.Run(r.port)
}
