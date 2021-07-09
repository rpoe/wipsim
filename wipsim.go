// Package main implements a ticket servicing system simulation.
// The simulation shows the effect of limiting work in progress
// on the lead time of tickets.
// The simulation runs for 100 days.
// Tickets arrive with a gaussian distribution, with mean 1 Ticket per day
// and standard deviation of 1d.
// Tickets have an effort in hours, with a gaussian distribution, with
// mean 7h and standard deviation of 4h.
// Troughput is fixed to 8h per day
// Two scheduling strategies are compared:
// 1. Work on each ticket max 2h per day.
// 2. Work on the tickets in order of arrival
//
// Ralf Poeppel 2021
//
package main

import (
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

func main() {
	days := 200
	meanNewPerDay := 1.0
	stddevNewPerDay := 1.0
	sumCount := 0
	meanEffortNew := 7.0
	stddevEffortNew := 4.0
	minEffort := 1
	sumEffort := 0
	incident := 0
	for d := 0; d < days; d++ {
		count := randomValueInt(meanNewPerDay, stddevNewPerDay, 0)
		sumCount += count
		for i := 0; i < count; i++ {
			incident++
			effort := randomValueInt(meanEffortNew, stddevEffortNew,
				minEffort)
			sumEffort += effort
			fmt.Println(d, count, incident, effort)
		}
		if count == 0 {
			fmt.Println(d, count, incident)
		}

	}
	meanCount := float64(sumCount) / float64(days)
	fmt.Println("mean count:", meanCount)
	meanEffort := float64(sumEffort) / float64(days)
	fmt.Println("mean effort:", meanEffort)
}
