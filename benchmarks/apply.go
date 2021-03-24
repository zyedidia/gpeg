package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
)

var dir = flag.String("dir", "", "dir to apply command to")
var suffix = flag.String("suffix", "", "file suffix")

func main() {
	flag.Parse()

	// target := flag.Args()[0]

	err := filepath.Walk(*dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), *suffix) && info.Size() > 50000 {
			args := []string{"java.lua", path}
			cmd := exec.Command("lua5.1", args...)
			buf := &bytes.Buffer{}
			cmd.Stdout = buf
			cmd.Run()
			if !strings.Contains(buf.String(), "nil") {
				lines := strings.Split(buf.String(), "\n")
				fmt.Println(lines[2], lines[1])
			}
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}
