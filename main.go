package main

import (
	"fmt"
	"net"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

func clearConsole() {
	print("\033[H\033[2J")
}

func makeRange(min, max int) []int {
	a := make([]int, max-min+1)
	for i := range a {
		a[i] = min + i
	}
	return a
}

func scanPort(ip string, port int, timeout time.Duration, result chan<- int) {
	address := fmt.Sprintf("%s:%d", ip, port)
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err == nil {
		result <- port
		conn.Close()
	} else {
		result <- 0
	}
}

func resolveDomain(input string) (string, error) {
	if strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://") {
		u, err := url.Parse(input)
		if err != nil {
			return "", err
		}
		input = u.Hostname()
	}

	ips, err := net.LookupIP(input)
	if err != nil {
		return "", err
	}

	return ips[0].String(), nil
}

func main() {
	red := color.New(color.FgRed).PrintfFunc()
	green := color.New(color.FgGreen).PrintfFunc()
	yellow := color.New(color.FgYellow).PrintfFunc()

	clearConsole()
	fmt.Println("Enter the IP address, domain or URL to scan:")
	var input string
	fmt.Println("")
	fmt.Scanln(&input)

	input, err := resolveDomain(input)
	if err != nil {
		red("Error resolving domain: %v\n", err)
		return
	}

	clearConsole()
	fmt.Println("\nWhat group of ports do you want to scan?")
	fmt.Println("[1] Every port (SLOW)")
	fmt.Println("[2] Basic ports")
	fmt.Println("[3] Web ports")
	var choice int
	fmt.Println("")
	fmt.Scanln(&choice)

	var ports []int
	switch choice {
	case 1:
		ports = makeRange(1, 65535)
	case 2:
		ports = []int{21, 22, 25, 26, 2525, 587, 80, 443, 110, 995, 143, 993, 3306}
	case 3:
		ports = []int{21, 22, 23, 25, 26, 2525, 587, 43, 53, 67, 68, 69, 80, 443, 110, 995, 143, 993, 123, 137, 138, 139, 161, 162, 389, 636, 989, 990, 3306, 2082, 2083, 2086, 2087, 2095, 2096, 2077, 2078}
	default:
		red("\nInvalid option specified.")
		return
	}

	yellow("\nScanning ports on IP address %s...\n", input)

	start := time.Now()

	timeout := 2 * time.Second
	result := make(chan int)
	var wg sync.WaitGroup

	for _, port := range ports {
		wg.Add(1)
		go func(p int) {
			scanPort(input, p, timeout, result)
			wg.Done()
		}(port)
	}

	go func() {
		wg.Wait()
		close(result)
	}()

	for r := range result {
		if r != 0 {
			green("\nPort %d is open.\n", r)
		}
	}

	elapsed := time.Since(start)
	green("\nFinished scan | Took %.2f seconds\n", elapsed.Seconds())
}
