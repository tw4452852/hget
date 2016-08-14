package main

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
)

type State struct {
	Url   string
	Name  string
	Parts []*Part
}

type Part struct {
	Path      string
	Current   int64
	RangeFrom int64
	RangeTo   int64
}

func (s *State) Save() error {
	//save state file
	j, err := json.Marshal(s)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filepath.Join(FolderOf(s.Name), stateFileName), j, 0644)
}

func Resume(task string) (*State, error) {
	file := filepath.Join(FolderOf(task), stateFileName)
	Printf("Getting data from %s\n", file)
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	s := new(State)
	err = json.Unmarshal(bytes, s)
	return s, err
}
