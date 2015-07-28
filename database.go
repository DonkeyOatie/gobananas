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

type post struct {
	ID            int       `db:"id"`
	Title         string    `db:"title"`
	Body          []byte    `db:"body"`
	DatePublished time.Time `db:"date_published"`
}

type comment struct {
	PostID        int       `db:"post_id"`
	Author        string    `db:"author"`
	Body          string    `db:"body"`
	DateCommented time.Time `db:"date_commented"`
}

// HTMLBody returns the html formatted body for the post
func (p post) HTMLBody() template.HTML {
	return template.HTML(p.Body)
}

// HTMLBodySample returns the first 150 characters of a post to display as
// summary in the list view
func (p post) HTMLBodySample() template.HTML {
	if len(p.Body) > 150 {
		sampleBody := string(p.Body[:150])
		return template.HTML(sampleBody + "<em> ... </em>")
	}
	return template.HTML(p.Body)
}

var db = initDb()

// initDb uses environment variables to connect to our database and return a
// connection handler
func initDb() *sqlx.DB {
	dbHost := "localhost"
	dbUser := os.Getenv("DBUSER")
	dbName := os.Getenv("DBNAME")
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

func checkNonFatalErr(err error, msg string) {
	if err != nil {
		log.Println(msg, err)
	}
}

func checkQueryErr(err error, query string) {
	msg := "Error running the following query: " + query
	checkNonFatalErr(err, msg)
}

// createBlogPost inserts the new blog post into the database
func createBlogPost(p post) error {
	const query = `INSERT INTO post (title, body, date_published)
	               VALUES ($1, $2, now())`

	_, err := db.Exec(query, p.Title, p.Body)
	checkQueryErr(err, query)
	return err
}

// updateBlogPost updates the post with the specified id
func updateBlogPost(id int, p post) error {
	const query = `UPDATE post
	               SET body=$2
		       WHERE id=$1`

	_, err := db.Exec(query, id, p.Body)
	checkQueryErr(err, query)
	return err
}

// getBlogPosts returns all posts from the database
func getBlogPosts(offset int) ([]post, error) {
	const query = `SELECT id, title, body, date_published
	               FROM post
		       ORDER BY date_published DESC
		       LIMIT $1
		       OFFSET $2`
	var posts []post

	err := db.Select(&posts, query, numberPostsPerPage, offset)
	checkQueryErr(err, query)

	return posts, err
}

// getBlogPost returns the blog post with with give id
func getBlogPost(id int) (post, error) {
	const query = `SELECT id, title, body, date_published
	               FROM post
		       WHERE id=$1`
	var p post

	err := db.Get(&p, query, id)
	checkQueryErr(err, query)
	return p, err
}

// returns the number of posts in the database
func getNumberOfPosts() (int, error) {
	var nPosts int
	const query = `SELECT COUNT(1)
		       FROM post`

	err := db.Get(&nPosts, query)
	checkQueryErr(err, query)
	return nPosts, err
}

// createsComment creates a comment for a post
func createComment(c comment) error {
	const query = `INSERT INTO comments (author, body, post_id)
	               VALUES ($1, $2, $3)`
	_, err := db.Exec(query, c.Author, c.Body, c.PostID)
	checkQueryErr(err, query)
	return err
}

// getCommentsForPost returns all of the comments for a particular post
func getCommentsForPost(pID int) ([]comment, error) {
	const query = `SELECT post_id, author, body, date_commented
	               FROM comments
		       WHERE post_id=$1 AND approved='t'
		       ORDER BY date_commented DESC`
	var comments []comment
	err := db.Select(&comments, query, pID)
	checkQueryErr(err, query)
	return comments, err
}
