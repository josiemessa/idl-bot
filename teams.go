package main

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
	return tf.Teams
}
