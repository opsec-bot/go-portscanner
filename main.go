package main

import (
	"bufio"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

const input = "input.txt"

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

func resolveDomain(target string) (string, error) {
	if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
		u, err := url.Parse(target)
		if err != nil {
			return "", err
		}
		target = u.Hostname()
	}

	ips, err := net.LookupIP(target)
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
	fmt.Printf("Do you want to scan a single target or a file with multiple targets?\n\n")
	fmt.Println("[1] Single target")
	fmt.Println("[2] Multiple targets from a file")
	fmt.Println("")
	var inputChoice int
	fmt.Scanln(&inputChoice)

	var targets []string
	if inputChoice == 2 {
		file, err := os.Open(input)
		if err != nil {
			red("Error opening file: %v\n", err)
			return
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			targets = append(targets, scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			red("Error reading file: %v\n", err)
			return
		}
	} else {
		clearConsole()
		fmt.Println("Enter the IP address, domain or URL to scan:")
		fmt.Println("")
		var target string
		fmt.Scanln(&target)
		targets = append(targets, target)
	}

	logFile, err := os.Create("log.txt")
	if err != nil {
		red("Error creating log file: %v\n", err)
		return
	}
	defer logFile.Close()

	for _, target := range targets {
		target, err := resolveDomain(target)
		if err != nil {
			red("Error resolving domain: %v\n", err)
			return
		}

		clearConsole()
		fmt.Printf("\nWhat group of ports do you want to scan for %v?\n\n", target)
		fmt.Println("[1] Every port (SLOW)")
		fmt.Println("[2] Basic ports")
		fmt.Println("[3] Web ports")
		fmt.Println("")
		var choice int
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

		yellow("\nScanning ports on IP address %s...\n", target)

		start := time.Now()

		timeout := 2 * time.Second
		result := make(chan int)
		var wg sync.WaitGroup

		for _, port := range ports {
			wg.Add(1)
			go func(p int) {
				scanPort(target, p, timeout, result)
				wg.Done()
			}(port)
		}

		go func() {
			wg.Wait()
			close(result)
		}()

		openPorts := []int{}
		for r := range result {
			if r != 0 {
				green("\nPort %d is open.\n", r)
				openPorts = append(openPorts, r)
			}
		}
		openPortsStr := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(openPorts)), ", "), "[]")
		logFile.WriteString(fmt.Sprintf("%s has these ports open: %s\n\n", target, openPortsStr))

		elapsed := time.Since(start)
		green("\nFinished scan | Took %.2f seconds\n", elapsed.Seconds())
	}
	fmt.Scanln()
}
