package types

type Variant string

const (
	SotwNonAggregated        Variant = "sotw non-aggregated"
	SotwAggregated           Variant = "sotw aggregated"
	IncrementalNonAggregated Variant = "incremental non-aggregated"
	IncrementalAggregated    Variant = "incremental aggregated"
)

type CukeComment struct {
	Value string `json:"value"`
	Line  int    `json:"line"`
}

type CukeDocstring struct {
	Value       string `json:"value"`
	ContentType string `json:"content_type"`
	Line        int    `json:"line"`
}

type CukeTag struct {
	Name string `json:"name"`
	Line int    `json:"line"`
}

type CukeResult struct {
	Status   string `json:"status"`
	Error    string `json:"error_message,omitempty"`
	Duration *int   `json:"duration,omitempty"`
}

type CukeMatch struct {
	Location string `json:"location"`
}

type CukeStep struct {
	Keyword   string              `json:"keyword"`
	Name      string              `json:"name"`
	Line      int                 `json:"line"`
	Docstring *CukeDocstring      `json:"doc_string,omitempty"`
	Match     CukeMatch           `json:"match"`
	Result    CukeResult          `json:"result"`
	DataTable []*CukeDataTableRow `json:"rows,omitempty"`
}

type CukeDataTableRow struct {
	Cells []string `json:"cells"`
}

type CukeElement struct {
	ID          string     `json:"id"`
	Keyword     string     `json:"keyword"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Line        int        `json:"line"`
	Type        string     `json:"type"`
	Tags        []CukeTag  `json:"tags,omitempty"`
	Steps       []CukeStep `json:"steps,omitempty"`
}

// CukeFeatureJSON ...
type CukeFeatureJSON struct {
	URI         string        `json:"uri"`
	ID          string        `json:"id"`
	Keyword     string        `json:"keyword"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Line        int           `json:"line"`
	Comments    []CukeComment `json:"comments,omitempty"`
	Tags        []CukeTag     `json:"tags,omitempty"`
	Elements    []CukeElement `json:"elements,omitempty"`
}

type VariantResults struct {
	Name            string           `json:"name"`
	Total           int              `json:"total"`
	Passed          int              `json:"passed"`
	Failed          int              `json:"failed"`
	Skipped         int              `json:"skipped"`
	Undefined       int              `json:"undefined"`
	Pending         int              `json:"pending"`
	FailedScenarios []FailedScenario `json:"failedScenarios"`
}

type FailedScenario struct {
	Name       string `json:"name"`
	FailedStep string `json:"failedStep"`
	Line       string `json:"line"`
}

type Results struct {
	Total            int64
	Passed           int64
	Failed           int64
	Skipped          int64
	Undefined        int64
	Pending          int64
	Variants         []string
	ResultsByVariant []VariantResults
}
