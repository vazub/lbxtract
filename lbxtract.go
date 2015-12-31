// Copyright 2015-2016 Vasyl Zubko. All rights reserved.
// Distributed under the terms of the MIT License, 
// that can be found in the LICENSE file.

// LBXtract reads and extracts data from the proprietary .LBX file format
// used in some Simtex/Microprose games, namely Master of Magic,
// Master of Orion and Master of Orion 2
package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	signLBX = []byte{0xAD, 0xFE, 0x00, 0x00}
	signSMK = []byte{0x53, 0x4D, 0x4B, 0x32}
	signVOC = []byte{0x43, 0x72, 0x65, 0x61}
	signWAV = []byte{0x52, 0x49, 0x46, 0x46}
	signXMI = []byte{0x46, 0x4F, 0x52, 0x4D}
	signDrv = []byte{0x2D, 0x00, 0x43, 0x6F}
	signMO2 = []byte{0x00, 0x08, 0x00, 0x00}
	lbxFile string
)

// decode() transforms an n-byte slice into an actual decimal
// data representation. Currently only 2 sizes of byte slice are
// supported: 2 and 4.
func decode(b []byte) int {
	var res2 uint16
	var res4 uint32
	btr := bytes.NewReader(b)
	switch {
	case len(b) == 2:
		if err := binary.Read(btr, binary.LittleEndian, &res2); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		return int(res2)
	case len(b) == 4:
		if err := binary.Read(btr, binary.LittleEndian, &res4); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		return int(res4)
	default:
		fmt.Fprintln(os.Stderr, "Unsupported length of byte slice to decode")
		os.Exit(1)
	}
	return 1
}

// checkType() looks for specific signatures inside the file and returns
// the named string to represent the type of files packed within it
func checkType(b []byte) string {
	offstart := decode(b[8:12])
	switch {
	case bytes.Contains(b[:4], signSMK):
		return "SMK"
	case bytes.Contains(b[offstart+16:offstart+20], signVOC):
		return "VOC"
	case bytes.Contains(b[offstart+16:offstart+20], signXMI):
		return "XMI"
	case bytes.Contains(b[offstart:offstart+4], signDrv): // SNDDRV.LBX has sound drivers at start and 2 XMI files at the end, so the type is mixed
		return "data+XMI"
	case bytes.Contains(b[offstart:offstart+4], signWAV): //?
		return "WAV"
	case bytes.Contains(b[2:6], signLBX):
		return "LBX"
	}
	return "unknown"
}

// getMeta() extracts file names and descriptions. 512 is the start of the
// names/descriptions block, 32 is their combined length: (8 + null)
// characters for file name and (22 + null) characters for file
// descriptions. Some files may have missing names or descriptions or both.
// Also, Master of Orion 2 files have garbage in this offset range and
// seemingly no names or descriptions whatsoever
func getMeta(b []byte, refc int, off []int) ([]string, []string) {
	var names, desc []string
	var nst, dst = 512, 521
	for i := 0; i < refc; i++ {
		if bytes.Contains(b[8:12], signMO2) {
			names = append(names, "")
			desc = append(desc, "")
			continue
		}
		if nst != off[0] {
			names = append(names, strings.Replace(string(b[nst:nst+8]), string(0), "", -1))
			desc = append(desc, strings.Replace(string(b[dst:dst+22]), string(0), "", -1))
			nst += 32
			dst += 32
		} else {
			names = append(names, "")
			desc = append(desc, "")
		}
	}
	return names, desc
}

