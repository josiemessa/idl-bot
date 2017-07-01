package main

import (
	"log"
)

type Player struct {
	Name        string `yaml:"name"`
	DiscordName string `yaml:"discord_user"`
}
type Team struct {
	Name    string    `yaml:"team_name"`
	Players []*Player `yaml:"players"`
}
type TeamFile struct {
	Players []*Player `yaml:"players"`
	Teams   []*Team   `yaml:"teams"`
}

func (tf *TeamFile) PopulateTeams() []*Team {
	ts := []*Team{}
	// In the yaml file we only specify the name of a player
	// in a team not the rest of their info, which is in the
	// player section, so build up this list properly
	for _, team := range tf.Teams {
		t := &Team{
			Name:    team.Name,
			Players: []*Player{},
		}
		for _, p := range team.Players {
			found := false
			for _, q := range tf.Players {
				if p.Name == q.Name {
					t.Players = append(t.Players, q)
					found = true
					break
				}
			}
			if !found {
				log.Printf("Player %s from %s not found in player list\n", p, team.Name)
			}
		}
		ts = append(ts, t)
	}
	return ts
}
