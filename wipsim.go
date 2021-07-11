// Package main implements a ticket servicing system simulation.
// The simulation shows the effect of limiting work in progress
// on the lead time of tickets.
// The simulation runs for 100 days.
// Tickets arrive with a gaussian distribution, with mean 1 Ticket per day
// and standard deviation of 1d.
// Tickets have an effort in hours, with a gaussian distribution, with
// mean 6h and standard deviation of 4h.
// Troughput is fixed to 8h per day
// Two scheduling strategies are compared:
// 1. Work on each ticket max 2h per day.
// 2. Work on the tickets in order of arrival
//
// Ralf Poeppel 2021
//
package main

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
)

// randomValueInt calculates a random int value from a
// gaussian distribution with mean and standard deviation
// not smaller as lowest
func randomValueInt(mean, stddev float64, lowest int) int {
	randomValue := rand.NormFloat64()*stddev + mean
	roundedValue := math.Round(randomValue)
	value := int(roundedValue)
	if value < lowest {
		value = lowest
	}
	return value
}

// ticket the state of a ticket
type ticket struct {
	startday int
	endday   int
	effort   int
	// burndown the remaining effort of a ticket at a day.
	// The day is the index in the array.
	burndown []int
}

// NewTicket create a new ticket
func NewTicket(startday, effort, totaldays int) *ticket {
	t := ticket{}
	t.startday = startday
	t.effort = effort
	t.burndown = make([]int, totaldays)
	t.burndown[startday] = effort
	return &t
}

// createTicketsForDay create count new tickets for a day with random effort
func createTicketsForDay(d, days, count int, meanEffortNew, stddevEffortNew float64,
	minEffort int) ([]*ticket, int) {

	tickets := make([]*ticket, count)
	sumEffort := 0
	for i := 0; i < count; i++ {
		effort := randomValueInt(meanEffortNew, stddevEffortNew,
			minEffort)
		sumEffort += effort
		ticket := NewTicket(d, effort, days)
		fmt.Println(d, count, effort, ticket)
		tickets[i] = ticket
	}
	if count == 0 {
		fmt.Println(d, count)
	}
	return tickets, sumEffort
}

// simulation the set of all tickets
type simulation []*ticket

// String create nice representation
func (sim simulation) String() string {
	var buf bytes.Buffer
	for i, t := range sim {
		buf.WriteString(fmt.Sprintln(i, *t))
	}
	return buf.String()
}

func main() {
	days := 20
	meanNewPerDay := 1.0
	stddevNewPerDay := 1.0
	sumCount := 0
	meanEffortNew := 6.0
	stddevEffortNew := 4.0
	minEffort := 1
	sumEffort := 0
	simulation := make(simulation, 0, days*3/2)
	for d := 0; d < days; d++ {
		count := randomValueInt(meanNewPerDay, stddevNewPerDay, 0)
		sumCount += count
		tickets, effort := createTicketsForDay(d, days, count,
			meanEffortNew, stddevEffortNew, minEffort)
		simulation = append(simulation, tickets...)
		sumEffort += effort

	}
	fmt.Println()
	fmt.Println(simulation)
	meanCount := float64(sumCount) / float64(days)
	fmt.Println("mean count:", meanCount)
	meanEffort := float64(sumEffort) / float64(days)
	fmt.Println("mean effort:", meanEffort)
}
