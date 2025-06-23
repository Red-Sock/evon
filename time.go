package evon

import (
	"time"
)

type customTime time.Time

func (t customTime) MarshalYAML() (interface{}, error) {
	return formatTime(time.Time(t)), nil
}
