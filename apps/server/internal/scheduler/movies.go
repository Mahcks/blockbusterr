package scheduler

import (
	"fmt"
	"time"
)

// MovieJobFunc defines the logic for the movie job
func MovieJobFunc() {
	fmt.Println("Movie Job is running at:", time.Now().Format(time.RFC3339))
}
