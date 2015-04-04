package main

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Post struct {
	Id             int
	Title          string
	Body           []byte
	Date_Published time.Time
}

type Comment struct {
	Post_Id        int
	Author         string
	Body           string
	Date_Commented time.Time
}

func (p Post) HtmlBody() template.HTML {
	return template.HTML(p.Body)
}

func (p Post) HtmlBodySample() template.HTML {
	if len(p.Body) > 150 {
		sample_body := string(p.Body[:150])
		return template.HTML(sample_body + "<em> ... </em>")
	} else {
		return template.HTML(p.Body)
	}
}

var db = initDb()

// initDb uses environment variables to connect to our database and return a
// connection handler
func initDb() *sqlx.DB {
	dbHost := "localhost"
	dbUser := "blogadmin"
	dbName := "blogdb"
	sslMode := "disable"
	dbPass := os.Getenv("DBPASS")

	// We have empirically shown that MAXIDLE=85 and MAXOPEN=90 are good,
	// conservative values. The goal is to make sure we do not run out of
	// database connections in case of high number of requests per second
	maxIdle := 85
	maxOpen := 90

	connstring := fmt.Sprintf("host=%s user=%s dbname=%s password=%s sslmode=%s",
		dbHost, dbUser, dbName, dbPass, sslMode)

	db := sqlx.MustConnect("postgres", connstring)
	db.SetMaxIdleConns(maxIdle)
	db.SetMaxOpenConns(maxOpen)

	fmt.Println(connstring, "MAXIDLE =", maxIdle, "MAXOPEN =", maxOpen)

	return db
}

func CheckNonFatalErr(err error, msg string) {
	if err != nil {
		log.Println(msg, err)
	}
}

func CheckQueryErr(err error, query string) {
	msg := "Error running the following query: " + query
	CheckNonFatalErr(err, msg)
}

// createBlogPost inserts the new blog post into the database
func createBlogPost(p Post) error {
	const query = `INSERT INTO post (title, body, date_published)
	               VALUES ($1, $2, now())`

	_, err := db.Exec(query, p.Title, p.Body)
	CheckQueryErr(err, query)
	return err
}

// updateBlogPost updates the post with the specified id
func updateBlogPost(id int, p Post) error {
	const query = `UPDATE post
	               SET body=$2
		       WHERE id=$1`

	_, err := db.Exec(query, id, p.Body)
	CheckQueryErr(err, query)
	return err
}

// getBlogPosts returns all posts from the database
func getBlogPosts(offset int) ([]Post, error) {
	const query = `SELECT id, title, body, date_published
	               FROM post
		       ORDER BY date_published DESC
		       LIMIT $1
		       OFFSET $2`
	var posts []Post

	err := db.Select(&posts, query, NUMBER_POSTS_PER_PAGE, offset)
	CheckQueryErr(err, query)

	return posts, err
}

// getBlogPost returns the blog post with with give id
func getBlogPost(id int) (Post, error) {
	const query = `SELECT id, title, body, date_published
	               FROM post
		       WHERE id=$1`
	var post Post

	err := db.Get(&post, query, id)
	CheckQueryErr(err, query)
	return post, err
}

// returns the number of posts in the database
func getNumberOfPosts() (int, error) {
	var nPosts int
	const query = `SELECT COUNT(1)
		       FROM post`

	err := db.Get(&nPosts, query)
	CheckQueryErr(err, query)
	return nPosts, err
}

// createsComment creates a comment for a post
func createComment(c Comment) error {
	const query = `INSERT INTO comments (author, body, post_id)
	               VALUES ($1, $2, $3)`
	_, err := db.Exec(query, c.Author, c.Body, c.Post_Id)
	CheckQueryErr(err, query)
	return err
}

// getCommentsForPost returns all of the comments for a particular post
func getCommentsForPost(pId int) ([]Comment, error) {
	const query = `SELECT post_id, author, body, date_commented
	               FROM comments
		       WHERE post_id=$1 AND approved='t'
		       ORDER BY date_commented DESC`
	var comments []Comment
	err := db.Select(&comments, query, pId)
	CheckQueryErr(err, query)
	return comments, err
}
