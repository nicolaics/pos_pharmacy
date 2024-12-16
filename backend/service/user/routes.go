package user

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/nicolaics/pharmacon/logger"
	"github.com/nicolaics/pharmacon/service/auth"
	"github.com/nicolaics/pharmacon/types"
	"github.com/nicolaics/pharmacon/utils"
)

type Handler struct {
	store types.UserStore
}

func NewHandler(store types.UserStore) *Handler {
	return &Handler{store: store}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/user/register", h.handleRegister).Methods(http.MethodPost)
	router.HandleFunc("/user/register", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)

	router.HandleFunc("/user/{params}/{val}", h.handleGetAll).Methods(http.MethodGet)
	router.HandleFunc("/user/{params}/{val}", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)

	router.HandleFunc("/user/current", h.handleGetCurrentUser).Methods(http.MethodGet)
	router.HandleFunc("/user/current", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)

	router.HandleFunc("/user/detail", h.handleGetUserDetail).Methods(http.MethodPost)
	router.HandleFunc("/user/detail", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)

	router.HandleFunc("/user", h.handleDelete).Methods(http.MethodDelete)
	router.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)

	router.HandleFunc("/user/modify", h.handleModify).Methods(http.MethodPatch)
	router.HandleFunc("/user/modify", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)

	router.HandleFunc("/user/logout", h.handleLogout).Methods(http.MethodGet)
	router.HandleFunc("/user/logout", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)

	router.HandleFunc("/user/admin", h.handleChangeAdminStatus).Methods(http.MethodPatch)
	router.HandleFunc("/user/admin", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
}

