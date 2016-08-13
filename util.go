package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

func FatalCheck(err error) {
	if err != nil {
		Errorf("%v", err)
		os.Exit(1)
	}
}

func MkdirIfNotExist(folder string) error {
	if _, err := os.Lstat(folder); err != nil {
		if err = os.MkdirAll(folder, 0700); err != nil {
			return err
		}
	}
	return nil
}

func RemoveAll(dir string) error {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Remove all in %s?[y/n]:", dir)
	response, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	response = strings.ToLower(strings.TrimSpace(response))
	if response[0] == 'n' {
		return nil
	}
	return os.RemoveAll(dir)
}

func ExistDir(folder string) bool {
	_, err := os.Lstat(folder)
	return err == nil
}

const (
	dataFolder    = ".hget/"
	stateFileName = "state.json"
)

var homeDir string

func init() {
	user, err := user.Current()
	FatalCheck(err)
	homeDir = user.HomeDir
}

func FolderOf(filename string) string {
	return filepath.Join(homeDir, dataFolder, filename)
}

func TaskPrint() error {
	downloading, err := ioutil.ReadDir(filepath.Join(homeDir, dataFolder))
	if err != nil {
		return err
	}

	folders := make([]string, 0)
	for _, d := range downloading {
		if d.IsDir() {
			folders = append(folders, d.Name())
		}
	}

	folderString := strings.Join(folders, "\n")
	Printf("Currently on going download: \n")
	fmt.Println(folderString)

	return nil
}
