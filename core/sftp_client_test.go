package core_test

import (
	"fmt"
	"io"
	"log"
	"net"
	"path"
	"testing"

	"github.com/APTrust/dart-runner/util"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

var sftpServer *sftp.Server
var netListener net.Listener

const sftpUserName = "demo"
const sftpPassword = "password"

func TestSftpUpload(t *testing.T) {
	StartSftpTestServer()

	// TODO: Set up StorageService and upload a file with and without progress.

	defer StopSftpTestServer()
}

func StopSftpTestServer() {
	if sftpServer != nil {
		sftpServer.Close()
	}
	if netListener != nil {
		netListener.Close()
	}
}

// StartSftpTestServer starts a local SFTP server for testing that
// accepts only local connections on 127.0.0.1:2022. This is
// used for testing SFTP connections and uploads.
//
// Adapted from https://github.com/pkg/sftp/blob/v1.13.6/examples/go-sftp-server/main.go
func StartSftpTestServer() {
	debugStream := io.Discard

	// An SSH server is represented by a ServerConfig, which holds
	// certificate details and handles authentication of ServerConns.
	config := &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			// Should use constant-time compare (or better, salt+hash) in
			// a production setting.
			fmt.Fprintf(debugStream, "Login: %s\n", c.User())
			if c.User() == sftpUserName && string(pass) == sftpPassword {
				return nil, nil
			}
			return nil, fmt.Errorf("password rejected for %q", c.User())
		},
	}

	pathToPrivateKey := path.Join(util.PathToTestData(), "files", "sftp_test_host_key")
	privateBytes, err := util.ReadFile(pathToPrivateKey)
	if err != nil {
		fmt.Println(err, "Failed to load SFTP private key")
		return
	}

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		fmt.Println(err, "Failed to parse SFTP private key")
		return
	}

	config.AddHostKey(private)

	// Once a ServerConfig has been configured, connections can be
	// accepted.
	listener, err := net.Listen("tcp", "127.0.0.1:2022")
	if err != nil {
		fmt.Println(err, "Failed to listen for connection")
		return
	}

	netListener = listener

	var nConn net.Conn
	go func() {
		nConn, err = listener.Accept()
		if err != nil {
			fmt.Println(err, "Cannot accept incoming connections")
			return
		}

		// Before use, a handshake must be performed on the incoming
		// net.Conn.
		_, chans, reqs, err := ssh.NewServerConn(nConn, config)
		if err != nil {
			fmt.Println(err, "SSH handshake failed")
			return
		}

		// The incoming Request channel must be serviced.
		go ssh.DiscardRequests(reqs)

		// Service the incoming Channel channel.
		for newChannel := range chans {
			// Channels have a type, depending on the application level
			// protocol intended. In the case of an SFTP session, this is "subsystem"
			// with a payload string of "<length=4>sftp"
			if newChannel.ChannelType() != "session" {
				newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
				continue
			}
			channel, requests, err := newChannel.Accept()
			if err != nil {
				fmt.Println(err, "Failed to accept channel")
				return
			}

			// Sessions have out-of-band requests such as "shell",
			// "pty-req" and "env".  Here we handle only the
			// "subsystem" request.
			go func(in <-chan *ssh.Request) {
				for req := range in {
					ok := false
					switch req.Type {
					case "subsystem":
						if string(req.Payload[4:]) == "sftp" {
							ok = true
						}
					}
					req.Reply(ok, nil)
				}
			}(requests)

			server, err := sftp.NewServer(channel)
			if err != nil {
				log.Fatal(err)
			}
			sftpServer = server

			if err := server.Serve(); err == io.EOF {
				server.Close()
			} else if err != nil {
				fmt.Println(err, "SFTP server completed with error")
			}
		}
	}()
}
