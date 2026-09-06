package enums

import "testing"

func TestMediaTypeIsValid(t *testing.T) {
	tests := []struct {
		name  string
		value MediaType
		want  bool
	}{
		{name: "movie", value: MediaTypeMovie, want: true},
		{name: "show", value: MediaTypeShow, want: true},
		{name: "invalid", value: MediaType("tv"), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.value.IsValid(); got != tt.want {
				t.Fatalf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseMediaType(t *testing.T) {
	v, ok := ParseMediaType(" Movie ")
	if !ok || v != MediaTypeMovie {
		t.Fatalf("ParseMediaType(movie) = (%q, %v), want (%q, true)", v, ok, MediaTypeMovie)
	}
	if _, ok := ParseMediaType("tv"); ok {
		t.Fatal("ParseMediaType(tv) should be invalid")
	}
}

func TestActivityStatusIsValid(t *testing.T) {
	tests := []struct {
		name  string
		value ActivityStatus
		want  bool
	}{
		{name: "added", value: ActivityStatusAdded, want: true},
		{name: "requested", value: ActivityStatusRequested, want: true},
		{name: "rejected", value: ActivityStatusRejected, want: true},
		{name: "skipped", value: ActivityStatusSkipped, want: true},
		{name: "failed", value: ActivityStatusFailed, want: true},
		{name: "blocked", value: ActivityStatusBlocked, want: true},
		{name: "invalid", value: ActivityStatus("ok"), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.value.IsValid(); got != tt.want {
				t.Fatalf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseActivityStatus(t *testing.T) {
	v, ok := ParseActivityStatus(" ReJeCtEd ")
	if !ok || v != ActivityStatusRejected {
		t.Fatalf("ParseActivityStatus(rejected) = (%q, %v), want (%q, true)", v, ok, ActivityStatusRejected)
	}
	if _, ok := ParseActivityStatus("ok"); ok {
		t.Fatal("ParseActivityStatus(ok) should be invalid")
	}
}

func TestJobRunStatusIsValid(t *testing.T) {
	tests := []struct {
		name  string
		value JobRunStatus
		want  bool
	}{
		{name: "running", value: JobRunStatusRunning, want: true},
		{name: "completed", value: JobRunStatusCompleted, want: true},
		{name: "failed", value: JobRunStatusFailed, want: true},
		{name: "invalid", value: JobRunStatus("done"), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.value.IsValid(); got != tt.want {
				t.Fatalf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseJobRunStatus(t *testing.T) {
	v, ok := ParseJobRunStatus(" completed ")
	if !ok || v != JobRunStatusCompleted {
		t.Fatalf("ParseJobRunStatus(completed) = (%q, %v), want (%q, true)", v, ok, JobRunStatusCompleted)
	}
	if _, ok := ParseJobRunStatus("done"); ok {
		t.Fatal("ParseJobRunStatus(done) should be invalid")
	}
}
