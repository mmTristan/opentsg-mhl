package ascmhl

import (
	"fmt"
	"path/filepath"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/afero"
)

func TestDirectorySearch(t *testing.T) {

	// Array of folders to generate and the expected folders to be found
	dictCheck := []string{"src/a/b/c", "src/a/../b/c", "src/../../afilethatistoohigh"}
	expected := [][]string{{"src", "src/a", "src/a/b", "src/a/b/c"}, {"src", "src/b", "src/b/c"}, {"src"}}
	// Map with these areas
	for i, dict := range dictCheck {
		// Mock the file system with a map
		appFS := afero.NewMemMapFs()
		// create test files and directories
		_ = appFS.MkdirAll(dict, 0755)
		// Make a file to intialise src
		_ = afero.WriteFile(appFS, "src/c", []byte("file c"), 0644)
		AppFS = appFS

		gen, err := findDirs(filepath.Clean("./src"))
		Convey("Checking the find folder finds all the required folders in a system", t, func() {
			Convey(fmt.Sprintf("using  %v as the folder system", dict), func() {
				Convey(fmt.Sprintf("The following folders %v are expected and the following are found %v", expected[i], gen), func() {
					So(err, ShouldBeNil)
					So(gen, ShouldResemble, expected[i])
				})
			})
		})
	}
}

func TestFileSearch(t *testing.T) {

	// First is the folder
	// Rest of the strings are folder to populate it
	fileCheck := []string{"src/a/b/c", "src/f", "src/a/fa", "src/a/b/fb", "src/a/b/c/fc", "src/../../afilethatistoohigh"}
	expected := [][]fileAndDir{{{Name: "fc", Dir: false}},
		{{Name: "c", Dir: true}, {Name: "fb", Dir: false}},
		{{Name: "b", Dir: true}, {Name: "fa", Dir: false}},
		{{Name: "a", Dir: true}, {Name: "f", Dir: false}}}
	// Map with these areas

	// Mock the file system with a map
	appFS := afero.NewMemMapFs()

	// create test files and directories
	_ = appFS.MkdirAll(fileCheck[0], 0777)
	for j := 1; j < len(fileCheck); j++ {
		// Make a file to intialise src
		_ = afero.WriteFile(appFS, fileCheck[j], []byte("file"), 0777)
	}
	AppFS = appFS

	// Repeat the test but for with a ./ and the same results are expected
	files, genFD, err := findFiles(filepath.Clean("./src"))
	for i, gen := range genFD {
		Convey("Checking the find files shows all the required files in a system and if they are directories", t, func() {
			Convey(fmt.Sprintf("using  %v as the folder system", files[i]), func() {
				Convey(fmt.Sprintf("The following folders %v are expected and the following are found %v", expected[i], gen), func() {
					So(gen, ShouldResemble, expected[i])
					So(err, ShouldBeNil)
				})
			})
		})
	}

	input := "srcer"
	_, _, err = findFiles(input)
	Convey("Checking the find files shows all the required files in a system and if they are directories", t, func() {
		Convey(fmt.Sprintf("using an invalid %v main directory to searched", input), func() {
			Convey("An array of 0 folders is expected to be generated", func() {
				So(err.Error(), ShouldEqual, "open srcer: file does not exist")
			})
		})
	})
}

func TestBadDir(t *testing.T) {

	// First is the folder
	// Rest of the strings are folder to populate it
	dictCheck := "src/a/b/ascmhl"
	// A nil is generated for b/ascmhl as the b folder exits and makes it through the initial search of directories
	badfile := []string{"./src/a/b/ascmhl", "./src/a/d", "./src/a/../../../f"}
	//badResult1 := nil //[][]fileAndDir{nil}
	badResult2 := [][]fileAndDir{}

	expec := [][][]fileAndDir{badResult2, nil, nil}
	// Mock the file system with a map
	appFS := afero.NewMemMapFs()
	// create test files and directories
	_ = appFS.MkdirAll(dictCheck, 0755)
	AppFS = appFS

	for i, bad := range badfile {

		_, gen, _ := findFiles(bad)
		Convey("Checking the find folder returns a nil result when non existent or forbidden folders are used", t, func() {
			Convey(fmt.Sprintf("using  %v as the folder system", dictCheck), func() {
				Convey(fmt.Sprintf("The following folders %v are searched for", bad), func() {
					So(gen, ShouldResemble, expec[i])
				})
			})
		})
	}
}
