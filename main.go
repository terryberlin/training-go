package main

import (
	"fmt"
	"net/http"
	//"html/template"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/jmoiron/sqlx"
	//run go get github.com/denisenkom/go-mssqldb from apiservice root folder

	"github.com/yosssi/ace"

	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"net/url"
)

//Book : Book is structure for holding a record of the book.
type Book struct {
	PK             int
	Title          string
	Author         string
	Classification string
}

//Page : Page is a slice of Book
type Page struct {
	Books []Book
}

//SearchResult : SearchResult is a structure for holding xml data.
type SearchResult struct {
	Title  string `xml:"title,attr"`
	Author string `xml:"author,attr"`
	Year   string `xml:"hyr,attr"`
	ID     string `xml:"owi,attr"`
}

// var db *sql.DB

// func verifyDatabase(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
//   if err := db.Ping(); err != nil {
//     http.Error(w, err.Error(), http.StatusInternalServerError)
//     return
//   }
//   next(w, r)
// }

func main() {
	//templates := template.Must(template.ParseFiles("templates/index.html"))

	// template, err := ace.Load("templates/index", "", nil)
	// if err != nil {http.Error(w, err.Error(), http.StatusInternalServerError)}
	// if err = template.Execute(w, p); err != nil {
	//     http.Error(w, err.Error(), http.StatusInternalServerError)
	//     }

	db, _ := sqlx.Open("mssql", "server=192.168.1.34;user id=sa;password=123456789;database=quikserve;log=64;encrypt=disable")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		template, err := ace.Load("templates/index", "", nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		p := Page{Books: []Book{}}
		rows, _ := db.Query("select pk, title, author, classification from books")
		for rows.Next() {
			var b Book
			rows.Scan(&b.PK, &b.Title, &b.Author, &b.Classification)
			p.Books = append(p.Books, b)
		}

		//if err := templates.ExecuteTemplate(w, "index.html", p); err != nil {
		if err = template.Execute(w, p); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		var results []SearchResult
		var err error

		if results, err = search(r.FormValue("search")); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		encoder := json.NewEncoder(w)
		if err := encoder.Encode(results); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	http.HandleFunc("/books/add", func(w http.ResponseWriter, r *http.Request) {
		var book ClassifyBookResponse
		var err error

		if book, err = find(r.FormValue("id")); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		result, err := db.Exec("insert into books (title, author, id, classification) values (?, ?, ?, ?)",
			book.BookData.Title, book.BookData.Author, book.BookData.ID, book.Classification.MostPopular)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		pk, _ := result.LastInsertId()
		b := Book{
			PK:             int(pk),
			Title:          book.BookData.Title,
			Author:         book.BookData.Author,
			Classification: book.Classification.MostPopular,
		}
		if err := json.NewEncoder(w).Encode(b); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	http.HandleFunc("/books/delete", func(w http.ResponseWriter, r *http.Request) {
		if _, err := db.Exec("delete from books where pk = ?", r.FormValue("pk")); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	fmt.Println(http.ListenAndServe(":8080", nil))
}

//ClassifySearchResponse : ClassifySearchResponse is a slice of the search results.
type ClassifySearchResponse struct {
	Results []SearchResult `xml:"works>work"`
}

//ClassifyBookResponse : ClassifyBookResponse is a structure for holding the xml of BookData.
type ClassifyBookResponse struct {
	BookData struct {
		Title  string `xml:"title,attr"`
		Author string `xml:"author,attr"`
		ID     string `xml:"owi,attr"`
	} `xml:"work"`
	Classification struct {
		MostPopular string `xml:"sfa,attr"`
	} `xml:"recommendations>ddc>mostPopular"`
}

func find(id string) (ClassifyBookResponse, error) {
	var c ClassifyBookResponse
	body, err := classifyAPI("http://classify.oclc.org/classify2/Classify?summary=true&owi=" + url.QueryEscape(id))

	if err != nil {
		return ClassifyBookResponse{}, err
	}

	err = xml.Unmarshal(body, &c)
	return c, err
}

func search(query string) ([]SearchResult, error) {
	var c ClassifySearchResponse
	body, err := classifyAPI("http://classify.oclc.org/classify2/Classify?summary=true&title=" + url.QueryEscape(query))
	//body, err := classifyAPI("http://classify.oclc.org/classify2/ClassifyDemo?search-title-txt=&title=" + url.QueryEscape(query))

	if err != nil {
		return []SearchResult{}, err
	}

	err = xml.Unmarshal(body, &c)
	return c.Results, err
}

func classifyAPI(url string) ([]byte, error) {
	var resp *http.Response
	var err error

	if resp, err = http.Get(url); err != nil {
		return []byte{}, err
	}

	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

//"log"

// func DB() *sqlx.DB {
//         db, err := sqlx.Connect("mssql", "server=192.168.1.34;userid=sa;password=123456789;database=quikserve;log64;encrypt=disable")
//         if err != nil {
//             log.Println(err)
//         }
//         return db
// }

// SearchResult{"Moby-Dick", "Herman Melville", "1851", "222222"},
// SearchResult{"The Adventures of Huckleberry Finn", "Mark Twain", "1854", "444444"},
// SearchResult{"The Catcher In The Rye", "JD Salinger", "1951", "333333"},
