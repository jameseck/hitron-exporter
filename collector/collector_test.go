package collector

import "fmt"

func Example_parseUptime() {
	fmt.Println(parseDuration("05 Days,21 Hours,33 Minutes,44 Seconds"))
	// Output: 509624
}

func Example_parsePkt() {
	fmt.Printf("%.2f", parsePkt("1.61G Bytes"))
	// Output: 1728724336.64
}

func Example_parsePkt_2() {
	fmt.Printf("%.2f", parsePkt("957.24M Bytes"))
	// Output: 1003738890.24
}
