package main

import (
	"bufio"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type User struct {
	id   uuid.UUID
	name string
}

type Category struct {
	id        uuid.UUID
	name      string
	parent_id uuid.UUID
}

type Message struct {
	id          uuid.UUID
	text        string
	category_id uuid.UUID
	posted_at   time.Time
	author_id   uuid.UUID
}

//type Timestamp struct {
//	year, month, day, hour, minute, second int
//}

const (
	usersCount      = 500
	messagesCount   = 1000
	categoriesCount = 500
	goroutinesCount = 100
)

var (
	firstNames         []string
	lastNames          []string
	words              []string
	existingUsers      []uuid.UUID
	existingCategories []uuid.UUID
)

func main() {
	firstNames, _ = readLines("assets/first-names.txt")
	lastNames, _ = readLines("assets/last-names.txt")
	words, _ = readLines("assets/words.txt")

	mutex := &sync.Mutex{}
	wg := &sync.WaitGroup{}

	// Writing users
	func() {
		start := time.Now()
		fmt.Printf("%v: Started recording users...\n", time.Now().Format(time.UnixDate))

		wg.Add(usersCount)
		iterationsNum := usersCount / goroutinesCount
		for i := 0; i < goroutinesCount; i++ {
			go func() {
				for j := 0; j < iterationsNum; j++ {
					go writeUser(mutex, wg)
				}
			}()
		}
		wg.Wait()
		fmt.Printf("%v: Recording is successfully completed and took %v.\n\n", time.Now().Format(time.UnixDate), time.Since(start))
	}()

	// Writing categories
	func() {
		start := time.Now()
		fmt.Printf("%v: Started recording categories...\n", time.Now().Format(time.UnixDate))

		f, _ := os.OpenFile("tables/users.sql", os.O_RDWR|os.O_APPEND, 0660)
		func(file *os.File) {
			defer f.Close()

			id, _ := uuid.NewV4()
			_, err := f.WriteString("INSERT INTO users(id, name) VALUES ('" + id.String() + "', 'Forum'); \n")
			if err != nil {
				panic(err)
			}
		}(f)

		wg.Add(usersCount)
		iterationsNum := categoriesCount / goroutinesCount
		for i := 0; i < goroutinesCount; i++ {
			go func() {
				for j := 0; j < iterationsNum; j++ {
					go writeCategory(mutex, wg)
				}
			}()
		}
		wg.Wait()
		fmt.Printf("%v: Recording is successfully completed and took %v.\n\n", time.Now().Format(time.UnixDate), time.Since(start))
	}()

	func() {
		start := time.Now()
		fmt.Printf("%v: Started recording messages...\n", time.Now().Format(time.UnixDate))

		wg.Add(usersCount)
		iterationsNum := messagesCount / goroutinesCount
		for i := 0; i < goroutinesCount; i++ {
			go func() {
				for j := 0; j < iterationsNum; j++ {
					go writeMessage(mutex, wg)
				}
			}()
		}
		wg.Wait()
		fmt.Printf("%v: Recording is successfully completed and took %v.\n\n", time.Now().Format(time.UnixDate), time.Since(start))
	}()

}

// writeUser writes users as INSERT INTO queries in the db
func writeUser(m *sync.Mutex, w *sync.WaitGroup) {
	m.Lock()
	f, err := os.OpenFile("tables/users.sql", os.O_RDWR|os.O_APPEND, 0660)
	if err != nil {
		panic(err)
	}
	defer func() {
		f.Close()
		m.Unlock()
		w.Done()
		//runtime.Gosched()
	}()

	id, _ := uuid.NewV4()
	existingUsers = append(existingUsers, id)

	firstName := firstNames[rand.Intn(len(firstNames))]
	lastName := lastNames[rand.Intn(len(lastNames))]

	user := &User{
		id:   id,
		name: firstName + " " + lastName,
	}
	_, err = f.WriteString("INSERT INTO users(id, name) VALUES ('" + id.String() + "', '" + user.name + "'); \n")
	if err != nil {
		panic(err)
	}
}

// writeCategory writes categories as INSERT INTO queries in the db
func writeCategory(m *sync.Mutex, w *sync.WaitGroup) {
	m.Lock()
	f, err := os.OpenFile("tables/categories.sql", os.O_RDWR|os.O_APPEND, 0660)
	if err != nil {
		panic(err)
	}
	defer func() {
		f.Close()
		m.Unlock()
		w.Done()
		//runtime.Gosched()
	}()

	id, _ := uuid.NewV4()
	existingCategories = append(existingCategories, id)

	var name []string
	nameLength := rand.Intn(4-2) + 2

	for i := 0; i < nameLength; i++ {
		name = append(name, words[rand.Intn(len(words))])
	}

	category := &Category{
		id:   id,
		name: strings.Title(strings.ToLower(strings.Join(name, " "))),
	}

	var query string
	if hasParent := rand.Float32() < 0.5; hasParent {
		category.parent_id = existingCategories[rand.Intn(len(existingCategories))]
		query = "INSERT INTO categories(id, name, parent_id) VALUES ('" + id.String() + "', '" + category.name + "', '" + category.parent_id.String() + "'); \n"
	} else {
		query = "INSERT INTO categories(id, name) VALUES ('" + id.String() + "', '" + category.name + "'); \n"
	}

	_, err = f.WriteString(query)
	if err != nil {
		panic(err)
	}
}

// writeMessage writes messages as INSERT INTO queries in the db
func writeMessage(m *sync.Mutex, w *sync.WaitGroup) {
	m.Lock()
	f, err := os.OpenFile("tables/messages.sql", os.O_RDWR|os.O_APPEND, 0660)
	if err != nil {
		panic(err)
	}
	defer func() {
		f.Close()
		m.Unlock()
		w.Done()
		//runtime.Gosched()
	}()

	id, _ := uuid.NewV4()

	var text []string
	textLength := rand.Intn(20-1) + 1

	for i := 0; i < textLength; i++ {
		text = append(text, words[rand.Intn(len(words))])
	}

	message := &Message{
		id:          id,
		text:        strings.Title(strings.ToLower(strings.Join(text, " "))),
		category_id: existingCategories[rand.Intn(len(existingCategories))],
		posted_at:   getRandomTimestamp(),
		author_id:   existingUsers[rand.Intn(len(existingUsers))],
	}

	_, err = f.WriteString("INSERT INTO messages(id, text, category_id, posted_at, author_id) VALUES ('" + message.id.String() +
		"', '" + message.text + "', '" + message.category_id.String() + "', '" + message.posted_at.String() + "', '" +
		message.author_id.String() + "'); \n")
	if err != nil {
		panic(err)
	}

}

// readLines reads a whole file into memory
// and returns a slice of its lines.
func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func getRandomTimestamp() time.Time {
	min := time.Date(2015, 1, 0, 0, 0, 0, 0, time.UTC).Unix()
	max := time.Date(2020, 1, 0, 0, 0, 0, 0, time.UTC).Unix()
	delta := max - min

	sec := rand.Int63n(delta) + min
	return time.Unix(sec, 0)
}

func test(i, j int, m *sync.Mutex, w *sync.WaitGroup) {
	m.Lock()
	f, err := os.OpenFile("tables/users.txt", os.O_RDWR|os.O_APPEND, 0660)
	if err != nil {
		panic(err)
	}

	defer func() {
		f.Close()
		m.Unlock()
		w.Done()
	}()

	_, err = f.WriteString("#" + strconv.Itoa(i) + " " + "#" + strconv.Itoa(j) + " " + firstNames[rand.Intn(len(firstNames))] + "\n")
	if err != nil {
		panic(err)
	}
}
