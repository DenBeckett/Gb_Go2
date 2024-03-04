package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/fs"
	"log"
	"os"
	"time"
)

func main() {
	var count int
	var buf bytes.Buffer
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("Введите текст: ")
		scanner.Scan()
		err := scanner.Err()
		if err != nil {
			log.Fatal(err)
		}

		text := scanner.Text()
		if text == "exit" {
			break
		}
		count++

		buf.WriteString(fmt.Sprint(count, " ", time.Now().Format(time.DateTime), " ", text))
		buf.WriteString(fmt.Sprintf("\n"))
	}

	err := os.WriteFile("strings.txt", buf.Bytes(), fs.ModePerm)
	if err != nil {
		log.Fatalf("Ошибка записи в файл %s: %s\n", "strings.txt", err)
	}

	text, err := os.ReadFile("strings.txt")
	if err != nil {
		log.Fatalf("Ошибка чтения файла %s: %s\n", "strings.txt", err)
	}
	fmt.Println(string(text))
}
