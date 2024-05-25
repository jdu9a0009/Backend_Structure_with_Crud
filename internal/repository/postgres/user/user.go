package user

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"math/rand"
	"net/http"
	"project/foundation/web"
	"project/internal/auth"
	"project/internal/entity"
	"project/internal/pkg/repository/postgresql"
	"project/internal/repository/postgres"
	"project/internal/service/hashing"
	"strings"
	"time"
)

type Repository struct {
	*postgresql.Database
}

func NewRepository(database *postgresql.Database) *Repository {
	return &Repository{Database: database}
}

func (r Repository) GetByLogin(ctx context.Context, login string) (entity.User, error) {
	var detail entity.User

	err := r.NewSelect().Model(&detail).Where("login = ? AND deleted_at IS NULL", login).Scan(ctx)
	if err != nil {
		return entity.User{}, &web.Error{
			Err:    errors.New("user not found!"),
			Status: http.StatusUnauthorized,
		}
	}

	return detail, err
}

func (r Repository) GetById(ctx context.Context, id int) (entity.User, error) {
	var detail entity.User

	err := r.NewSelect().Model(&detail).Where("id = ?", id).Scan(ctx)

	return detail, err
}

func (r Repository) Create(ctx context.Context, request CreateRequest) (CreateResponse, error) {
	claims, err := r.CheckClaims(ctx, auth.RoleAdmin)
	if err != nil {
		return CreateResponse{}, err
	}

	if err := r.ValidateStruct(&request, "Login", "Password", "Role"); err != nil {
		return CreateResponse{}, err
	}
	rand.Seed(time.Now().UnixNano())

	loginStatus := true
	if err := r.QueryRowContext(ctx,
		fmt.Sprintf(`SELECT 
    						CASE WHEN 
    						(SELECT id FROM users WHERE login = '%s' AND deleted_at IS NULL) IS NOT NULL 
    						THEN true ELSE false END`, *request.Login)).Scan(&loginStatus); err != nil {
		return CreateResponse{}, web.NewRequestError(errors.Wrap(err, "login check"), http.StatusInternalServerError)
	}
	if loginStatus {
		return CreateResponse{}, web.NewRequestError(errors.Wrap(errors.New(""), "login is used"), http.StatusInternalServerError)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(*request.Password), bcrypt.DefaultCost)
	if err != nil {
		return CreateResponse{}, web.NewRequestError(errors.Wrap(err, "hashing password"), http.StatusInternalServerError)
	}
	hashedPassword := string(hash)

	var response CreateResponse
	role := strings.ToUpper(*request.Role)
	if (role != "ADMIN") && (role != "EMPLOYEE") {
		return CreateResponse{}, web.NewRequestError(errors.New("incorrect role. role should be ADMIN or EMPLOYEE"), http.StatusBadRequest)
	}

	response.Role = &role
	response.FullName = request.FullName
	response.Login = request.Login
	response.Avatar = request.AvatarLink
	response.Password = &hashedPassword
	response.Phone = request.Phone
	response.CreatedAt = time.Now()
	response.CreatedBy = claims.UserId

	_, err = r.NewInsert().Model(&response).Returning("id").Exec(ctx, &response.ID)
	if err != nil {
		return CreateResponse{}, web.NewRequestError(errors.Wrap(err, "creating user"), http.StatusBadRequest)
	}

	if response.Avatar != nil {
		link := r.ServerBaseUrl + hashing.GenerateHash(*response.Avatar)
		response.Avatar = &link
	}

	return response, nil
}

