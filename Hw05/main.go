package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

type User struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type server struct {
	db *sql.DB
}

func main() {
	db, err := sql.Open("sqlite3", "data.db")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer db.Close()

	s := server{db: db}

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		http.ListenAndServe("localhost:8080", http.HandlerFunc(s.handler))
	}()

	go func() {
		defer wg.Done()
		http.ListenAndServe("localhost:8081", http.HandlerFunc(s.handler))
	}()

	go func() {
		defer wg.Done()
		http.ListenAndServe("localhost:8082", http.HandlerFunc(proxyHandler))
	}()

	wg.Wait()
}

func (s *server) handler(w http.ResponseWriter, r *http.Request) {
	mux := http.NewServeMux()
	var id string

	mux.HandleFunc("/create", s.Create)
	mux.HandleFunc("/make_friends", s.MakeFriends)
	mux.HandleFunc("/user", s.Delete)
	mux.HandleFunc("/friends/"+id, s.Friends)
	mux.HandleFunc("/"+id, s.AgeUpd)

	mux.ServeHTTP(w, r)
}

func parseURL(target string) *url.URL {
	url, err := url.Parse(target)
	if err != nil {
		panic(err)
	}
	return url
}

func proxyHandler(w http.ResponseWriter, r *http.Request) {
	targetURL := "http://localhost:8082"
	proxy := httputil.NewSingleHostReverseProxy(parseURL(targetURL))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	})
}

func (u *User) toString() string {
	return fmt.Sprintf("%s %d", u.Name, u.Age)
}

func (s *server) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		data, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		defer r.Body.Close()

		var u User
		if err := json.Unmarshal(data, &u); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		add, err := s.db.Exec(`
				BEGIN;
				CREATE TABLE IF NOT EXISTS Users(
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					Name TEXT,
					Age INTEGER
				);
				INSERT INTO Users(Name, Age) 
					VALUES ($1, $2);
				COMMIT;
				`, u.Name, u.Age)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		userID, _ := add.LastInsertId()

		friendsTable := "friends_" + strconv.Itoa(int(userID))
		script := fmt.Sprintf(`
				BEGIN;
				CREATE TABLE IF NOT EXISTS %s(
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					Friend INTEGER UNIQUE NOT NULL
				);
				COMMIT;
				`, friendsTable)
		_, err = s.db.Exec(script)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(fmt.Sprintf("User %s has just been added with ID=%d\n", u.Name, userID)))
		return
	}
	w.WriteHeader(http.StatusBadRequest)
}

func (s *server) MakeFriends(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		data, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		defer r.Body.Close()

		var acceptance map[string]int
		if err := json.Unmarshal(data, &acceptance); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		sourceID, ok := acceptance["source_id"]
		if !ok {
			w.Write([]byte("Incorrect source_id.\n"))
			return
		}
		targetID, ok := acceptance["target_id"]
		if !ok {
			w.Write([]byte("Incorrect target_id.\n"))
			return
		}

		tabSource := "friends_" + strconv.Itoa(sourceID)
		tabTarget := "friends_" + strconv.Itoa(targetID)
		script := fmt.Sprintf(`
				BEGIN;
				INSERT INTO %s (Friend) 
					VALUES(%d);
				INSERT INTO %s(Friend) 
					VALUES(%d);
				COMMIT;
				`, tabSource, targetID, tabTarget, sourceID)

		_, err = s.db.Exec(script)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("id=%d and id=%d have just become friends.\n", sourceID, targetID)))
		return
	}
	w.WriteHeader(http.StatusBadRequest)
}

func (s *server) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method == "DELETE" {
		data, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		defer r.Body.Close()

		var targetID map[string]int
		if err := json.Unmarshal(data, &targetID); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		userID, ok := targetID["target_id"]
		if !ok {
			w.Write([]byte("Incorrect source_id.\n"))
			return
		}

		script := fmt.Sprintf(`
					SELECT Name
					FROM Users
					WHERE id = %d		
					`, userID)
		getName, err := s.db.Query(script)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		defer getName.Close()

		var name string
		for getName.Next() {
			err = getName.Scan(&name)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
		}

		script = fmt.Sprintf(`
					SELECT Friend
					FROM %s		
					`, "friends_"+strconv.Itoa(userID))
		getFriends, err := s.db.Query(script)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		defer getFriends.Close()

		var friends []string
		for getFriends.Next() {
			var f string
			err := getFriends.Scan(&f)
			if err != nil {
				fmt.Println(err)
				continue
			}
			friends = append(friends, f)
		}

		for _, f := range friends {
			table := "friends_" + f
			script = fmt.Sprintf(`
						DELETE FROM %s
						WHERE Friend = %d			
						`, table, userID)
			_, err = s.db.Exec(script)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err.Error()))
				return
			}
		}

		userFriends := "friends_" + strconv.Itoa(userID)
		script = fmt.Sprintf(`
					BEGIN;
					DELETE FROM Users
						WHERE id = %d;
					DROP TABLE IF EXISTS %s;
					COMMIT;
					`, userID, userFriends)
		_, err = s.db.Exec(script)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		fmt.Printf("User id=%d(%s) and table %s have been deleted.\n", userID, name, userFriends)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("%s has just deleted his/her account.\n", name)))
		return
	}
	w.WriteHeader(http.StatusBadRequest)
}

func (s *server) Friends(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) < 3 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Incorrect request."))
			return
		}
		userID, err := strconv.Atoi(parts[2])
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		fTab := "friends_" + strconv.Itoa(userID)
		script := fmt.Sprintf(`
					SELECT Name, Age FROM Users 
					JOIN %s ON Users.id =  friend`, fTab)
		friends, err := s.db.Query(script)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		defer friends.Close()

		var friendsList string
		users := []User{}

		for friends.Next() {
			f := User{}
			err := friends.Scan(&f.Name, &f.Age)
			if err != nil {
				fmt.Println(err)
				continue
			}
			users = append(users, f)
		}
		for _, f := range users {
			friendsList += f.toString() + "\n"
		}

		w.Header().Set("Connection", "close")
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(friendsList))
		return
	}
	w.WriteHeader(http.StatusBadRequest)
}

func (s *server) AgeUpd(w http.ResponseWriter, r *http.Request) {
	if r.Method == "PUT" {
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) < 2 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Incorrect request."))
			return
		}
		userID, err := strconv.Atoi(parts[1])
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		data, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		defer r.Body.Close()

		var newAge map[string]int
		if err := json.Unmarshal(data, &newAge); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		age, ok := newAge["new age"]
		if !ok {
			w.Write([]byte("Incorrect source_id.\n"))
			return
		}

		script := fmt.Sprintf(`
					SELECT Name
					FROM Users
					WHERE id = %d		
					`, userID)
		getName, err := s.db.Query(script)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		defer getName.Close()

		var name string
		for getName.Next() {
			err = getName.Scan(&name)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
		}

		if name == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("Couldn't update age: User id=%d doesn't exist\n", userID)))
			return
		}
		script = fmt.Sprintf(`
					UPDATE Users
					SET Age = %d
					WHERE id = %d			
					`, age, userID)
		_, err = s.db.Exec(script)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("%s is now %d.\n", name, age)))
		return
	}
	w.WriteHeader(http.StatusBadRequest)
}
