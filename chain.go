package ascmhl

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"

	"github.com/spf13/afero"
)

// go:embed schema/ascmhldirectory.xsd 	_ "embed"
//var chainSchema []byte

// ChainEnc encodes the ascfile bytes as a C4 hash and adds it to the chain file.
// It requires the previous .mhl file name generated and the root folder, it appends the information to the current chain file,
// if no chain file is found a new one is generated.
func chainEnc(ascfile []byte, prev, root string) error {
	chaHash := contentHasher(ascfile, []string{"C4"})

	fchain, err := AppFS.OpenFile(root+"/ascmhl/chain.xml", os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	defer fchain.Close()

	// Generate the chain struct using information from the chain file
	chain, err := chainUp(fchain, chaHash, prev)
	if err != nil {
		return err
	}

	bychain, err := xml.MarshalIndent(chain, "", "   ")
	if err != nil {
		return err
	}
	// Write at the exact point to prevent appending of dile information
	_, err = fchain.WriteAt([]byte(xml.Header), 0)
	if err != nil {
		return err
	}
	_, err = fchain.WriteAt(bychain, 39)
	if err != nil {
		return err
	}

	return nil
}

func chainUp(fchain afero.File, chaHash map[string]string, name string) (ascmhldirectory, error) {
	chainStat, _ := fchain.Stat()
	var chain ascmhldirectory
	// Check if the file has information or was made by go
	if chainStat.Size() != 0 {
		// Read the bytes and then populate the struct
		chainBody, _ := io.ReadAll(fchain)
		var preSeq ascmhldirectory
		err := xml.Unmarshal(chainBody, &preSeq)
		if err != nil {
			return ascmhldirectory{}, fmt.Errorf("error extracting %s: %v", fchain.Name(), err)
		}

		chain = chainGen(chaHash, *preSeq.Hashlist, name)
		/*
			// Check the file is valid
			if schemaCheck(chainSchema, chainBody) == nil {
				chain = chainGen(chaHash, *preSeq.Hashlist, name)
			} else {

				// Throw an error saying the chain is corrupted
				// And remake it
				return chainGen(chaHash, nil, name), fmt.Errorf("chain file corrupted, new chain file generated")
			}*/
	} else {
		chain = chainGen(chaHash, nil, name)
	}

	return chain, nil
}

func chainGen(id map[string]string, seq []hashlistChain, path string) ascmhldirectory {
	// Populate the struct for the chain
	var hc hashlistChain

	hc.C4 = id["C4"] // Search only for the c4 value as that's the only value allowed
	hc.Sequence = len(seq) + 1
	hc.Path = path
	hcAr := []hashlistChain{hc}

	if len(seq) != 0 {
		hcAr = append(seq, hcAr...)
	}
	var am ascmhldirectory
	am.Hashlist = &hcAr
	am.Xmlns = "urn:ASC:MHL:DIRECTORY:v2.0"

	return am
}

// xml layout in struct form
type ascmhldirectory struct {
	Xmlns    string           `xml:"xmlns,attr"`
	Hashlist *[]hashlistChain `xml:"hashlist,omitempty"`
}

type hashlistChain struct {
	Sequence int    `xml:"sequencenr,attr"`
	Path     string `xml:"path,omitempty"`
	C4       string `xml:"c4,omitempty"`
}
