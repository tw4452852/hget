package main

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestGetFileName(t *testing.T) {
	for i, c := range []struct {
		cd, url, expect string
	}{
		{"", "", "."},
		{"", "http://test.com/1.txt", "1.txt"},
		{"", "http://test.com/../1.txt", "1.txt"},
		{"", "http://test.com/", "test.com"},
		{"", "http://test.com", "test.com"},
		{`test; filename="test.txt"`, "", "test.txt"},
		{`test; filename = "test.txt"; test2`, "", "test.txt"},
		{`test; filename=test.txt`, "http://test.com/1.txt", "1.txt"},
		{`test; filename='test.txt'`, "http://test.com/1.txt", "1.txt"},
		{`test; filename="test.txt'`, "http://test.com/1.txt", "1.txt"},
		{`test; filename='test.txt"`, "http://test.com/1.txt", "1.txt"},
		{`test`, "http://test.com/1.txt", "1.txt"},
		{`test; filename="test.txt"`, "http://test.com/1.txt", "test.txt"},
	} {
		if got := getFileName(c.cd, c.url); got != c.expect {
			t.Errorf("case %d failed: expect[%s], got[%s]", i, c.expect, got)
		}
	}
}

func TestPartCalculate(t *testing.T) {
	for i, c := range []struct {
		par, len int64
		filename string
		expect   []*Part
	}{
		{1, 100, "test", []*Part{
			{Path: filepath.Join(FolderOf("test"), "test.part0"), Current: 0, RangeFrom: 0, RangeTo: 100}}},
		{2, 100, "test", []*Part{
			{Path: filepath.Join(FolderOf("test"), "test.part0"), Current: 0, RangeFrom: 0, RangeTo: 49},
			{Path: filepath.Join(FolderOf("test"), "test.part1"), Current: 50, RangeFrom: 50, RangeTo: 100}}},
	} {
		got := partCalculate(c.par, c.len, c.filename)
		if len(got) != len(c.expect) {
			t.Errorf("case %d failed: length not equal got[%d], expect[%d]", i, len(got), len(c.expect))
			continue
		}
		for p := range got {
			if !reflect.DeepEqual(got[p], c.expect[p]) {

				t.Errorf("case %d failed: %dth entry not equal got[%#v], expect[%#v]",
					i, p, got[p], c.expect[p])
				break
			}
		}
	}
}