func (r Repository) GetList(ctx context.Context, filter Filter) ([]GetListResponse, int, error) {
	_, err := r.CheckClaims(ctx, auth.RoleAdmin)
	if err != nil {
		return nil, 0, err
	}

	whereQuery := fmt.Sprintf(`
			WHERE 
				deleted_at IS NULL
			`)

	if filter.Role != nil {
		role := strings.ToUpper(*filter.Role)
		whereQuery += fmt.Sprintf(` AND role = '%s' `, role)
	}

	if filter.Search != nil {
		search := strings.Replace(*filter.Search, " ", "", -1)
		search = strings.Replace(search, "'", "", -1)

		whereQuery += fmt.Sprintf(` AND full_name ilike '%s'`, "%"+search+"%")
	}
	orderQuery := "ORDER BY created_at desc"

	var limitQuery, offsetQuery string

	if filter.Page != nil && filter.Limit != nil {
		offset := (*filter.Page - 1) * (*filter.Limit)
		filter.Offset = &offset
	}

	if filter.Limit != nil {
		limitQuery += fmt.Sprintf(" LIMIT %d", *filter.Limit)
	}

	if filter.Offset != nil {
		offsetQuery += fmt.Sprintf(" OFFSET %d", *filter.Offset)
	}

	query := fmt.Sprintf(`
		SELECT 
			id,
			avatar,
			full_name,
			login,
			status,
			phone,
			role
		FROM users
		%s %s %s %s
	`, whereQuery, orderQuery, limitQuery, offsetQuery)

	rows, err := r.QueryContext(ctx, query)
	if err == sql.ErrNoRows {
		return nil, 0, web.NewRequestError(postgres.ErrNotFound, http.StatusBadRequest)
	}
	if err != nil {
		return nil, 0, web.NewRequestError(errors.Wrap(err, "selecting users"), http.StatusBadRequest)
	}

	var list []GetListResponse

	for rows.Next() {
		var detail GetListResponse
		if err = rows.Scan(
			&detail.ID,
			&detail.Avatar,
			&detail.FullName,
			&detail.Login,
			&detail.Status,
			&detail.Phone,
			&detail.Role); err != nil {
			return nil, 0, web.NewRequestError(errors.Wrap(err, "scanning user list"), http.StatusBadRequest)
		}
		if detail.Avatar != nil {
			link := r.ServerBaseUrl + hashing.GenerateHash(*detail.Avatar)
			detail.Avatar = &link
		}

		list = append(list, detail)
	}

	countQuery := fmt.Sprintf(`
		SELECT
			count(id)
		FROM  users
			%s
	`, whereQuery)

	countRows, err := r.QueryContext(ctx, countQuery)
	if err == sql.ErrNoRows {
		return nil, 0, web.NewRequestError(postgres.ErrNotFound, http.StatusNotFound)
	}
	if err != nil {
		return nil, 0, web.NewRequestError(errors.Wrap(err, "selecting users"), http.StatusBadRequest)
	}

	count := 0

	for countRows.Next() {
		if err = countRows.Scan(&count); err != nil {
			return nil, 0, web.NewRequestError(errors.Wrap(err, "scanning user count"), http.StatusBadRequest)
		}
	}

	return list, count, nil
}

func (r Repository) GetDetailById(ctx context.Context, id int) (GetDetailByIdResponse, error) {
	_, err := r.CheckClaims(ctx, auth.RoleAdmin)
	if err != nil {
		return GetDetailByIdResponse{}, err
	}

	query := fmt.Sprintf(`
		SELECT
			id,
			avatar,
			full_name,
			login,
			status,
			phone,
			role
		FROM
		    users
		WHERE deleted_at IS NULL AND id = %d
	`, id)

	var detail GetDetailByIdResponse

	err = r.QueryRowContext(ctx, query).Scan(
		&detail.ID,
		&detail.Avatar,
		&detail.FullName,
		&detail.Login,
		&detail.Status,
		&detail.Phone,
		&detail.Role,
	)

	if err == sql.ErrNoRows {
		return GetDetailByIdResponse{}, web.NewRequestError(postgres.ErrNotFound, http.StatusNotFound)
	}
	if err != nil {
		return GetDetailByIdResponse{}, web.NewRequestError(errors.Wrap(err, "selecting user detail"), http.StatusBadRequest)
	}

	if detail.Avatar != nil {
		link := r.ServerBaseUrl + hashing.GenerateHash(*detail.Avatar)
		detail.Avatar = &link
	}

	return detail, nil
}

func (r Repository) UpdateAll(ctx context.Context, request UpdateRequest) error {
	claims, err := r.CheckClaims(ctx, auth.RoleAdmin)
	if err != nil {
		return err
	}

	if err := r.ValidateStruct(&request, "ID", "Login", "FullName", "Phone", "AvatarLink", "Password", "Role"); err != nil {
		return err
	}
	loginStatus := true
	if err := r.QueryRowContext(ctx, fmt.Sprintf("SELECT CASE WHEN (SELECT id FROM users WHERE login = '%s' AND deleted_at IS NULL AND id != %d) IS NOT NULL THEN true ELSE false END", *request.Login, request.ID)).Scan(&loginStatus); err != nil {
		return web.NewRequestError(errors.Wrap(err, "login check"), http.StatusInternalServerError)
	}
	if loginStatus {
		return web.NewRequestError(errors.Wrap(errors.New(""), "login is used"), http.StatusInternalServerError)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(*request.Password), bcrypt.DefaultCost)
	if err != nil {
		return web.NewRequestError(errors.Wrap(err, "hashing password"), http.StatusInternalServerError)
	}
	hashedPassword := string(hash)

	q := r.NewUpdate().Table("users").Where("deleted_at IS NULL AND id = ?", request.ID)

	role := strings.ToUpper(*request.Role)
	if (role != "ADMIN") && (role != "EMPLOYEE") {
		return web.NewRequestError(errors.New("incorrect role. role should be ADMIN or EMPLOYEE"), http.StatusBadRequest)
	}

	q.Set("role = ?", role)
	q.Set("full_name = ?", request.FullName)
	q.Set("login = ?", request.Login)
	q.Set("phone = ?", request.Phone)
	q.Set("avatar = ?", request.AvatarLink)
	q.Set("updated_at = ?", time.Now())
	q.Set("updated_by = ?", claims.UserId)
	q.Set("password = ?", hashedPassword)

	_, err = q.Exec(ctx)
	if err != nil {
		return web.NewRequestError(errors.Wrap(err, "updating user"), http.StatusBadRequest)
	}

	return nil
}

