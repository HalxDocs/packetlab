package scenario

import (
	"strings"
	"testing"
)

func TestValidateProblems(t *testing.T) {
	cases := map[string]func(*Scenario){
		"schema":           func(s *Scenario) { s.SchemaVersion = "2.0" },
		"id":               func(s *Scenario) { s.ID = "" },
		"name":             func(s *Scenario) { s.Name = "" },
		"nodes":            func(s *Scenario) { s.Nodes = nil },
		"node kind":        func(s *Scenario) { s.Nodes[0].Kind = "switch" },
		"duplicate node":   func(s *Scenario) { s.Nodes[1].ID = s.Nodes[0].ID },
		"duplicate iface":  func(s *Scenario) { s.Nodes[1].Interfaces[0].ID = s.Nodes[0].Interfaces[0].ID },
		"duplicate addr":   func(s *Scenario) { s.Nodes[1].Interfaces[0].Address = "10.0.0.1/24" },
		"bad address":      func(s *Scenario) { s.Nodes[0].Interfaces[0].Address = "10.0.0.999/24" },
		"unknown link":     func(s *Scenario) { s.Links[0].B = "missing:eth0" },
		"self link":        func(s *Scenario) { s.Links[0].B = s.Links[0].A },
		"negative delay":   func(s *Scenario) { s.Links[0].DelayMs = -1 },
		"unknown action":   func(s *Scenario) { s.Actions[0].From = "missing" },
		"bad action kind":  func(s *Scenario) { s.Actions[0].Kind = "send_udp" },
		"bad destination":  func(s *Scenario) { s.Actions[0].To = "nope" },
		"action count":     func(s *Scenario) { s.Actions[0].Count = 1001 },
		"action interval":  func(s *Scenario) { s.Actions[0].IntervalMs = -1 },
		"route interface":  func(s *Scenario) { s.Nodes[0].Routes = []Route{{Network: "0.0.0.0/0", Interface: ""}} },
		"bad route prefix": func(s *Scenario) { s.Nodes[0].Routes = []Route{{Network: "bad", Interface: "a:eth0"}} },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			s, err := Decode(strings.NewReader(validJSON()))
			if err != nil {
				t.Fatalf("Decode: %v", err)
			}
			mutate(s)
			if err := s.Validate(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}
