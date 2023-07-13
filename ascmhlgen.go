package ascmhl

import (
	"encoding/xml"
	"fmt"
	"hash/crc32"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/howeyc/crc16"
	"github.com/spf13/afero"
)

// var actions = []string{"original", "verified", "failed"}

// var ignored = []string{".DS_Store", "ascmhl", "ascmhl/"}
// var procType = []string{"in-place", "transfer", "flatten"}

// Var genTime = time.Now().Round(time.Second)

// Insert a main function here
/*
func main() {
	want := ToHash{C4: true}
	MhlGenAll("./test/Output/scenario_01/travel_01/A002R2EC", want, []string{".git", "pkg"})
	// MhlGenAll("./bot-tlh/staging/", want, []string{".git", "pkg"})
	// MhlGenAll(".", want, []string{".git", "pkg"})
	// MhlGen(".", want, []string{".git", "pkg"})
}*/

// MhlGen generates a ascmhl folder for the given root
func MhlGen(root string, targetHash ToHash, newIgnore []string) error {
	ignoreUpdate(newIgnore)
	// Return check if this is a file that can be opened and return file if so

	return mhlGen(root, targetHash)
}

// MhlGenAll generates ascmhl folders for a root folder and all of its subdirectories
func MhlGenAll(root string, targetHash ToHash, newIgnore []string) error {
	ignoreUpdate(newIgnore)
	// Run mhl gen for every folder in the directory

	// update this to use filepath to clean it all and go from there

	folders, err := findDirs(root)
	if err != nil {
		return err
	}

	// fmt.Println(folders)
	for _, dir := range folders {
		//	fmt.Printf("Generating ascmhl folder for %v\n", dir)
		err := mhlGen(dir, targetHash)
		if err != nil {
			return err
		}
	}

	return nil
}

// MhlGenFile generates the mhl with added crc32/16 of an image, if the pixels are supplied
func MhlGenFile(target *os.File, want ToHash, imgPix []byte, pixelBitdepth int) error {
	var pix []byte

	if pixelBitdepth == 16 || pixelBitdepth == 8 {
		pix = removeAlpha(imgPix, pixelBitdepth)
	} else {
		return fmt.Errorf("invalid bit depth used of %v. Only 8 bit or 16 bit pixels are acceptable", pixelBitdepth)
	}
	ignoreUpdate(nil) // As it is just targeting an empty file we still need to init the standard ignore files

	root, prename, name := rootSorter(target.Name())
	ascmhlFol := root + "/ascmhl"

	reqHash := wantToMap(want)
	s, _ := target.Stat()
	prevCont, _, decHash := decode(s.Name(), ascmhlFol+string(os.PathSeparator)+prename)
	var needHash []string
	for _, h := range availableHashes {
		if decHash[h] || reqHash[h] {
			needHash = append(needHash, h)
		}
	}

	f, _ := getFileHash(target, needHash)
	var em FileHash
	i := processGen(em, nil, nil, nil)

	m := hashAttr
	m.Creatorinfo = &creator
	m.Creatorinfo.Tool = &toolAttr
	m.Processinfo = &i
	// Calculate the map and folder hashes
	mg := make(map[string]FileHash)

	if want.Crc16RGB {
		crc16num := crc16.Checksum(pix, crc16.IBMTable)
		f.Hash["Crc16RGB"] = fmt.Sprintf("%0x", crc16num)
	}
	if want.Crc32RGB {
		crc32num := crc32.Checksum(pix, crc32.IEEETable)
		f.Hash["Crc32RGB"] = fmt.Sprintf("%0x", crc32num)
	}

	mg[s.Name()] = f
	fh := mapToHash(mg, nil, prevCont, nil, "")
	m.Hashes = &fh
	// Generate files to save
	by, err := xml.MarshalIndent(m, "", "   ")
	if err != nil {
		return err
	}

	// Generate or use the folder
	_, exist := afero.ReadDir(AppFS, ascmhlFol)
	if exist != nil {
		err = AppFS.Mkdir(ascmhlFol, os.ModePerm)
		if err != nil {
			return fmt.Errorf("error generating %s : %v", ascmhlFol, err)
		}
	}
	// Save the mhl file
	base, err := AppFS.OpenFile(ascmhlFol+"/"+name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		return err
	}
	defer base.Close()
	fileByt := append([]byte(xml.Header), by...)
	_, err = base.Write(fileByt)
	if err != nil {
		return err
	}
	//path, _ := filepath.Abs(base.Name())
	//fmt.Printf("%s has been generated\n", path)

	return nil

}

