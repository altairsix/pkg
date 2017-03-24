package epoch

import (
	"log"
	"time"
)

var (
	ET *time.Location
	PT *time.Location
)

func init() {
	v, err := time.LoadLocation("America/New_York")
	if err != nil {
		log.Fatalln(err)
	}
	ET = v

	v, err = time.LoadLocation("America/Los_Angeles")
	if err != nil {
		log.Fatalln(err)
	}
	PT = v
}
