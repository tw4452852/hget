package main

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestJoinFile(t *testing.T) {
	err := ioutil.WriteFile("./part1.txt", []byte("hello "), os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove("./part1.txt")

	err = ioutil.WriteFile("./part2.txt", []byte("world!"), os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove("./part2.txt")

	err = JoinFile([]string{"./part2.txt", "./part1.txt"}, "./out.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove("./out.txt")

	got, err := ioutil.ReadFile("./out.txt")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "hello world!" {
		t.Errorf("got[%s] isn't expect[%s]\n", string(got), "hello world")
	}
}
