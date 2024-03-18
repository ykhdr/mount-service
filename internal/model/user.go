package model

import "net"

type User struct {
	Username string `db:"username"`
	Password string `db:"password"`
	IpAddr   net.IP `db:"addr"`
}
