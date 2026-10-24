package services

import (
	"bufio"
	"net"
	"strings"
	"testing"
	"time"

	"backend/internal/shared/application"
)

// fakeSMTP is a tiny in-process SMTP server (no STARTTLS) that records the DATA it receives.
func fakeSMTP(t *testing.T, got chan<- string) (host, port string) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ln.Close() })
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		r := bufio.NewReader(conn)
		write := func(s string) { _, _ = conn.Write([]byte(s + "\r\n")) }
		write("220 fake ESMTP")
		var data strings.Builder
		inData := false
		for {
			line, err := r.ReadString('\n')
			if err != nil {
				return
			}
			line = strings.TrimRight(line, "\r\n")
			if inData {
				if line == "." {
					inData = false
					got <- data.String()
					write("250 queued")
					continue
				}
				data.WriteString(line + "\n")
				continue
			}
			switch cmd := strings.ToUpper(line); {
			case strings.HasPrefix(cmd, "EHLO"):
				write("250-fake")
				write("250 AUTH PLAIN")
			case strings.HasPrefix(cmd, "AUTH"):
				write("235 ok")
			case strings.HasPrefix(cmd, "MAIL"), strings.HasPrefix(cmd, "RCPT"):
				write("250 ok")
			case cmd == "DATA":
				inData = true
				write("354 go")
			case cmd == "QUIT":
				write("221 bye")
				return
			default:
				write("502 unsupported")
			}
		}
	}()
	h, p, _ := net.SplitHostPort(ln.Addr().String())
	return h, p
}

func TestMailerSendDeliversMessage(t *testing.T) {
	got := make(chan string, 1)
	host, port := fakeSMTP(t, got)
	svc, _ := NewMailerService(application.MailerConfig{SMTP_HOST: host, SMTP_PORT: port, SMTP_FROM: "chat@example.com", SMTP_PASSWORD: "pw"})

	err := svc.Send(&application.Message{To: []string{"ana@example.com"}, Subject: "Hi", Body: "hello there"})
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	select {
	case data := <-got:
		if !strings.Contains(data, "Subject: Hi") || !strings.Contains(data, "hello there") {
			t.Fatalf("unexpected DATA: %q", data)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("server never received the message")
	}
}

func TestMailerSendGivesUpOnAServerThatNeverAnswers(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	go func() { // accept and stay silent: no 220 greeting
		conn, err := ln.Accept()
		if err == nil {
			defer conn.Close()
			time.Sleep(5 * time.Second)
		}
	}()
	host, port, _ := net.SplitHostPort(ln.Addr().String())
	svc, _ := NewMailerService(application.MailerConfig{SMTP_HOST: host, SMTP_PORT: port, SMTP_FROM: "chat@example.com"})
	svc.(*MailerService).deadline = 300 * time.Millisecond

	start := time.Now()
	if err := svc.Send(&application.Message{To: []string{"a@example.com"}, Subject: "s", Body: "b"}); err == nil {
		t.Fatal("expected a timeout error")
	}
	if elapsed := time.Since(start); elapsed > 3*time.Second {
		t.Fatalf("send blocked for %s", elapsed)
	}
}
