package reporting

import (
	"io/ioutil"
	"log"
)

// Create a file containing the provided content
// It is the caller's responsibility to delete the file.
// eg
//   tempFile, err := CreateTempFileWithContent(contentBytes)
//   if err != nil {
//     log.Fatal(err)
//   }
//   defer os.Remove(tempFile)
func createTempFileWithContent(content string) (string, error) {
	tmpfile, err := ioutil.TempFile("", "analyze*.rb")
	if err != nil {
		log.Fatal(err)
	}
	if _, err := tmpfile.Write([]byte(content)); err != nil {
		return "", err
	}
	if err := tmpfile.Close(); err != nil {
		return "", err
	}
	return tmpfile.Name(), nil

}
