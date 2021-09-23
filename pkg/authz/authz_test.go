package authz

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestIsAllowed(t *testing.T) {
	data := "aaa\nbbb\n"
	if err := ioutil.WriteFile("userfile", []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	defer os.Remove("userfile")
	ok, err := IsAllowed("aaa", "userfile")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Errorf("expected user %s to be allowed", "aaa")
	}

	ok, err = IsAllowed("ccc", "userfile")
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Errorf("expected user %s to be disallowed", "ccc")
	}
}
