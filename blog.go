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

// FieldMap maps the POST parameters to the comment struct members
func (c *comment) FieldMap(httpRequest *http.Request) binding.FieldMap {
	return binding.FieldMap{
		&c.Author: "author",
		&c.Body:   "comment",
		&c.PostID: "post_id",
	}
}

var adminUser = os.Getenv("BA_USER")
var adminPass = os.Getenv("BA_PASS")

var templateDir = os.Getenv("TEMPLATE_DIR")
var staticDir = os.Getenv("STATIC_DIR")

const numberPostsPerPage = 5

// renderer is our global renderer, used for returning pretty JSON
var renderer = render.New(render.Options{IndentJSON: true})

// parse the templates once and hold them in memory
var templates = template.Must(template.ParseGlob(fmt.Sprintf("%s/*", templateDir)))

func handlers() *mux.Router {
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

	return r
}

func main() {
	if adminUser == "" || adminPass == "" {
		log.Fatalln("need to set admin username and password")
	}
	fs := http.FileServer(http.Dir(staticDir))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.Handle("/", handlers())

	http.ListenAndServe(":8000", nil)
}

// handleViewHomePage renders the index.html template
func handleViewHomePage(w http.ResponseWriter, req *http.Request) {
	templates.ExecuteTemplate(w, "index", nil)
}

// handleViewBlogList renders the list.html template
func handleViewBlogList(w http.ResponseWriter, req *http.Request) {
	var pageInt int
	var offset int
	var err error

	vars := mux.Vars(req)
	if page, ok := vars["page"]; ok {
		pageInt, err = strconv.Atoi(page)
		if err != nil {
			renderer.JSON(w, 400, "Invalid page request")
			return
		}
	} else {
		pageInt = 1
	}

	offset = (pageInt - 1) * numberPostsPerPage
	nPosts, _ := getNumberOfPosts()
	posts, _ := getBlogPosts(offset)

	postList := struct {
		Posts       []post
		NextPage    int
		PrevPage    int
		CurrentPage int
		More        bool
		Less        bool
	}{
		posts,
		pageInt + 1,
		pageInt - 1,
		pageInt,
		(offset + numberPostsPerPage) < nPosts,
		pageInt > 1,
	}

	templates.ExecuteTemplate(w, "blog_list", postList)
}

// handleViewBlogPost renders a single blog post corresponding to the id in the URL
func handleViewBlogPost(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	id, err := strconv.Atoi(vars["id"])

	if err != nil {
		renderer.JSON(w, 400, "Invalid article request")
		return
	}

	p, _ := getBlogPost(id)
	comments, _ := getCommentsForPost(id)

	postPage := struct {
		Post     post
		Comments []comment
	}{
		p,
		comments,
	}

	templates.ExecuteTemplate(w, "blog", postPage)
}

// handleViewContactPage renders the contact template
func handleViewContactPage(w http.ResponseWriter, req *http.Request) {
	templates.ExecuteTemplate(w, "contact", nil)
}

// handleAddCommentToPost takes comment data and adds it to the post
func handleAddCommentToPost(w http.ResponseWriter, req *http.Request) {
	var c comment

	binding.Bind(req, &c)
	err := createComment(c)
	if err != nil {
		renderer.JSON(w, 500, "Failed to create comment")
		return
	}
	url := fmt.Sprintf("/article/%d", c.PostID)
	http.Redirect(w, req, url, http.StatusFound)
}

// checkBasicAuth checks the admin username and password
func checkBasicAuth(w http.ResponseWriter, req *http.Request) bool {
	username, password, _ := req.BasicAuth()
	if username != adminUser || password != adminPass {
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

	var post post

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
