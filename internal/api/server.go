package api

import (
	"crypto/sha256"
	"crypto/subtle"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"mount-service/internal/mount"
	//"log"
	"mount-service/internal/db"
	"mount-service/internal/log"
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
	username, _, ok := req.BasicAuth()
	if ok {
		ipAddr, err := readUserIP(req)
		if err != nil {
			log.Logger.WithError(err).Errorln("Error on reading user IP")
			log.Logger.Warningln("Unknown user tried register")
			http.Error(w, "Unknown user ip", http.StatusBadRequest)
			return
		}

		user := s.credRepo.GetUser(username)
		if user == nil {
			log.Logger.WithFields(logrus.Fields{
				"username": username,
				"ip_addr":  ipAddr,
			}).Warningln("User not found")
			http.Error(w, "User not found", http.StatusBadRequest)
			return
		}
		user.IpAddr = ipAddr

		log.Logger.WithFields(logrus.Fields{
			"username": user.Username,
			"ip_addr":  user.IpAddr,
		}).Infoln("User start mount...")

		err = s.mounter.MountAll(user, s.activeUsers)
		if err != nil {
			log.Logger.WithError(err).
				WithFields(logrus.Fields{
					"username": user.Username,
					"ip_addr":  user.IpAddr,
				}).
				Warningln("Error on mount active users")
			http.Error(w, "Mounting error", http.StatusInternalServerError)
			return
		}

		log.Logger.WithFields(logrus.Fields{
			"username": user.Username,
			"ip_addr":  user.IpAddr,
		}).Infoln("User successfully mount")
		w.WriteHeader(http.StatusOK)

	} else {
		http.Error(w, "Error on authorization", http.StatusBadRequest)
	}
}

func (s *MountServer) RegisterHandler(w http.ResponseWriter, req *http.Request) {
	username, password, ok := req.BasicAuth()
	if ok {
		ipAddr, err := readUserIP(req)
		if err != nil {
			log.Logger.WithError(err).Errorln("Error on reading user IP")
			log.Logger.Warningln("Unknown user tried register")
			http.Error(w, "Unknown user ip", http.StatusBadRequest)
			return
		}

		user := s.credRepo.GetUser(username)
		if user != nil {
			log.Logger.Infoln("Registered user try register")
			http.Error(w, "User already register", http.StatusBadRequest)
			return
		}

		log.Logger.WithFields(logrus.Fields{
			"username": username,
			"ip_addr":  ipAddr,
		}).Infoln("User register...")

		user = &models.User{Username: username, Password: password, IpAddr: ipAddr}
		err = s.credRepo.AddUser(user)
		if err != nil {
			log.Logger.WithFields(logrus.Fields{
				"username": username,
				"ip_addr":  ipAddr,
			}).WithError(err).Errorln("Error on creating user")

			http.Error(w, "Can't create user", http.StatusInternalServerError)
			return
		}
		s.activeUsers = append(s.activeUsers, user)

		log.Logger.WithFields(logrus.Fields{
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
		log.Logger.WithFields(logrus.Fields{
			"username": username,
			"ip_addr":  ipAddr,
		}).Infoln("User logout...")

		if ipAddr == "127.0.0.1" {
			ipAddr = getOutboundIP().String()
		}

		user, ind := findUserByUsername(s.activeUsers, username)
		if user == nil {
			log.Logger.WithFields(logrus.Fields{
				"username": username,
				"ip_addr":  ipAddr,
			}).Warningln("User not in active users")

			http.Error(w, "User not in active users", http.StatusBadRequest)
			return
		}

		s.activeUsers = append(s.activeUsers[:ind], s.activeUsers[ind+1:]...)

		log.Logger.WithFields(logrus.Fields{
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
				log.Logger.Infoln("User", username, "doesn't exists")
				s.handleUnauthorized(w)
				return
			}

			expectedUsernameHash := sha256.Sum256([]byte(user.Username))
			expectedPasswordHash := sha256.Sum256([]byte(user.Password))

			usernameMatch := subtle.ConstantTimeCompare(usernameHash[:], expectedUsernameHash[:]) == 1
			passwordMatch := subtle.ConstantTimeCompare(passwordHash[:], expectedPasswordHash[:]) == 1

			if usernameMatch && passwordMatch {
				next.ServeHTTP(w, r)
				log.Logger.Infoln("User", username, "successfully authorized")
				return
			}
		}
		log.Logger.Infoln("User", username, "has incorrect login or password")
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
	log.Logger.Fatal(http.ListenAndServe(":8080", s.router))
}
