package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"
	"sync"
)

var (
	wg = sync.WaitGroup{}
)

var (
	inputPath  = flag.String("input", "", "Path to the input directory/file")
	outputPath = flag.String("output", "output.txt", "Output file path")
	needle     = flag.String("needle", "", "Needle in a haystack")
)

func main() {
	flag.Parse()

	if *inputPath == "" {
		log.Fatalf("no \"input\" argument was provided")
	}

	if *needle == "" {
		log.Fatal("no \"needle\" argument was provided")
	}

	inputFile, err := os.Open(*inputPath)
	if err != nil {
		log.Fatalf("unable to open input directory/file: %s", err)
	}

	fileStat, _ := inputFile.Stat()

	results := make(chan string)
	done := make(chan bool)

	go WriteMatchesToFile(results, done, *outputPath)

	if fileStat.IsDir() {
		dirFiles, err := inputFile.ReadDir(-1)
		if err != nil {
			log.Fatalf("unexpected error during directory read: %s", err)
		}

		for _, file := range dirFiles {
			wg.Add(1)
			go SearchInFile(path.Join(*inputPath, file.Name()), *needle, results)
		}

	} else {
		wg.Add(1)
		go SearchInFile(*inputPath, *needle, results)
	}

	wg.Wait()
	close(results)

	<-done
	close(done)
}

func SearchInFile(filepath string, needle string, results chan string) {
	defer wg.Done()

	file, _ := os.Open(filepath)
	defer file.Close()

	scanner := bufio.NewReader(file)

	for lineNum := 1; ; lineNum++{
		line, err := scanner.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}

			log.Fatalf("unexpected error during file reading: %s", err)
		}

		if strings.Contains(line, needle) {
			fmt.Printf("[%s]: match found at line %d\n", filepath, lineNum)
			results <- line
		}
	}
}

func WriteMatchesToFile(results chan string, done chan bool, filepath string) {
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
		log.Fatalf("error during output file creation: %s", err)
	}
	defer file.Close()


	for line := range results {
		_, err := file.WriteString(line)
		if err != nil {
			log.Fatalf("unable to write to output file: %s", err)
		}
	}

	done <- true
}