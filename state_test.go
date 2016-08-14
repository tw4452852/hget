package main

import (
	"net/http"
	"os"
	"reflect"
	"testing"
)

func TestState(t *testing.T) {
	name := "test.txt"
	err := os.MkdirAll(FolderOf(name), os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(FolderOf(name))

	expect := &State{
		Url:  "https://test.com/test.txt",
		Name: name,
		Parts: []*Part{
			{
				Path:      "./part1",
				Current:   123,
				RangeFrom: 0,
				RangeTo:   345,
			},
			{
				Path:      "./part2",
				Current:   123,
				RangeFrom: 0,
				RangeTo:   345,
			},
		},
		Cookies: []*http.Cookie{
			{Name: "hello", Value: "world"},
			{Name: "foo", Value: "bar"},
		},
	}

	err = expect.Save()
	if err != nil {
		t.Fatal(err)
	}

	got, err := Resume(expect.Name)
	if err != nil {
		t.Fatal(err)
	}

	if !stateEqual(got, expect) {
		t.Errorf("got[%#v] isn't expect[%#v]\n", got, expect)
	}
}

func stateEqual(a, b *State) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil && b != nil {
		return false
	}
	if a != nil && b == nil {
		return false
	}

	if a.Url != b.Url {
		return false
	}
	if a.Name != b.Name {
		return false
	}
	if len(a.Parts) != len(b.Parts) {
		return false
	}
	for i := range a.Parts {
		if !reflect.DeepEqual(a.Parts[i], b.Parts[i]) {
			return false
		}
	}

	return true
}
