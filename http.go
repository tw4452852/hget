package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/cheggaaa/pb"
	"github.com/fatih/color"
)

var (
	tr = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client = &http.Client{Transport: tr}
)

type HttpDownloader struct {
	url       string
	name      string
	size      int64
	parts     []*Part
	resumable bool
}

func RestoreHttpDownloader(s *State) (*HttpDownloader, error) {
	return &HttpDownloader{
		url:       s.Url,
		size:      s.Parts[len(s.Parts)-1].RangeTo,
		parts:     s.Parts,
		name:      s.Name,
		resumable: true,
	}, nil
}

func NewHttpDownloader(url string, par int) (*HttpDownloader, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept-Encoding", "gzip")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	fmt.Printf("%#v\n", resp)

	const (
		acceptRangeHeader        = "Accept-Ranges"
		contentLengthHeader      = "Content-Length"
		contentDispositionHeader = "Content-Disposition"
	)

	if resp.Header.Get(acceptRangeHeader) == "" {
		Warnf("Target url is not supported range download\n")
		//fallback to par = 1
		par = 1
	}

	resumable := true
	//get download range
	clen := resp.Header.Get(contentLengthHeader)
	if clen == "" {
		Warnf("Target url not contain Content-Length header\n")
		clen = "-1"
		resumable = false
	}

	size, err := strconv.ParseInt(clen, 10, 64)
	if err != nil {
		return nil, err
	}

	fileName := getFileName(resp.Header.Get(contentDispositionHeader), url)
	fmt.Printf("filename: %q\n", fileName)
	folder := FolderOf(fileName)
	if ExistDir(folder) {
		state, err := Resume(fileName)
		if err != nil {
			return nil, err
		}
		return RestoreHttpDownloader(state)
	}

	if err := MkdirIfNotExist(folder); err != nil {
		return nil, err
	}

	return &HttpDownloader{
		url:       url,
		size:      size,
		name:      fileName,
		parts:     partCalculate(int64(par), size, fileName),
		resumable: resumable,
	}, nil
}

func getFileName(cd, url string) string {
	if start := strings.Index(cd, "filename"); start > 0 {
		start := strings.IndexByte(cd[start:], '"') + start
		end := strings.IndexByte(cd[start+1:], '"') + start
		if start != -1 && end != -1 {
			return cd[start+1 : end]
		}
	}
	return path.Base(url)
}

func partCalculate(par, len int64, filename string) []*Part {
	ret := make([]*Part, 0)
	for j := int64(0); j < par; j++ {
		from := (len / par) * j
		var to int64
		if j < par-1 {
			to = (len/par)*(j+1) - 1
		} else {
			to = len
		}

		ret = append(ret, &Part{
			Path:      filepath.Join(FolderOf(filename), fmt.Sprintf("%s.part%d", filename, j)),
			Current:   from,
			RangeFrom: from,
			RangeTo:   to})
	}
	return ret
}

func (d *HttpDownloader) Capture() *State {
	return &State{
		Url:   d.url,
		Name:  d.name,
		Parts: d.parts,
	}
}

func (d *HttpDownloader) Do(doneChan chan struct{}, errorChan chan error, interruptChan chan struct{}) {
	size := d.size
	switch {
	case size <= 0:
		Printf("Dowload size: not specified\n")
	case 0 < size && size < 1024:
		Printf("Download target size: %dB\n", size)
	case 1024 <= size && size < 1024*1024:
		Printf("Download target size: %.1f KB\n", float64(size)/1024)
	case 1024*1024 <= size && size < 1024*1024*1024:
		Printf("Download target size: %.1f MB\n", float64(size)/1024/1024)
	default:
		Printf("Download target size: %.1f GB\n", float64(size)/1024/1024/1024)
	}

	bars := make([]*pb.ProgressBar, 0)
	for i, part := range d.parts {
		newbar := pb.New64(part.RangeTo - part.RangeFrom).SetUnits(pb.U_BYTES).Prefix(color.YellowString(fmt.Sprintf("part%d", i+1)))
		newbar.ShowSpeed = true
		bars = append(bars, newbar)
	}
	barpool, err := pb.StartPool(bars...)
	FatalCheck(err)

	var ws sync.WaitGroup
	for i, part := range d.parts {
		// do nothing if we are done
		if part.RangeTo <= part.Current {
			bars[i].Set64(bars[1].Total)
			bars[i].Finish()
			continue
		}

		ws.Add(1)
		go func(d *HttpDownloader, loop int) {
			defer ws.Done()

			bar := bars[loop]
			part := d.parts[loop]
			bar.Set64(part.Current - part.RangeFrom)

			//send request
			req, err := http.NewRequest("GET", d.url, nil)
			if err != nil {
				errorChan <- err
				return
			}

			if len(d.parts) > 1 { // support range download just in case parallel factor is over 1
				ranges := fmt.Sprintf("bytes=%d-%d", part.Current, part.RangeTo)
				if loop == (len(d.parts))-1 {
					ranges = fmt.Sprintf("bytes=%d-", part.Current)
				}
				req.Header.Add("Range", ranges)
				if err != nil {
					errorChan <- err
					return
				}
			}

			//write to file
			resp, err := client.Do(req)
			if err != nil {
				errorChan <- err
				return
			}
			defer resp.Body.Close()
			f, err := os.OpenFile(part.Path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)

			defer f.Close()
			if err != nil {
				errorChan <- err
				return
			}

			writer := io.MultiWriter(f, bar)

			//make copy interruptable by copy 100 bytes each loop
			for {
				select {
				case <-interruptChan:
					return
				default:
					written, err := io.CopyN(writer, resp.Body, 100)
					part.Current += written
					if err != nil {
						if err != io.EOF {
							errorChan <- err
						}
						bar.Finish()
						return
					}
				}
			}
		}(d, i)
	}

	ws.Wait()
	barpool.Stop()
	doneChan <- struct{}{}
}
