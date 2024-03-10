package api

import "github.com/gorilla/mux"

func FillRoute(server *MountServer) {
	router := mux.NewRouter()

	router.HandleFunc("/register", server.RegisterHandler).Methods("POST")
	router.HandleFunc("/logout", server.BasicAuth(server.LogoutHandler)).Methods("POST")
	router.HandleFunc("/mount", server.BasicAuth(server.MountHandler)).Methods("GET")
}
