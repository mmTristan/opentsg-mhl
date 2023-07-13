package ascmhl

import "github.com/spf13/afero"

func init() {
	AppFS = afero.NewOsFs()
}

// AppFS is the file system the mhlgeneration is taking place in.
var AppFS afero.Fs

/*
// Make a function here that pulls based on the information
// Test to ensure the afero.fs type is correct for each one
func OsAssign(ftype, source string) error {
	switch ftype {
	case "os", "":
		AppFS = afero.NewOsFs()
	case "s3":

		// get a way to extract the relevant information


		sess, err := session.NewSession(&aws.Config{
			Region:      aws.String(region),
			Credentials: credentials.NewStaticCredentials(keyID, secretAccessKey, ""),
		})

		svc := s3.New(sess)
		result, err := svc.ListBuckets(nil)
		if err != nil {
			fmt.Printf("Unable to list buckets, %v", err)
		}

		fmt.Println("Buckets:")

		for _, b := range result.Buckets {
			fmt.Printf("* %s created on %s\n",
				aws.StringValue(b.Name), aws.TimeValue(b.CreationDate))
		}

		resp, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{Bucket: aws.String(bucket),
			Prefix: aws.String("bot-tlh/")})
		if err != nil {
			exitErrorf("Unable to list items in bucket %q, %v", bucket, err)
		}
		for _, item := range resp.Contents {
			fmt.Println("Name:         ", *item.Key)
			fmt.Println("Last modified:", *item.LastModified)
			fmt.Println("Size:         ", *item.Size)
			fmt.Println("Storage class:", *item.StorageClass)
			fmt.Println("")
		}
		_, e := AppFS.Open("pkg/chain.go")
		fmt.Println(e, "FOR OPening")

		// Fmt.Println(resp.Contents)
		// Initialize the file system
		s3Fs := s3Afero.NewFs(bucket, sess)
		fmt.Println(s3Fs, "mybucket")
		AppFS = s3Fs
		fmt.Println(FindDirs("bot-tlh/"))
		fmt.Println(AppFS.Open("bot-tlh/"))
	}
	return nil
}

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
}
*/
