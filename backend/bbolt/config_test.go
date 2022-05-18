package bbolt

import (
	"reflect"
	"testing"
)

func TestConfig(t *testing.T) {

	c0 := &Config{}

	c1 := &Config{
		Path: "some/file/path",
	}

	c2 := &Config{
		Path:     "some/other/file/path",
		Filemode: 0777,
	}

	c3 := &Config{
		Path:     "some/file/path",
		Filemode: 0777,
	}

	c4 := c1.Merge(c0)

	c5 := c2.Merge(c1)

	c6 := DefaultConfig().Merge(c2)

	if c4.Path != c1.Path {
		t.Fatalf("expected both config paths to be %s, but merged path was %s", c1.Path, c4.Path)
	}

	if !reflect.DeepEqual(c3, c5) {
		t.Fatalf("expected merged config to be %#v, but got %#v", c2, c5)
	}

	if !reflect.DeepEqual(c2, c6) {
		t.Fatalf("expected default merged config to be %#v, but got %#v", c2, c6)
	}

}
