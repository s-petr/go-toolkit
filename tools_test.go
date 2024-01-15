package toolkit

import (
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"mime/multipart"
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
