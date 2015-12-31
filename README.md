# LBXtract #
[![GoDoc](https://godoc.org/github.com/vazub/lbxtract?status.svg)](https://godoc.org/github.com/vazub/lbxtract)

LBXtract is an application designed to read and extract data from the proprietary .LBX archive format, used by Simtex/Microprose in some of their old games, namely those in the Master of Magic/Orion series.

The tool itself is console-based and written in Go programming language, and therefore should be portable among most of current operating systems. It was developed under Mac OS X, but other *nix as well as Windows users should be able to use it without issues as well.

This application has been made possible thanks to the efforts of those who had previously investigated and documented major parts of the LBX format, as provided by [this wiki](http://www.shikadi.net/moddingwiki/LBX_Format) page.

## Games supported ##
- Master of Magic (1994)
- Master of Orion (1993)
- Master of Orion 2: Battle at Antares (1996)

## Usage ##
It is assumed, that you have a working Go environment already set up. It not, consult the official guidelines at http://golang.org

Get the source code:

```
$ go get github.com/vazub/lbxtract
```
Enter the source code folder and build the executable:
```
$ go build
```
Now copy the produced executable into the folder containing your game data (.lbx files) and run it from there.

Otherwise, you can provide an explicit path to your game data folder as an argument to the executable, like this:
```
$ ./lbxtract path-to-data-folder
```
You will find your output data in <**EXTRACTED/name-of-original-lbx**> folders within the data directory you pointed LBXtract to. The files will be named according to the original metadata info provided and given specific file extentions (.SMK, .VOC, .WAV, .XMI) when applicable.

## License ##
The source code for this application is provided under the terms of MIT License found in the [LICENSE](./LICENSE) file.