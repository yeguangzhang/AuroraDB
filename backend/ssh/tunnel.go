package ssh

import (
	"fmt"
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

type Tunnel struct {
	Local    string
	Remote   string
	Server   string
	Config   *ssh.ClientConfig
	client   *ssh.Client
	listener net.Listener
}

func NewTunnel(sshHost string, sshPort int, sshUser, sshPassword, sshKey, passphrase string, dbHost string, dbPort int) (*Tunnel, error) {
	var authMethods []ssh.AuthMethod

	if sshKey != "" {
		var signer ssh.Signer
		var err error
		if passphrase != "" {
			signer, err = ssh.ParsePrivateKeyWithPassphrase([]byte(sshKey), []byte(passphrase))
		} else {
			signer, err = ssh.ParsePrivateKey([]byte(sshKey))
		}
		if err != nil {
			return nil, fmt.Errorf("parse private key error: %v", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	} else {
		authMethods = append(authMethods, ssh.Password(sshPassword))
	}

	config := &ssh.ClientConfig{
		User:            sshUser,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}

	return &Tunnel{
		Local:  "127.0.0.1:0", // 随机��地端口
		Remote: fmt.Sprintf("%s:%d", dbHost, dbPort),
		Server: fmt.Sprintf("%s:%d", sshHost, sshPort),
		Config: config,
	}, nil
}

func (t *Tunnel) Start() (string, error) {
	client, err := ssh.Dial("tcp", t.Server, t.Config)
	if err != nil {
		return "", fmt.Errorf("ssh dial error: %v", err)
	}
	t.client = client

	listener, err := net.Listen("tcp", t.Local)
	if err != nil {
		return "", fmt.Errorf("local listen error: %v", err)
	}
	t.listener = listener

	go t.forward()

	// 返回实际分配的本地地址
	return listener.Addr().String(), nil
}

func (t *Tunnel) forward() {
	for {
		local, err := t.listener.Accept()
		if err != nil {
			return
		}

		go func() {
			remote, err := t.client.Dial("tcp", t.Remote)
			if err != nil {
				local.Close()
				return
			}

			go func() {
				defer local.Close()
				defer remote.Close()
				copyData(local, remote)
			}()

			go func() {
				defer local.Close()
				defer remote.Close()
				copyData(remote, local)
			}()
		}()
	}
}

func (t *Tunnel) Stop() {
	if t.listener != nil {
		t.listener.Close()
	}
	if t.client != nil {
		t.client.Close()
	}
}

func copyData(dst net.Conn, src net.Conn) {
	defer dst.Close()
	defer src.Close()
	buf := make([]byte, 32*1024)
	for {
		n, err := src.Read(buf)
		if n > 0 {
			if _, err := dst.Write(buf[0:n]); err != nil {
				return
			}
		}
		if err != nil {
			return
		}
	}
}
