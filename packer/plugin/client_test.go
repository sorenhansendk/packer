package plugin

import (
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	process := helperProcess("mock")
	c := NewClient(&ClientConfig{Cmd: process})
	defer c.Kill()

	// Test that it parses the proper address
	addr, err := c.Start()
	if err != nil {
		t.Fatalf("err should be nil, got %s", err)
	}

	if addr != ":1234" {
		t.Fatalf("incorrect addr %s", addr)
	}

	// Test that it exits properly if killed
	c.Kill()

	if process.ProcessState == nil {
		t.Fatal("should have process state")
	}

	// Test that it knows it is exited
	if !c.Exited() {
		t.Fatal("should say client has exited")
	}
}

func TestClient_Start_Timeout(t *testing.T) {
	config := &ClientConfig{
		Cmd:          helperProcess("start-timeout"),
		StartTimeout: 50 * time.Millisecond,
	}

	c := NewClient(config)
	defer c.Kill()

	_, err := c.Start()
	if err == nil {
		t.Fatal("err should not be nil")
	}
}

func TestClient_Stdin(t *testing.T) {
	// Overwrite stdin for this test with a temporary file
	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(tf.Name())
	defer tf.Close()

	if _, err = tf.WriteString("hello"); err != nil {
		t.Fatalf("error: %s", err)
	}

	if err = tf.Sync(); err != nil {
		t.Fatalf("error: %s", err)
	}

	if _, err = tf.Seek(0, 0); err != nil {
		t.Fatalf("error: %s", err)
	}

	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()
	os.Stdin = tf

	process := helperProcess("stdin")
	c := NewClient(&ClientConfig{Cmd: process})
	defer c.Kill()

	_, err = c.Start()
	if err != nil {
		t.Fatalf("error: %s", err)
	}

	for {
		if c.Exited() {
			break
		}

		time.Sleep(50 * time.Millisecond)
	}

	if !process.ProcessState.Success() {
		t.Fatal("process didn't exit cleanly")
	}
}
