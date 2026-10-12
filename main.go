// Rickroll TUI is a tiny SSH service with a deliberately over-produced terminal UI.
package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/binary"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

const defaultPort = "2222"
const defaultBindAddress = "0.0.0.0"

type dimensions struct{ w, h int }

func main() {
	addr := getenv("RICKROLL_BIND_ADDRESS", defaultBindAddress) + ":" + getenv("RICKROLL_PORT", defaultPort)
	signer, err := loadHostKey(getenv("RICKROLL_HOST_KEY", "host_key.pem"))
	if err != nil {
		log.Fatal(err)
	}
	hostKey := ssh.FingerprintSHA256(signer.PublicKey())
	config := &ssh.ServerConfig{NoClientAuth: true}
	config.AddHostKey(signer)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("rickroll TUI listening on %s (anonymous SSH, host key %s)", addr, hostKey)
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("accept: %v", err)
			continue
		}
		go serve(conn, config)
	}
}

func serve(raw net.Conn, config *ssh.ServerConfig) {
	defer raw.Close()
	_, chans, reqs, err := ssh.NewServerConn(raw, config)
	if err != nil {
		return
	}
	go ssh.DiscardRequests(reqs)
	for channel := range chans {
		if channel.ChannelType() != "session" {
			channel.Reject(ssh.UnknownChannelType, "sessions only")
			continue
		}
		ch, requests, err := channel.Accept()
		if err != nil {
			continue
		}
		go session("anonymous", ch, requests)
	}
}

func session(user string, ch ssh.Channel, requests <-chan *ssh.Request) {
	defer ch.Close()
	sizes := make(chan dimensions, 1)
	started := make(chan bool, 1)
	done := make(chan struct{})
	defer close(done)
	go func() {
		hasPTY, running := false, false
		defer func() {
			if !running {
				select {
				case started <- false:
				default:
				}
			}
		}()
		for req := range requests {
			ok := false
			switch req.Type {
			case "pty-req":
				var p struct {
					Term                                   string
					Width, Height, PixelWidth, PixelHeight uint32
					Modes                                  string
				}
				if ssh.Unmarshal(req.Payload, &p) == nil {
					hasPTY, ok = true, true
					select {
					case sizes <- dimensions{int(p.Width), int(p.Height)}:
					default:
					}
				}
			case "window-change":
				if len(req.Payload) == 16 {
					size := dimensions{int(binary.BigEndian.Uint32(req.Payload)), int(binary.BigEndian.Uint32(req.Payload[4:]))}
					select {
					case <-sizes:
					default:
					}
					select {
					case sizes <- size:
					default:
					}
					ok = true
				}
			case "shell":
				ok = hasPTY && !running && len(req.Payload) == 0
				if ok {
					running = true
					started <- true
				}
			}
			if req.WantReply {
				req.Reply(ok, nil)
			}
		}
	}()
	if !<-started {
		return
	}
	keys := make(chan byte, 16)
	go func() {
		defer close(keys)
		buf := make([]byte, 1)
		for {
			if _, err := ch.Read(buf); err != nil {
				return
			}
			select {
			case keys <- buf[0]:
			case <-done:
				return
			}
		}
	}()
	io.WriteString(ch, "\x1b[?1049h\x1b[?25l")
	defer io.WriteString(ch, "\x1b[0m\x1b[?25h\x1b[?1049l")
	ticker := time.NewTicker(150 * time.Millisecond)
	defer ticker.Stop()
	size, frame, paused := dimensions{80, 24}, 0, false
	for {
		select {
		case key, ok := <-keys:
			if !ok {
				return
			}
			// A rickroll must persist. Quit keys are intentionally swallowed;
			// closing the SSH connection is the only way out.
			if key == 'q' || key == 'Q' || key == 3 {
				continue
			}
			if key == ' ' {
				paused = !paused
			}
		case size = <-sizes:
		case <-ticker.C:
			if !paused {
				frame++
			}
		}
		if _, err := io.WriteString(ch, asciiScreen(frame, size)); err != nil {
			return
		}
	}
}

