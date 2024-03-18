package api

import (
	"crypto/sha256"
	"crypto/subtle"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net"

	//"log"
	"mount-service/internal/db"
	"mount-service/internal/models"
	"net/http"
)

type MountServer struct {
	credRepo    *db.UserRepository
	router      *mux.Router
	activeUsers []*models.User
}

func CreateNewServer(config *models.Config) *MountServer {
	credRepo := db.NewUserRepository(config)

	server := &MountServer{credRepo: credRepo, router: &mux.Router{}}
	server.setupRouter()

	return server
}

func (s *MountServer) MountHandler(w http.ResponseWriter, req *http.Request) {

}

func (s *MountServer) RegisterHandler(w http.ResponseWriter, req *http.Request) {
	username, password, ok := req.BasicAuth()
	if ok {
		ipAddr := req.RemoteAddr
		log.WithFields(log.Fields{
			"username": username,
			"password": password,
			"ip_addr":  ipAddr,
		}).Infoln("User register...")

		user := models.User{Username: username, Password: password, IpAddr: net.ParseIP(ipAddr)}
		err := s.credRepo.AddUser(user)
		if err != nil {
			log.WithFields(log.Fields{
				"username": username,
				"password": password,
				"ip_addr":  ipAddr,
			}).WithError(err).Errorln("Error on creating user")

			http.Error(w, "Can't create user", http.StatusInternalServerError)
			return
		}
		s.activeUsers = append(s.activeUsers)

		log.WithFields(log.Fields{
			"username": username,
			"password": password,
			"ip_addr":  ipAddr,
		}).Infoln("User successfully register")
		w.WriteHeader(http.StatusOK)

	} else {
		http.Error(w, "Error on authorization", http.StatusBadRequest)
	}

}

func (s *MountServer) LogoutHandler(w http.ResponseWriter, req *http.Request) {
	username, password, ok := req.BasicAuth()
	if ok {
		ipAddr := req.RemoteAddr
		log.WithFields(log.Fields{
			"username": username,
			"password": password,
			"ip_addr":  ipAddr,
		}).Infoln("User logout...")

		user, ind := findUserByUsername(s.activeUsers, username)
		if user == nil {
			log.WithFields(log.Fields{
				"username": username,
				"password": password,
				"ip_addr":  ipAddr,
			}).Warningln("User not in active users")

			http.Error(w, "User not in active users", http.StatusBadRequest)
			return
		}

		s.activeUsers = append(s.activeUsers[:ind], s.activeUsers[ind:]...)

		log.WithFields(log.Fields{
			"username": username,
			"password": password,
			"ip_addr":  ipAddr,
		}).Infoln("User successfully logout")

		w.WriteHeader(http.StatusOK)

	} else {
		http.Error(w, "Error on authenticate", http.StatusBadRequest)
	}

}

func (s *MountServer) BasicAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if ok {
			usernameHash := sha256.Sum256([]byte(username))
			passwordHash := sha256.Sum256([]byte(password))

			user := s.credRepo.GetUser(username)
			if user == nil {
				s.handleUnauthorized(w)
				return
			}

			expectedUsernameHash := sha256.Sum256([]byte(user.Username))
			expectedPasswordHash := []byte(user.Password)

			usernameMatch := subtle.ConstantTimeCompare(usernameHash[:], expectedUsernameHash[:]) == 1
			passwordMatch := subtle.ConstantTimeCompare(passwordHash[:], expectedPasswordHash[:]) == 1

			if usernameMatch && passwordMatch {
				next.ServeHTTP(w, r)
				return
			}
		}
		s.handleUnauthorized(w)
	}
}

func (s *MountServer) handleUnauthorized(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
	http.Error(w, "Incorrect password or login", http.StatusUnauthorized)
}

func (s *MountServer) setupRouter() {
	s.router.HandleFunc("/register", s.RegisterHandler).Methods("POST")
	s.router.HandleFunc("/logout", s.BasicAuth(s.LogoutHandler)).Methods("POST")
	s.router.HandleFunc("/mount", s.BasicAuth(s.MountHandler)).Methods("GET")
}

func (s *MountServer) Run() {
	log.Fatal(http.ListenAndServe(":8080", s.router))
}

func findUserByUsername(users []*models.User, username string) (*models.User, int) {
	for i, user := range users {
		if user.Username == username {
			return user, i
		}
	}
	return nil, -1
}
