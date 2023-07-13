package ascmhl

// Generate the files and test how many iterations we can run, slap go functions on etc

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/afero"
)

func TestHashBenchmark(t *testing.T) {

	prev, _ := os.ReadFile("testdata/results.json")

	var prevJSON testResults
	_ = json.Unmarshal(prev, &prevJSON)

	x := prevJSON.Result[len(prevJSON.Result)-1]
	lastRun, _ := time.Parse(xmlTFormat, x.RunDate)
	ly, lm, ld := lastRun.Date()
	y, m, d := time.Now().Date()
	if y == ly && lm == m && ld == d {
		fmt.Fprintln(os.Stdout, "Benchmark test already ran today")
	} else {
		fmt.Fprintln(os.Stdout, "This test will take 3minutes 30 seconds to run, use this time wisely (to go get a drink)")
		rand.Seed(20027)

		filesToCheck := []string{"testdata/run1.txt", "testdata/run2.txt", "testdata/run3.txt"}
		fileMake(filesToCheck)

		hashToCheck := []string{"C4", "Md5", "Sha256", "Sha1", "Xxh3", "Xxh64", "Xxh128"}
		// Results are the same as the random has the same seed
		Expec := [][]string{{"c44MTPEWxGBp4iujjPKck4PrJ6rMoCqFgnPJKyAvqBr3eve2K4y2pp8kx6HocvJVm3WubXTN6eGpRR2cZ97DTL4B2w", "c42UqvyUWPm6vfNhuMNTSRDicc4ZhywA5YVCVqCC5KHFz4W4zkx7XjGGuniV3cJpZogjrXuKmF9hjQKVgVky6QxsiN", "c42F9KN9m5ptDrcuRa9DpJwgXTwxFqi1EZXydM2mFrdkS3xyMxN8dto3XqvRfsYWhkbuPGZp7x6AYWtP4Dj4GAmSmJ"},
			{"cdaa74d26930f4b11a635a00174283ea", "d5ba6269f2d98f962c3a3c49a64a3f12", "677cb880bd576557577bfae8a7ba7809"},
			{"bb3c2b7ea823b0d38cb75c28d6abaeceb2020def10f26a19ad907391bfec3c34", "f3de12a946e85578ee7434352bdc4f1076f9eb37ea68f7f2db90403ce82f713c", "c6c84c5a167a4d23cacd667917acc5a5f58ad3a45fbc7a679402188aa9410cb9"},
			{"8606eca8ec0dc55111a7b45c4f6d8747fbd06c83", "f4b5f43fd3a7a0cd8ac83737736a4b3d47a80b2a", "09c7794906f31061fed67d898f0244bfcfa3b078"},
			{"fe64498cb5b77f9b", "f85eb5bc724fe282", "76a309f2596ffb9c"},
			{"b0a92242096e0449", "197eda497258c7b5", "7bcbca44b1ff4348"},
			{"45fa043ca74a238dfe64498cb5b77f9b", "7fd3b66dd38aebe9f85eb5bc724fe282", "40a922fcca029ca876a309f2596ffb9c"},
		}

		hashNumber := make([]check, len(hashToCheck))
		for k, hash := range hashToCheck {
			AppFS = afero.NewOsFs()
			var hashCheck check
			hashGroup := make([]hashTypeTest, len(filesToCheck))
			for j, f := range filesToCheck {
				var h hashTypeTest
				i := 0
				var gen string
				loop := time.Now()
				for start := time.Now(); time.Since(start) < 10*time.Second; {
					i++
					g, _, _ := getMapHash(f, []string{hash})
					gen = g[f].Hash[hash]
				}
				runtime := time.Since(loop).Seconds()
				h.Name = f
				h.NumberOf = i
				h.TimePer = runtime / float64(i)
				hashGroup[j] = h
				// GenResult, genErr := intTo4(numberToCheck[i])
				Convey("Checking the benchmark of each hash type", t, func() {
					Convey(fmt.Sprintf("using  %v as  the hash type", hash), func() {
						Convey(fmt.Sprintf("A hash of %v is expected and a hash of %v is generated", Expec[k][j], gen), func() {
							So(gen, ShouldEqual, Expec[k][j])
						})
					})
				})
			}

			hashCheck.RunInfo = hashGroup
			hashCheck.Name = hash
			hashNumber[k] = hashCheck
		}

		// Assign values to be saved in the json
		var test testResult
		host, _ := os.Hostname()
		test.Host = host
		test.Checked = hashNumber
		timeNow := time.Now()
		test.RunDate = timeNow.Format(xmlTFormat)

		prevJSON.Result = append(prevJSON.Result, test)

		b, _ := json.MarshalIndent(prevJSON, "", "   ")

		f, _ := os.OpenFile("testdata/results.json", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
		_, _ = f.Write(b)

		for i, hash := range hashToCheck {
			g, _, _ := getMapHash("./testdata", []string{hash})
			// fmt.Println(g)
			gen := g["./testdata/run1.txt"].Hash[hash]
			Convey("Checking the benchmark of each hash type", t, func() {
				Convey(fmt.Sprintf("using  %v as  the hash type", hash), func() {
					Convey(fmt.Sprintf("A hash of %v is expected and a hash of %v is generated", Expec[i][0], gen), func() {
						So(gen, ShouldEqual, Expec[i][0])
					})
				})
			})
		}
	}
}

func fileMake(files []string) {
	for _, fn := range files {
		fb := make([]byte, 25000000)
		rand.Read(fb)
		f, _ := os.OpenFile(fn, os.O_RDWR|os.O_CREATE, 0755)
		defer f.Close()
		_, _ = f.Write(fb)
	}
}

type testResults struct {
	Result []testResult `json:"Test Run Results"`
}

type testResult struct {
	Host    string  `json:"Host"`
	RunDate string  `json:"Time of test"`
	Checked []check `json:"Results by hash"`
}

type check struct {
	Name    string         `json:"Hash Type"`
	RunInfo []hashTypeTest `json:"Results"`
}

type hashTypeTest struct {
	Name     string  `json:"File name"`
	NumberOf int     `json:"Number of iterations"`
	TimePer  float64 `json:"time per hash(s)"`
}
