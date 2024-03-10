package api

import (
	"crypto/sha256"
	"crypto/subtle"
	"mount-service/internal/db"
	"mount-service/internal/model"
	"net/http"
)

type MountServer struct {
	credRepo *db.UserRepository
}

func NewMountServer(config *model.Config) *MountServer {
	credRepo := db.NewUserRepository(config)
	return &MountServer{credRepo: credRepo}
}

func (s *MountServer) MountHandler(w http.ResponseWriter, req *http.Request) {

}

func (s *MountServer) RegisterHandler(w http.ResponseWriter, req *http.Request) {

}

func (s *MountServer) LogoutHandler(w http.ResponseWriter, req *http.Request) {

}

func (s *MountServer) BasicAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if ok {
			usernameHash := sha256.Sum256([]byte(username))
			passwordHash := sha256.Sum256([]byte(password))

			user := s.credRepo.GetUser(username)

			expectedUsernameHash := sha256.Sum256([]byte(user.Username))
			expectedPasswordHash := []byte(user.Password)

			usernameMatch := subtle.ConstantTimeCompare(usernameHash[:], expectedUsernameHash[:]) == 1
			passwordMatch := subtle.ConstantTimeCompare(passwordHash[:], expectedPasswordHash[:]) == 1

			if usernameMatch && passwordMatch {
				next.ServeHTTP(w, r)
				return
			}
		}

		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	}
}
