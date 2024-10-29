package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

type APIServer struct {
	listenAddr string
	store      Storage
}

func NewAPIServer(listenAddr string, store Storage) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
		store:      store,
	}
}

func (s *APIServer) Run() {
	router := mux.NewRouter()

	router.HandleFunc("/login", makeHTTPHandleFunc(s.handleLogin))
	router.HandleFunc("/account", makeHTTPHandleFunc(s.handleAccount))
	router.HandleFunc("/account/{id}", withJwtAuth(makeHTTPHandleFunc(s.handleGetAccountById), s.store))
	router.HandleFunc("/transfer", makeHTTPHandleFunc(s.handleTransfer)).Methods("POST")
	router.HandleFunc("/jwt", makeHTTPHandleFunc(s.handleGenJwt)).Methods("POST")

	log.Printf("API server listening on %s", s.listenAddr)
	http.ListenAndServe(s.listenAddr, router)
}

func (s *APIServer) handleLogin(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		return fmt.Errorf("method not allowed %s", r.Method)
	}
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return err
	}

	// validate password
	account, err := s.store.GetAccountByNumber(int64(req.Number))
	if err != nil {
		return err
	}

	err = account.ValidatePassword([]byte(req.Password))
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return fmt.Errorf("invalid password")
	}
	fmt.Printf("%+v\n", account)

	return WriteJSON(w, http.StatusOK, account)

}

func (a *Account) ValidatePassword(pw []byte) error {
	fmt.Printf("%s\n", a.EncryptedPassword)
	fmt.Printf("%s\n", pw)
	return bcrypt.CompareHashAndPassword([]byte(a.EncryptedPassword), pw)
}

func (s *APIServer) handleAccount(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		return s.handleGetAccount(w, r)
	}
	if r.Method == "POST" {
		return s.handleCreateAccount(w, r)
	}

	return fmt.Errorf("unsupported method %s", r.Method)

}

// GET /account
func (s *APIServer) handleGetAccount(w http.ResponseWriter, r *http.Request) error {
	accounts, err := s.store.GetAccounts()
	if err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, accounts)
}

// GET /account/{id}
func (s *APIServer) handleGetAccountById(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		id, err := getId(r)
		if err != nil {
			return err
		}

		account, err := s.store.GetAccountById(id)
		if err != nil {
			return err
		}
		return WriteJSON(w, http.StatusOK, account)
	}
	if r.Method == "DELETE" {
		return s.handleDeleteAccount(w, r)
	}
	return fmt.Errorf("method not allowed: %s", r.Method)
}

func (s *APIServer) handleCreateAccount(w http.ResponseWriter, r *http.Request) error {
	createAccountReq := new(CreateAccountRequest)
	if err := json.NewDecoder(r.Body).Decode(createAccountReq); err != nil {
		return err
	}
	account, err := NewAccount(createAccountReq.FirstName, createAccountReq.LastName, createAccountReq.Password)
	if err != nil {
		return err
	}
	err = s.store.CreateAccount(account)
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, account)
}

// DELETE /account/{id}
func (s *APIServer) handleDeleteAccount(w http.ResponseWriter, r *http.Request) error {
	id, err := getId(r)
	if err != nil {
		return err
	}
	err = s.store.DeleteAccount(id)
	if err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, map[string]int{"deleted_id": id})
}

func (s *APIServer) handleGenJwt(w http.ResponseWriter, r *http.Request) error {
	account := new(Account)
	if err := json.NewDecoder(r.Body).Decode(account); err != nil {
		return err
	}
	// check info against db
	user, err := s.store.GetAccountById(account.ID)
	if err != nil || account.Number != user.Number {
		return WriteJSON(w, http.StatusForbidden, ApiError{Error: "permission denied"})
	}

	defer r.Body.Close()

	tokenString, err := createJWT(account)
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, map[string]string{"token": tokenString})
}

func (s *APIServer) handleTransfer(w http.ResponseWriter, r *http.Request) error {
	transferReq := new(TransferRequest)
	if err := json.NewDecoder(r.Body).Decode(transferReq); err != nil {
		return err
	}
	defer r.Body.Close()
	fmt.Printf("%+v\n", transferReq)
	return nil
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func createJWT(account *Account) (string, error) {
	claims := &jwt.MapClaims{
		"expiresAt":     15000,
		"accountNumber": account.Number,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := os.Getenv("JWT_SECRET")

	return token.SignedString([]byte(secret))
}

func permissionDenied(w http.ResponseWriter) {
	WriteJSON(w, http.StatusForbidden, ApiError{Error: "permission denied"})
}

func withJwtAuth(handlerFunc http.HandlerFunc, s Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("calling JWT auth middleware")

		tokenString := r.Header.Get("x-jwt-token")
		token, err := validateJWT(tokenString)
		if err != nil {
			permissionDenied(w)
			return
		}
		if !token.Valid {
			permissionDenied(w)
			return
		}
		userID, err := getId(r)
		if err != nil {
			permissionDenied(w)
			return
		}
		account, err := s.GetAccountById(userID)
		if err != nil {
			permissionDenied(w)
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		if account.Number != int64(claims["accountNumber"].(float64)) {
			permissionDenied(w)
			return
		}

		handlerFunc(w, r)
	}
}

func validateJWT(tokenString string) (*jwt.Token, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})
}

type apiFunc func(w http.ResponseWriter, r *http.Request) error

type ApiError struct {
	Error string `json:"error"`
}

func makeHTTPHandleFunc(fn apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := fn(w, r); err != nil {
			WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
	}
}

func getId(r *http.Request) (int, error) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, fmt.Errorf("invalid id given %s", idStr)
	}
	return id, nil
}
