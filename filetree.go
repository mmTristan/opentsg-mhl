package ascmhl

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/afero"
)

// APPFS is the filesystem in use
// At the moment it is a substitute with the os for testing and abstraction
// Var AppFs = afero.NewOsFs()

type fileAndDir struct {
	Name string
	Dir  bool
}

/*
type fpath struct {
	path     string
	children []string
}*/

// FindFiles takes a location and returns an array of the folders and an array of the contents and if they
// Are a directory or not. It recusrivley searches through the system.
func findFiles(loc string) ([]string, [][]fileAndDir, error) {

	// Ensure the file follows the format
	//format := regexp.MustCompile(`^\./`)
	//if loc == "." || format.MatchString(loc) {
	// Get directories
	// Then list the children
	dirs, err := findDirs(loc)
	if err != nil {
		return nil, nil, err
	}

	dirs = reverse(dirs)
	contents := make([][]fileAndDir, len(dirs))
	for i, name := range dirs {
		contents[i], err = findfiles(name)
		if err != nil {
			return dirs, contents, err
		}
	}

	return dirs, contents, nil
}

//return nil, nil, fmt.Errorf("invalid folder name, must be \".\" or begin with a \"./\"")
//}

// FindDirs returns a list of the folders in a given system
func findDirs(loc string) ([]string, error) {

	var dirs []string
	err := afero.Walk(AppFS, loc, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if ignoreFence(path) {
			if info.IsDir() {
				dirs = append(dirs, path)
			}
		}

		return err
	})

	if err != nil {
		return dirs, err
	}

	for i, name := range dirs {
		if name != loc {
			dirs[i] = name // "./" + name
		}
	}

	return dirs, nil
}

// findFiles searches a folder for all its contents
// , it does not recursivley search any children.
func findfiles(dir string) ([]fileAndDir, error) {
	var files []fileAndDir
	maxDepth := strings.Count(dir, string(os.PathSeparator)) + 1

	err := afero.Walk(AppFS, dir, func(s string, d fs.FileInfo, err error) error {
		var f fileAndDir
		if err != nil {
			return err
		}
		fmt.Println(s)
		// Ignore the requesite files, skip the directory itself and keep it to the depth of the folder
		if ignoreFence(s) && s != dir && strings.Count(s, string(os.PathSeparator)) == maxDepth {

			//split := strings.Split(s, string(os.PathSeparator))
			f.Name = filepath.Base(s) //split[len(split)-1]
			if d.IsDir() {
				f.Dir = true

			}
			files = append(files, f)
		}

		return nil
	})

	if err != nil {
		return files, err
	}

	return files, nil
}

// reverse reverses an array
// https:// Github.com/golang/go/wiki/SliceTricks#reversing
func reverse(s []string) []string {
	a := make([]string, len(s))
	copy(a, s)

	for i := len(a)/2 - 1; i >= 0; i-- {
		opp := len(a) - 1 - i
		a[i], a[opp] = a[opp], a[i]
	}

	return a
}

var constFolds = []string{".DS_Store", "ascmhl", "ascmhl/"}

var foldersToIgnore []string

// IgnoreUpdate adds extra files to be ignored by the file searcher'
// It resets the folders on each call
func ignoreUpdate(extra []string) {
	// This was appending constfolds the extra parts
	foldersToIgnore = make([]string, len(constFolds))
	copy(foldersToIgnore, constFolds)
	foldersToIgnore = append(foldersToIgnore, extra...)
}

func getIgnore() []string {
	return foldersToIgnore
}

// ignore fence checks if a folder is to be searched for or not and is skipped as part of the analysis
func ignoreFence(check string) bool {

	for _, fold := range foldersToIgnore {
		r := regexp.MustCompile(fold)
		if r.MatchString(check) {
			return false
		}
	}

	return true
}
