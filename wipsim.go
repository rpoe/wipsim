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
	"sort"
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
	leadtime int
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

// Clone create a deep copy of a ticket
func (t *ticket) Clone() *ticket {
	cp := ticket{}
	cp.startday = t.startday
	cp.effort = t.effort
	cp.burndown = make([]int, 0, len(t.burndown))
	cp.burndown = append(cp.burndown, t.burndown...)
	return &cp
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

// burndown burn down a ticket, max for the given hours and return updated hoursleft
func (t *ticket) burndownhours(day, hoursleft, hours int) int {
	d1 := day + 1
	workremain := t.burndown[day]
	if workremain > 0 {
		// calculate possible burndown
		if hoursleft > 0 {
			if workremain < hours {
				hours = workremain
			}
			if hoursleft < hours {
				hours = hoursleft
			}
			workremain -= hours
			hoursleft -= hours
		}
		// update ticket stats for actual day for ticket in work
		t.endday = day
		t.leadtime = d1 - t.startday
	}
	t.burndown[d1] = workremain
	return hoursleft
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

// addTickets add a copy of the given tickets to the simulation
func (sim simulation) addTickets(ts []*ticket) simulation {
	s := sim
	for _, t := range ts {
		tcp := t.Clone()
		s = append(s, tcp)
	}
	return s
}

// statsLeadTime return average and standard deviation
// and sum of mean and stdev of tickets leadtime
func (sim simulation) statsLeadTime() (float64, float64, float64) {
	var sum float64 = 0.0
	var sumSq float64 = 0.0
	for _, t := range sim {
		l := float64(t.leadtime)
		sum += l
		sumSq += l * l
	}
	// calculate the mean/std.dev
	l := float64(len(sim))
	meanSq := sumSq / l
	mean := sum / l
	stdev := math.Sqrt(meanSq - mean*mean)
	return mean, stdev, mean + stdev
}

// workhoursday working hours per day
const workhoursday = 8

// burndownMaxWip burn down maximum number of tickets, each 2h for a day
func (sim *simulation) burndownMaxWip(day int) {
	d1 := day + 1
	// no burndown without ticket or on last day
	if len(*sim) <= 0 || d1 >= len((*(*sim)[0]).burndown) {
		return
	}
	hourswork := 2
	hoursleft := workhoursday
	for _, t := range *sim {
		hoursleft = t.burndownhours(day, hoursleft, hourswork)
	}
	if hoursleft > 0 {
		// burn hours left
		for _, t := range *sim {
			hoursleft = t.burndownhours(day, hoursleft, hoursleft)
		}
	}
}

// burndownMinWip burn down the oldest tickets first
func (sim *simulation) burndownMinWip(day int) {
	d1 := day + 1
	// no burndown without ticket or on last day
	if len(*sim) <= 0 || d1 >= len((*(*sim)[0]).burndown) {
		return
	}
	hoursleft := workhoursday
	for _, t := range *sim {
		hoursleft = t.burndownhours(day, hoursleft, hoursleft)
	}
}

// burndownSjf burn down shortest job first
func (sim *simulation) burndownSjf(day int) {
	d1 := day + 1
	// no burndown without ticket or on last day
	if len(*sim) <= 0 || d1 >= len((*(*sim)[0]).burndown) {
		return
	}
	// copy sim and sort copy, then burn down
	scp := make(simulation, len(*sim))
	for i, t := range *sim {
		scp[i] = t
	}
	sort.Slice(scp, func(i, j int) bool {
		ti := scp[i]
		tj := scp[j]
		return ti.burndown[day] < tj.burndown[day]
	})
	hoursleft := workhoursday
	for _, t := range scp {
		hoursleft = t.burndownhours(day, hoursleft, hoursleft)
	}
}

// burndownWsjf burn down weightest shortest job first older jobs have priority
func (sim *simulation) burndownWsjf(day int) {
	d1 := day + 1
	// no burndown without ticket or on last day
	if len(*sim) <= 0 || d1 >= len((*(*sim)[0]).burndown) {
		return
	}
	// copy sim and sort copy, then burn down
	scp := make(simulation, len(*sim))
	for i, t := range *sim {
		scp[i] = t
	}
	sort.Slice(scp, func(i, j int) bool {
		ti := scp[i]
		tj := scp[j]
		if ti.startday < tj.startday {
			return true
		}
		return ti.burndown[day] < tj.burndown[day]
	})
	hoursleft := workhoursday
	for _, t := range scp {
		hoursleft = t.burndownhours(day, hoursleft, hoursleft)
	}
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
	simMaxWip := make(simulation, 0, days*3/2)
	simMinWip := make(simulation, 0, days*3/2)
	simSjf := make(simulation, 0, days*3/2)
	simWsjf := make(simulation, 0, days*3/2)
	for d := 0; d < days; d++ {
		count := randomValueInt(meanNewPerDay, stddevNewPerDay, 0)
		sumCount += count
		tickets, effort := createTicketsForDay(d, days, count,
			meanEffortNew, stddevEffortNew, minEffort)
		simMaxWip = simMaxWip.addTickets(tickets)
		simMaxWip.burndownMaxWip(d)
		simMinWip = simMinWip.addTickets(tickets)
		simMinWip.burndownMinWip(d)
		simSjf = simSjf.addTickets(tickets)
		simSjf.burndownSjf(d)
		simWsjf = simWsjf.addTickets(tickets)
		simWsjf.burndownWsjf(d)
		sumEffort += effort

	}
	fmt.Println()
	meanCount := float64(sumCount) / float64(days)
	fmt.Println("mean count:", meanCount)
	meanEffort := float64(sumEffort) / float64(days)
	fmt.Println("mean effort:", meanEffort)
	fmt.Println()
	fmt.Println("Max WIP")
	fmt.Println(simMaxWip.statsLeadTime())
	fmt.Println(simMaxWip)
	fmt.Println("Min WIP")
	fmt.Println(simMinWip.statsLeadTime())
	fmt.Println(simMinWip)
	fmt.Println("Sjf")
	fmt.Println(simSjf.statsLeadTime())
	fmt.Println(simSjf)
	fmt.Println("Wsjf")
	fmt.Println(simWsjf.statsLeadTime())
	fmt.Println(simWsjf)
}
