package user

import (
	"fmt"
	"net/http"
	"strconv"

	env "github.com/Franzcasttr/ecom/config"
	"github.com/Franzcasttr/ecom/service/auth"
	"github.com/Franzcasttr/ecom/types"
	"github.com/Franzcasttr/ecom/utils"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

type Handler struct{
 store types.UserStore
}

func NewHandler(store types.UserStore) *Handler{
	return &Handler{
		store: store,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router){
	router.HandleFunc("/login", h.HandleLogin).Methods("POST")
	router.HandleFunc("/register", h.HandleRegister).Methods("POST")

	router.HandleFunc("/users/{userID}", auth.WithJWTAuth(h.handleGetUser, h.store)).Methods(http.MethodGet)
}

func (h *Handler) HandleLogin(w http.ResponseWriter, r *http.Request){
	var user types.LoginUserPayload
	if err := utils.ParseJSON(r, &user); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := utils.Validate.Struct(user); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", errors))
		return
	}

	u, err := h.store.GetUserByEmail(user.Email)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("not found, invalid email or password"))
		return
	}

	if !auth.ComparePasswords(u.Password, []byte(user.Password)) {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid email or password"))
		return
	}

	secret := []byte(env.Envs.JWTSecret)
	token, err := auth.CreateJWT(secret, u.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"token": token})
}
func (h *Handler) HandleRegister(w http.ResponseWriter, r *http.Request){
	var payload types.RegisterUserPayload
 	if err := utils.ParseJSON(r, payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
 }

	_, err := h.store.GetUserByEmail(payload.Email)

	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("User already exist"))
		return
	}

	hashedPassword, err := auth.HashPassword(payload.Password)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
	}
	
	err = h.store.CreateUser(types.User{
	FirstName: payload.FirstName,
	LastName: payload.LastName,
	Email: payload.Email,
	Password: hashedPassword,
	})

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	utils.WriteJSON(w, http.StatusCreated, nil)
}

func (h *Handler) handleGetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	str, ok := vars["userID"]
	if !ok {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("missing user ID"))
		return
	}

	userID, err := strconv.Atoi(str)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid user ID"))
		return
	}

	user, err := h.store.GetUserByID(userID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, user)
}