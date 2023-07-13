package ascmhl

import (
	"crypto/sha256"
	"fmt"
	"os"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func init() {
	filesToCheck := []string{"testdata/run1.txt", "testdata/run2.txt", "testdata/run3.txt"}
	fileMake(filesToCheck)
}

func TestFile(t *testing.T) {
	fileToCheck := []string{"testdata/run1.txt", "testdata/run2.txt"}
	results := []string{"./testdata/ascmhl/0001_run1.txt_2006-01-02_150405.mhl", "./testdata/ascmhl/0001_run2.txt_2006-01-02_150405.mhl"}
	expecResults := []string{"./testdata/ascmhl/RES_run1.txt_2006-01-02_150405.mhl", "./testdata/ascmhl/RES_run2.txt_2006-01-02_150405.mhl"}

	//	expec := []string{"105349e8038026fbbde19f415988399cb13c0ff10247fcfe32952a30e083ce3a",
	//		"f9ac630ffda9bd00fd43ae25fed9b366bdeacefd2637e15d191d99119d618622"}

	crcTarget := []string{
		"xL4zmPDhhhWpEVzE24M6LjnfvdNmWZ2wdrwkGsaGQgA6bh53BkC2pQqRY5pE8JZpdasXSDvb",
		"3jXSjgjLSVczw4SRq2H94fujHcTEGNHHRAdxPCpBaAYyc7sMXneTtU8jPsRXQvPLxTspLGAK",
	}
	depth := []int{16, 8}

	// Keep all the times constant
	timeFile = "2006-01-02_150405"
	timeXML = "2006-01-02T15:04:05+00:00"
	creator.DateTimeC = timeXML
	constant := time.Unix(0, 0)

	for i, file := range fileToCheck {
		_ = os.Chtimes(file, constant, constant)
		f, _ := os.Open(file)

		defer f.Close()
		err := MhlGenFile(f, ToHash{C4: true, Sha1: true, Crc32RGB: true, Crc16RGB: true}, []byte(crcTarget[i]), depth[i])

		res, _ := os.ReadFile(results[i])
		expected, _ := os.ReadFile(expecResults[i])
		htestWrite := sha256.New()
		htestWrite.Write(res)
		hExpectedWrite := sha256.New()
		hExpectedWrite.Write(expected)

		Convey("Checking that hashes can be generated for a single file", t, func() {
			Convey("using an os.file", func() {
				Convey(fmt.Sprintf("Nil error is expected got %v", err), func() {
					So(err, ShouldBeNil)
					So(fmt.Sprintf("%0x", htestWrite.Sum(nil)), ShouldResemble, fmt.Sprintf("%0x", hExpectedWrite.Sum(nil)))
				})
			})
		})
		os.Remove(results[i])
	}
}

/*
import (
	"fmt"
	"io"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/afero"
)

// How do i test instead of running?
// Make virtual space
// Decode both into structs and compare them?

var testTime = time.Unix(0, 0)

func TestMhlRun(t *testing.T) {
	AppFS = afero.NewOsFs()
	// Test all and run
	want := ToHash{Xxh64: true, Sha1: true}
	MhlGenAll(".", want, []string{"git"})
	timeFile = testTime.Format(folTformat)
	timeXml = testTime.Format(xmlTFormat)
	// Array of folders to generate and the expected folders to be found

	dictCheck := []string{"./A002R2EC", "./A002R2EC/Clips"}
	expect := []string{"/ascmhl/0001_A002R2EC_1970-01-01_000000.mhl", "/ascmhl/0001_Clips_1970-01-01_000000.mhl"}
	// Search := []string{}
	// Map with these areas
	for i, dict := range dictCheck {
		AppFS = replicateASCtest()
		//0001_A002R2EC_1970-01-01_000000.mhl
		err := MhlGenAll(".", want, nil)
		// Make a body here of what we want and what we get
		gq, _ := AppFS.Open(dict + expect[i])

		g, _ := io.ReadAll(gq)
		fmt.Println(string(g))
		fmt.Println(AppFS)
		Convey("Checking the find folder finds all the required folders in a system", t, func() {
			Convey(fmt.Sprintf("using  %v as the folder system", dict), func() {
				Convey(fmt.Sprintf("The following folders %v are expected and the following are found %v", dict[i], err), func() {
					// So(gen, ShouldResemble, expected[i])
					So(err, ShouldBeNil)
				})
			})
		})
	}
}

func replicateASCtest() afero.Fs {
	appFS := afero.NewMemMapFs()

	dir := "./A002R2EC/Clips"
	appFS.MkdirAll(dir, 0755)

	files := []string{"./A002R2EC/sidecar.txt", "./A002R2EC/Clips/A002C007_141024_R2EC.mov", "./A002R2EC/Clips/A002C006_141024_R2EC.mov"}
	content := []string{"BLOREM ipsum dolor sit amet, consetetur sadipscing elitr.\n", "def\n", "abcd\n"}

	for i, fn := range files {

		afero.WriteFile(appFS, fn, []byte(content[i]), 0644)
		// Keep times constant to prevent errors
		appFS.Chtimes(fn, testTime, testTime)
	}
	appFS.Chtimes(dir, testTime, testTime)
	return appFS
}

*/
