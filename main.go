package main

import (
	"fmt"
	"easyconfig/core"
)

func main() {
	setter, err := core.GetConfigController()
	if err != nil {
		fmt.Println(err)
		return
	}
	setter.Set("lebron.james.number", "23")
	value, err := setter.Get("lebron.james.number")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(value)

	getter, err := core.GetConfigGetter()
	if err != nil {
		fmt.Println(err)
		return
	}
	value, err = getter.Get("lebron.james.number")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(value)
}
