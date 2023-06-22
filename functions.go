package main

import (
	"fmt"
)

// define a struct for a person
type Person struct {
	Name string
	Age  int
}

// define a function to create a new person
func NewPerson(name string, age int) *Person {
	return &Person{
		Name: name,
		Age:  age,
	}
}

// define a method to greet a person
func (p *Person) Greet() {
	fmt.Printf("Hello, my name is %s and I am %d years old.\n", p.Name, p.Age)
}