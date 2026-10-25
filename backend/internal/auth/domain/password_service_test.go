package domain

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func TestPasswordServiceQueueTimeout(t *testing.T) {
	svc := newBcryptPasswordService(1, 30*time.Millisecond)
	hash, err := bcrypt.GenerateFromPassword([]byte("secret-pass"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}

	if !svc.CheckPasswordHash("secret-pass", string(hash)) {
		t.Fatal("a free slot must verify normally")
	}

	svc.slots <- struct{}{} // every slot is busy
	start := time.Now()
	if _, err := svc.HashPassword("whatever-123"); !errors.Is(err, ErrPasswordBusy) {
		t.Fatalf("hash while busy: err = %v, want ErrPasswordBusy", err)
	}
	if svc.CheckPasswordHash("secret-pass", string(hash)) {
		t.Fatal("compare while busy must not succeed")
	}
	if time.Since(start) < 30*time.Millisecond {
		t.Fatal("callers must wait for the queue timeout before giving up")
	}

	svc.release()
	if !svc.CheckPasswordHash("secret-pass", string(hash)) {
		t.Fatal("verification must work again once a slot is free")
	}
}

func TestPasswordServiceBoundsConcurrency(t *testing.T) {
	svc := newBcryptPasswordService(2, time.Second)
	var running, peak atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if !svc.acquire() {
				t.Error("acquire timed out")
				return
			}
			n := running.Add(1)
			for {
				p := peak.Load()
				if n <= p || peak.CompareAndSwap(p, n) {
					break
				}
			}
			time.Sleep(5 * time.Millisecond)
			running.Add(-1)
			svc.release()
		}()
	}
	wg.Wait()
	if peak.Load() > 2 {
		t.Fatalf("peak concurrency %d exceeds the limit of 2", peak.Load())
	}
}
