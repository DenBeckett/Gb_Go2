package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	file, err := os.OpenFile("strings.txt", os.O_RDONLY, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer file.Close()

	data := make([]byte, 64)

	for {
		n, err := file.Read(data)
		if err == io.EOF {
			break
		}
		fmt.Print(string(data[:n]))
	}
}
