package toolkit

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
)

var tools Tools

func TestTools_RandomString(t *testing.T) {
	s := tools.RandomString(10)
	if len(s) != 10 {
		t.Error("returned string has the wrong length")
	}
}

var uploadTests = []struct {
	name          string
	allowedTypes  []string
	testImage     string
	renameFile    bool
	errorExpected bool
}{
	{name: "allowed no rename",
		testImage:     "./testdata/image.jpg",
		allowedTypes:  []string{"image/jpeg", "image/png"},
		renameFile:    false,
		errorExpected: false},
	{name: "allowed rename",
		testImage:     "./testdata/image.jpg",
		allowedTypes:  []string{"image/jpeg", "image/png"},
		renameFile:    true,
		errorExpected: false},
	{name: "file type not allowed",
		testImage:     "./testdata/image.jpg",
		allowedTypes:  []string{"image/png"},
		renameFile:    false,
		errorExpected: true},
}

func TestTools_UploadMultipleFiles(t *testing.T) {
	for _, e := range uploadTests {
		pr, pw := io.Pipe()
		writer := multipart.NewWriter(pw)
		wg := sync.WaitGroup{}
		wg.Add(1)

		go func() {
			defer writer.Close()
			defer wg.Done()

			part, err := writer.CreateFormFile("file", e.testImage)
			if err != nil {
				t.Error(err)
			}

			f, err := os.Open(e.testImage)
			if err != nil {
				t.Error(err)
			}
			defer f.Close()

			img, _, err := image.Decode(f)
			if err != nil {
				t.Error("error decoding image", err)
			}

			err = jpeg.Encode(part, img, nil)
			if err != nil {
				t.Error(err)
			}
		}()

		request := httptest.NewRequest("POST", "/", pr)
		request.Header.Add("Content-Type", writer.FormDataContentType())

		var testTools Tools
		testTools.AllowedFileTypes = e.allowedTypes

		uploadedFiles, err := testTools.UploadMultipleFiles(request,
			"./testdata/uploads/", e.renameFile)
		if err != nil && !e.errorExpected {
			t.Error(err)
		}

		if !e.errorExpected {
			newFilePath := fmt.Sprintf("./testdata/uploads/%s", uploadedFiles[0].NewFileName)
			if _, err := os.Stat(newFilePath); os.IsNotExist(err) {
				t.Errorf("%s: expected file to exist: %s", e.name, err.Error())
			}

		}

		_ = os.RemoveAll("./testdata/uploads")

		if !e.errorExpected && err != nil {
			t.Errorf("%s: error expected but none received", e.name)
		}

		wg.Wait()
	}
}

func TestTools_UploadOneFile(t *testing.T) {
	for _, e := range uploadTests {
		pr, pw := io.Pipe()
		writer := multipart.NewWriter(pw)
		wg := sync.WaitGroup{}
		wg.Add(1)

		go func() {
			defer writer.Close()
			defer wg.Done()

			part, err := writer.CreateFormFile("file", e.testImage)
			if err != nil {
				t.Error(err)
			}

			f, err := os.Open(e.testImage)
			if err != nil {
				t.Error(err)
			}
			defer f.Close()

			img, _, err := image.Decode(f)
			if err != nil {
				t.Error("error decoding image", err)
			}

			err = jpeg.Encode(part, img, nil)
			if err != nil {
				t.Error(err)
			}
		}()

		request := httptest.NewRequest("POST", "/", pr)
		request.Header.Add("Content-Type", writer.FormDataContentType())

		var testTools Tools
		testTools.AllowedFileTypes = e.allowedTypes

		uploadedFiles, err := testTools.UploadOneFile(request,
			"./testdata/uploads/", e.renameFile)
		if err != nil && !e.errorExpected {
			t.Error(err)
		}

		if !e.errorExpected {
			newFilePath := fmt.Sprintf("./testdata/uploads/%s", uploadedFiles.NewFileName)
			if _, err := os.Stat(newFilePath); os.IsNotExist(err) {
				t.Errorf("%s: expected file to exist: %s", e.name, err.Error())
			}

		}

		_ = os.RemoveAll("./testdata/uploads")

		if !e.errorExpected && err != nil {
			t.Errorf("%s: error expected but none received", e.name)
		}

		wg.Wait()
	}
}

func TestTools_CreateDirIfNotExist(t *testing.T) {
	var testTool Tools

	if err := testTool.CreateDirIfNotExists("./testdir/l1/l2"); err != nil {
		t.Error(err)
	}

	if err := testTool.CreateDirIfNotExists("./testdir/l1/l2"); err != nil {
		t.Error(err)
	}

	_ = os.RemoveAll("./testdir")
}

var slugTests = []struct {
	name          string
	s             string
	expected      string
	errorExpected bool
}{
	{name: "valid string", s: "now is  THE&(time)   ",
		expected: "now-is-the-time", errorExpected: false},
	{name: "empty string", s: "",
		expected: "", errorExpected: true},
	{name: "only non-latin string", s: "道可道，非常道。名可名，非常名",
		expected: "", errorExpected: true},
	{name: "non-latin and latin string", s: "道可道，非常道。名可名，非常名 - The Dao",
		expected: "the-dao", errorExpected: false},
}

