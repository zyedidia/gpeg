package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var dir = flag.String("dir", "", "dir to apply command to")
var suffix = flag.String("suffix", "", "file suffix")

const cutoff = 50000

func main() {
	flag.Parse()

	target := flag.Args()[0]

	err := filepath.Walk(*dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), *suffix) && info.Size() >= cutoff {
			args := []string{path}
			cmd := exec.Command(target, args...)
			cmd.Stdout = os.Stdout
			cmd.Run()
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}
