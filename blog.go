// golang_blog

package main

import (
	"fmt"
	"html/template"
	"net/http"
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

const NUMBER_POSTS_PER_PAGE = 5

// renderer is our global renderer, used for returning pretty JSON
var renderer = render.New(render.Options{IndentJSON: true})

// parse the templates once and hold them in memory
var templates = template.Must(template.ParseGlob("templates/*"))

func main() {
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	r := mux.NewRouter().StrictSlash(false)

	// Home Page
	r.HandleFunc("/", viewHomePage).Methods("GET")

	// blog list view
	r.HandleFunc("/blog/", viewBlogListView).Methods("GET")
	r.HandleFunc("/blog/{page}", viewBlogListView).Methods("GET")

	// post detail view
	r.HandleFunc("/article/{id}", viewBlogPost).Methods("GET")

	// contact view
	r.HandleFunc("/contact/", viewContactPage).Methods("GET")

	// create a comment
	r.HandleFunc("/comment/", addCommentToPost).Methods("POST")

	http.Handle("/", r)

	http.ListenAndServe(":8000", nil)
}

// viewHomePage renders the index.html template
func viewHomePage(w http.ResponseWriter, req *http.Request) {
	templates.ExecuteTemplate(w, "index", nil)
}

// viewHomePage renders the list.html template
func viewBlogListView(w http.ResponseWriter, req *http.Request) {
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

// viewBlogPost renders a single blog post corresponding to the id in the URL
func viewBlogPost(w http.ResponseWriter, req *http.Request) {
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

// viewContactPage renders the contact template
func viewContactPage(w http.ResponseWriter, req *http.Request) {
	templates.ExecuteTemplate(w, "contact", nil)
}

// addCommentToPost takes comment data and adds it to the post
func addCommentToPost(w http.ResponseWriter, req *http.Request) {
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
