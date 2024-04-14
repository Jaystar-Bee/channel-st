package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/fatih/color"
)

const NumberOfPizzas = 10

var pizzasMade, pizzasFailed, total int

type Producer struct {
	data chan PizzaOrder
	quit chan chan error
}

func (p *Producer) Close() error {
	ch := make(chan error)
	p.quit <- ch
	return <-ch
}

type PizzaOrder struct {
	PizzaNumber int
	Message     string
	IsSuccess   bool
}

func makePizza(pizzaNumber int) *PizzaOrder {
	pizzaNumber++
	if pizzaNumber <= NumberOfPizzas {
		delay := rand.Intn(5) + 1
		fmt.Printf("Received order #%d!\n", pizzaNumber)

		rnd := rand.Intn(12) + 1
		msg := ""
		success := false

		fmt.Printf("Making of pizza #%d will take %d seconds...\n", pizzaNumber, delay)

		if rnd < 5 {
			pizzasFailed++
		} else {
			pizzasMade++
		}
		total++
		// Delay for a while
		time.Sleep(time.Duration(delay) * time.Second)

		// Decisions per the given random number
		if rnd <= 2 {
			msg = fmt.Sprintf("*** We ran out of ingredients for pizza #%d!\n", pizzaNumber)
		} else if rnd <= 4 {
			msg = fmt.Sprintf("*** The cook quit while making pizza #%d!\n", pizzaNumber)
		} else {
			success = true
			msg = fmt.Sprintf("Pizza order #%d is ready!\n", pizzaNumber)
		}

		return &PizzaOrder{
			PizzaNumber: pizzaNumber,
			Message:     msg,
			IsSuccess:   success,
		}

	}
	return &PizzaOrder{
		PizzaNumber: pizzaNumber,
		Message:     "Done making pizzas",
		IsSuccess:   false,
	}
}

func pizzeria(pizzaMaker *Producer) {
	// Keep track of the pizza we are making
	var i = 0

	// run forever or until we receive a quit notification

	// try to make pizzas
	for {
		// try to make a pizza
		currentPizza := makePizza(i)

		if currentPizza != nil {
			i = currentPizza.PizzaNumber
			select {

			// we tried to make a pizza (we sent something to the data channel)
			case pizzaMaker.data <- *currentPizza:

			// we want to quit. so we send pizzaMaker.quit to the quit channel
			case quitChan := <-pizzaMaker.quit:

				// Close the channels
				close(pizzaMaker.data)
				close(quitChan)
				return
			}
		}
	}

}

func main() {
	// seed the random number generator
	// rand.Seed(time.Now().UnixNano())
	rand.New(rand.NewSource(time.Now().UnixNano()))

	// print out a message
	color.Cyan("The Pizzeria is open for business!")
	color.Cyan("----------------------------------")

	// create a producer
	pizzaJob := &Producer{
		data: make(chan PizzaOrder),
		quit: make(chan chan error),
	}

	// run the producer in the background
	go pizzeria(pizzaJob)

	// create and run consumer
	for data := range pizzaJob.data {

		if data.PizzaNumber <= NumberOfPizzas {

			if data.IsSuccess {
				color.Green(data.Message)
				color.Green("Order #%d is out for delivery!\n", data.PizzaNumber)
				color.Green("----------------------------------\n\n")
			} else {
				color.Red(data.Message)
				color.Red("Sorry, We can't deliver your order!\n")
				color.Red("----------------------------------\n\n")
			}

			stringData, _ := json.Marshal(data)
			os.WriteFile(fmt.Sprintf("%d-data.json", data.PizzaNumber), stringData, 0644)
		} else {
			pizzaJob.Close()
			color.Yellow("Done making pizzas")
		}
	}

	// print out the ending message

	color.Cyan("-----------------")
	color.Cyan("Done for the day.")
	color.Cyan("We made %d pizzas, but failed to make %d, with %d attempts in total.", pizzasMade, pizzasFailed, total)
	color.Cyan("-----------------")
}
