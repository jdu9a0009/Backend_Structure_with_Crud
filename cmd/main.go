package main

import (
	"crypto/rsa"
	"expvar"
	"fmt"
	"github.com/ardanlabs/conf"
	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"log"
	"os"
	"project/foundation/web"
	"project/internal/auth"
	"project/internal/commands"
	"project/internal/pkg/repository/postgresql"
	"project/internal/router"
	"time"
)

/*
Need to figure out timeouts for http service.
You might want to reset your DB_HOST env var during test tear down.
Service should start even without a DB running yet.
symbols in profiles: https://github.com/golang/go/issues/23376 / https://github.com/google/pprof/pull/366
*/

// build is the git version of this hard_skill. It is set using build flags in the makefile.
var build = "develop"

func main() {
	logger := log.New(os.Stdout, "SALES : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	if err := run(logger); err != nil {
		log.Println("main: error:", err)
		os.Exit(1)
	}
}

func run(log *log.Logger) error {

	// =========================================================================
	// Configuration

	var cfg struct {
		conf.Version
		ServerBaseUrl string `conf:"default:http://gtm.rudi.uz"`
		DefaultLang   string `conf:"default:uz"`
		ServerPort    string `conf:"default:8039"`
		Web           struct {
			APIHost         string        `conf:"default:0.0.0.0:3000"`
			DebugHost       string        `conf:"default:0.0.0.0:4000"`
			ReadTimeout     time.Duration `conf:"default:5s"`
			WriteTimeout    time.Duration `conf:"default:5s"`
			ShutdownTimeout time.Duration `conf:"default:5s"`
		}
		Auth struct {
			KeyID          string `conf:"default:54bb2165-71e1-41a6-af3e-7da4a0e1e2c1"`
			PrivateKeyFile string `conf:"default:./private.pem"`
			Algorithm      string `conf:"default:RS256"`
		}
		Postgres struct {
			User       string `conf:"default:postgres"`
			Password   string `conf:"default:1"`
			Host       string `conf:"default:0.0.0.0"`
			Name       string `conf:"default:project"`
			DisableTLS bool   `conf:"default:true"`
		}
		Zipkin struct {
			ReporterURI string  `conf:"default:http://zipkin:9411/api/v2/spans"`
			ServiceName string  `conf:"default:sale-api"`
			Probability float64 `conf:"default:0.05"`
		}
		Redis struct {
			Host string `conf:"default:localhost"`
			Port string `conf:"default:6379"`
			DB   int    `conf:"default:0"`
		}
	}
	cfg.Version.SVN = build
	cfg.Version.Desc = "copyright information here"

	if err := conf.Parse(os.Args[1:], "krember", &cfg); err != nil {
		switch err {
		case conf.ErrHelpWanted:
			usage, err := conf.Usage("krember", &cfg)
			if err != nil {
				return errors.Wrap(err, "generating config usage")
			}
			fmt.Println(usage)
			return nil
		case conf.ErrVersionWanted:
			version, err := conf.VersionString("krember", &cfg)
			if err != nil {
				return errors.Wrap(err, "generating config version")
			}
			fmt.Println(version)
			return nil
		}
		return errors.Wrap(err, "parsing config")
	}

	// =========================================================================
	// App Starting

	// Print the build version for our logs. Also expose it under /debug/vars.
	expvar.NewString("build").Set(build)
	log.Printf("main : Started : Application initializing : version %q", build)
	defer log.Println("main: Completed")

	out, err := conf.String(&cfg)
	if err != nil {
		return errors.Wrap(err, "generating config for output")
	}
	log.Printf("main: Config :\n%v\n", out)

	// =========================================================================
	// Initialize authentication support

	log.Println("main : Started : Initializing authentication support")

	privatePEM, err := os.ReadFile(cfg.Auth.PrivateKeyFile)
	if err != nil {
		return errors.Wrap(err, "reading auth private key")
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privatePEM)
	if err != nil {
		return errors.Wrap(err, "parsing auth private key")
	}

	lookup := func(kid string) (*rsa.PublicKey, error) {
		switch kid {
		case cfg.Auth.KeyID:
			return &privateKey.PublicKey, nil
		}
		return nil, fmt.Errorf("no public key found for the specified kid: %s", kid)
	}

	auth, err := auth.New(cfg.Auth.Algorithm, lookup, auth.Keys{cfg.Auth.KeyID: privateKey})
	if err != nil {
		return errors.Wrap(err, "constructing auth")
	}

	// =========================================================================
	// Start Database: postgresql

	log.Println("main: Initializing database support")

	postgresDB := postgresql.NewDB(postgresql.Config{
		User:          cfg.Postgres.User,
		Password:      cfg.Postgres.Password,
		Host:          cfg.Postgres.Host,
		Name:          cfg.Postgres.Name,
		DisableTLS:    cfg.Postgres.DisableTLS,
		ServerBaseUrl: cfg.ServerBaseUrl,
		DefaultLang:   cfg.DefaultLang,
	})
	if err != nil {
		return errors.Wrap(err, "connecting to db")
	}
	defer func() {
		log.Printf("main: Database Stopping : %s", cfg.Postgres.Host)
		postgresDB.Close()
	}()

	// =====================

	// =========================================================================
	// Start Cache: redis

	log.Println("main: Initializing cache support")

	redisDB := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: "",
		DB:       cfg.Redis.DB,
	})

	// ======================

	shutdown := make(chan os.Signal, 1)

	// gin engine
	webApp := web.NewApp(shutdown, cfg.DefaultLang)

	// migrations
	commands.MigrateUP(postgresDB)
	//commands.Migrate(postgresDB)

	r := router.NewRouter(webApp, postgresDB, redisDB, fmt.Sprintf(":%s", cfg.ServerPort), auth, cfg.ServerBaseUrl)

	return r.Init()
}
