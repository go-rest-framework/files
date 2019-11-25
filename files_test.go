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
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/go-rest-framework/core"
	"github.com/go-rest-framework/files"
	"github.com/go-rest-framework/users"
	"github.com/icrowley/fake"
)

var OneNewID uint
var OneANewID uint
var AdminToken string
var OneGroup string
var OneTitle string
var NewOneTitle string
var NewAOneTitle string
var TestFileUserID int
var TestFileID uint
var TestFileHash string
var Murl = "http://localhost/api/files"
var AMurl = "http://localhost/api/attachments"

type TestFiles struct {
	Errors []core.ErrorMsg `json:"errors"`
	Data   files.Files     `json:"data"`
}

type TestFile struct {
	Errors []core.ErrorMsg `json:"errors"`
	Data   files.File      `json:"data"`
}

type TestAttachments struct {
	Errors []core.ErrorMsg   `json:"errors"`
	Data   files.Attachments `json:"data"`
}

type TestAttachment struct {
	Errors []core.ErrorMsg  `json:"errors"`
	Data   files.Attachment `json:"data"`
}

type TestUser struct {
	Errors []core.ErrorMsg `json:"errors"`
	Data   users.User      `json:"data"`
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

func doUpload(url, proto, filepath string) *http.Response {
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

			fw, err := w.CreateFormFile(key, x.Name())

			if err != nil {
				log.Fatal(err)
			}

			if _, err := io.Copy(fw, r); err != nil {
				log.Fatal(err)
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
	req.Header.Set("Authorization", "Bearer "+AdminToken)

	// Submit the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	return resp
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

func readAttachmentBody(r *http.Response, t *testing.T) TestAttachment {
	var u TestAttachment
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal([]byte(body), &u)
	return u
}

func readAttachmentsBody(r *http.Response, t *testing.T) TestAttachments {
	var u TestAttachments
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal([]byte(body), &u)
	return u
}

func readUserBody(r *http.Response, t *testing.T) TestUser {
	var u TestUser
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal([]byte(body), &u)
	defer r.Body.Close()
	return u
}

func toUrlcode(str string) (string, error) {
	u, err := url.Parse(str)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func deleteFile(t *testing.T, id uint) {
	url := fmt.Sprintf("%s%s%d", Murl, "/", id)

	resp := doRequest(url, "DELETE", "", AdminToken)

	if resp.StatusCode != 200 {
		t.Errorf("Success expected: %d", resp.StatusCode)
	}

	u := readFileBody(resp, t)

	if len(u.Errors) != 0 {
		t.Errorf("Error when delete id = %d", id)
		t.Fatal(u.Errors)
	}

	return
}

func TestAdminLogin(t *testing.T) {

	url := "http://localhost/api/users/login"
	var userJson = `{"email":"admin@admin.a", "password":"adminpass"}`

	resp := doRequest(url, "POST", userJson, "")

	if resp.StatusCode != 200 {
		t.Errorf("Success expected: %d", resp.StatusCode)
	}

	u := readUserBody(resp, t)

	AdminToken = u.Data.Token

	return
}

func TestUpload(t *testing.T) {
	resp := doUpload(Murl, "POST", "test_pic.png")

	if resp.StatusCode != http.StatusOK {
		t.Errorf("bad status: %s", resp.Status)
	}

	u := readFileBody(resp, t)

	if len(u.Errors) != 0 {
		t.Fatal(u.Errors)
	}

	TestFileID = u.Data.ID
	TestFileUserID = u.Data.UserID
	TestFileHash = u.Data.Hash
}

func TestAttachmentCreate(t *testing.T) {
	url := AMurl
	OneGroup = fake.Word()
	OneTitle = fake.Title()
	el := &files.Attachment{
		UserID:      TestFileUserID,
		Group:       OneGroup,
		FileID:      int(TestFileID),
		Title:       OneTitle,
		Description: fake.Paragraphs(),
		IsMain:      0,
		Hash:        TestFileHash,
		Index:       0,
	}

	uj, err := json.Marshal(el)
	if err != nil {
		fmt.Printf("Error: %s", err)
		return
	}

	resp := doRequest(url, "POST", string(uj), AdminToken)

	if resp.StatusCode != 200 {
		t.Errorf("Success expected: %d", resp.StatusCode)
	}

	u := readAttachmentBody(resp, t)

	if len(u.Errors) != 0 {
		t.Fatal(u.Errors)
	}

	OneANewID = u.Data.ID

	return
}

func TestGetOne(t *testing.T) {
	url := Murl + "/0"
	resp := doRequest(url, "GET", "", " ")

	if resp.StatusCode != 200 {
		t.Errorf("Success expected: %d", resp.StatusCode)
	}

	u := readFileBody(resp, t)

	if len(u.Errors) == 0 {
		t.Fatal("element not found dont work")
	}

	url = fmt.Sprintf("%s%s%d", Murl, "/", TestFileID)

	resp = doRequest(url, "GET", "", " ")

	if resp.StatusCode != 200 {
		t.Errorf("Success expected: %d", resp.StatusCode)
	}

	u = readFileBody(resp, t)

	if len(u.Errors) != 0 {
		t.Fatal(u.Errors)
	}

	return
}

func TestAttachmentGetOne(t *testing.T) {
	url := AMurl + "/0"
	resp := doRequest(url, "GET", "", " ")

	if resp.StatusCode != 200 {
		t.Errorf("Success expected: %d", resp.StatusCode)
	}

	u := readAttachmentBody(resp, t)

	if len(u.Errors) == 0 {
		t.Fatal("element not found dont work")
	}

	url = fmt.Sprintf("%s%s%d", AMurl, "/", OneANewID)

	resp = doRequest(url, "GET", "", " ")

	if resp.StatusCode != 200 {
		t.Errorf("Success expected: %d", resp.StatusCode)
	}

	u = readAttachmentBody(resp, t)

	if len(u.Errors) != 0 {
		t.Fatal(u.Errors)
	}

	if u.Data.File.ID != TestFileID {
		t.Errorf("No file object in attachment: %+v", u.Data)
	}

	return

}

func TestGetAll(t *testing.T) {
	// get count
	url := Murl

	resp := doRequest(url, "GET", "", " ")

	if resp.StatusCode != 200 {
		t.Errorf("Success expected: %d", resp.StatusCode)
	}

	u := readFilesBody(resp, t)

	if len(u.Errors) != 0 {
		t.Fatal(u.Errors)
	}

	if len(u.Data) == 0 {
		t.Errorf("Wrong elements count: %d", len(u.Data))
	}

	//---------------

	uname, _ := toUrlcode("test_pic")

	url1 := Murl + "?name=" + uname

	resp1 := doRequest(url1, "GET", "", " ")

	if resp1.StatusCode != 200 {
		t.Errorf("Success expected: %d%s", resp1.StatusCode, url1)
	}

	u1 := readFilesBody(resp1, t)

	if len(u1.Errors) != 0 {
		t.Fatal(u1.Errors)
	}

	if u1.Data[0].Name != "test_pic" {
		t.Errorf("Wrong name search - : %s", u1.Data[0].Name)
	}

	//---------------

	url2 := Murl + "?limit=1"

	resp2 := doRequest(url2, "GET", "", " ")

	if resp2.StatusCode != 200 {
		t.Errorf("Success expected: %d %s", resp2.StatusCode, url2)
	}

	u2 := readFilesBody(resp2, t)

	if len(u2.Errors) != 0 {
		t.Fatal(u2.Errors)
	}

	if len(u2.Data) != 1 {
		t.Errorf("Wrong search limit: %d %s", len(u2.Data), url2)
	}

	return
}

func TestAttachmentGetGroup(t *testing.T) {
	// get count
	url := AMurl

	resp := doRequest(url, "GET", "", " ")

	if resp.StatusCode != 200 {
		t.Errorf("Success expected: %d", resp.StatusCode)
	}

	u := readAttachmentsBody(resp, t)

	if len(u.Errors) != 0 {
		t.Fatal(u.Errors)
	}

	if len(u.Data) == 0 {
		t.Errorf("Wrong elements count: %d", len(u.Data))
	}

	//---------------

	group, _ := toUrlcode(OneGroup)

	url1 := AMurl + "?group=" + group

	resp1 := doRequest(url1, "GET", "", " ")

	if resp1.StatusCode != 200 {
		t.Errorf("Success expected: %d%s", resp1.StatusCode, url1)
	}

	u1 := readAttachmentsBody(resp1, t)

	if len(u1.Errors) != 0 {
		t.Fatal(u1.Errors)
	}

	if u1.Data[0].Title != OneTitle {
		t.Errorf("Wrong group search - : %s", u1.Data[0].Title)
	}

	return
}

func TestUpdate(t *testing.T) {
	url := fmt.Sprintf("%s%s%d", Murl, "/", TestFileID)
	resp := doUpload(url, "PATCH", "test_pic2.png")

	if resp.StatusCode != http.StatusOK {
		t.Errorf("bad status: %s", resp.Status)
	}

	u := readFileBody(resp, t)

	if len(u.Errors) != 0 {
		t.Fatal(u.Errors)
	}

	OneNewID = u.Data.ID
}

func TestAttachmentUpdate(t *testing.T) {
	NewAOneTitle = fake.Title()
	url := fmt.Sprintf("%s%s%d", AMurl, "/", OneANewID)
	userJson := `{"title":"` + NewAOneTitle + `"}`

	resp := doRequest(url, "PATCH", userJson, AdminToken)

	if resp.StatusCode != 200 {
		t.Errorf("Success expected: %d", resp.StatusCode)
	}

	u := readAttachmentBody(resp, t)

	if len(u.Errors) != 0 {
		t.Fatal(u.Errors)
	}

	if u.Data.Title != NewAOneTitle {
		t.Errorf("Wrong new title : %+v", u.Data)
	}

	return
}

func TestDelete(t *testing.T) {
	url := fmt.Sprintf("%s%s%d", Murl, "/", 0)

	resp := doRequest(url, "DELETE", "", AdminToken)

	if resp.StatusCode != 200 {
		t.Errorf("Success expected: %d", resp.StatusCode)
	}

	u := readFileBody(resp, t)

	if len(u.Errors) == 0 {
		t.Fatal("wrong id validation dont work")
	}

	deleteFile(t, TestFileID)

	return
}

func TestAttachmentDelete(t *testing.T) {
	url := fmt.Sprintf("%s%s%d", AMurl, "/", 0)

	resp := doRequest(url, "DELETE", "", AdminToken)

	if resp.StatusCode != 200 {
		t.Errorf("Success expected: %d", resp.StatusCode)
	}

	u := readAttachmentBody(resp, t)

	if len(u.Errors) == 0 {
		t.Fatal("wrong id validation dont work")
	}

	url = fmt.Sprintf("%s%s%d", AMurl, "/", OneANewID)

	resp = doRequest(url, "DELETE", "", AdminToken)

	if resp.StatusCode != 200 {
		t.Errorf("Success expected: %d", resp.StatusCode)
	}

	u = readAttachmentBody(resp, t)

	if len(u.Errors) != 0 {
		t.Fatal(u.Errors)
	}

	return
}