func TestTools_Slugify(t *testing.T) {
	for _, e := range slugTests {
		slug, err := tools.Slugify(e.s)
		if err != nil && !e.errorExpected {
			t.Errorf("%s: error received when none expected: %s\n", e.name, err.Error())
		}
		if slug != e.expected && !e.errorExpected {
			t.Errorf("%s: wrong slug returned; expected %s, got %s\n", e.name, e.expected, slug)
		}
	}
}

func TestTools_DownloadStaticFile(t *testing.T) {
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)

	var testTool Tools
	const pathName = "./testdata/image.jpg"
	const downloadName = "download.jpg"

	fileStats, err := os.Stat(pathName)
	if err != nil {
		t.Error(err)
	}

	fileSize := fmt.Sprintf("%d", fileStats.Size())

	testTool.DownloadStaticFile(rr, req, pathName, "download.jpg")

	res := rr.Result()
	defer res.Body.Close()

	lengthHeader := res.Header["Content-Length"][0]

	if lengthHeader != fileSize {
		t.Error("wrong content length of", lengthHeader)
	}

	if res.Header["Content-Disposition"][0] !=
		fmt.Sprintf("attachment; filename=\"%s\"", downloadName) {
		t.Error("wrong content disposition")
	}
}

var jsonTests = []struct {
	name          string
	json          string
	errorExpected bool
	maxSize       int64
	allowUnknown  bool
}{
	{name: "good JSON", json: `{"message": "hello world"}`,
		errorExpected: false, maxSize: 1024, allowUnknown: true},
	{name: "badly formatted JSON", json: `{message: "hello world"}`,
		errorExpected: true, maxSize: 1024, allowUnknown: true},
	{name: "syntax error in JSON", json: `{"message": hello world"}`,
		errorExpected: true, maxSize: 1024, allowUnknown: true},
	{name: "size too large", json: `{"message": "hello world"}`,
		errorExpected: true, maxSize: 2, allowUnknown: true},
	{name: "wrong type", json: `{"message": true}`,
		errorExpected: true, maxSize: 1024, allowUnknown: true},
	{name: "empty body", json: "",
		errorExpected: true, maxSize: 1024, allowUnknown: true},
	{name: "unknown key, not allowed", json: `{"message": "hello world", "id": 1}`,
		errorExpected: true, maxSize: 1024, allowUnknown: false},
	{name: "unknown key, allowed", json: `{"message": "hello world", "id": 1}`,
		errorExpected: false, maxSize: 1024, allowUnknown: true},
	{name: "multiple JSON bodies", json: `{"message": "hello world"}{"message": "hello world"}`,
		errorExpected: true, maxSize: 1024, allowUnknown: false},
	{name: "sent XML", json: `<message>hello world</message>`,
		errorExpected: true, maxSize: 1024, allowUnknown: false},
	{name: "unknown field", json: `{"msg": "hello world"}`,
		errorExpected: true, maxSize: 1024, allowUnknown: false},
}

func TestTools_ReadJSON(t *testing.T) {
	var testTool Tools
	for _, e := range jsonTests {
		testTool.MaxJSONSize = e.maxSize
		testTool.AllowUnknownFields = e.allowUnknown

		var decodedJSON struct {
			Message string `json:"message"`
		}

		req, err := http.NewRequest("POST", "/", bytes.NewReader([]byte(e.json)))
		if err != nil {
			t.Log("Error:", err)
		}

		rr := httptest.NewRecorder()

		err = testTool.ReadJSON(rr, req, &decodedJSON)
		if err == nil && e.errorExpected {
			t.Errorf("%s: error expected but none received\n", e.name)
		}

		if !e.errorExpected && err != nil {
			t.Errorf("%s: error received when none expected: %s\n", e.name, err.Error())
		}

		req.Body.Close()
	}
}

func TestTools_WriteJSON(t *testing.T) {
	var testTool Tools

	rr := httptest.NewRecorder()

	payload := struct {
		Message string `json:"message"`
	}{
		Message: "hello world",
	}

	headers := make(http.Header)
	headers.Add("hello", "world")

	err := testTool.WriteJSON(rr, http.StatusOK, payload, headers)
	if err != nil {
		t.Errorf("failed to write JSON: %v\n", err)
	}
}

func TestTools_ErrorJSON(t *testing.T) {
	var testTool Tools

	rr := httptest.NewRecorder()
	statusCode := http.StatusTeapot

	err := testTool.ErrorJSON(rr, errors.New("test error"), statusCode)
	if err != nil {
		t.Error(err)
	}

	var payload JSONResponse
	decoder := json.NewDecoder(rr.Body)
	err = decoder.Decode(&payload)
	if err != nil {
		t.Error("received error when decoding JSON:", err)
	}

	if !payload.Error {
		t.Error("error sete to false in JSON, should be set to true")
	}

	if rr.Code != statusCode {
		t.Errorf("wrong status code returned; expected %d, got %d", statusCode, rr.Code)
	}
}

type RoundTripFunc func(req *http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: fn,
	}
}

func TestTools_PushJSONToRemote(t *testing.T) {
	client := NewTestClient(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString("ok")),
			Header:     make(http.Header),
		}
	})

	var testTool Tools

	payload := struct {
		Message string `json:"message"`
	}{
		Message: "hello world",
	}

	_, _, err := testTool.PushJSONToRemote("http://example.com/test",
		payload, client)
	if err != nil {
		t.Error("failed to call remote url - ", err)
	}
}
