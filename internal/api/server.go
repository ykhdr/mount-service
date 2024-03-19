package api

import (
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"mount-service/internal/mount"
	"net"
	"strings"

	//"log"
	"mount-service/internal/db"
	"mount-service/internal/models"
	"net/http"
)

type MountServer struct {
	credRepo    *db.UserRepository
	router      *mux.Router
	mounter     *mount.Mounter
	activeUsers []*models.User
}

func CreateNewServer(config *models.Config) *MountServer {
	credRepo := db.NewUserRepository(config)
	mounter := mount.NewMounter(config.HostUser, config.HostPassword)

	server := &MountServer{credRepo: credRepo, router: &mux.Router{}, mounter: mounter, activeUsers: make([]*models.User, 0)}
	server.setupRouter()

	return server
}

func (s *MountServer) MountHandler(w http.ResponseWriter, req *http.Request) {

}

func (s *MountServer) RegisterHandler(w http.ResponseWriter, req *http.Request) {
	username, password, ok := req.BasicAuth()
	if ok {
		ipAddr, err := readUserIP(req)
		if err != nil {
			log.WithError(err).Errorln("Error on reading user IP")
			log.Warningln("Unknown user tried register")
			return
		}
		log.WithFields(log.Fields{
			"username": username,
			"ip_addr":  ipAddr,
		}).Infoln("User register...")

		user := &models.User{Username: username, Password: password, IpAddr: ipAddr}
		err = s.credRepo.AddUser(user)
		if err != nil {
			log.WithFields(log.Fields{
				"username": username,
				"ip_addr":  ipAddr,
			}).WithError(err).Errorln("Error on creating user")

			http.Error(w, "Can't create user", http.StatusInternalServerError)
			return
		}
		s.activeUsers = append(s.activeUsers, user)

		log.WithFields(log.Fields{
			"username": username,
			"ip_addr":  ipAddr,
		}).Infoln("User successfully register")
		w.WriteHeader(http.StatusOK)

	} else {
		http.Error(w, "Error on authorization", http.StatusBadRequest)
	}

}

func (s *MountServer) LogoutHandler(w http.ResponseWriter, req *http.Request) {
	username, _, ok := req.BasicAuth()
	if ok {
		ipAddr := req.RemoteAddr
		log.WithFields(log.Fields{
			"username": username,
			"ip_addr":  ipAddr,
		}).Infoln("User logout...")

		user, ind := findUserByUsername(s.activeUsers, username)
		if user == nil {
			log.WithFields(log.Fields{
				"username": username,
				"ip_addr":  ipAddr,
			}).Warningln("User not in active users")

			http.Error(w, "User not in active users", http.StatusBadRequest)
			return
		}

		s.activeUsers = append(s.activeUsers[:ind], s.activeUsers[ind+1:]...)

		log.WithFields(log.Fields{
			"username": username,
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
			expectedPasswordHash := sha256.Sum256([]byte(user.Password))

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

func readUserIP(r *http.Request) (net.IP, error) {
	ipAddr := r.Header.Get("X-Real-Ip")
	if ipAddr == "" {
		ipAddr = r.Header.Get("X-Forwarded-For")
	}
	if ipAddr == "" {
		ipAddr = r.RemoteAddr
	}

	if strings.ContainsAny(ipAddr, "[::1]") {
		return net.ParseIP("127.0.0.1"), nil
	}

	resolvedAddr, err := net.ResolveTCPAddr("tcp4", ipAddr)
	if err != nil {
		return nil, errors.New("error on resolving user IP addr")
	}

	return resolvedAddr.IP, nil
}

func findUserByUsername(users []*models.User, username string) (*models.User, int) {
	for i, user := range users {
		if user.Username == username {
			return user, i
		}
	}
	return nil, -1
}
