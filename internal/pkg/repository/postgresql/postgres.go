package postgresql

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
	"log"
	"net/http"
	"project/foundation/web"
	"project/internal/auth"
	"reflect"
	"strings"
	"time"
)

type CurrencyValue struct {
	ID        string   `json:"id"`
	Value     *float32 `json:"value"`
	PriceDate *string  `json:"price_date"`
	Currency  *string  `json:"currency"`
	Icon      *string  `json:"icon"`
}

// Config is the required properties to use the database.
type Config struct {
	User          string
	Password      string
	Host          string
	Name          string
	DisableTLS    bool
	ServerBaseUrl string
	DefaultLang   string
}

type Database struct {
	*bun.DB
	DBName        string
	DBPassword    string
	DBUser        string
	ServerBaseUrl string
	DefaultLang   string
}

func NewDB(cfg Config) *Database {
	dsn := fmt.Sprintf("postgres://%v:%v@localhost:5432/%v?sslmode=disable", cfg.User, cfg.Password, cfg.Name)

	sqlDB := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))

	db := bun.NewDB(sqlDB, pgdialect.New())

	db.AddQueryHook(bundebug.NewQueryHook(
		bundebug.WithVerbose(true),
		bundebug.FromEnv("BUNDEBUG"),
	))

	return &Database{DB: db, DBName: cfg.Name, DBPassword: cfg.Password, DBUser: cfg.User, ServerBaseUrl: cfg.ServerBaseUrl, DefaultLang: cfg.DefaultLang}
}

func (d Database) DeleteRow(ctx context.Context, table string, id int) error {
	claims, err := d.CheckClaims(ctx)
	if err != nil {
		return err
	}

	q := d.NewUpdate().
		Table(table).
		Where("id = ?", id).
		Set("deleted_at = ?", time.Now()).
		Set("deleted_by = ?", claims.UserId)

	_, err = q.Exec(ctx)
	if err != nil {
		return web.NewRequestError(errors.Wrapf(err, "deleting %s", table), http.StatusBadRequest)
	}

	return nil
}

func (d Database) CheckClaims(ctx context.Context, role ...string) (auth.Claims, error) {
	claims, ok := ctx.Value(auth.Key).(auth.Claims)
	if !ok {
		return auth.Claims{}, web.NewRequestError(errors.New("claims missing from context"), http.StatusBadRequest)
	}

	for _, r := range role {
		if strings.Compare(r, claims.Role) == 0 {
			return claims, nil
		}
	}

	if len(role) == 0 {
		return claims, nil
	}

	return auth.Claims{}, errors.New("no permission")
}

func (d Database) GetLang(ctx context.Context) string {
	if value, ok := ctx.Value("lang").(string); ok {
		return value
	}

	return d.DefaultLang
}

func (d Database) ValidateStruct(s interface{}, requiredFields ...string) error {
	structVal := reflect.Value{}
	if reflect.Indirect(reflect.ValueOf(s)).Kind() == reflect.Struct {
		structVal = reflect.Indirect(reflect.ValueOf(s))
	} else {
		return errors.New("input param should be a struct")
	}

	errFields := make([]web.FieldError, 0)

	structType := reflect.Indirect(reflect.ValueOf(s)).Type()
	fieldNum := structVal.NumField()

	for i := 0; i < fieldNum; i++ {
		field := structVal.Field(i)
		fieldName := structType.Field(i).Name

		isSet := field.IsValid() && !field.IsZero()
		if !isSet {
			log.Print(isSet, fieldName, reflect.ValueOf(field))
			for _, f := range requiredFields {
				if f == fieldName {
					errFields = append(errFields, web.FieldError{
						Error: "field is required!",
						Field: fieldName,
					})
				}
			}
		}
	}

	if len(errFields) > 0 {
		return &web.Error{
			Err:    errors.New("required fields"),
			Fields: errFields,
			Status: http.StatusBadRequest,
		}
	}

	return nil
}
