// main
package main

import (
	"fmt"
)

type SalarySummator interface {
	Sum() float32
}

type Plumper struct {
	Age     int
	Name    string
	Salary  float32
	Advance float32
}

type Teacher struct {
	Age        int
	Salary     float32
	Advance    float32
	Speciality string
}

func (p Plumper) Sum() float32 {
	return p.Salary + p.Advance
}

func (t Teacher) Sum() float32 {
	return t.Salary + t.Advance
}

func getFullSalary(s SalarySummator) {
	fmt.Println(s.Sum())
}

func main() {
	plumper := Plumper{25, "Vasya", 20000, 10000}
	getFullSalary(plumper)

	teacher := Teacher{25, 25000, 15000, "History"}
	getFullSalary(teacher)

}
