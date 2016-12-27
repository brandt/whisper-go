package main

import (
	"flag"
	"fmt"
	"github.com/brandt/whisper-go/whisper"
	"log"
	"os"
	"path/filepath"
)

func usage() {
	fmt.Println("Usage: whisper-aggregate DEST MATCH")
	os.Exit(1)
}

func main() {
	flag.Parse()
	if flag.NArg() != 2 {
		usage()
	}

	dstPath := flag.Args()[0]
	match := flag.Args()[1]
	srcFiles, err := filepath.Glob(match)
	if err != nil {
		log.Fatalln("error looking for matching src files:", err)
	}

	// We need at least one matching source file
	if len(srcFiles) == 0 {
		log.Fatal("no matching files")
	}

	var dst *whisper.Whisper

	if _, err := os.Stat(dstPath); err == nil {
		// The destination file already exists; just open it
		d, openErr := whisper.Open(dstPath)
		if openErr != nil {
			log.Fatalln("error opening dest file:", openErr)
		}
		dst = d
	} else {
		// The destination file doesn't yet exist.

		// Create a new whisper file with the same structure as the first source
		// file we found.
		ref, openErr := whisper.Open(srcFiles[0])
		if openErr != nil {
			log.Fatalln("error opening dest file:", openErr)
		}
		d, cloneErr := ref.Clone(dstPath)
		if cloneErr != nil {
			log.Fatalf("error cloning src into new dest. src: %s: %s\n", srcFiles[0], cloneErr)
		}
		ref.Close()
		dst = d
	}

	// For each matching source file, add it to our dst file
	for _, f := range srcFiles {
		log.Println("Source file:", f)

		src, err := whisper.Open(f)
		if err != nil {
			log.Fatalf("error opening source file: %s: %s\n", f, err)
		}

		errRef := dst.AddWhisper(src)
		if errRef != nil {
			log.Fatalln("error writing to destination file:", errRef)
		}

		src.Close()
	}

	dst.Close()

	return
}
