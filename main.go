package main

import (
	"os"
	"io"
	"sync"
	"math/rand"
	"time"
	"io/ioutil"
	"path/filepath"
	"net/http"
	"strings"
	"log"
	
	"github.com/fatih/color"
)

type fnColor func(w io.Writer, format string, a ...interface{})

type CarrotLive struct {
	mutex  sync.RWMutex
	Frames []string
	index  int
	
	colors []fnColor
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func New(p string) (*CarrotLive, error) {
	carrot := &CarrotLive{
		colors: []fnColor{
			color.New(color.FgRed).FprintfFunc(),
			color.New(color.FgYellow).FprintfFunc(),
			color.New(color.FgGreen).FprintfFunc(),
			color.New(color.FgBlue).FprintfFunc(),
			color.New(color.FgMagenta).FprintfFunc(),
			color.New(color.FgCyan).FprintfFunc(),
			color.New(color.FgWhite).FprintfFunc(),
		},
	}
	
	if err := carrot.readFiles(p); err != nil {
		return nil, err
	}
	
	return carrot, nil
}

func (c *CarrotLive) readFiles(p string) (error) {
	fnRead := func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		
		bs, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		
		c.Frames = append(c.Frames, string(bs))
		return nil
	}
	
	return filepath.Walk(p, fnRead)
}

func (c *CarrotLive) NextFrame() string {
	defer c.mutex.RUnlock()
	c.mutex.RLock()
	c.index = c.index + 1
	
	if len(c.Frames) <= c.index {
		c.index = 0
	}
	
	return c.Frames[c.index]
}

func (c *CarrotLive) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if !strings.Contains(req.Header.Get("user-agent"), "curl") {
		io.WriteString(rw, "Use the curl to access this link.")
		return
	}
	
	hj, ok := rw.(http.Hijacker)
	if !ok {
		http.Error(rw, "webserver doesn't support hijacking", http.StatusInternalServerError)
		return
	}
	
	conn, bufrw, err := hj.Hijack()
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	
	defer conn.Close()
	
	for {
		select {
		case <-time.After(70 * time.Millisecond):
			c.colors[rand.Intn(len(c.colors))](bufrw, "\033[2J\033[H")
			c.colors[rand.Intn(len(c.colors))](bufrw, c.NextFrame())
			c.colors[rand.Intn(len(c.colors))](bufrw, "\nWelcome PARROT LIVE Mother fucker!!!\n")
			bufrw.Flush()
		}
	}
}

func main() {
	c, err := New("./frames")
	if err != nil {
		log.Fatalf("Error to read frames because: %v", err)
		os.Exit(1)
	}
	
	if err = http.ListenAndServe(":8080", c); err != nil {
		log.Fatalf("Error to start server because: %v", err)
		os.Exit(1)
	}
}
