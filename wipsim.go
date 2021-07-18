// Package main implements a ticket servicing system simulation.
// The simulation shows the effect of limiting work in progress
// on the lead time of tickets.
// The simulation runs for 100 days.
// Tickets arrive with a gaussian distribution, with mean 1 Ticket per day
// and standard deviation of 1d.
// Tickets have an effort in hours, with a gaussian distribution, with
// mean 6h and standard deviation of 4h.
// Troughput is fixed to 8h per day
// Five scheduling strategies are compared:
// 1. Work on each ticket max 2h per day.
// 2. Work on the tickets in order of arrival
// 3. Work on the ticket with the shortest remaining work first
// 4. Work on the yesterdays tickets first, then on shortest
// 5. Divide remaining work by number of days open and work on ticket with
//    smallest weight first
//
// Ralf Poeppel 2021
//
package main

import (
	"bytes"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"sort"
	"strconv"
)

const maxPrint = 20 // when to print details

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
	leadtime int
	endday   int
	effort   int
	// remaining the remaining effort of a ticket at a day.
	// The day is the index in the array.
	remaining []int
}

// NewTicket create a new ticket
func NewTicket(startday, effort, totaldays int) *ticket {
	t := ticket{}
	t.startday = startday
	t.effort = effort
	t.remaining = make([]int, totaldays)
	t.remaining[startday] = effort
	return &t
}

// Clone create a deep copy of a ticket
func (t *ticket) Clone() *ticket {
	cp := ticket{}
	cp.startday = t.startday
	cp.effort = t.effort
	cp.remaining = make([]int, 0, len(t.remaining))
	cp.remaining = append(cp.remaining, t.remaining...)
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
		if days <= maxPrint {
			fmt.Println(d, count, effort, ticket)
		}
		tickets[i] = ticket
	}
	if count == 0 && days <= maxPrint {
		fmt.Println(d, count)
	}
	return tickets, sumEffort
}

// burndownhours burn down a ticket, max for the given hours
// and return updated hoursleft
func (t *ticket) burndownhours(day, hoursleft, hours int) int {
	d1 := day + 1
	workremain := t.remaining[day]
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
	t.remaining[d1] = workremain
	return hoursleft
}

// simulation the set of all tickets
//type simulation []*ticket
type simulation struct {
	name         string
	burndownaday func(*simulation, int)
	tickets      []*ticket
}

// NewSimulation create a simulation
func NewSimulation(name string, burndownaday func(*simulation, int), size int) simulation {
	sim := simulation{}
	sim.name = name
	sim.burndownaday = burndownaday
	sim.tickets = make([]*ticket, 0, size)
	return sim
}

// addTickets add a copy of the given tickets to the simulation
func (sim simulation) addTickets(ts []*ticket) simulation {
	sts := sim.tickets
	for _, t := range ts {
		tcp := t.Clone()
		sts = append(sts, tcp)
	}
	sim.tickets = sts
	return sim
}

// copyTickets return sim.tickets copy
func (sim *simulation) copyTickets() []*ticket {
	tscp := make([]*ticket, len((*sim).tickets))
	for i, t := range (*sim).tickets {
		tscp[i] = t
	}
	return tscp
}

// statsLeadTime return average and standard deviation
// and sum of mean and stdev of tickets leadtime
func (sim simulation) statsLeadTime() (float64, float64, float64) {
	var sum float64 = 0.0
	var sumSq float64 = 0.0
	for _, t := range sim.tickets {
		l := float64(t.leadtime)
		sum += l
		sumSq += l * l
	}
	// calculate the mean/std.dev
	l := float64(len(sim.tickets))
	meanSq := sumSq / l
	mean := sum / l
	stdev := math.Sqrt(meanSq - mean*mean)
	return mean, stdev, mean + stdev
}

// String create nice representation
func (sim simulation) String() string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintln(sim.name))
	m, s, ms := sim.statsLeadTime()
	frmt := "Leadtime of tickets mean: %.2f stdev: %.2f mean+stdev: %.2f\n"
	buf.WriteString(fmt.Sprintf(frmt, m, s, ms))
	if len(sim.tickets) <= maxPrint {
		header := "# start leadtime end effort [remaining per day]\n"
		buf.WriteString(header)
		for i, t := range sim.tickets {
			buf.WriteString(fmt.Sprintln(i, *t))
		}
	}
	return buf.String()
}

// workhoursday working hours per day
const workhoursday = 8

// burndownMaxWip burn down maximum number of tickets in work, try each 2h for a day
func burndownMaxWip(sim *simulation, day int) {
	hourswork := 2
	hoursleft := workhoursday
	for _, t := range (*sim).tickets {
		hoursleft = t.burndownhours(day, hoursleft, hourswork)
	}
	if hoursleft > 0 {
		// burn hours left
		for _, t := range (*sim).tickets {
			hoursleft = t.burndownhours(day, hoursleft, hoursleft)
		}
	}
}

