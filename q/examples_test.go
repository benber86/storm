package q_test

import (
	"errors"
	"fmt"
	"log"

	"time"

	"os"

	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
)

func ExampleRe() {
	dir, db := prepareDB()
	defer os.RemoveAll(dir)
	defer db.Close()

	var users []User

	// Find all users with name that starts with the letter D.
	if err := db.Select(q.Re("Name", "^D")).Find(&users); err != nil {
		log.Println("error: Select failed:", err)
		return
	}

	// Donald and Dilbert
	fmt.Println("Found", len(users), "users.")

	// Output:
	// Found 2 users.
}

func ExampleCustom() {
	dir, db := prepareDB()
	defer os.RemoveAll(dir)
	defer db.Close()

	var users []User
	testTime := NullTime{time.Now().AddDate(0, -1*2, 0), true}
	if err := db.Select(q.Gte("UpdatedAt", testTime)).Find(&users); err != nil {
		fmt.Println("error: Select failed:", err)
		return
	}
	fmt.Println("Found", len(users), "users.")

	// Output:
	// Found 2 users.
}

type NullTime struct {
	Time  time.Time
	Valid bool // Valid is true if Time is not NULL
}

func (n NullTime) Compare(other interface{}) (int, error) {
	o, ok := other.(NullTime)
	if !ok {
		return 0, errors.New("Can only compare with variable of type NullTime")
	}
	if !n.Valid && !o.Valid {
		return 0, nil
	}
	if n.Valid && o.Valid {
		if n.Time.Equal(o.Time) {
			return 0, nil
		} else if n.Time.Before(o.Time) {
			return -1, nil
		}
		return 1, nil
	}
	return 0, errors.New("Can not compare Null Time and non Null Time")
}

type User struct {
	ID        int    `storm:"id,increment"`
	Group     string `storm:"index"`
	Email     string `storm:"unique"`
	Name      string
	Age       int       `storm:"index"`
	CreatedAt time.Time `storm:"index"`
	UpdatedAt NullTime  `storm:"index"`
}

func prepareDB() (string, *storm.DB) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	db, _ := storm.Open(filepath.Join(dir, "storm.db"))

	for i, name := range []string{"John", "Norm", "Donald", "Eric", "Dilbert"} {
		email := strings.ToLower(name + "@provider.com")
		user := User{
			Group:     "staff",
			Email:     email,
			Name:      name,
			Age:       21 + i,
			CreatedAt: time.Now(),
			UpdatedAt: NullTime{time.Now().AddDate(0, -1*i, 0), true},
		}
		err := db.Save(&user)

		if err != nil {
			log.Fatal(err)
		}
	}

	return dir, db
}
