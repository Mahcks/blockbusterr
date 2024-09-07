package scheduler

import (
	"fmt"
	"time"
)

// ShowJobFunc defines the logic for the show job
func ShowJobFunc() {
	fmt.Println("Show Job is running at:", time.Now().Format(time.RFC3339))
}
