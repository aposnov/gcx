package pyroscope

import "time"

// QueryRequest represents a Pyroscope profile query request.
type QueryRequest struct {
	LabelSelector string
	ProfileTypeID string
	Start         time.Time
	End           time.Time
	MaxNodes      int64
}

// IsRange returns true if this is a range query with explicit time bounds.
func (r QueryRequest) IsRange() bool {
	return !r.Start.IsZero() && !r.End.IsZero()
}

// QueryResponse represents the response from a Pyroscope profile query.
type QueryResponse struct {
	Flamegraph *Flamegraph `json:"flamegraph,omitempty"`
}

// Flamegraph represents a flame graph structure.
type Flamegraph struct {
	Names   []string `json:"names"`
	Levels  []Level  `json:"levels"`
	Total   int64    `json:"total,string"`
	MaxSelf int64    `json:"maxSelf,string"`
}

// Level represents a single level in the flame graph.
type Level struct {
	Values []string `json:"values"` // API returns strings that need to be parsed
}

// ProfileTypesRequest represents a request to list profile types.
type ProfileTypesRequest struct {
	Start time.Time
	End   time.Time
}

// ProfileTypesResponse represents the response from a profile types query.
type ProfileTypesResponse struct {
	ProfileTypes []ProfileType `json:"profileTypes"`
}

// ProfileType represents a profile type in Pyroscope.
type ProfileType struct {
	ID         string `json:"ID"`
	Name       string `json:"name"`
	SampleType string `json:"sampleType"`
	SampleUnit string `json:"sampleUnit"`
	PeriodType string `json:"periodType"`
	PeriodUnit string `json:"periodUnit"`
}

// LabelNamesRequest represents a request to list label names.
type LabelNamesRequest struct {
	Matchers []string
	Start    time.Time
	End      time.Time
}

// LabelNamesResponse represents the response from a label names query.
type LabelNamesResponse struct {
	Names []string `json:"names"`
}

// LabelValuesRequest represents a request to list label values.
type LabelValuesRequest struct {
	Name     string
	Matchers []string
	Start    time.Time
	End      time.Time
}

// LabelValuesResponse represents the response from a label values query.
type LabelValuesResponse struct {
	Names []string `json:"names"` // Pyroscope uses "names" for both labels and values
}

// FunctionSample represents a function in the flame graph with computed stats.
type FunctionSample struct {
	Name       string
	Self       int64
	Total      int64
	Percentage float64
}
