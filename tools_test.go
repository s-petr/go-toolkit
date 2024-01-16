package toolkit

import (
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
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
	const filename = "image.jpg"
	const downloadName = "download.jpg"
	const folder = "./testdata"

	fileStats, err := os.Stat(path.Join(folder, filename))
	if err != nil {
		t.Error(err)
	}

	fileSize := fmt.Sprintf("%d", fileStats.Size())

	testTool.DownloadStaticFile(rr, req, folder, filename, "download.jpg")

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
