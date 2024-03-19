package mount

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"mount-service/internal/models"
	"sync"
)

type Mounter struct {
	hostUser     string
	hostPassword string
}

func NewMounter(hostUser, hostPassword string) *Mounter {
	return &Mounter{
		hostUser:     hostUser,
		hostPassword: hostPassword,
	}
}

func (m *Mounter) MountAll(user *models.User, mountUsers []*models.User) error {

	sshConfig := &ssh.ClientConfig{
		User: user.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(user.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", user.IpAddr.String(), sshConfig)
	if err != nil {
		// TODO log
		return err
	}

	defer client.Close()

	wg := &sync.WaitGroup{}

	for _, mUser := range mountUsers {
		if user != mUser {
			wg.Add(1)
			go mountUser(client, mUser)
		}
	}

	wg.Wait()
	return nil
}

func mountUser(client *ssh.Client, mountUser *models.User) {
	session, err := client.NewSession()
	if err != nil {
		// TODO log
		return
	}
	defer session.Close()

	mkdirCommand := fmt.Sprintf("mkdir %s", mountUser.Username)
	mountCommand := fmt.Sprintf("sshfs -o password_stdin %s@%s:~ ~/%s", mountUser.Username, mountUser.IpAddr.String(), mountUser.Username)

	mkdirOutput, err := session.CombinedOutput(mkdirCommand)
	if err != nil {
		// TODO log
		return
	}

	fmt.Println("mountOutput:", string(mkdirOutput))

	mountOutput, err := session.CombinedOutput(mountCommand)
	if err != nil {
		// TODO log
		return
	}

	fmt.Println("mountOutput:", string(mountOutput))
}
