package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"

	windowCRC32 "github.com/sjpotter/window-crc32"
)

var (
	file           = flag.String("file", "", "file to search in")
	size           = flag.Uint("size", 0, "size in bytes for moving crc window")
	hash           = flag.String("hash", "", "hash value in hex to search for")
	checkpoint     = flag.Uint("checkpoint", 0, "checkpoint roll table state every N loops")
	checkpointFile = flag.String("checkpointFile", "", "roll table checkpoint file to (re)store state with")
	threads        = flag.Uint("threads", 1, "number of threads to use to compute rolling table")
	all            = flag.Bool("all", false, "find all matches")
)

func main() {
	flag.Parse()

	if *file == "" || *size == 0 || *hash == "" || *checkpointFile == "" {
		flag.Usage()
		os.Exit(1)
	}

	fh, err := os.Open(*file)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	defer func() {
		err := fh.Close()
		if err != nil {
			fmt.Printf("Failed to close fh: %v\n", err)
		}
	}()

	crc, err := strconv.ParseInt(*hash, 16, 64)
	if err != nil {
		fmt.Println(err)
		os.Exit(3)
	}

	var s windowCRC32.Serializer

	if *checkpoint == 0 {
		s = windowCRC32.NewJsonSerializer(0)
		data, err := ioutil.ReadFile(*checkpointFile)
		if err != nil {
			fmt.Println(err)
			os.Exit(4)
		}
		err = s.ReadIn(data)
		if err != nil {
			fmt.Println(err)
			os.Exit(5)
		}
	} else if *checkpoint > 0 {
		fmt.Printf("Creating a new checkpoint state file with output every %v loops\n", *checkpoint)
		s = windowCRC32.NewJsonSerializer(*checkpoint)
	} else {
		fmt.Printf("checkpoint (%v) needs to be greater than 0\n", *checkpoint)
		os.Exit(6)
	}

	// save state at end
	defer func(s windowCRC32.Serializer, fileName string) {
		data, err := s.WriteOut()
		if err != nil {
			fmt.Println(err)
			os.Exit(8)
		}
		err = ioutil.WriteFile(fileName, data, 0644)
		if err != nil {
			fmt.Println(err)
			os.Exit(9)
		}
	}(s, *checkpointFile)

	crc32 := windowCRC32.NewCRCThreaded(windowCRC32.CrcPoly, *size, s, *threads)

	var pos uint
	buf := make([]byte, *size)
	for {
		_, err := fh.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println(err)
			os.Exit(7)
		}

		for _, b := range buf {
			pos++
			crc32.UpdateCRC32(b)
			if pos > *size {
				if crc32.Finish() == uint(crc) {
					lastByte := pos
					firstByte := lastByte - *size
					fmt.Printf("found data at %v to %v\n", firstByte, lastByte)
					fmt.Printf("dd if=%v of=test skip=%v bs=1 count=%v\n", *file, firstByte, *size)
					fmt.Println()
					if !*all {
						return
					}
				}
			}
		}
	}
}
