package main

import (
	"errors"
	"fmt"
	"time"
)

type Message struct { // данные сообщения
	Data      []byte
	MimeType  string
	Timestamp time.Time
}

type TwitterSource struct { // источник сообщений Twitter
	Username string
}

type SkypeSource struct { // источник сообщений Skype
	Login         string
	MSBackdoorKey string
}

// Finder реализует интерфейс для поиска слов в любом из источников выше
type Finder interface {
	Find(word string) ([]Message, error)
}

// создаем реализацию функции Find интерфейса Finder для источника TwitterSource (реализуется отдельно от интерфейса Finder)
func (s TwitterSource) Find(word string) ([]Message, error) {
	return s.searchAPICall(s.Username, word) // внутри реализации выполняем отдельный для твиттера код, в частности, searchAPICall-функцию
}

// создаем реализацию функции Find интерфейса Finder для источника SkypeSource (реализуется отдельно от интерфейса Finder)
func (s SkypeSource) Find(word string) ([]Message, error) {
	return s.searchSkypeServers(s.MSBackdoorKey, s.Login, word) // внутри реализации выполняем отдельный для скайпа код, в частности, searchSkypeServers-функцию
}

type Sources []Finder // добавили новый тип Sources — массив из Finder-ов — всего, что может давать нам
// возможность поиска (чтобы искать ВЕЗДЕ СРАЗУ, а не в каждом источнике отдельно).

func (s Sources) SearchWords(word string) []Message { // Метод SearchWords() для этого типа (Sources) возвращает массив с сообщениями.
	var messages []Message
	for _, source := range s {
		msgs, err := source.Find(word) // ищем в любом источнике, удовлетворяющем интерфейсу Finder (делаем это в цикле для каждого разного источника по одному стандарту)
		if err != nil {
			fmt.Println("WARNING:", err)
			continue
		}
		messages = append(messages, msgs...)
	}

	return messages
}

type Person struct { // создаем Person как абстрактный объект, в который всттроим и будем использовать интерфейс Finder (через Sources, созданный выше)
	FullName string
	Sources
}

var (
	sources = Sources{
		TwitterSource{
			Username: "@rickhickey",
		},
		SkypeSource{
			Login:         "rich.hickey",
			MSBackdoorKey: "12345",
		},
	}

	person = Person{
		FullName: "Rick Hickey",
		Sources:  sources, // Добавив безымянное поле типа Sources мы используем embedding — «встраивание»,
		//которое без лишних манипуляций позволяет объекту типа Person напрямую использовать функции Sources:
	}
)

// Реализация поиска конкретно для твиттера
func (s TwitterSource) searchAPICall(username, word string) ([]Message, error) {
	return []Message{
		Message{
			Data:      ([]byte)("Remember, remember, the fifth of November, если бы бабушка..."),
			MimeType:  "text/plain",
			Timestamp: time.Now(),
		},
	}, nil
}

// Реализация поиска конкретно для сккайпа
func (s SkypeSource) searchSkypeServers(key, login, word string) ([]Message, error) {
	return []Message{}, errors.New("NSA can't read your skype messages ;)")
}

func (m Message) String() string {
	return string(m.Data) + " @ " + m.Timestamp.Format(time.RFC822)
}

func main() {
	msgs := person.SearchWords("если бы бабушка") // ищем у person во всех источниках сразу
	fmt.Println(msgs)
}
