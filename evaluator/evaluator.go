package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	fmt.Println("MODULE_READY")

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input := scanner.Text()

		fmt.Println(input)

	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}

	fmt.Fprintln(os.Stderr, "terminating error")
}
