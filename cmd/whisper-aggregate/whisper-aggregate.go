// Offline aggregation of multiple Whisper files.
//
// Currently, there are two aggregation modes:
// 1. "sum" which works like sumSeries
// 2. "delta" which works like nonNegativeDerivative
//

package main

import (
	"flag"
	"fmt"
	"github.com/brandt/whisper-go/whisper"
	"log"
	"os"
	"path/filepath"
)

type Mode int

const (
	SumSeries Mode = iota
	Delta
)

var mode Mode

func init() {
	var modeFlag string

	flag.StringVar(&modeFlag, "mode", "sum", "aggregation mode")
	flag.Parse()

	switch modeFlag {
	case "sum":
		mode = SumSeries
	case "delta":
		mode = Delta
	default:
		log.Fatalln("invalid mode selected")
	}

}

func usage() {
	fmt.Println("Usage: whisper-aggregate DEST MATCH")
	os.Exit(1)
}

func main() {
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

	// Exit with an error if the dest file already exists
	if _, err := os.Stat(dstPath); err == nil {
		log.Fatalln("dest file already exists:", dstPath)
	}

	// Create a new file with the same structure as the first source file we found
	ref, openErr := whisper.Open(srcFiles[0])
	if openErr != nil {
		log.Fatalln("error opening dest file:", openErr)
	}
	dst, cloneErr := ref.Clone(dstPath)
	if cloneErr != nil {
		log.Fatalf("error cloning src into new dest. src: %s: %s\n", srcFiles[0], cloneErr)
	}
	ref.Close()

	// For each matching source file, add it to our dst file
	for _, f := range srcFiles {
		log.Println("Source file:", f)

		src, err := whisper.Open(f)
		if err != nil {
			log.Fatalf("error opening source file: %s: %s\n", f, err)
		}

		var errRef error

		switch mode {
		case Delta:
			errRef = dst.AddWhisperDelta(src)
		case SumSeries:
			errRef = dst.AddWhisper(src)
		default:
			errRef = dst.AddWhisper(src)
		}

		if errRef != nil {
			log.Fatalln("error writing to destination file:", errRef)
		}

		src.Close()
	}

	dst.Close()

	return
}
