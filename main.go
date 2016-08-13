package main

import (
	"flag"

	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
)

type options struct {
	conn   int
	output string
}

func main() {
	var ops options
	flag.IntVar(&ops.conn, "n", runtime.NumCPU(), "concurrent connection")
	flag.StringVar(&ops.output, "o", "", "output path")

	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		usage()
		os.Exit(1)
	}

	switch args[0] {
	case "tasks":
		FatalCheck(TaskPrint())
		return
	case "resume":
		if len(args) < 2 {
			Errorln("downloading task name is required")
			usage()
			os.Exit(1)
		}
		state, err := Resume(args[1])
		FatalCheck(err)
		FatalCheck(execute(state.Url, state, ops))
		return
	}

	//otherwise is hget <URL> command
	if ExistDir(FolderOf(args[0])) {
		Warnf("Downloading task already exist, remove first \n")
		FatalCheck(os.RemoveAll(FolderOf(args[0])))
	}
	FatalCheck(execute(args[0], nil, ops))
}

func execute(url string, state *State, ops options) error {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM)

	var (
		isInterrupted = false
		doneChan      = make(chan struct{})
		errorChan     = make(chan error)
		interruptChan = make(chan struct{})
		downloader    *HttpDownloader
		err           error
	)

	if state != nil {
		downloader, err = RestoreHttpDownloader(state)
	} else {
		downloader, err = NewHttpDownloader(url, ops.conn)
	}
	if err != nil {
		return err
	}

	go downloader.Do(doneChan, errorChan, interruptChan)

	for {
		select {
		case <-signalChan:
			if !isInterrupted {
				isInterrupted = true
				close(interruptChan)
			}
		case err := <-errorChan:
			Errorf("%v", err)
			isInterrupted = true
		case <-doneChan:
			s := downloader.Capture()
			if isInterrupted {
				if downloader.resumable {
					Printf("Interrupted, saving state ... \n")
					return s.Save()
				} else {
					Warnf("Interrupted, but downloading url is not resumable, silently die")
					return nil
				}
			} else {
				parts := s.Parts
				files := make([]string, 0, len(parts))
				for i := range parts {
					// fmt.Printf("get a part: %#v\n", parts[i])
					files = append(files, parts[i].Path)
				}
				err := s.Save()
				if err != nil {
					return err
				}
				if ops.output == "" {
					ops.output = filepath.Join(".", s.Name)
				}
				err = JoinFile(files, ops.output)
				if err != nil {
					return err
				}
				dir := filepath.Dir(files[0])
				Printf("Deleting temp files in %s\n", dir)
				return RemoveAll(dir)
			}
		}
	}
}

func usage() {
	Printf(`Usage:
hget [-n connection] -o path [URL]
hget tasks
hget resume -o path [TaskName]
`)
}
