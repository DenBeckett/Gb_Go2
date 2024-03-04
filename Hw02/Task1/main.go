package main

import (
	"bufio"
	"fmt"
	"os"
	"time"
)

func main() {
	var text string
	count := 1
	file, err := os.Create("strings.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	for {
		fmt.Println("Введите текст")
		fmt.Scan(&text)
		if text == "exit" {
			break
		} else {
			newStr(text, count)
			count++
		}
	}
	defer file.Close()
}

func newStr(text string, count int) {

	file, err := os.OpenFile("strings.txt", os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}
	writer := bufio.NewWriter(file)
	writer.WriteString(fmt.Sprint(count, " ", time.Now().Format(time.DateTime), " ", text))
	writer.WriteString(fmt.Sprintf("\n"))
	writer.Flush()

	defer file.Close()
}
