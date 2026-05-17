package main

type TimerService struct{}

func (t *TimerService) Greet(name string) string {
	return "Hello " + name + "!"
}
