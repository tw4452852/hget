package main

import (
	"net/http"
	"reflect"
	"testing"
)

func TestGetCookies(t *testing.T) {
	for i, c := range []struct {
		input  string
		expect []*http.Cookie
	}{
		{"nothing", []*http.Cookie{}},
		{"foo=bar", []*http.Cookie{{Name: "foo", Value: "bar"}}},
		{"foo=bar;", []*http.Cookie{{Name: "foo", Value: "bar"}}},
		{"foo=bar;hello=world", []*http.Cookie{
			{Name: "foo", Value: "bar"},
			{Name: "hello", Value: "world"}}},
	} {
		got := getCookies(c.input)
		if len(got) != len(c.expect) {
			t.Errorf("case %d failed: length not equal, got[%d], expect[%d]",
				i, len(got), len(c.expect))
			continue
		}
		for j := range got {
			if !reflect.DeepEqual(got[j], c.expect[j]) {
				t.Errorf("case %d failed: %dth entry not equal, got[%#v], c.expect[%#v]",
					i, j, got[j], c.expect[j])
				break
			}
		}
	}
}
