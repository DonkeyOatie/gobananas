// golang_blog

package main

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/mholt/binding"
	"github.com/unrolled/render"
)

type PostList struct {
	Posts        []Post
	Next_Page    int
	Prev_Page    int
	Current_Page int
	More         bool
	Less         bool
}

type PostPage struct {
	Post     Post
	Comments []Comment
}

func (c *Comment) FieldMap() binding.FieldMap {
	return binding.FieldMap{
		&c.Author:  "author",
		&c.Body:    "comment",
		&c.Post_Id: "post_id",
	}
}

var ADMIN_USER = os.Getenv("BA_USER")
var ADMIN_PASS = os.Getenv("BA_PASS")

var TEMPLATE_DIR = os.Getenv("TEMPLATE_DIR")
var STATIC_DIR = os.Getenv("STATIC_DIR")

const NUMBER_POSTS_PER_PAGE = 5

// renderer is our global renderer, used for returning pretty JSON
var renderer = render.New(render.Options{IndentJSON: true})

// parse the templates once and hold them in memory
var templates = template.Must(template.ParseGlob(fmt.Sprintf("%s/*", TEMPLATE_DIR)))

func main() {
	if ADMIN_USER == "" || ADMIN_PASS == "" {
		log.Fatalln("need to set admin username and password")
	}
	fs := http.FileServer(http.Dir(STATIC_DIR))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	r := mux.NewRouter().StrictSlash(false)

	// Home Page
	r.HandleFunc("/", handleViewHomePage).Methods("GET")

	// blog list view
	r.HandleFunc("/blog/", handleViewBlogList).Methods("GET")
	r.HandleFunc("/blog/{page}", handleViewBlogList).Methods("GET")

	// post detail view
	r.HandleFunc("/article/{id}", handleViewBlogPost).Methods("GET")

	// post add view
	r.HandleFunc("/article/", handleAddBlogPost).Methods("POST")

	r.HandleFunc("/article/update/{id}", handleUpdateBlogPost).Methods("POST")

	// contact view
	r.HandleFunc("/contact/", handleViewContactPage).Methods("GET")

	// create a comment
	r.HandleFunc("/comment/", handleAddCommentToPost).Methods("POST")

	http.Handle("/", r)

	http.ListenAndServe(":8000", nil)
}

// handleViewHomePage renders the index.html template
func handleViewHomePage(w http.ResponseWriter, req *http.Request) {
	templates.ExecuteTemplate(w, "index", nil)
}

// handleViewBlogList renders the list.html template
func handleViewBlogList(w http.ResponseWriter, req *http.Request) {
	var page_int int
	var offset int
	var post_list PostList
	var err error

	vars := mux.Vars(req)
	if page, ok := vars["page"]; ok {
		page_int, err = strconv.Atoi(page)
		if err != nil {
			renderer.JSON(w, 400, "Invalid page request")
			return
		}
	} else {
		page_int = 1
	}

	offset = (page_int - 1) * NUMBER_POSTS_PER_PAGE
	nPosts, _ := getNumberOfPosts()

	post_list.Posts, _ = getBlogPosts(offset)

	post_list.More = (offset + NUMBER_POSTS_PER_PAGE) < nPosts
	post_list.Less = page_int > 1

	post_list.Next_Page = page_int + 1
	post_list.Prev_Page = page_int - 1
	post_list.Current_Page = page_int

	templates.ExecuteTemplate(w, "blog_list", post_list)
}

// handleViewBlogPost renders a single blog post corresponding to the id in the URL
func handleViewBlogPost(w http.ResponseWriter, req *http.Request) {
	var post_page PostPage

	vars := mux.Vars(req)
	id, err := strconv.Atoi(vars["id"])

	if err != nil {
		renderer.JSON(w, 400, "Invalid article request")
		return
	}

	post, _ := getBlogPost(id)
	comments, _ := getCommentsForPost(id)

	post_page.Post = post
	post_page.Comments = comments

	templates.ExecuteTemplate(w, "blog", &post_page)
}

// handleViewContactPage renders the contact template
func handleViewContactPage(w http.ResponseWriter, req *http.Request) {
	templates.ExecuteTemplate(w, "contact", nil)
}

// handleAddCommentToPost takes comment data and adds it to the post
func handleAddCommentToPost(w http.ResponseWriter, req *http.Request) {
	var comment Comment

	binding.Bind(req, &comment)
	err := createComment(comment)
	if err != nil {
		renderer.JSON(w, 500, "Failed to create comment")
		return
	}
	url := fmt.Sprintf("/article/%d", comment.Post_Id)
	http.Redirect(w, req, url, http.StatusFound)
}

// checkBasicAuth checks the admin username and password
func checkBasicAuth(w http.ResponseWriter, req *http.Request) bool {
	username, password, _ := req.BasicAuth()
	if username != os.Getenv("BA_USER") || password != os.Getenv("BA_PASS") {
		renderer.JSON(w, 403, "Be gone, pest")
		return false
	}
	return true
}

// handleAddBlogTask tasks a file and information posted to this endpoint and saves
// it in the DB as a blog post
func handleAddBlogPost(w http.ResponseWriter, req *http.Request) {
	if !checkBasicAuth(w, req) {
		return
	}

	var post Post

	title := req.FormValue("title")
	if title == "" {
		renderer.JSON(w, 400, "you have to have a title!")
		return
	}

	content, err := processFileUpload(w, req)
	if err != nil {
		return
	}
	post.Body = content
	post.Title = title
	createBlogPost(post)
}

// handleUpdateBlogPost updates the post with the given id
func handleUpdateBlogPost(w http.ResponseWriter, req *http.Request) {
	if !checkBasicAuth(w, req) {
		return
	}

	vars := mux.Vars(req)
	id, err := strconv.Atoi(vars["id"])

	if err != nil {
		renderer.JSON(w, 400, "Invalid article request")
		return
	}

	post, _ := getBlogPost(id)

	content, err := processFileUpload(w, req)
	if err != nil {
		return
	}

	post.Body = content
	updateBlogPost(id, post)
}

// processFileUpload takes an uploaded file and returns a byte stream
func processFileUpload(w http.ResponseWriter, req *http.Request) ([]byte, error) {
	var content []byte

	// the FormFile function takes in the POST input id file
	file, _, err := req.FormFile("file")

	if err != nil {
		renderer.JSON(w, 500, "Failed to receive file")
		return content, err
	}

	defer file.Close()

	out, err := os.Create("/tmp/new_post")
	if err != nil {
		renderer.JSON(w, 500, "Failed to create tmp file")
		return content, err
	}

	defer out.Close()

	//  write the content from POST to the file
	_, err = io.Copy(out, file)
	if err != nil {
		return content, err
	}

	content, err = ioutil.ReadFile("/tmp/new_post")
	return content, err
}
