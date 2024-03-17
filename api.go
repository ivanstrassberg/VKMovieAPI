package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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
	mux := http.NewServeMux()

	mux.HandleFunc("/login", makeHTTPHandleFunc(s.handleLogin))
	mux.HandleFunc("/actor", makeHTTPHandleFunc(s.handleActor))
	mux.HandleFunc("/movie", makeHTTPHandleFunc(s.handleMovie))
	// mux.HandleFunc("/transfer", makeHTTPHandleFunc(s.handleLogin))

	log.Println("JSON API server running on port", s.listenAddr)

	if err := http.ListenAndServe(s.listenAddr, mux); err != nil {
		log.Fatalf("Error starting server: %s\n", err)
	}
}

func (s *APIServer) handleMovie(w http.ResponseWriter, r *http.Request) error {

	if r.Method == "GET" {
		return s.handleGetMovies(w, r)
	}

	if r.Method == "POST" {
		return s.handleCreateMovie(w, r)
	}

	if r.Method == "DELETE" {
		return s.handleDeleteMovie(w, r)
	}

	if r.Method == "PUT" {
		return s.handleUpdateMovie(w, r)
	}

	return fmt.Errorf("method not allowed %s", r.Method)
}

func (s *APIServer) handleGetMovies(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (s *APIServer) handleCreateMovie(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (s *APIServer) handleDeleteMovie(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (s *APIServer) handleUpdateMovie(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (s *APIServer) handleLogin(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (s *APIServer) handleActor(w http.ResponseWriter, r *http.Request) error {

	if r.Method == "GET" {
		return s.handleGetActors(w, r)
	}

	if r.Method == "POST" {
		return s.handleCreateActor(w, r)
	}

	if r.Method == "DELETE" {
		return s.handleDeleteActor(w, r)
	}

	if r.Method == "PUT" {
		return s.handleUpdateActor(w, r)
	}

	return fmt.Errorf("method not allowed %s", r.Method)
}

func (s APIServer) handleCreateActor(w http.ResponseWriter, r *http.Request) error {
	createAccountReq := new(CreateActorReq)
	if err := json.NewDecoder(r.Body).Decode(createAccountReq); err != nil {
		return err
	}
	if createAccountReq.FirstName == "" || createAccountReq.LastName == "" || createAccountReq.Sex == "" {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: "parameters missing"})
	}
	actor := NewActor(createAccountReq.FirstName, createAccountReq.LastName, createAccountReq.Sex)
	if err := s.store.CreateActor(actor); err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, actor)

}

func (s *APIServer) handleUpdateActor(w http.ResponseWriter, r *http.Request) error {
	updateReq := new(UpdateActorReq)
	if err := json.NewDecoder(r.Body).Decode(updateReq); err != nil {
		return err
	}
	// fmt.Println(updateReq, "handle")
	if err := s.store.UpdateActor(updateReq); err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, updateReq)
}

func (s *APIServer) handleDeleteActor(w http.ResponseWriter, r *http.Request) error {
	credentials, err := s.getActorDeletionCredentials(w, r)
	// fmt.Printf("%+v", credentials)
	// dcredentials = json.NewDecoder(r.Body).Decode(credentials)
	if err != nil {
		return err
	}

	if err := s.store.DeleteActor(int(credentials.ID), credentials.FirstName, credentials.LastName); err != nil {
		return err
	}

	responseData := make(map[string]interface{})
	if credentials.ID != 0 {
		responseData["id"] = int(credentials.ID)
	}
	// if credentials.FirstName != "" {
	// 	responseData["firstName"] = credentials.FirstName
	// }
	// if credentials.LastName != "" {
	// 	responseData["lastName"] = credentials.LastName
	// }

	return WriteJSON(w, http.StatusOK, map[string]interface{}{"deleted": responseData})

}

func handleUpdateActor() error {
	return nil
}

func (s *APIServer) handleGetActors(w http.ResponseWriter, r *http.Request) error {
	actors, err := s.store.GetActors()
	if err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, actors)
}

//

func WriteJSON(w http.ResponseWriter, status int, v any) error {

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

type apiFunc func(http.ResponseWriter, *http.Request) error

type ApiError struct {
	Error string `json:"error"`
}

func makeHTTPHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
	}
}

// // error here FIX or not idk // fixed?
func (s *APIServer) getActorDeletionCredentials(w http.ResponseWriter, r *http.Request) (DeleteActorReq, error) {
	deleteActorReq := new(DeleteActorReq)
	if err := json.NewDecoder(r.Body).Decode(deleteActorReq); err != nil {
		return DeleteActorReq{}, fmt.Errorf("permission denied")
	}
	return *deleteActorReq, nil
}

/*

func getId(r *http.Request) (int, error) {
	urlPath := r.URL.Path
	pathParts := strings.Split(urlPath, "/")
	idStr := pathParts[len(pathParts)-1]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, fmt.Errorf("permission denied: invalid ID '%s'", idStr)
	}
	return id, nil
}

*/
