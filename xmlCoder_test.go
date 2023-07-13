package ascmhl

import (
	"encoding/xml"
	"fmt"
	"math/rand"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/afero"
)

func TestHashChange(t *testing.T) {
	// Init the system at the start
	// AppFS = afero.NewOsFs()

	AppFS = replicateSystem()
	root := "./src"
	creator.HostName = "xmlCoder test"
	timeXML = "2006-01-02T15:04:05+00:00"
	creator.DateTimeC = timeXML
	wants := []ToHash{{C4: true}, {Md5: true}, {Sha1: true}, {Xxh128: true}, {Xxh3: true}, {Xxh64: true}}
	// Run this initially to have a map to build over
	_ = MhlGen(root, wants[0], nil)
	// Make a map of false hashes to write over
	accrued := make(map[string]bool)
	for _, k := range availableHashes {
		accrued[k] = false
	}
	for _, want := range wants {
		// Conver the new hash to a map then add it to the accrued map
		newHash := wantToMap(want)
		for _, k := range availableHashes {
			if newHash[k] == true {
				accrued[k] = true
			}
		}

		// Make the body with the new hash
		err := MhlGen(root, want, nil)
		// Decode the file you just made and get the map of hash types
		_, preName, _ := rootSorter(root)
		_, _, got := decode(root, root+"/ascmhl/"+preName)
		fmt.Println(err, root)
		Convey("Checking that hashes are added to the generated file", t, func() {
			Convey(fmt.Sprintf("using a mock file system and adding a hash of %v", want), func() {
				Convey(fmt.Sprintf("The code is identified as accrued hashes become %v", got), func() {
					So(accrued, ShouldResemble, got)
					So(err, ShouldBeNil)
				})
			})
		})
	}
}

func TestEncode(t *testing.T) {
	AppFS = replicateSystem()
	root := "src"
	creator.HostName = "xmlCoder test"
	timeXML = "2006-01-02T15:04:05+00:00"
	creator.DateTimeC = timeXML
	reqHash := make(map[string]bool)
	for _, k := range availableHashes {
		reqHash[k] = true
	}

	prevFiles := []string{"", "test.mhl", "test1.mhl"}
	nextFiles := []string{"test.mhl", "test1.mhl", "test2.mhl"}
	want := []string{"original", "verified", "failed"}
	fileContents := []string{"same", "same", "different"}

	for i, next := range nextFiles {
		// Rewrite the file to be checked at every turn

		_ = afero.WriteFile(AppFS, "src/sidecar.txt", []byte(fileContents[i]), 0644)

		ignoreUpdate([]string{"second"})
		// Fill with a byte to be id and then a dummy name for the previous mhl
		body, err := encode(root, prevFiles[i], next, reqHash)
		// Get file as a go struct
		var made hashlist
		_ = xml.Unmarshal(body, &made)
		toComp := *made.Hashes.Fhash
		act := toComp[0].C4.Action
		Convey("Checking that xml files are generated and the correct action is attributed when checking previous codes", t, func() {
			Convey(fmt.Sprintf("using a mock file system and a previous file of %v", prevFiles[i]), func() {
				Convey(fmt.Sprintf("The code is identified as %v", want[i]), func() {
					So(act, ShouldEqual, want[i])
					So(err, ShouldBeNil)
				})
			})
		})
	}
}

func replicateSystem() afero.Fs {
	rand.Seed(20027)

	appFS := afero.NewMemMapFs()

	// create test files and directories
	dict := "./src/second/"
	_ = appFS.MkdirAll(dict, 0755)
	files := []string{"src/sidecar.txt"} // , "./src/second/secondfile", "./src/second/thirdfile"}
	constant := time.Unix(0, 0)
	for _, fn := range files {
		fb := make([]byte, 200)
		rand.Read(fb) // Fill with predetermined data

		_ = afero.WriteFile(appFS, fn, fb, 0644)
		// Keep times constant to prevent errors
		_ = appFS.Chtimes(fn, constant, constant)
	}
	_ = appFS.Chtimes(dict, constant, constant)

	return appFS
}