// Hand-drawn ASCII poses: pompadour, trench coat, striped shirt and microphone.
var poses = []string{
	`              .-======-.
             / /////// \
             | /    \  |
             (  -  -  )
              \  __  /
               '.__.'
          ____ /|  |\ ____
         /    / |==| \    \
        / /  /  |==|  \  \ \       ___
       / /  /   |==|   \  \ \     (###)
      / /  /    |==|    \  \ '----' | |
     /_/  /     |==|     \  '------'| |
    (__) /______|__|______\         | |
           |     /\     |          | |
           |    /  \    |          | |
           |   /    \   |          | |
          /   /      \   \         | |
         /___/        \___\        | |
        (____)        (____)       _|_|_`,
	`                .-======-.
               / /////// \
               | /    \  |
               (  -  -  )
                \  __  /
                 '.__.'
            ____ /|  |\ ____
           /    / |==| \    \
    ___   / /  /  |==|  \  \ \     ___
   (___)-' /  /   |==|   \  \ \   (###)
    '-----'  /    |==|    \  \ '--' | |
            /     |==|     \  '---'| |
           /______|__|______\       | |
              |    /\    |         | |
              |   /  \   |         | |
             /   /   /   /         | |
            /   /   /   /          | |
           /___/   /___/           | |
          (____)  (____)          __|_|__`,
	`              .-======-.
             / /////// \
             | /    \  |
             (  -  -  )
              \  __  /
               '.__.'
          ____ /|  |\ ____
         /    / |==| \    \
        / /  /  |==|  \  \ \       ___
       / /  /   |==|   \  \ \     (###)
      / /  /    |==|    \  \ '----' | |
      \ \ /     |==|     \  '------'| |
       \(_)_____|__|______\         | |
           |     /\     |          | |
           |    /  \    |          | |
            \   \   \   \          | |
             \   \   \   \         | |
              \___\   \___\        | |
              (____)  (____)      __|_|__`,
}

func asciiScreen(frame int, size dimensions) string {
	w, h := size.w, size.h
	if w < 1 {
		w = 80
	}
	if h < 1 {
		h = 24
	}
	if w > 240 {
		w = 240
	}
	if h > 100 {
		h = 100
	}
	lines := []string{"RICK ASTLEY - NEVER GONNA GIVE YOU UP", ""}
	lines = append(lines, strings.Split(poses[[]int{0, 1, 0, 2}[(frame/2)%4]], "\n")...)
	subtitles := []string{
		"♪ Never gonna give you up ♪",
		"♪ Never gonna let you down ♪",
		"♪ Never gonna run around and desert you ♪",
		"♪ Never gonna make you cry ♪",
		"♪ Never gonna say goodbye ♪",
	}
	lines = append(lines, "", subtitles[(frame/8)%len(subtitles)])
	var out strings.Builder
	out.WriteString("\x1b[H\x1b[2J")
	for i, line := range lines {
		if i >= h {
			break
		}
		if len(line) > w {
			line = line[:w]
		}
		pad := (w - len(line)) / 2
		// Anchor all artwork to the same column to preserve its proportions.
		if i >= 2 && i < 21 {
			pad = (w - 48) / 2
			if pad < 0 {
				pad = 0
			}
		}
		fmt.Fprintf(&out, "\x1b[%d;%dH%s", i+1, pad+1, line)
	}
	return out.String()
}

func getenv(k, fallback string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return fallback
}

func loadHostKey(path string) (ssh.Signer, error) {
	data, err := os.ReadFile(path)
	if err == nil {
		return ssh.ParsePrivateKey(data)
	}
	if !os.IsNotExist(err) {
		return nil, err
	}
	key, err := rsa.GenerateKey(rand.Reader, 3072)
	if err != nil {
		return nil, err
	}
	data = pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return nil, err
	}
	_, writeErr := f.Write(data)
	closeErr := f.Close()
	if writeErr != nil {
		return nil, writeErr
	}
	if closeErr != nil {
		return nil, closeErr
	}
	return ssh.NewSignerFromKey(key)
}
