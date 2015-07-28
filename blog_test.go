package main

import (
	"bytes"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
)

var server *httptest.Server

// Util functions to aid testing

func checkError(t *testing.T, err error) {
	if err != nil {
		t.Error(err)
	}
}

func checkStatus(t *testing.T, res *http.Response, expected int) {
	if res.StatusCode != expected {
		t.Errorf("Status=%d, expected=%d", res.StatusCode, expected)
	}
}

func checkString(t *testing.T, value string, expected string) {
	if value != expected {
		t.Errorf("value=%s, expected=%s", value, expected)
	}
}

func doHTTP(t *testing.T, method string, url string) (*testing.T, *http.Response) {
	req, err := http.NewRequest(method, url, nil)
	checkError(t, err)

	res, err := http.DefaultClient.Do(req)
	checkError(t, err)
	return t, res
}

// SetUpDatabase loads the schema.sql file into the test database and loads
// fixtures from fixtures/load_fixtures.sql
func SetUpDatabase(db *sqlx.DB) {
	//load the latest schema
	_, err := sqlx.LoadFile(db, "schema.sql")

	if err != nil {
		log.Fatalln("Failed to load schema", err)
	}

	// load our test data
	_, err = sqlx.LoadFile(db, "fixtures/load_fixtures.sql")

	if err != nil {
		TearDownDatabase(db)
		log.Fatalln("Failed to load fixtures", err)
	}
}

// TearDownDatabase delets all tables in test database so that the next test run
// starts from a clearn database;
func TearDownDatabase(db *sqlx.DB) {
	// destroy all tables to reset sequences and remove test data
	_, err := sqlx.LoadFile(db, "fixtures/delete_fixtures.sql")

	if err != nil {
		log.Fatalln("Failed to delete fixtures, run delete_fixtures.sql manually to continue", err)
	}
}

// TestMain runs before and after the tests, we use this to create our database
// with the latest schema and load fixture data, then destroy it after the
// tests finish
func TestMain(m *testing.M) {
	server = httptest.NewServer(handlers())
	db = initDb()
	SetUpDatabase(db)

	res := m.Run()

	TearDownDatabase(db)

	os.Exit(res)
}

// GET /
func TestGetIndexView(t *testing.T) {
	t, res := doHTTP(t, "GET", server.URL+"/")
	checkStatus(t, res, 200)
}

// GET /blog/
func TestGetBlogListView(t *testing.T) {
	t, res := doHTTP(t, "GET", server.URL+"/blog/")
	checkStatus(t, res, 200)
}

// GET /contact/
func TestGetContactView(t *testing.T) {
	t, res := doHTTP(t, "GET", server.URL+"/contact/")
	checkStatus(t, res, 200)
}

// GET /blog/:page_number
func TestGetBlogSecondPageView(t *testing.T) {
	t, res := doHTTP(t, "GET", server.URL+"/blog/1")
	checkStatus(t, res, 200)
}

// GET /article/:id
func TestGetBlogArticleView(t *testing.T) {
	t, res := doHTTP(t, "GET", server.URL+"/article/1")
	checkStatus(t, res, 200)
}

// POST /arcticle/
func TestAddBlogArticle(t *testing.T) {
	path := "./fixtures/test_upload.html"
	file, _ := os.Open(path)
	defer file.Close()

	body := &bytes.Buffer{}

	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", filepath.Base(path))
	io.Copy(part, file)

	fw, _ := writer.CreateFormField("title")
	fw.Write([]byte("test upload"))
	writer.Close()

	req, _ := http.NewRequest("POST", server.URL+"/article/", body)
	req.SetBasicAuth("test", "test")
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	res, _ := client.Do(req)

	checkStatus(t, res, 200)
}

// POST /article/update/:id
func TestUpdateBlogArticle(t *testing.T) {
	path := "./fixtures/test_upload.html"
	file, _ := os.Open(path)
	defer file.Close()

	body := &bytes.Buffer{}

	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", filepath.Base(path))
	io.Copy(part, file)

	fw, _ := writer.CreateFormField("title")
	fw.Write([]byte("test upload updated"))
	writer.Close()

	req, _ := http.NewRequest("POST", server.URL+"/article/update/1", body)
	req.SetBasicAuth("test", "test")
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	res, _ := client.Do(req)

	checkStatus(t, res, 200)
}

// POST /comment/
func TestAddCommentToArticle(t *testing.T) {
	form := url.Values{}
	form.Set("post_id", "1")
	form.Set("author", "comment@example.org")
	form.Set("body", "golang testing")

	req, _ := http.NewRequest("POST", server.URL+"/comment/", strings.NewReader(form.Encode()))
	req.Form = form
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	res, _ := client.Do(req)

	checkStatus(t, res, 200)
}
