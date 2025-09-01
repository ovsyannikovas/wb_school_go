package l11

import "fmt"

type Human struct{}

func (h Human) SayHello() { fmt.Println("Hello") }

type Action struct {
	Human
}

func main() {
	action := Action{}
	action.SayHello()
}
