package user

import (
	"github.com/pkg/errors"
	"net/http"
	"project/foundation/web"
	"project/internal/commands"
	"project/internal/repository/postgres/user"
	"reflect"
)

type Controller struct {
	user User
}

func NewController(user User) *Controller {
	return &Controller{user}
}

// user

func (uc Controller) GetUserList(c *web.Context) error {
	var filter user.Filter

	if limit, ok := c.GetQueryFunc(reflect.Int, "limit").(*int); ok {
		filter.Limit = limit
	}
	if offset, ok := c.GetQueryFunc(reflect.Int, "offset").(*int); ok {
		filter.Offset = offset
	}
	if page, ok := c.GetQueryFunc(reflect.Int, "page").(*int); ok {
		filter.Page = page
	}
	if search, ok := c.GetQueryFunc(reflect.String, "search").(*string); ok {
		filter.Search = search
	}
	if role, ok := c.GetQueryFunc(reflect.String, "role").(*string); ok {
		filter.Role = role
	}
	if err := c.ValidQuery(); err != nil {
		return c.RespondError(err)
	}

	list, count, err := uc.user.GetList(c.Ctx, filter)
	if err != nil {
		return c.RespondError(err)
	}

	return c.Respond(map[string]interface{}{
		"data": map[string]interface{}{
			"results": list,
			"count":   count,
		},
		"status": true,
	}, http.StatusOK)
}

func (uc Controller) GetUserDetailById(c *web.Context) error {
	id := c.GetParam(reflect.Int, "id").(int)

	if err := c.ValidParam(); err != nil {
		return c.RespondError(err)
	}

	response, err := uc.user.GetDetailById(c.Ctx, id)
	if err != nil {
		return c.RespondError(err)
	}

	return c.Respond(map[string]interface{}{
		"data":   response,
		"status": true,
	}, http.StatusOK)
}

func (uc Controller) CreateUser(c *web.Context) error {
	var request user.CreateRequest

	if err := c.BindFunc(&request, "Username", "Password", "Role"); err != nil {
		return c.RespondError(err)
	}

	if request.Avatar != nil {
		if ok := commands.CheckFileType(c.Ctx, request.Avatar, "image"); !ok {
			return c.RespondError(web.NewRequestError(errors.New("avatar must be image"), http.StatusBadRequest))
		}
		fileUrl, _, _, err := commands.Upload(c.Ctx, request.Avatar, "users/avatar", commands.AvatarSize)
		if err != nil {
			return c.RespondError(web.NewRequestError(errors.Wrap(err, "upload avatar"), http.StatusBadRequest))
		}
		request.AvatarLink = &fileUrl
	}

	response, err := uc.user.Create(c.Ctx, request)
	if err != nil {
		return c.RespondError(err)
	}

	return c.Respond(map[string]interface{}{
		"data":   response,
		"status": true,
	}, http.StatusOK)
}

func (uc Controller) UpdateUserAll(c *web.Context) error {
	id := c.GetParam(reflect.Int, "id").(int)

	if err := c.ValidParam(); err != nil {
		return c.RespondError(err)
	}

	var request user.UpdateRequest

	if err := c.BindFunc(&request, "Username", "Phone", "FirstName", "Surname", "BirthDate"); err != nil {
		return c.RespondError(err)
	}

	if request.Avatar != nil {
		if ok := commands.CheckFileType(c.Ctx, request.Avatar, "image"); !ok {
			return c.RespondError(web.NewRequestError(errors.New("avatar must be image"), http.StatusBadRequest))
		}
		fileUrl, _, _, err := commands.Upload(c.Ctx, request.Avatar, "users/avatar", commands.AvatarSize)
		if err != nil {
			return c.RespondError(web.NewRequestError(errors.Wrap(err, "upload avatar"), http.StatusBadRequest))
		}
		request.AvatarLink = &fileUrl
	}

	request.ID = id

	err := uc.user.UpdateAll(c.Ctx, request)
	if err != nil {
		return c.RespondError(err)
	}

	return c.Respond(map[string]interface{}{
		"data":   "ok!",
		"status": true,
	}, http.StatusOK)
}

func (uc Controller) UpdateUserColumns(c *web.Context) error {
	id := c.GetParam(reflect.Int, "id").(int)

	if err := c.ValidParam(); err != nil {
		return c.RespondError(err)
	}

	var request user.UpdateRequest

	if err := c.BindFunc(&request); err != nil {
		return c.RespondError(err)
	}

	if request.Avatar != nil {
		if ok := commands.CheckFileType(c.Ctx, request.Avatar, "image"); !ok {
			return c.RespondError(web.NewRequestError(errors.New("avatar must be image"), http.StatusBadRequest))
		}
		fileUrl, _, _, err := commands.Upload(c.Ctx, request.Avatar, "users/avatar", commands.AvatarSize)
		if err != nil {
			return c.RespondError(web.NewRequestError(errors.Wrap(err, "upload avatar"), http.StatusBadRequest))
		}
		request.AvatarLink = &fileUrl
	}

	request.ID = id

	err := uc.user.UpdateColumns(c.Ctx, request)
	if err != nil {
		return c.RespondError(err)
	}

	return c.Respond(map[string]interface{}{
		"data":   "ok!",
		"status": true,
	}, http.StatusOK)
}

func (uc Controller) DeleteUser(c *web.Context) error {
	id := c.GetParam(reflect.Int, "id").(int)

	if err := c.ValidParam(); err != nil {
		return c.RespondError(err)
	}

	err := uc.user.Delete(c.Ctx, id)
	if err != nil {
		return c.RespondError(err)
	}

	return c.Respond(map[string]interface{}{
		"data":   "ok!",
		"status": true,
	}, http.StatusOK)
}
