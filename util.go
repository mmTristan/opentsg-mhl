package ascmhl

/*
The schemacheck util has been removed as libxml is very difficult to get on windows

import (
	"github.com/lestrrat-go/libxml2"
	"github.com/lestrrat-go/libxml2/xsd"
)

func schemaCheck(schema, file []byte) error {
	// Schema parsing section
	s, err := xsd.Parse(schema)

	defer s.Free()

	if err != nil {
		return err
	}
	d, _ := libxml2.Parse(file)
	err = s.Validate(d)

	if err != nil {
		/* fmt.Println(string(file))
		for _, e := range err.(xsd.SchemaValidationError).Errors() {
			fmt.Println(e)
		}

		return err
	}

	return nil
}
*/
