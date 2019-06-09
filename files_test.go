package files_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/go-rest-framework/core"
	"github.com/go-rest-framework/files"
)

var Murl = "http://gorest.ga/api/files"

type TestFiles struct {
	Errors []core.ErrorMsg `json:"errors"`
	Data   files.Files     `json:"data"`
}

type TestFile struct {
	Errors []core.ErrorMsg `json:"errors"`
	Data   files.File      `json:"data"`
}

func doRequest(url, proto, userJson, token string) *http.Response {
	reader := strings.NewReader(userJson)
	request, err := http.NewRequest(proto, url, reader)
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(request)

	if err != nil {
		log.Fatal(err)
	}
	return resp
}

func doUpload(url, proto, filepath string) {
	//prepare the reader instances to encode
	values := map[string]io.Reader{
		"file": mustOpen(filepath), // lets assume its this file
		//"other": strings.NewReader("hello world!"),
	}

	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for key, r := range values {
		if x, ok := r.(io.Closer); ok {
			defer x.Close()
		}
		// Add an image file
		if x, ok := r.(*os.File); ok {
			if fw, err := w.CreateFormFile(key, x.Name()); err != nil {
				log.Fatal(err)
				if _, err := io.Copy(fw, r); err != nil {
					log.Fatal(err)
				}
			}

		}
		//} else {
		// Add other fields
		//if fw, err := w.CreateFormField(key); err != nil {
		//log.Fatal(err)
		//}
		//}
	}
	// Don't forget to close the multipart writer.
	// If you don't close it, your request will be missing the terminating boundary.
	w.Close()

	// Now that you have a form, you can submit it to your handler.
	req, err := http.NewRequest(proto, url, &b)
	if err != nil {
		log.Fatal(err)
	}
	// Don't forget to set the content type, this will contain the boundary.
	req.Header.Set("Content-Type", w.FormDataContentType())

	// Submit the request
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	// Check the response
	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("bad status: %s", res.Status)
	}
}

func mustOpen(f string) *os.File {
	r, err := os.Open(f)
	if err != nil {
		panic(err)
	}
	return r
}

func readFileBody(r *http.Response, t *testing.T) TestFile {
	var u TestFile
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal([]byte(body), &u)
	return u
}

func readFilesBody(r *http.Response, t *testing.T) TestFiles {
	var u TestFiles
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal([]byte(body), &u)
	return u
}

func TestUpload(t *testing.T) {
	doUpload(Murl, "POST", "test.png")
}
