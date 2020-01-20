package main

import (
	"log"
	"errors"
	"strconv"
	"strings"
	"encoding/json"
	"github.com/mitchellh/go-z3"
)

type Person struct {
	Name string `json:"name"`// a person's name
	Hours [5][]int `json:"hours"` // monday - friday hours
}

type Schedule struct {
	People []Person `json:"people"` // all the people 
	PerPersonHours int `json:"dutiesPerPerson"` // how many hours each person fills
}

func (schedule *Schedule) Assign() error {
	/*
	variables:
		each person has some variables representing their hours
			e.g. 
			jonas_1 ~ Jonas's first hour
			roz_2 ~ Roz's second hour
		some people can not be certain hours
			e.g.
			andre can't do the fitfh hour, thus andre_1 =/= 5 and andre_2 =/=5

	constraints:
		all variables must be assigned
		no two variables can be equal
	*/

	// setting up the z3 model 
	config := z3.NewConfig()
	ctx := z3.NewContext(config)
	config.Close()
	defer ctx.Close()

	// everyone gets a variable for each of their hours
	vars := make([]*z3.AST, len(schedule.People)*schedule.PerPersonHours)
	v := 0
	for _, person := range schedule.People {
		for i := 0; i < schedule.PerPersonHours; i++ {
			vars[v] = ctx.Const(ctx.Symbol(strings.Join([]string{person.Name, strconv.Itoa(i)}, "_")), ctx.IntSort())
			v++
		}
	}

	// creating a solver
	s := ctx.NewSolver()
	defer s.Close()

	// people can only make certain hours
	hourList := make([](*z3.AST), 31) // UPL is open 30 hours a week
	v = 0
	for _, person := range schedule.People {		
		h := 0	
		for dayno, hours := range person.Hours {
			for _, hour := range hours {
				hourList[h] = ctx.Int((dayno+1)*100+hour, ctx.IntSort())
				h++
			}
		}
		for i := 0; i < schedule.PerPersonHours; i++ {
			openHours := ctx.False()
			for j := 0; j < h; j++ {
				openHours = openHours.Or(vars[v].Eq(hourList[j]))
			}
			s.Assert(openHours)
			v++
		}
	}

	// duties can not overlap
	for i, v := range vars {
		for j, w := range vars {
			if i != j {
				s.Assert(v.Eq(w).Not())
			}
		}
	}

	// todo return this as a value
	if sat := s.Check(); sat == z3.False {
		return errors.New("not sat")
	}
	log.Println(s.Model())

	return nil
 }

func main() {
	b := []byte(`
{
	"dutiesPerPerson": 2,
	"people": [
		{
			"name":"Harrison",
			"hours":[[10, 12, 1, 2, 3, 4], [9, 10, 4], [10, 12, 1, 2, 3, 4], [9, 10, 4], [10, 11, 12, 1, 2, 3, 4]]
		},
		{
			"name":"Karl",
			"hours":[[4],[3,4],[4],[3,4],[11,12,4]]
		},
		{
			"name":"Anisa",
			"hours":[[],[12,1],[12,1],[12,1],[]]
		}
	]
}
`)

	var s Schedule
	error := json.Unmarshal(b, &s)
	s.Assign()

	log.Println(s, error)
}