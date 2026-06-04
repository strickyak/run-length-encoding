// 0 means stop.
// 1..127 mean copy that many following bytes verbatim.
// 128..255 mean copy the following byte (n-128) times.
//
// Usage:
//
//	rle < plain > compressed
//	rle -d < compressed > restored-plain
//
// This is for files small enough to fit both into RAM.
// func `compressRLE()` written by Gemini.
// The rest by Strick Yak.
package main

import (
	"flag"
	"log"
	"os"
)

var D = flag.Bool("d", false, "use -d to decompressRLE")

func decompressRLE(compressed []byte) (z []byte) {
	for i := 0; i < len(compressed); i++ {
		var a byte = compressed[i]
		switch {
		case a == 0:
			break
		case a < 128:
			for j := 0; j < int(a); j++ {
				z = append(z, compressed[i+1+j])
			}
			i += int(a)
		default:
			for j := 0; j < int(a)-128; j++ {
				z = append(z, compressed[i+1])
			}
			i++
		}
	}
	return
}

// compressRLE implements the custom Run-Length Encoding logic
func compressRLE(raw []byte) []byte {
	var comp []byte
	var rawBuf []byte

	// Helper to flush accumulated uncompressed bytes
	flushRaw := func() {
		for len(rawBuf) > 0 {
			chunkSize := len(rawBuf)
			if chunkSize > 127 {
				chunkSize = 127
			}
			comp = append(comp, byte(chunkSize))
			comp = append(comp, rawBuf[:chunkSize]...)
			rawBuf = rawBuf[chunkSize:]
		}
	}

	i := 0
	n := len(raw)
	for i < n {
		// Look ahead to count repeats of the current byte
		count := 1
		for i+count < n && raw[i+count] == raw[i] && count < 127 {
			count++
		}

		if count >= 3 {
			// We found a run of 3 or more. Flush any pending uncompressed data first.
			flushRaw()

			// Append repeat prefix (128 + count) and the byte value itself
			comp = append(comp, byte(128+count), raw[i])
			i += count
		} else {
			// Less than 3 repeats, treat as raw data
			rawBuf = append(rawBuf, raw[i])
			if len(rawBuf) == 127 {
				flushRaw() // Max chunk size reached, force flush
			}
			i++
		}
	}

	// Flush any remaining uncompressed data at the end of the image
	flushRaw()

	// Append the special terminating 0
	comp = append(comp, 0)

	return comp
}

func main() {
	flag.Parse()
	input, err := os.ReadFile("/dev/stdin")
	if err != nil {
		log.Fatalf("Cannot read /dev/stdin: %v", err)
		panic(0)
	}

	var output []byte
	if *D {
		output = decompressRLE(input)
	} else {
		output = compressRLE(input)
	}

	_, err = os.Stdout.Write(output)
	if err != nil {
		log.Fatalf("Cannot write stdout: %v", err)
		panic(0)
	}
}
