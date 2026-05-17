package roadmap

import (
	"io"
	"regexp"
	"time"
)

type config struct {
	Owner         string
	Repo          string
	ProjectNumber int
	OutputDir     string
}

type projectFetcher func(cfg config) (*projectData, error)

type fieldValueNode struct {
	Name  string `json:"name"`
	Text  string `json:"text"`
	Title string `json:"title"`
	Field struct {
		Name string `json:"name"`
	} `json:"field"`
}

type projectItemNode struct {
	Content struct {
		Number     int    `json:"number"`
		Title      string `json:"title"`
		URL        string `json:"url"`
		Repository struct {
			NameWithOwner string `json:"nameWithOwner"`
		} `json:"repository"`
	} `json:"content"`
	FieldValues struct {
		Nodes []fieldValueNode `json:"nodes"`
	} `json:"fieldValues"`
}

type projectData struct {
	Title string
	Items []projectItemNode
}

type projectPage struct {
	Title     string
	Items     []projectItemNode
	HasNext   bool
	EndCursor string
}

type issueRow struct {
	ID         string
	Number     int
	Title      string
	URL        string
	Type       string
	Phase      string
	Sprint     string
	Parent     string
	DependsOn  string
	Status     string
	Estimate   string
	Difficulty int
}

type lane struct {
	ParentKey       string
	Phase           string
	Label           string
	Start           int
	TotalDifficulty int
	CompletedEnd    int
	IsPhase         bool
	Tickets         []ticket
}

type ticket struct {
	ID             string
	Title          string
	URL            string
	Estimate       string
	DependsOn      string
	Status         string
	Completed      bool
	Difficulty     int
	Start          int
	End            int
	StartPct       float64
	WidthPct       float64
	ExecutionOrder int
	TicketCode     string
}

type htmlData struct {
	GeneratedAt      string
	ProjectTitle     string
	Owner            string
	Repo             string
	ProjectNumber    int
	OutputPath       string
	TotalItems       int
	TotalTickets     int
	MaxDifficulty    int
	DifficultyLegend string
	Lanes            []lane
}

// runFunc は Run の関数シグネチャ（テスト用）
type runFunc func(args []string, stdout io.Writer, fetcher projectFetcher, now func() time.Time) error

var (
	ticketOrderPattern = regexp.MustCompile(`TK-\d+-(\d+)`)
	ticketCodePattern  = regexp.MustCompile(`(TK-\d+-\d+)`)
	spPattern          = regexp.MustCompile(`SP(\d+)-(\d+)`)
	phPattern          = regexp.MustCompile(`PH(\d+)`)
	invalidFileChars   = regexp.MustCompile(`[\\/:*?"<>|\x00-\x1f]`)
)

const (
	defaultOwner     = "BossApe"
	defaultRepo      = "Musuhi"
	defaultProjectNo = 2
	outputTimeLayout = "20060102150405"
)

var estimateScore = map[string]int{
	"XS": 1,
	"S":  2,
	"M":  3,
	"L":  5,
	"XL": 8,
}
