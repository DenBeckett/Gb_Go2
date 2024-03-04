package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	file, err := os.Create("abc.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	os.Chmod("abc.txt", 0444)
	file.Close()

	f, err := os.Open("abc.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	writer := bufio.NewWriter(f)
	writer.WriteString("abc")
	writer.Flush()

	defer file.Close()
}
