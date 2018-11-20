package trial

import (
	"fmt"
	"log"
	"os"
	"testing"
)

func Test_CaptureLog(t *testing.T) {
	// test single line read
	c := CaptureLog()
	log.Println("log line")
	if equal, diff := Equal(c.ReadAll(), "log line"); !equal {
		t.Fatalf("FAIL: read single log line %s", diff)
	}

	//verify the log output has been reset
	log.Println("test")
	if equal, _ := Equal(c.ReadAll(), "log line"); !equal {
		t.Error("log output not reset")
	}

	// test multi-line reads
	c = CaptureLog()
	log.Println("log1")
	log.Println("log2")
	log.Println("log3")
	if equal, diff := Equal(c.ReadLines(), []string{"log1", "log2", "log3"}); !equal {
		t.Errorf("FAIL: multi-line %s", diff)
	}
}

func Test_CaptureStdErr(t *testing.T) {
	// test single line read
	c := CaptureStdErr()
	fmt.Fprint(os.Stderr, "log line")
	if equal, diff := Equal(c.ReadAll(), "log line"); !equal {
		t.Fatalf("FAIL: read single log line %s", diff)
	}

	//verify the log output has been reset
	fmt.Fprint(os.Stderr, "stderr")
	if equal, _ := Equal(c.ReadAll(), "log line"); !equal {
		t.Error("log output not reset")
	}

	// test multi-line reads
	c = CaptureStdErr()
	fmt.Fprint(os.Stderr, "log1\n")
	fmt.Fprint(os.Stderr, "log2\n")
	fmt.Fprint(os.Stderr, "log3\n")
	if equal, diff := Equal(c.ReadLines(), []string{"log1", "log2", "log3"}); !equal {
		t.Errorf("FAIL: multi-line %s", diff)
	}
}

func Test_CaptureStdOut(t *testing.T) {
	// test single line read
	c := CaptureStdOut()
	fmt.Println("log line")
	if equal, diff := Equal(c.ReadAll(), "log line"); !equal {
		t.Fatalf("FAIL: read single log line %s", diff)
	}

	//verify the log output has been reset
	fmt.Println("stdout")
	if equal, _ := Equal(c.ReadAll(), "log line"); !equal {
		t.Error("log output not reset")
	}

	// test multi-line reads
	c = CaptureStdOut()
	fmt.Println("log1")
	fmt.Println("log2")
	fmt.Println("log3")
	if equal, diff := Equal(c.ReadLines(), []string{"log1", "log2", "log3"}); !equal {
		t.Errorf("FAIL: multi-line %s", diff)
	}
}
