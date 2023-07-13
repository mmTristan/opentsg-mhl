package ascmhl

import (
	"encoding/xml"
	"fmt"
	"io"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/afero"
)

// Make a struct and then fill it with stuff to get the bytes// Loop through each method!

func TestChainGen(t *testing.T) {
	// Sequence is a hashlist to test against
	seq := [][]hashlistChain{{{1, "testPath", "code"}, {2, "testPath", "code"}, {3, "testPath", "code"}},
		{{23, "testPath", "code"}},
		{}}

	// Expected end point
	end := hashlistChain{Path: "test01.mhl", C4: "c45mAN4LS8t6YxHNJ9MkxEKivJXv287BsXEdPd8UBgsReG1ALx5RfqaRQLRP3Jpr5xsHAcYhixyVFggBKcaucdPq6E"}
	// Array of folders to generate and the expected folders to be found
	dict := "src/ascmhl"

	for _, s := range seq {
		// Mock the file system with a map
		appFS := afero.NewMemMapFs()
		// create test files and directories
		_ = appFS.MkdirAll(dict, 0755)
		var chainDummy ascmhldirectory
		var textb []byte
		if len(s) != 0 {
			chainDummy.Hashlist = &s
			chainDummy.Xmlns = "urn:ASC:MHL:DIRECTORY:v2.0"
			textb, _ = xml.Marshal(chainDummy)
		} // Means we can test empty arrays as well
		// Make the original chain file
		_ = afero.WriteFile(appFS, "src/ascmhl/chain.xml", textb, 0644)
		AppFS = appFS

		// Fill with a byte to be id and then a dummy name for the previous mhl
		err := chainEnc([]byte("test"), "test01.mhl", "./src")

		x, _ := AppFS.Open("src/ascmhl/chain.xml")
		b, _ := io.ReadAll(x)
		var endDummy ascmhldirectory
		_ = xml.Unmarshal(b, &endDummy)
		// Ad the expected end to the input
		var expec []hashlistChain
		if len(s) != 0 {
			expec = *chainDummy.Hashlist
			end.Sequence = len(expec) + 1
			expec = append(expec, end)
		} else {
			end.Sequence = 1
			expec = []hashlistChain{end}
		}

		Convey("Checking the chain file hashlist is generated correctly", t, func() {
			Convey(fmt.Sprintf("using  %v as the initial chain file contents", s), func() {
				Convey(fmt.Sprintf("The hashlist has the following value appended to it %v", end), func() {
					So(err, ShouldBeNil)
					So(*endDummy.Hashlist, ShouldResemble, expec)
				})
			})
		})
	}
}

func TestBadInput(t *testing.T) {
	// Sequence is a hashlist to test against
	want := fmt.Errorf("error extracting src/ascmhl/chain.xml: EOF")
	// Mock the file system with a map
	appFS := afero.NewMemMapFs()

	// create test files and directories
	dict := "src/ascmhl"
	_ = appFS.MkdirAll(dict, 0755)
	_ = afero.WriteFile(appFS, "src/ascmhl/chain.xml", []byte("corrupt"), 0644)
	AppFS = appFS

	// Fill with a byte to be id and then a dummy name for the previous mhl
	err := chainEnc([]byte("test"), "test01.mhl", "./src")
	fmt.Println(appFS)
	Convey("Checking the chain file hashlist catches errors", t, func() {
		Convey("using a corrputed xml file", func() {
			Convey(fmt.Sprintf("An error of %v is returned", want), func() {
				So(err, ShouldResemble, want)
			})
		})
	})
}
