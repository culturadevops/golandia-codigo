package main

import "fmt"

func main() {
	stock := map[string]int{"café": 10, "arepa": 4}
	stock["té"] = 7
	stock["café"] += 2
	fmt.Println(stock, "len", len(stock))

	fmt.Println("sushi:", stock["sushi"])
	if n, ok := stock["sushi"]; !ok {
		fmt.Println("sushi no está (n =", n, ")")
	}
	stock["agua"] = 0
	n, ok := stock["agua"]
	fmt.Println("agua:", n, "existe:", ok)

	delete(stock, "té")
	fmt.Println("tras delete:", stock)
}
