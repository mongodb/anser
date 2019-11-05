package apm

import "time"

type MonitorConfig struct {
	NumWindows     int
	WindowDuration time.Duration
}

func (c *MonitorConfig) Validate() error {
	if c.NumWindows <= 0 {
		c.NumWindows = 100
	}
	if c.WindowDuration < time.Minute {
		c.WindowDuration = time.Minute
	}

	return nil
}