func (r Repository) UpdateColumns(ctx context.Context, request UpdateRequest) error {
	claims, err := r.CheckClaims(ctx, auth.RoleAdmin)
	if err != nil {
		return err
	}

	if err := r.ValidateStruct(&request, "ID"); err != nil {
		return err
	}

	q := r.NewUpdate().Table("users").Where("deleted_at IS NULL AND id = ? ", request.ID)

	if request.FullName != nil {
		q.Set("full_name = ?", request.FullName)
	}
	if request.Login != nil {
		loginStatus := true
		if err := r.QueryRowContext(ctx, fmt.Sprintf("SELECT CASE WHEN (SELECT id FROM users WHERE login = '%s' AND deleted_at IS NULL AND id != %d) IS NOT NULL THEN true ELSE false END", *request.Login, request.ID)).Scan(&loginStatus); err != nil {
			return web.NewRequestError(errors.Wrap(err, "login check"), http.StatusInternalServerError)
		}
		if loginStatus {
			return web.NewRequestError(errors.Wrap(errors.New(""), "login is used"), http.StatusInternalServerError)
		}
		q.Set("login = ?", request.Login)
	}
	if request.Phone != nil {
		q.Set("phone = ?", request.Phone)
	}
	if request.AvatarLink != nil {
		q.Set("avatar = ?", request.AvatarLink)
	}
	if request.Password != nil {
		hash, err := bcrypt.GenerateFromPassword([]byte(*request.Password), bcrypt.DefaultCost)
		if err != nil {
			return web.NewRequestError(errors.Wrap(err, "hashing password"), http.StatusInternalServerError)
		}
		hashedPassword := string(hash)
		q.Set("password = ?", hashedPassword)
	}
	if request.Role != nil {
		role := strings.ToUpper(*request.Role)
		if (role != "ADMIN") && (role != "EMPLOYEE") {
			return web.NewRequestError(errors.New("incorrect role. role should be ADMIN or EMPLOYEE"), http.StatusBadRequest)
		}

		q.Set("role = ?", role)
	}

	q.Set("updated_at = ?", time.Now())
	q.Set("updated_by = ?", claims.UserId)

	_, err = q.Exec(ctx)
	if err != nil {
		return web.NewRequestError(errors.Wrap(err, "updating user"), http.StatusBadRequest)
	}

	return nil
}

func (r Repository) Delete(ctx context.Context, id int) error {
	return r.DeleteRow(ctx, "users", id)
}

func (r Repository) UploadAvatar(ctx context.Context, request UploadAvatarRequest) error {
	_, err := r.CheckClaims(ctx)
	if err != nil {
		return err
	}

	_, err = r.NewUpdate().
		Table("users").
		Set("avatar = ?", request.AvatarLink).
		Where("deleted_at IS NULL AND id = ?", request.ID).
		Exec(ctx)
	if err != nil {
		return web.NewRequestError(errors.Wrap(err, "updating avatar"), http.StatusBadRequest)
	}

	return nil
}

func (r Repository) GetMe(ctx context.Context) (GetMeResponse, error) {
	claims, err := r.CheckClaims(ctx)
	if err != nil {
		return GetMeResponse{}, err
	}

	query := fmt.Sprintf(`
	SELECT
	    		users.id,
	    		users.avatar,
	    		users.full_name,
	    		users.phone,
	    		users.login
	FROM
	    users
	WHERE
	    users.deleted_at IS NULL AND
	    users.id = %d
	`, claims.UserId)

	var detail GetMeResponse

	err = r.QueryRowContext(ctx, query).Scan(
		&detail.ID,
		&detail.Avatar,
		&detail.FullName,
		&detail.Phone,
		&detail.Login,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return GetMeResponse{}, web.NewRequestError(postgres.ErrNotFound, http.StatusBadRequest)
	}
	if err != nil {
		return GetMeResponse{}, web.NewRequestError(errors.Wrap(err, "selecting user detail"), http.StatusBadRequest)
	}

	if detail.Avatar != nil {
		link := r.ServerBaseUrl + hashing.GenerateHash(*detail.Avatar)
		detail.Avatar = &link
	}

	return detail, nil
}