func (h *Handler) RegisterUnprotectedRoutes(router *mux.Router) {
	router.HandleFunc("/user/login", h.handleLogin).Methods(http.MethodPost)
	router.HandleFunc("/user/login", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.LoginUserPayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		logFile, _ := logger.WriteServerErrorLog("login", 0, nil, fmt.Errorf("parsing payload failed: %v", err))
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: "Failed parsing payload",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		logFile, _ := logger.WriteServerErrorLog("login", 0, nil, fmt.Errorf("invalid payload: %v", err))
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid payload",
			Log:     logFile,
			Error:   errors.Error(),
		}
		resp.WriteError(w)
		return
	}

	user, err := h.store.GetUserByName(payload.Name)
	if err != nil {
		data := map[string]interface{}{"name": payload.Name}
		logFile, _ := logger.WriteServerErrorLog("login", 0, data, fmt.Errorf("error get user by name: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}
	if user == nil {
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("User %s not found", payload.Name),
		}
		resp.WriteError(w)
		return
	}

	// check password match
	if !(auth.ComparePassword(user.Password, []byte(payload.Password))) {
		resp := utils.Response{
			Code:    http.StatusUnauthorized,
			Message: "Invalid password",
		}
		resp.WriteError(w)
		return
	}

	tokenDetails, err := auth.CreateJWT(user.ID, user.Admin)
	if err != nil {
		data := map[string]interface{}{"name": payload.Name}
		logFile, _ := logger.WriteServerErrorLog("login", user.ID, data, fmt.Errorf("error create jwt: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	err = h.store.SaveToken(user.ID, tokenDetails)
	if err != nil {
		data := map[string]interface{}{"name": payload.Name}
		logFile, _ := logger.WriteServerErrorLog("login", user.ID, data, fmt.Errorf("error save token: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	err = h.store.UpdateLastLoggedIn(user.ID)
	if err != nil {
		data := map[string]interface{}{"name": payload.Name}
		logFile, _ := logger.WriteServerErrorLog("login", user.ID, data, fmt.Errorf("error update last logged in: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	tokens := map[string]string{
		"token": tokenDetails.Token,
	}

	resp := utils.Response{
		Code:   http.StatusOK,
		Result: tokens,
	}
	resp.WriteSuccess(w)
}

func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.RegisterUserPayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		logFile, _ := logger.WriteServerErrorLog("register user", 0, nil, fmt.Errorf("parsing payload failed: %v", err))
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: "Failed parsing payload",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		logFile, _ := logger.WriteServerErrorLog("register user", 0, nil, fmt.Errorf("invalid payload: %v", err))
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid payload",
			Log:     logFile,
			Error:   errors.Error(),
		}
		resp.WriteError(w)
		return
	}

	// validate token
	admin, err := h.store.ValidateUserToken(w, r, true)
	if err != nil {
		data := map[string]interface{}{"name": payload.Name}
		logFile, _ := logger.WriteServerErrorLog("register user", 0, data, fmt.Errorf("user token invalid: %v", err))
		resp := utils.Response{
			Code:    http.StatusUnauthorized,
			Message: "User token invalid or not admin!",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	// validate admin password
	if !(auth.ComparePassword(admin.Password, []byte(payload.AdminPassword))) {
		resp := utils.Response{
			Code:    http.StatusUnauthorized,
			Message: "Invalid admin password",
		}
		resp.WriteError(w)
		return
	}

	// check if the newly created user exists
	temp, err := h.store.GetUserByName(payload.Name)
	if err != nil {
		data := map[string]interface{}{"name": payload.Name}
		logFile, _ := logger.WriteServerErrorLog("register user", admin.ID, data, fmt.Errorf("error get user by name: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}
	if temp != nil {
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("User %s already exists", payload.Name),
		}
		resp.WriteError(w)
		return
	}

	// if it doesn't, we create new user
	hashedPassword, err := auth.HashPassword(payload.Password)
	if err != nil {
		data := map[string]interface{}{"name": payload.Name}
		logFile, _ := logger.WriteServerErrorLog("register user", admin.ID, data, fmt.Errorf("error hash password: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	err = h.store.CreateUser(types.User{
		Name:        payload.Name,
		Password:    hashedPassword,
		PhoneNumber: payload.PhoneNumber,
		Admin:       payload.Admin,
	})
	if err != nil {
		data := map[string]interface{}{"name": payload.Name}
		logFile, _ := logger.WriteServerErrorLog("register user", admin.ID, data, fmt.Errorf("error create user: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
	}

	resp := utils.Response{
		Code:    http.StatusCreated,
		Message: fmt.Sprintf("User %s successfully created", payload.Name),
	}
	resp.WriteSuccess(w)
}

func (h *Handler) handleGetAll(w http.ResponseWriter, r *http.Request) {
	// validate token
	user, err := h.store.ValidateUserToken(w, r, true)
	if err != nil {
		logFile, _ := logger.WriteServerErrorLog("get all users", 0, nil, fmt.Errorf("user token invalid: %v", err))
		resp := utils.Response{
			Code:    http.StatusUnauthorized,
			Message: "User token invalid!\nPlease login again!",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	vars := mux.Vars(r)
	params := vars["params"]
	val := vars["val"]

	var users []types.User

	if val == "all" {
		users, err = h.store.GetAllUsers()
		if err != nil {
			logFile, _ := logger.WriteServerErrorLog("get all users", user.ID, nil, fmt.Errorf("error get all users: %v", err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
			return
		}
	} else if params == "name" {
		data := map[string]interface{}{"searched_name": val}
		users, err = h.store.GetUserBySearchName(val)
		if err != nil {
			logFile, _ := logger.WriteServerErrorLog("get all users", user.ID, data, fmt.Errorf("error get user by search name: %v", err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
			return
		}
	} else if params == "id" {
		id, err := strconv.Atoi(val)
		if err != nil {
			data := map[string]interface{}{"searched_id": val}
			logFile, _ := logger.WriteServerErrorLog("get all users", user.ID, data, fmt.Errorf("error parse id: %v", err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
			return
		}

		user, err := h.store.GetUserByID(id)
		if err != nil {
			data := map[string]interface{}{"searched_id": id}
			logFile, _ := logger.WriteServerErrorLog("get all users", user.ID, data, fmt.Errorf("error get user by id: %v", err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
			return
		}
		if user == nil {
			resp := utils.Response{
				Code:    http.StatusBadRequest,
				Message: fmt.Sprintf("User ID %d doesn't exist", id),
			}
			resp.WriteError(w)
			return
		}

		users = append(users, *user)
	} else if params == "phone-number" {
		users, err = h.store.GetUserBySearchPhoneNumber(val)
		if err != nil {
			data := map[string]interface{}{"searched_phone_number": val}
			logFile, _ := logger.WriteServerErrorLog("get all users", user.ID, data, fmt.Errorf("error get user by search phone number: %v", err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
			return
		}
	} else {
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: "Unknown query",
		}
		resp.WriteError(w)
		return
	}

	resp := utils.Response{
		Code:   http.StatusOK,
		Result: users,
	}
	resp.WriteSuccess(w)
}

func (h *Handler) handleGetCurrentUser(w http.ResponseWriter, r *http.Request) {
	// validate token
	user, err := h.store.ValidateUserToken(w, r, true)
	if err != nil {
		logFile, _ := logger.WriteServerErrorLog("get current user", 0, nil, fmt.Errorf("user token invalid: %v", err))
		resp := utils.Response{
			Code:    http.StatusUnauthorized,
			Message: "User token invalid!\nPlease login again!",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	resp := utils.Response{
		Code:   http.StatusOK,
		Result: user,
	}
	resp.WriteSuccess(w)
}

// get one user other than the current user
func (h *Handler) handleGetUserDetail(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.GetUserDetailPayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		logFile, _ := logger.WriteServerErrorLog("get user detail", 0, nil, fmt.Errorf("parsing payload failed: %v", err))
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: "Failed parsing payload",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		logFile, _ := logger.WriteServerErrorLog("get user detail", 0, nil, fmt.Errorf("invalid payload: %v", err))
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid payload",
			Log:     logFile,
			Error:   errors.Error(),
		}
		resp.WriteError(w)
		return
	}

	// validate token
	user, err := h.store.ValidateUserToken(w, r, true)
	if err != nil {
		data := map[string]interface{}{"id": payload.ID}
		logFile, _ := logger.WriteServerErrorLog("get user detail", 0, data, fmt.Errorf("user token invalid: %v", err))
		resp := utils.Response{
			Code:    http.StatusUnauthorized,
			Message: "User token invalid!\nPlease login again!",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	// check if the newly created user exists
	requested, err := h.store.GetUserByID(payload.ID)
	if err != nil {
		data := map[string]interface{}{"id": payload.ID}
		logFile, _ := logger.WriteServerErrorLog("get user detail", user.ID, data, fmt.Errorf("error get user by id: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}
	if requested == nil {
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("User ID %d doesn't exist", payload.ID),
		}
		resp.WriteError(w)
		return
	}

	resp := utils.Response{
		Code:   http.StatusOK,
		Result: requested,
	}
	resp.WriteSuccess(w)
}

func (h *Handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	var payload types.RemoveUserPayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		logFile, _ := logger.WriteServerErrorLog("delete user", 0, nil, fmt.Errorf("parsing payload failed: %v", err))
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: "Failed parsing payload",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		logFile, _ := logger.WriteServerErrorLog("delete user", 0, nil, fmt.Errorf("invalid payload: %v", err))
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid payload",
			Log:     logFile,
			Error:   errors.Error(),
		}
		resp.WriteError(w)
		return
	}

	// validate token
	admin, err := h.store.ValidateUserToken(w, r, true)
	if err != nil {
		data := map[string]interface{}{"id": payload.ID}
		logFile, _ := logger.WriteServerErrorLog("delete user", 0, data, fmt.Errorf("user token invalid: %v", err))
		resp := utils.Response{
			Code:    http.StatusUnauthorized,
			Message: "User token invalid or not admin",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	// validate admin password
	if !(auth.ComparePassword(admin.Password, []byte(payload.AdminPassword))) {
		resp := utils.Response{
			Code:    http.StatusUnauthorized,
			Message: "Invalid admin password",
		}
		resp.WriteError(w)
		return
	}

	// users, err := h.store.GetAllUsers()
	// if err != nil {
	// 	logFile, _ := logger.WriteServerErrorLog("get delete user", 0, nil, fmt.Errorf("parsing payload failed: %v", err))
	// 	resp := utils.Response{
	// 		Code:    http.StatusBadRequest,
	// 		Message: "Failed parsing payload",
	// 		Log:     logFile,
	// 		Error:   err.Error(),
	// 	}
	// 	resp.WriteError(w)
	// 	utils.WriteError(w, http.StatusInternalServerError, err, nil)
	// 	return
	// }

	// if len(users) == 1 || payload.ID == 1 {
	// 	utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("cannot delete initial admin"), nil)
	// 	return
	// }

	err = h.store.DeleteUser(payload.ID, admin)
	if err != nil {
		data := map[string]interface{}{"id": payload.ID}
		logFile, _ := logger.WriteServerErrorLog("delete user", admin.ID, data, fmt.Errorf("error delete: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	resp := utils.Response{
		Code:    http.StatusOK,
		Message: fmt.Sprintf("User ID %d successfully deleted", payload.ID),
	}
	resp.WriteSuccess(w)
}

func (h *Handler) handleModify(w http.ResponseWriter, r *http.Request) {
	var payload types.ModifyUserPayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		logFile, _ := logger.WriteServerErrorLog("modify user", 0, nil, fmt.Errorf("parsing payload failed: %v", err))
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: "Failed parsing payload",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		logFile, _ := logger.WriteServerErrorLog("modify user", 0, nil, fmt.Errorf("invalid payload: %v", err))
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid payload",
			Log:     logFile,
			Error:   errors.Error(),
		}
		resp.WriteError(w)
		return
	}

	// validate token
	admin, err := h.store.ValidateUserToken(w, r, true)
	if err != nil {
		data := map[string]interface{}{"id": payload.ID}
		logFile, _ := logger.WriteServerErrorLog("modify user", 0, data, fmt.Errorf("user token invalid: %v", err))
		resp := utils.Response{
			Code:    http.StatusUnauthorized,
			Message: "User token invalid or not admin",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	// validate admin password
	if !(auth.ComparePassword(admin.Password, []byte(payload.NewData.AdminPassword))) {
		resp := utils.Response{
			Code:    http.StatusUnauthorized,
			Message: "Invalid admin password",
		}
		resp.WriteError(w)
		return
	}

	user, err := h.store.GetUserByID(payload.ID)
	if err != nil {
		data := map[string]interface{}{"id": payload.ID}
		logFile, _ := logger.WriteServerErrorLog("modify user", admin.ID, data, fmt.Errorf("error get user by id: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}
	if user == nil {
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("User ID %d doesn't exist", payload.ID),
		}
		resp.WriteError(w)
		return
	}

	if user.Name != payload.NewData.Name {
		temp, err := h.store.GetUserByName(payload.NewData.Name)
		if err != nil {
			data := map[string]interface{}{"id": payload.ID}
			logFile, _ := logger.WriteServerErrorLog("modify user", admin.ID, data, fmt.Errorf("error get user by name: %v", err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
			return
		}
		if temp != nil {
			resp := utils.Response{
				Code:    http.StatusBadRequest,
				Message: fmt.Sprintf("User %s already exists", payload.NewData.Name),
			}
			resp.WriteError(w)
			return
		}
	}

	err = h.store.ModifyUser(user.ID, types.User{
		Name:        payload.NewData.Name,
		Password:    payload.NewData.Password,
		Admin:       payload.NewData.Admin,
		PhoneNumber: payload.NewData.PhoneNumber,
	}, admin)
	if err != nil {
		data := map[string]interface{}{"id": payload.ID}
		logFile, _ := logger.WriteServerErrorLog("modify user", admin.ID, data, fmt.Errorf("error modify: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	resp := utils.Response{
		Code:    http.StatusOK,
		Message: fmt.Sprintf("User ID %d updated", payload.ID),
	}
	resp.WriteSuccess(w)
}

func (h *Handler) handleLogout(w http.ResponseWriter, r *http.Request) {
	accessDetails, err := auth.ExtractTokenFromClient(r)
	if err != nil {
		logFile, _ := logger.WriteServerErrorLog("logout", 0, nil, fmt.Errorf("error extract token: %v", err))
		resp := utils.Response{
			Code:    http.StatusUnauthorized,
			Message: "Invalid token",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	// check user exists or not
	user, err := h.store.GetUserByID(accessDetails.UserID)
	if err != nil {
		logFile, _ := logger.WriteServerErrorLog("logout", 0, nil, fmt.Errorf("error get user by id: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}
	if user == nil {
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("User ID %d doesn't exist", accessDetails.UserID),
		}
		resp.WriteError(w)
		return
	}

	err = h.store.UpdateLastLoggedIn(accessDetails.UserID)
	if err != nil {
		logFile, _ := logger.WriteServerErrorLog("logout", user.ID, nil, fmt.Errorf("error update last logged in: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	err = h.store.DeleteToken(accessDetails.UserID)
	if err != nil {
		logFile, _ := logger.WriteServerErrorLog("logout", user.ID, nil, fmt.Errorf("error delete token: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	resp := utils.Response{
		Code:    http.StatusOK,
		Message: "Successfully logged out",
	}
	resp.WriteSuccess(w)
}

func (h *Handler) handleChangeAdminStatus(w http.ResponseWriter, r *http.Request) {
	var payload types.ChangeAdminStatusPayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		logFile, _ := logger.WriteServerErrorLog("change admin status", 0, nil, fmt.Errorf("parsing payload failed: %v", err))
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: "Failed parsing payload",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		logFile, _ := logger.WriteServerErrorLog("change admin status", 0, nil, fmt.Errorf("invalid payload: %v", err))
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid payload",
			Log:     logFile,
			Error:   errors.Error(),
		}
		resp.WriteError(w)
		return
	}

	// validate token
	admin, err := h.store.ValidateUserToken(w, r, true)
	if err != nil {
		data := map[string]interface{}{"id": payload.ID, "new_admin_value": payload.Admin}
		logFile, _ := logger.WriteServerErrorLog("change admin status", 0, data, fmt.Errorf("user token invalid: %v", err))
		resp := utils.Response{
			Code:    http.StatusUnauthorized,
			Message: "User token invalid or not admin",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	// validate admin password
	if !(auth.ComparePassword(admin.Password, []byte(payload.AdminPassword))) {
		resp := utils.Response{
			Code:    http.StatusUnauthorized,
			Message: "Invalid admin password",
		}
		resp.WriteError(w)
		return
	}

	// check whether user exists or not
	user, err := h.store.GetUserByID(payload.ID)
	if err != nil {
		data := map[string]interface{}{"id": payload.ID, "new_admin_value": payload.Admin}
		logFile, _ := logger.WriteServerErrorLog("change admin status", admin.ID, data, fmt.Errorf("error get user by id: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}
	if user == nil {
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("User ID %d doesn't exist", payload.ID),
		}
		resp.WriteError(w)
		return
	}

	err = h.store.ModifyUser(user.ID, types.User{
		Name:        user.Name,
		Password:    user.Password,
		Admin:       payload.Admin,
		PhoneNumber: user.PhoneNumber,
	}, admin)
	if err != nil {
		data := map[string]interface{}{"id": payload.ID, "new_admin_value": payload.Admin}
		logFile, _ := logger.WriteServerErrorLog("change admin status", admin.ID, data, fmt.Errorf("error modify user: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	resp := utils.Response{
		Code:    http.StatusOK,
		Message: fmt.Sprintf("User ID %d's admin status is now %t", payload.ID, payload.Admin),
	}
	resp.WriteSuccess(w)
}
