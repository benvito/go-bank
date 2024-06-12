package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
)

func WriteJSON(writer http.ResponseWriter, status int, value any) error {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	return json.NewEncoder(writer).Encode(value)
}

type apiFunc func(http.ResponseWriter, *http.Request) error

// G2aS3ben1pwWSd22lL9
type ApiError struct {
	Error string `json:"error"`
}

func makeHttpHandleFunc(f apiFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if err := f(writer, request); err != nil {
			WriteJSON(writer, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
	}
}

type ApiServer struct {
	listenAddress string
	database      Database
	log           *log.Logger
	errorLog      *log.Logger
}

func NewApiServer(listenAddress string, database Database, log, errorLog *log.Logger) *ApiServer {
	return &ApiServer{
		listenAddress: listenAddress,
		database:      database,
		log:           log,
		errorLog:      errorLog,
	}
}

func (s *ApiServer) Run() {
	router := mux.NewRouter()

	router.HandleFunc("/account", makeHttpHandleFunc(s.handleAccount))

	router.HandleFunc("/account/{id}", withJWTAuth(makeHttpHandleFunc(s.handleGetAccountByID), s.database))
	router.HandleFunc("/transfer", makeHttpHandleFunc(s.handleTransfer))

	s.log.Println("API running on port: ", s.listenAddress)

	srv := &http.Server{
		Addr:     s.listenAddress,
		Handler:  router,
		ErrorLog: s.errorLog,
	}

	srv.ListenAndServe()
}

func (s *ApiServer) handleAccount(writer http.ResponseWriter, request *http.Request) error {
	switch request.Method {
	case http.MethodGet:
		return s.handleGetAccount(writer, request)
	case http.MethodPost:
		return s.handleCreateAccount(writer, request)
	case http.MethodDelete:
		return s.handleDeleteAccount(writer, request)
	}

	writer.Header().Set("Allow", "GET, POST, DELETE")
	return WriteJSON(writer, http.StatusMethodNotAllowed, map[string]string{"error": fmt.Sprintf("method %s not allowed", request.Method)})
}

func (s *ApiServer) handleGetAccount(writer http.ResponseWriter, request *http.Request) error {
	accounts, err := s.database.GetAccounts()

	if err != nil {
		return err
	}

	return WriteJSON(writer, http.StatusOK, accounts)
}

func (s *ApiServer) handleGetAccountByID(writer http.ResponseWriter, request *http.Request) error {
	if request.Method == http.MethodGet {
		id, err := getID(request)

		if err != nil {
			return err
		}

		account, err := s.database.GetAccountByID(id)

		if err != nil {
			return err
		}

		return WriteJSON(writer, http.StatusOK, account)
	}

	if request.Method == http.MethodDelete {
		return s.handleDeleteAccount(writer, request)
	}

	writer.Header().Set("Allow", "GET, DELETE")
	return WriteJSON(writer, http.StatusMethodNotAllowed, map[string]string{"error": fmt.Sprintf("method %s not allowed", request.Method)})
}

func (s *ApiServer) handleCreateAccount(writer http.ResponseWriter, request *http.Request) error {
	createAccReq := CreateAccountRequest{}

	if err := json.NewDecoder(request.Body).Decode(&createAccReq); err != nil {
		return err
	}
	defer request.Body.Close()

	account := NewAccount(createAccReq.FirstName, createAccReq.LastName, createAccReq.Number)

	tokenString, err := createJWTToken(account)

	if err != nil {
		return err
	}

	fmt.Println(tokenString)

	inserted, err := s.database.CreateAccount(account)

	if err != nil {
		return err
	}

	return WriteJSON(writer, http.StatusOK, inserted)
}

func (s *ApiServer) handleDeleteAccount(writer http.ResponseWriter, request *http.Request) error {
	id, err := getID(request)

	if err != nil {
		return err
	}

	err = s.database.DeleteAccount(id)

	if err != nil {
		return err
	}

	return WriteJSON(writer, http.StatusOK, map[string]int{"deleted": id})
}

func (s *ApiServer) handleTransfer(writer http.ResponseWriter, request *http.Request) error {
	transerReq := &TransferRequest{}

	if err := json.NewDecoder(request.Body).Decode(transerReq); err != nil {
		return err
	}
	defer request.Body.Close()

	return WriteJSON(writer, http.StatusOK, transerReq)
}

func getID(request *http.Request) (int, error) {
	idStr := mux.Vars(request)["id"]
	id, err := strconv.Atoi(idStr)

	if err != nil || id < 1 {
		return id, fmt.Errorf("invalid ID: %s", idStr)
	}

	return id, nil
}

func writePermissionDenied(writter http.ResponseWriter) {
	WriteJSON(writter, http.StatusForbidden, ApiError{Error: "permission denied"})
}

func withJWTAuth(handlerFunc http.HandlerFunc, database Database) http.HandlerFunc {
	return func(writter http.ResponseWriter, request *http.Request) {
		tokenString := request.Header.Get("jwt-token")
		token, err := verifyToken(tokenString)

		if err != nil {
			writePermissionDenied(writter)
			return
		}

		userID, err := getID(request)

		if err != nil {
			writePermissionDenied(writter)
			return
		}

		account, err := database.GetAccountByID(userID)

		if err != nil {
			writePermissionDenied(writter)
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		if account.Number != claims["accountNumber"] {
			writePermissionDenied(writter)
			return
		}

		handlerFunc(writter, request)
	}
}

// eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2NvdW50TnVtYmVyIjoiNTUyMTIyIiwiZXhwIjoiMjAyNC0wNi0xM1QwMjoyNjowMS4zNDI0MzMzKzA3OjAwIn0.XgE6ozq4nqsPFLZQ2ylGOpVL2etMRoFPXBSqh3cejBA

func createJWTToken(account *Account) (string, error) {
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"accountNumber": account.Number,
		"exp":           time.Now().Add(12 * time.Hour).Unix(),
	})

	secret := []byte(os.Getenv("JWT_SECRET"))
	tokenString, err := claims.SignedString(secret)

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func verifyToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("permission denied")
	}

	return token, nil
}
