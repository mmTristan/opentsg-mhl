# opentsg-mhl

`opentsg-mhl` is a golang implementation of the [Media Hash List](https://mediahashlist.org/).

## Documentation

### Background

Media Hash Lists are used to ensure changes to your files and folder systems are
recognized with a chain of files documenting their history, in a human readable
xml format.

Inspired by the [python library](https://github.com/ascmitc/mhl) for generating
mhl files via the command line, we aim to replicate the functionality as a go
package, rather than a command line tool.

Check out the ascmhl [specification
page.](https://theasc.com/asc/asc-media-hash-list)

ASC MHL supports the following hash formats
- xxHash (64-bit, and latest XXH3 with 64-bit and 128-bit)
- MD5
- SHA1
- C4

### Installing and importing

`go get` the latest version of the library.

```sh
    go get github.com/mrmxf/opentsg-mhl
```

then include in your application.

```go
import "github.com/mrmxf/opentsg-mhl"
```

### Usage

Simply declare which hash types you'd like to use on the file system and any folders you may want to ignore as variables and then call one of the following functions. If no hashes are given then an empty file is generated.

To run in the current folder use a location of "." else use the path to the folder.

To generate ascmhl folder for a every folder in a sub system call:

```go
want := ToHash{C4: true, Sha1: true}
MhlGenAll("./test/Output/scenario_01/travel_01/A002R2EC", want, []string{".git", "pkg"})
```

To generate an ascmhl folder for just the location called call:

```go
want := ToHash{Md5: true, Xxh3: true}
MhlGen("test/Output/scenario_01/travel_01/A002R2EC", want, []string{".git", "pkg"})
```

To generate an ascmhl folder for a specific file run, where a crc is calculated the pixel bytes and image bytes need to be parsed as well. This methohd is still subject to change

The ascmhl files are generated on a per file basis and are not part of the offical ascmhl specification.

```go
want := ToHash{Md5: true, Xxh3: true, CRC16RGB}
MhlGen(file, want, imageBytes, 16)
```

The folders, "ascmhl" and "ascmhl/" are always ignored as these hold the .mhl files. ".DS_Store" Is also not searched.

### Features to add to make `opentsg-mhl` more generic

- [ ] Function to compare two mhl files without an output
- [ ] Ability to add author information
- [ ] API Functionality
- [ ] The ability to use almost any file system (s3, azure etc)

#### Known Bugs

empty folders do not generate any hashes
