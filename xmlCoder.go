package ascmhl

import (
	"encoding/xml"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	_ "embed"

	"github.com/spf13/afero"
)

//go:embed schema/ascmhl.xsd
var mhlSchema []byte

var timeXML string
var timeFile string
var folTformat string
var xmlTFormat string

func init() {

	xmlTFormat = "2006-01-02T15:04:05+00:00"
	folTformat = "2006-01-02_150405"

	genTime := time.Now().Round(time.Second)
	timeFile = genTime.Format(folTformat)
	timeXML = genTime.Format(xmlTFormat)
}

var availableHashes = []string{"C4", "Md5", "Sha1", "Xxh128", "Xxh3", "Xxh64", "Crc16RGB", "Crc32RGB"}

func init() {

	xmlTFormat = "2006-01-02T15:04:05+00:00"
	folTformat = "2006-01-02_150405"

	genTime := time.Now().Round(time.Second)
	timeFile = genTime.Format(folTformat)
	timeXML = genTime.Format(xmlTFormat)
	host, _ := os.Hostname()
	creator.HostName = host
	creator.DateTimeC = timeXML
}

var hashAttr = hashlist{Version: "2.0", Xmlns: "urn:ASC:MHL:v2.0"}
var hashAttrCRC = hashlistCRC{Version: "2.0", Xmlns: "urn:ASC:MHL:v2.0"}
var toolAttr = tool{"ascmhl.go", "0.0.1"}
var creator = creatorinfo{DateTimeC: timeXML, Tool: &toolAttr}

// Put the structs and all endevaours facing maps to structs etc
// Want to decode and encode maps to the xml to save!