func main() {

	var dirPath string
	switch {
	case len(os.Args) == 1:
		dirPath = filepath.Dir(os.Args[0])
	case len(os.Args) == 2:
		dirPath = os.Args[1]
	case len(os.Args) > 2:
		fmt.Fprintln(os.Stderr, "Provide exactly one folder to process")
	}

	err := os.Chdir(dirPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	dirList, err := filepath.Glob("*.[l|L][b|B][x|X]")
	if err != nil {
		panic(err)
	}

	if dirList == nil {
		fmt.Fprintln(os.Stderr, "No .LBX files found at this location")
	}

	for _, v := range dirList {
		lbxFile = strings.TrimSuffix(strings.ToUpper(v), ".LBX")
		outFolderPath := "EXTRACTED/" + lbxFile

		file, err := ioutil.ReadFile(v)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}

		refCount := decode(file[:2])

		var off = 8 // Get resource offsets. First resource offset reference starts at 8
		var fileOffsets []int
		for i := 1; i <= refCount; i++ {
			fileOffsets = append(fileOffsets, decode(file[off:off+4]))
			off += 4
		}

		err = os.MkdirAll(outFolderPath, 0777)
		if err != nil {
			panic(err)
		}

		currDir, err := os.Getwd()
		if err != nil {
			panic(err)
		}

		err = os.Chdir(outFolderPath)
		if err != nil {
			panic(err)
		}

		ftype := checkType(file)
		if ftype == "SMK" {
			err = ioutil.WriteFile(lbxFile+".SMK", file[:], 0666)
			if err != nil {
				panic(err)
			}
			err = os.Chdir(currDir)
			if err != nil {
				panic(err)
			}
			fmt.Printf("Extracted 1 file(s) from %s.LBX\n", lbxFile)
			continue
		}

		names, desc := getMeta(file, refCount, fileOffsets)

		var outFileName string
		var outFile []byte
		for i, v := range fileOffsets {

			if bytes.Contains(file[v:v+4], signWAV) {
				outFileName = strconv.Itoa(i+1) + "_" + names[i] + "_" + strings.Replace(desc[i], "/", "_", -1) + ".WAV"

				if i == len(fileOffsets)-1 {
					outFile = file[v:len(file)]
				} else {
					outFile = file[v:fileOffsets[i+1]]
				}
				err = ioutil.WriteFile(outFileName, outFile, 0666)
				if err != nil {
					panic(err)
				}
				continue
			}

			switch ftype {
			case "VOC":
				outFileName = strconv.Itoa(i+1) + "_" + names[i] + "_" + strings.Replace(desc[i], "/", "_", -1) + ".VOC"

				if i == len(fileOffsets)-1 {
					outFile = file[v+16 : len(file)]
				} else {
					outFile = file[v+16 : fileOffsets[i+1]]
				}
				err = ioutil.WriteFile(outFileName, outFile, 0666)
				if err != nil {
					panic(err)
				}

			case "XMI":
				outFileName = strconv.Itoa(i+1) + "_" + names[i] + "_" + strings.Replace(desc[i], "/", "_", -1) + ".XMI"

				if i == len(fileOffsets)-1 {
					outFile = file[v+16 : len(file)]
				} else {
					outFile = file[v+16 : fileOffsets[i+1]]
				}
				err = ioutil.WriteFile(outFileName, outFile, 0666)
				if err != nil {
					panic(err)
				}

			case "data+XMI":
				if i == len(fileOffsets)-2 {
					outFile = file[v+16 : fileOffsets[i+1]]
				} else if i == len(fileOffsets)-1 {
					outFile = file[v+16 : len(file)]
				} else {
					continue
				}
				outFileName = strconv.Itoa(i+1) + "_" + names[i] + "_" + strings.Replace(desc[i], "/", "_", -1) + ".XMI"
				err = ioutil.WriteFile(outFileName, outFile, 0666)
				if err != nil {
					panic(err)
				}

			case "LBX":
				outFileName = strconv.Itoa(i+1) + "_" + names[i] + "_" + strings.Replace(desc[i], "/", "_", -1)

				if i == len(fileOffsets)-1 {
					outFile = file[v:len(file)]
				} else {
					outFile = file[v:fileOffsets[i+1]]
				}
				err = ioutil.WriteFile(outFileName, outFile, 0666)
				if err != nil {
					panic(err)
				}
			}
		}

		fmt.Printf("Extracted %d file(s) from %s.LBX\n", refCount, lbxFile)

		err = os.Chdir(currDir)
		if err != nil {
			panic(err)
		}
	}
}