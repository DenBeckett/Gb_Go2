package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
)

func squareNum(wg *sync.WaitGroup, num int) int {
	result := num * num
	fmt.Println(num, "^2 = ", result)
	go doubleNum(wg, result)
	defer wg.Done()
	return result
}

func doubleNum(wg *sync.WaitGroup, num int) int {
	result := 2 * num
	fmt.Println("2x", num, " = ", result)
	defer wg.Done()
	return result
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	var wg sync.WaitGroup
	for {
		fmt.Print("Введите целое число или 'stop' для выхода: ")
		scanner.Scan()
		err := scanner.Err()
		if err != nil {
			log.Fatal(err)
		}
		num := scanner.Text()
		if num == "stop" {
			break
		}
		s, err := strconv.Atoi(num)
		if err != nil {
			fmt.Println("Ошибка. Введите число")
		}

		wg.Add(2)
		go squareNum(&wg, s)

		wg.Wait()
	}
}
