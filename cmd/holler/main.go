package main

import "github.com/malnick/holler"

func main() {
	myHoller, err := holler.New()
	if err != nil {
		panic(err)
	}

	myHoller.Start()
}