// Encode generates the maps and then saves the ascmhl file in the specified root folder
func encode(root, prename, name string, reqHash map[string]bool) ([]byte, error) {
	ascmhlFol := root + string(os.PathSeparator) + "ascmhl"

	m := hashAttr
	m.Creatorinfo = &creator
	m.Creatorinfo.Tool = &toolAttr

	prevCont, prevStruc, decHash := decode(root, ascmhlFol+string(os.PathSeparator)+prename)
	var needHash []string
	for _, h := range availableHashes {
		if decHash[h] || reqHash[h] {
			needHash = append(needHash, h)
		}
	}

	contentHash, structureHash, err := getMapHash(root, needHash)
	if err != nil {
		return nil, err
	}

	i := processGen(contentHash[root], structureHash[root], prevCont[root], prevStruc[root])
	delete(contentHash, root) // Delete the root the folders so they are not represented twice
	delete(structureHash, root)
	m.Processinfo = &i
	// Calculate the map and folder hashes
	fh := mapToHash(contentHash, structureHash, prevCont, prevStruc, root)
	m.Hashes = &fh

	// Generate files to save
	by, err := xml.MarshalIndent(m, "", "   ")
	if err != nil {
		return nil, err
	}

	// Generate or use the folder
	_, exist := afero.ReadDir(AppFS, ascmhlFol)
	if exist != nil {
		err = AppFS.Mkdir(ascmhlFol, os.ModePerm)
		if err != nil {
			return nil, fmt.Errorf("error generating %s : %v", ascmhlFol, err)
		}
	}
	// Save the mhl file
	f, _ := AppFS.OpenFile(ascmhlFol+"/"+name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
	defer f.Close()
	fileByt := append([]byte(xml.Header), by...)

	_, err = f.Write(fileByt)
	if err != nil {
		return nil, err
	}

	return fileByt, nil
}

// decode returns a map of hash values from a mhl file
func Decode(root, loc string) (map[string]map[string]string, map[string]map[string]string, map[string]bool) {
	return decode(root, loc)
}

func decode(root, loc string) (map[string]map[string]string, map[string]map[string]string, map[string]bool) {
	// Open the file and convert it to the struct
	var hl hashlist
	// TODO error checking files and returns
	mhl, err := afero.ReadFile(AppFS, loc)

	toCheckf := make(map[string]map[string]string)
	toCheckd := make(map[string]map[string]string)
	if err != nil { // Return nothing if there is no schema to check
		return toCheckf, toCheckd, nil
	}

	/*/ Ensure we are parsing valid schema
	if err := schemaCheck(mhlSchema, mhl); err != nil {
		fmt.Printf("Error validating the %v against a schema: %v\n", loc, err)

		return toCheckf, toCheckd, nil
	}*/

	err = xml.Unmarshal(mhl, &hl)
	if err != nil {
		return toCheckf, toCheckd, nil
	}

	// Extract the files and directories only if there's something in it
	var fh []hashes
	if *hl.Hashes.Fhash != nil {
		fh = *hl.Hashes.Fhash
	}
	var dh []dhashes
	if hl.Hashes.Dhash != nil {
		dh = *hl.Hashes.Dhash
	}
	var rh roothash
	if hl.Processinfo.Roothash != nil {
		rh = *hl.Processinfo.Roothash
		// Put the root folder back into the map of directories for later comparison
		var rootToFol dhashes
		rootToFol.Content = rh.Content
		rootToFol.Structure = rh.Structure
		dh = append(dh, rootToFol)
	}

	// Get a map of the hashes to check
	key := []string{"C4", "Md5", "Sha1", "Xxh3", "Xxh64", "Xxh128", "Crc16RGB", "Crc32RGB"}
	// Htypes is a map of the hash types and if they are called in the previous file
	hTypes := make(map[string]bool)

	for _, h := range fh {
		var p string
		if h.Path.Code == root {
			p = h.Path.Code
		} else {
			p = root + "/" + h.Path.Code
		}
		toCheckf[p] = make(map[string]string)
		for _, k := range key {
			hTypes[k] = refToMap(p, k, reflect.ValueOf(&h).Elem(), toCheckf)
		}
	}

	// Repeat for different structs
	for _, h := range dh {
		var p string
		if h.Path != nil {
			p = root + string(os.PathSeparator) + h.Path.Code
		} else {
			p = root
		}
		toCheckd[p] = make(map[string]string)
		toCheckf[p] = make(map[string]string)
		for _, k := range key {
			// Take the pointer to prevent errors with reflect
			c := *h.Content
			s := *h.Structure
			// Assign values
			refToMap(p, k, reflect.ValueOf(&c).Elem(), toCheckf)
			refToMap(p, k, reflect.ValueOf(&s).Elem(), toCheckd)
		}
	}

	return toCheckf, toCheckd, hTypes
}

// Compare will compare folders when it's made
func Compare() {
	// This will be for later once more planning is done
}

func processGen(cont FileHash, stru, preCont, preStru map[string]string) processinfo {
	var p processinfo
	p.ProcessType = "in-place" // Inplace for the moment
	var i pattern
	p.Ignore = &i
	p.Ignore.Pattern = getIgnore() // Get from file

	if cont.Hash != nil && stru != nil {
		var r roothash
		var c hashes
		var s hashes
		c.hashAssign(cont.Hash, preCont)
		s.hashAssign(stru, preStru)
		r.Content = &c
		r.Structure = &s
		p.Roothash = &r
	}

	return p
}

func mapToHash(mc map[string]FileHash, ms, tc, ts map[string]map[string]string, root string) hashType {

	dirs := make([]dhashes, len(ms))
	// Calculate structure for directories
	j := 0
	for key, dir := range ms {
		var d dhashes
		var conHash hashes
		var strHash hashes

		conHash.hashAssign(mc[key].Hash, tc[key])
		strHash.hashAssign(dir, ts[key])
		// Size is 0 for the directories
		pathSkel := pathGen(mc[key].Time, root, key, 0)

		d.Path = &pathSkel
		d.Content = &conHash
		d.Structure = &strHash
		dirs[j] = d
		// Assign the modification time before deleting it and using it in the contents as well
		delete(mc, key)
		j++
	}

	// Then for the files here
	files := make([]hashes, len(mc))

	i := 0
	for p, file := range mc {

		var genHash hashes
		genHash.hashAssign(file.Hash, tc[p])
		// Generate the path information for each file
		pathSkel := pathGen(file.Time, root, p, int(file.Size))
		genHash.Path = &pathSkel
		files[i] = genHash
		i++

	}
	// Assign code and size
	// Set up a reflect
	var ht hashType
	ht.Dhash = &dirs
	ht.Fhash = &files

	return ht
}

// path gen generates a path variable to be used for each file/folder
func pathGen(time, root, fPath string, size int) (p path) {
	p.DateTimeL = time
	p.Size = size
	p.Code = strings.Replace(fPath, root+string(os.PathSeparator), "", 1)

	return
}

// hash assign uses reflect to match the hash in algorithm names to
// their position in the struct
func (genHash *hashes) hashAssign(dir, comp map[string]string) {
	// The key names  match the exported struct fields
	// We can make a skeleton of the hash format and assign it pointer of hashes
	// At any point if this fails then we'll put an error in
	var action string
	// If there's nothing to compare to then the map is original
	if comp == nil {
		action = "original"
	} else {
		action = "verified"
	}
	// Run through the generated keys
	for key, code := range dir {
		var skel hashFormat
		// If the hash matches or is new and has nothing to compare to
		if comp[key] == code || comp[key] == "" {
			skel.Action = action
		} else { // If they don't match
			skel.Action = "failed"
		}

		skel.DateTime = timeXML
		// Get the value of the hashes
		// And assign the hash type using value reflect
		structVal := reflect.ValueOf(genHash).Elem()
		// Find if the key is c4/Xxh64 etc then assign the value
		structField := structVal.FieldByName(key)
		skel.Code = code
		structField.Set(reflect.ValueOf(&skel))
	}
}

// refToMap takes the reflect value and checks if there is a value for that key
// if there is then the code is assigned to the path key of the map
func refToMap(p, k string, structVal reflect.Value, m map[string]map[string]string) bool {
	structField := structVal.FieldByName(k)
	keyHash := structField.Interface().(*hashFormat)
	if keyHash != nil { // Check there's actually a code to parse
		m[p][k] = keyHash.Code

		return true
	}

	return false
}

// This is the struct for the xml layout
type hashlistCRC struct {
	Creatorinfo *creatorinfo `xml:"creatorinfo,omitempty"`
	Processinfo *processinfo `xml:"processinfo,omitempty"`
	Hashes      *hashType    `xml:"hashes,omitempty"`
	MetaData    interface{}  `xml:"metadata,omitempty"`
	// Reference   *reference       `xml:"references,omitempty"`
	Version string `xml:"version,attr"`
	Xmlns   string `xml:"xmlns,attr"`
}

// This is the struct for the xml layout
type hashlist struct {
	Creatorinfo *creatorinfo `xml:"creatorinfo,omitempty"`
	Processinfo *processinfo `xml:"processinfo,omitempty"`
	Hashes      *hashType    `xml:"hashes,omitempty"`
	MetaData    interface{}  `xml:"metadata,omitempty"`
	// Reference   *reference       `xml:"references,omitempty"`
	Version string `xml:"version,attr"`
	Xmlns   string `xml:"xmlns,attr"`
}

type creatorinfo struct {
	DateTimeC string       `xml:"creationdate,omitempty"`
	HostName  string       `xml:"hostname,omitempty"`
	Tool      *tool        `xml:"tool,omitempty"`
	Author    *[]authorInf `xml:"author,omitempty"`
	Location  string       `xml:"location,omitempty"`
	Comment   string       `xml:"comment,omitempty"`
}

type authorInf struct {
	Email  string `xml:"email,attr,omitempty"`
	Phone  string `xml:"phone,attr,omitempty"`
	Role   string `xml:"role,attr,omitempty"`
	Author string `xml:",innerxml"`
}

type tool struct {
	ToolText string `xml:",innerxml"`
	Version  string `xml:"version,attr"`
}

type processinfo struct {
	ProcessType string    `xml:"process,omitempty"`
	Roothash    *roothash `xml:"roothash,omitempty"` // Update roothash
	Ignore      *pattern  `xml:"ignore,omitempty"`
}

type roothash struct {
	Content   *hashes `xml:"content,omitempty"`
	Structure *hashes `xml:"structure,omitempty"`
}

type pattern struct {
	Pattern []string `xml:"pattern,omitempty"`
}

type hashFormat struct {
	Action    string `xml:"action,attr,omitempty"`
	DateTime  string `xml:"hashdate,attr,omitempty"`
	Structure string `xml:"structure,attr,omitempty"`
	Code      string `xml:",innerxml"`
}

/*
type hashesOLD struct {
	Path   *path       `xml:"path,omitempty"`
	C4     *hashFormat `xml:"c4,omitempty"`
	Md5    *hashFormat `xml:"md5,omitempty"`
	Sha1   *hashFormat `xml:"sha1,omitempty"`
	Xxh128 *hashFormat `xml:"xxh128,omitempty"`
	Xxh3   *hashFormat `xml:"xxh3,omitempty"`
	Xxh64  *hashFormat `xml:"xxh64,omitempty"`
}*/

type hashes struct {
	Path     *path       `xml:"path,omitempty"`
	C4       *hashFormat `xml:"c4,omitempty"`
	Md5      *hashFormat `xml:"md5,omitempty"`
	Sha1     *hashFormat `xml:"sha1,omitempty"`
	Xxh128   *hashFormat `xml:"xxh128,omitempty"`
	Xxh3     *hashFormat `xml:"xxh3,omitempty"`
	Xxh64    *hashFormat `xml:"xxh64,omitempty"`
	Crc16RGB *hashFormat `xml:"crc16RGB,omitempty"`
	Crc32RGB *hashFormat `xml:"crc32RGB,omitempty"`
}

/*
type hashTypeCRC struct {
	Fhash *[]hashesOLD `xml:"hash,omitempty"`
	Dhash *[]dhashes   `xml:"directoryhash,omitempty"`
}*/

type hashType struct {
	Fhash *[]hashes  `xml:"hash,omitempty"`
	Dhash *[]dhashes `xml:"directoryhash,omitempty"`
}

type dhashes struct {
	Path      *path       `xml:"path,omitempty"`
	Content   *hashes     `xml:"content,omitempty"`
	Structure *hashes     `xml:"structure,omitempty"`
	PPath     *path       `xml:"previousPath,omitempty"`
	MetaData  interface{} `xml:"metadata,omitempty"`
}

type path struct {
	Size      int    `xml:"size,attr,omitempty"`
	DateTimeC string `xml:"creationdate,attr,omitempty"`
	DateTimeL string `xml:"lastmodificationdate,attr,omitempty"`
	Code      string `xml:",innerxml"`
}