// burndownOldestFirst burn down the oldest tickets first
func burndownOldestFirst(sim *simulation, day int) {
	hoursleft := workhoursday
	for _, t := range (*sim).tickets {
		hoursleft = t.burndownhours(day, hoursleft, hoursleft)
	}
}

// burndownSjf burn down shortest job first
func burndownSjf(sim *simulation, day int) {
	// copy sim.tickets and sort copy, then burn down
	tscp := sim.copyTickets()
	sort.Slice(tscp, func(i, j int) bool {
		ti := tscp[i]
		tj := tscp[j]
		return ti.remaining[day] < tj.remaining[day]
	})
	hoursleft := workhoursday
	for _, t := range tscp {
		hoursleft = t.burndownhours(day, hoursleft, hoursleft)
	}
}

// burndownOsjf burn down shortest job first, older jobs have priority
func burndownOsjf(sim *simulation, day int) {
	// copy sim and sort copy, then burn down
	tscp := sim.copyTickets()
	sort.Slice(tscp, func(i, j int) bool {
		ti := tscp[i]
		tj := tscp[j]
		if ti.startday < tj.startday {
			return true
		}
		return ti.remaining[day] < tj.remaining[day]
	})
	hoursleft := workhoursday
	for _, t := range tscp {
		hoursleft = t.burndownhours(day, hoursleft, hoursleft)
	}
}

// burndownAwsjf burn down age weighted, shortest job first
func burndownAwsjf(sim *simulation, day int) {
	// copy sim and sort copy, then burn down
	tscp := sim.copyTickets()
	sort.Slice(tscp, func(i, j int) bool {
		ti := tscp[i]
		tj := tscp[j]
		wi := day + 1 - ti.startday
		wj := day + 1 - tj.startday
		return ti.remaining[day]/wi < tj.remaining[day]/wj
	})
	hoursleft := workhoursday
	for _, t := range tscp {
		hoursleft = t.burndownhours(day, hoursleft, hoursleft)
	}
}

// simulationset the set of simulations
type simulationset []simulation

// NewSimulationset create the set of simulations
func NewSimulationset(days int) simulationset {
	sz := days * 3 / 2 // some more size avoid reallocation
	cnt := 5
	simset := make(simulationset, cnt)
	simset[0] = NewSimulation("Equal working", burndownMaxWip, sz)
	simset[1] = NewSimulation("Oldest first", burndownOldestFirst, sz)
	simset[2] = NewSimulation("Shortest first", burndownSjf, sz)
	simset[3] = NewSimulation("Oldest, shortest first", burndownOsjf, sz)
	simset[4] = NewSimulation("Age weighted, shortest first", burndownAwsjf, sz)
	return simset
}

func (simset simulationset) String() string {
	var buf bytes.Buffer
	for _, s := range simset {
		buf.WriteString(fmt.Sprintln(s))
	}
	return buf.String()
}

// addTickets add the tickets to each simulation
func (simset simulationset) addTickets(ts []*ticket) simulationset {
	for i, s := range simset {
		simset[i] = s.addTickets(ts)
	}
	return simset
}

// burndown the tickets in each simulation
func (simset simulationset) burndown(day int) {
	for _, s := range simset {
		s.burndownaday(&s, day)
	}
}

// simdays read number of days to simulate from args, use default if none is given,
// log fatal if not readable
func simdays() int {
	a := os.Args
	if len(a) <= 1 {
		return maxPrint // the default
	}
	d, err := strconv.Atoi(a[1])
	if err != nil || len(a) > 2 {
		log.Fatal("usage: " + a[0] + " <n>")
	}
	return d
}

func printSimulatedDataHeader(days int) {
	fmt.Println("Simulating", days, "days")
	if days <= maxPrint {
		header := "day, count, effort, ticket{start leadtime end effort" +
			" [remaining/day]}"
		fmt.Println(header)
	}
}

func main() {
	days := simdays()
	printSimulatedDataHeader(days)
	meanNewPerDay := 1.0
	stddevNewPerDay := 1.0
	sumCount := 0
	meanEffortNew := 6.0
	stddevEffortNew := 4.0
	minEffort := 1
	sumEffort := 0
	simset := NewSimulationset(days)
	for d := 0; d < days; d++ {
		count := randomValueInt(meanNewPerDay, stddevNewPerDay, 0)
		sumCount += count
		tickets, effort := createTicketsForDay(d, days, count,
			meanEffortNew, stddevEffortNew, minEffort)
		simset = simset.addTickets(tickets)
		// burndown on all days except last day
		if d < days-1 {
			simset.burndown(d)
		}
		sumEffort += effort

	}
	fmt.Println()
	meanCount := float64(sumCount) / float64(days)
	fmt.Println("mean ticket count per day:", meanCount)
	meanEffort := float64(sumEffort) / float64(days)
	fmt.Println("mean ticket effort per day:", meanEffort)
	fmt.Println()
	fmt.Println(simset)
}