func removeAlpha(in []byte, depth int) []byte {
	out := make([]byte, (3*len(in))/4)
	offset := 0
	increment := (depth / 2)
	offInc := (3 * increment) / 4
	if depth == 8 {
		for i := 0; i < len(in); i += increment {
			out[offset] = in[i]
			out[offset+1] = in[i+1]
			out[offset+2] = in[i+2]
			offset += offInc
		}
	} else if depth == 16 {
		for i := 0; i < len(in); i += increment {
			out[offset] = in[i]
			out[offset+1] = in[i+1]
			out[offset+2] = in[i+2]
			out[offset+3] = in[i+3]
			out[offset+4] = in[i+4]
			out[offset+5] = in[i+5]
			offset += offInc
		}
	}

	return out
}

/*
type fileHash struct {
	hash map[string]string
	size int64
	time string
}*/

// mhlgen generates an ascmhl folder for a single file or folder
func mhlGen(root string, want ToHash) error {
	// Make sure the root is in a valid for
	// Find the previous chain file and the name of the next one
	root, prename, name := rootSorter(root)
	// Generate a map of the hashes we want
	reqHash := wantToMap(want)

	// Encode the file
	enc, err := encode(root, prename, name, reqHash)
	if err != nil {
		return err
	}
	// Encode the chain
	err = chainEnc(enc, name, root)
	if err != nil {
		return err
	}

	return nil
}

// change the struct to a more flexible array of strings
func wantToMap(want ToHash) map[string]bool {
	reqHash := make(map[string]bool)
	v := reflect.ValueOf(&want).Elem()
	vtype := v.Type()

	for i := 0; i < v.NumField(); i++ {
		val := v.Field(i)
		// If it is present from the input or the previous mhl file
		if val.Interface().(bool) {
			reqHash[vtype.Field(i).Name] = true
		}
	}

	return reqHash
}

// rootsorter returns the previous mhl, the root and the name of th ascmhl file
func rootSorter(root string) (string, string, string) {
	if root[len(root)-1] == byte('/') {
		root = root[:len(root)-1]
	}

	// Check for previous mhl files before proceeding
	// Find the top most folder if declared in the folder
	var rootFol []string
	if root == "." {
		newRoot := afero.GetTempDir(AppFS, root)
		rootFol = strings.Split(newRoot, string(os.PathSeparator))
	} else {
		rootFol = strings.Split(root, string(os.PathSeparator))
	}

	// If it is a file then trim to the first /
	fileCheck := strings.Split(root, "/")
	// Check if the final area is a file or a folder
	if len(strings.Split(fileCheck[len(fileCheck)-1], ".")) > 1 {
		split := 0
		for i, c := range root {
			if c == os.PathSeparator {
				split = i
			}
		}
		// If there's no folders to split
		if split == 0 {
			root = "."
		} else {
			root = root[:split] // Leave at the directory
		}
	}

	// Generate the aschmlfolder location
	aschmlFol := root + string(os.PathSeparator) + "ascmhl"
	target := rootFol[len(rootFol)-1]
	// Calculate the number of previous mhl files that have been generated (if any)
	pos, prename := hashNum(aschmlFol, target)
	name := pos + "_" + target + "_" + timeFile + ".mhl"

	return root, prename, name
}

// hashnum finds the next number for the list of mhl files
func hashNum(location, contents string) (string, string) {
	// Read the directories and extract the mhl files
	files, _ := afero.ReadDir(AppFS, location)
	var all []string
	// If it's a file replace the . with regex friendly one
	contents = strings.ReplaceAll(contents, ".", `\.`)
	start := regexp.MustCompile(`^(\d){4}_` + contents + `[\w\-]{1,251}(.)(m)(h)(l)$`)

	for _, file := range files {
		// Follow this format for the naming of the files
		// ("^(\\d){4}"+filename+"[\\w\\-]{1,251}(.)(m)(h)(l)$")
		f := file.Name()
		if start.Match([]byte(f)) {
			// Map all files that match the mhl requirements
			all = append(all, f)
		}
	}
	// Sort them in order to find the number
	sort.Strings(all)
	if len(all) == 0 {
		return "0001", ""
	}

	finalC := all[len(all)-1][:4]
	// Convert the 4 digit value to an integer
	integ, _ := strconv.ParseInt(finalC, 10, 8)
	// Find the next value as a string
	num := fmt.Sprintf("%v", integ+1)
	pos := strings.Repeat("0", 4-len(num))
	pos += num

	// Return the final file
	return pos, all[len(all)-1]
}

// Tohash are the hash types available.
// Crc hashes only work with MhlGenFile function and are not calculated for MhlGen and MhlGenAll
type ToHash struct {
	C4       bool
	Md5      bool
	Sha1     bool
	Xxh128   bool
	Xxh3     bool
	Xxh64    bool
	Crc32RGB bool
	Crc16RGB bool
}
