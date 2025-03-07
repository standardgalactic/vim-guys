package config

import "time"

type AuthConfig struct {
	AuthenticationTimeout time.Duration
	Port uint16
}

// I know this is dangerous.  But thsi is my project and you can straight up
// give me your twitch prime
var Config = AuthConfig{
	AuthenticationTimeout: time.Second * 3,
	Port: uint16(42000),
}
