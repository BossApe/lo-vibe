package source

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
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
	Iteration  string
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
	IsPhase         bool
	CompletedEnd    int
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
	ExecutionOrder int
	TicketCode     string
	StartPct       float64
	WidthPct       float64
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

var (
	ticketOrderPattern = regexp.MustCompile(`TK-\d+-(\d+)`)
	ticketCodePattern  = regexp.MustCompile(`(TK-\d+-\d+)`)
	itPattern          = regexp.MustCompile(`IT(\d+)-(\d+)`)
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

func Main() {
	if err := Run(os.Args[1:], os.Stdout, fetchProjectByGH, time.Now); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
}

func Run(args []string, stdout io.Writer, fetcher projectFetcher, now func() time.Time) error {
	cfg, err := parseConfig(args)
	if err != nil {
		return err
	}

	project, err := fetcher(cfg)
	if err != nil {
		return err
	}

	rows := normalizeRows(project, cfg)
	if len(rows) == 0 {
		return errors.New("指定プロジェクト内に対象リポジトリのIssueが見つかりません")
	}

	lanes := buildLanes(rows)
	if len(lanes) == 0 {
		return errors.New("Type=Ticket のIssueが見つかりません")
	}

	maxDifficulty, totalTickets := applyPercentages(lanes)
	if maxDifficulty <= 0 {
		maxDifficulty = 1
	}

	if err := os.MkdirAll(cfg.OutputDir, 0o755); err != nil {
		return fmt.Errorf("出力ディレクトリ作成失敗: %w", err)
	}

	fileName := buildOutputFileName(project.Title, now())
	outputPath := filepath.Join(cfg.OutputDir, fileName)
	if err := writeHTML(outputPath, htmlData{
		GeneratedAt:      now().Format("2006-01-02 15:04:05 -0700"),
		ProjectTitle:     project.Title,
		Owner:            cfg.Owner,
		Repo:             cfg.Repo,
		ProjectNumber:    cfg.ProjectNumber,
		OutputPath:       outputPath,
		TotalItems:       len(rows),
		TotalTickets:     totalTickets,
		MaxDifficulty:    maxDifficulty,
		DifficultyLegend: "XS=1, S=2, M=3, L=5, XL=8",
		Lanes:            lanes,
	}); err != nil {
		return err
	}

	fmt.Fprintf(stdout, "プロジェクト: %s\n", project.Title)
	fmt.Fprintf(stdout, "レーン数: %d / チケット数: %d\n", len(lanes), totalTickets)
	fmt.Fprintf(stdout, "出力: %s\n", outputPath)
	return nil
}

func parseConfig(args []string) (config, error) {
	fs := flag.NewFlagSet("generate_issue_roadmap", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	cfg := config{}
	fs.StringVar(&cfg.Owner, "owner", defaultOwner, "GitHub owner (user or organization)")
	fs.StringVar(&cfg.Repo, "repo", defaultRepo, "GitHub repository name")
	fs.IntVar(&cfg.ProjectNumber, "project-number", defaultProjectNo, "GitHub Project (v2) number")
	fs.StringVar(&cfg.OutputDir, "output-dir", defaultOutputDir(), "出力先ディレクトリ")

	if err := fs.Parse(args); err != nil {
		return config{}, fmt.Errorf("設定の初期化失敗: %w", err)
	}
	if cfg.Owner == "" || cfg.Repo == "" {
		return config{}, errors.New("設定の初期化失敗: owner/repo は必須です")
	}
	if cfg.ProjectNumber <= 0 {
		return config{}, errors.New("設定の初期化失敗: project-number は1以上で指定してください")
	}
	return cfg, nil
}

func defaultOutputDir() string {
	return filepath.Join("..", "..", "_document", "003.設計・開発・テストフェーズ", "002.開発進捗")
}

func fetchProjectByGH(cfg config) (*projectData, error) {
	userProj, userErr := fetchProjectByGHPaged(cfg, true)
	if userErr == nil && userProj != nil {
		return userProj, nil
	}

	orgProj, orgErr := fetchProjectByGHPaged(cfg, false)
	if orgErr == nil && orgProj != nil {
		return orgProj, nil
	}

	if userErr != nil && orgErr != nil {
		return nil, fmt.Errorf("gh api 実行失敗 (user=%v, org=%v)", userErr, orgErr)
	}
	return nil, errors.New("projectV2 が見つかりませんでした")
}

func fetchProjectByGHPaged(cfg config, isUser bool) (*projectData, error) {
	query := graphQLQueryOrg
	if isUser {
		query = graphQLQueryUser
	}

	after := ""
	seenCursor := map[string]struct{}{}
	allItems := make([]projectItemNode, 0, 128)
	projectTitle := ""

	for {
		payload, err := runGraphQLQuery(cfg.Owner, cfg.ProjectNumber, query, after)
		if err != nil {
			return nil, err
		}

		page := parseProjectPageFromPayload(payload, isUser)
		if page == nil {
			return nil, nil
		}
		if projectTitle == "" {
			projectTitle = page.Title
		}
		allItems = append(allItems, page.Items...)

		if !page.HasNext || strings.TrimSpace(page.EndCursor) == "" {
			break
		}
		if _, exists := seenCursor[page.EndCursor]; exists {
			break
		}
		seenCursor[page.EndCursor] = struct{}{}
		after = page.EndCursor
	}

	return &projectData{Title: projectTitle, Items: allItems}, nil
}

func runGraphQLQuery(owner string, projectNumber int, query string, after string) ([]byte, error) {
	cmd := exec.Command(
		"gh", "api", "graphql",
		"-F", "owner="+owner,
		"-F", fmt.Sprintf("number=%d", projectNumber),
		"-f", "query="+query,
	)
	if strings.TrimSpace(after) != "" {
		cmd.Args = append(cmd.Args, "-F", "after="+after)
	}
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if stderr.Len() > 0 {
			return nil, fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
		}
		return nil, err
	}
	return out.Bytes(), nil
}

func parseProjectPageFromPayload(payload []byte, user bool) *projectPage {
	var resp struct {
		Data struct {
			User struct {
				ProjectV2 *struct {
					Title string `json:"title"`
					Items struct {
						Nodes    []projectItemNode `json:"nodes"`
						PageInfo struct {
							HasNextPage bool   `json:"hasNextPage"`
							EndCursor   string `json:"endCursor"`
						} `json:"pageInfo"`
					} `json:"items"`
				} `json:"projectV2"`
			} `json:"user"`
			Organization struct {
				ProjectV2 *struct {
					Title string `json:"title"`
					Items struct {
						Nodes    []projectItemNode `json:"nodes"`
						PageInfo struct {
							HasNextPage bool   `json:"hasNextPage"`
							EndCursor   string `json:"endCursor"`
						} `json:"pageInfo"`
					} `json:"items"`
				} `json:"projectV2"`
			} `json:"organization"`
		} `json:"data"`
	}
	if err := json.Unmarshal(payload, &resp); err != nil {
		return nil
	}
	if user {
		if resp.Data.User.ProjectV2 == nil {
			return nil
		}
		return &projectPage{
			Title:     resp.Data.User.ProjectV2.Title,
			Items:     resp.Data.User.ProjectV2.Items.Nodes,
			HasNext:   resp.Data.User.ProjectV2.Items.PageInfo.HasNextPage,
			EndCursor: resp.Data.User.ProjectV2.Items.PageInfo.EndCursor,
		}
	}
	if resp.Data.Organization.ProjectV2 == nil {
		return nil
	}
	return &projectPage{
		Title:     resp.Data.Organization.ProjectV2.Title,
		Items:     resp.Data.Organization.ProjectV2.Items.Nodes,
		HasNext:   resp.Data.Organization.ProjectV2.Items.PageInfo.HasNextPage,
		EndCursor: resp.Data.Organization.ProjectV2.Items.PageInfo.EndCursor,
	}
}

func normalizeRows(project *projectData, cfg config) []issueRow {
	repoFilter := cfg.Owner + "/" + cfg.Repo
	rows := make([]issueRow, 0, len(project.Items))
	for _, item := range project.Items {
		content := item.Content
		if content.Number == 0 || content.Repository.NameWithOwner != repoFilter {
			continue
		}
		fm := fieldMap(item.FieldValues.Nodes)
		est := strings.ToUpper(strings.TrimSpace(fm["Estimate"]))
		diff, ok := estimateScore[est]
		if !ok {
			diff = 1
		}
		rows = append(rows, issueRow{
			ID:         fmt.Sprintf("#%d", content.Number),
			Number:     content.Number,
			Title:      content.Title,
			URL:        content.URL,
			Type:       fallback(fm["Type"], "Unknown"),
			Phase:      fallback(fm["Phase"], "Unknown"),
			Iteration:  fallback(fm["Iteration"], "Unknown"),
			Parent:     fallback(fm["Parent"], "No Parent"),
			DependsOn:  fm["Depends on"],
			Status:     fm["Status"],
			Estimate:   fallback(est, "N/A"),
			Difficulty: diff,
		})
	}
	return rows
}

func fieldMap(nodes []fieldValueNode) map[string]string {
	m := map[string]string{}
	for _, n := range nodes {
		name := strings.TrimSpace(n.Field.Name)
		if name == "" {
			continue
		}
		if v := strings.TrimSpace(n.Name); v != "" {
			m[name] = v
			continue
		}
		if v := strings.TrimSpace(n.Text); v != "" {
			m[name] = v
			continue
		}
		if v := strings.TrimSpace(n.Title); v != "" {
			m[name] = v
		}
	}
	return m
}

func fallback(v, alt string) string {
	if strings.TrimSpace(v) == "" {
		return alt
	}
	return v
}

func isCompletedStatus(v string) bool {
	s := strings.ToLower(strings.TrimSpace(v))
	s = strings.ReplaceAll(s, "_", "")
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, " ", "")
	return s == "done" || s == "completed" || s == "complete" || s == "closed" || s == "完了"
}

func buildLanes(rows []issueRow) []lane {
	tickets := make([]issueRow, 0)
	parents := make([]issueRow, 0)
	for _, r := range rows {
		switch r.Type {
		case "Ticket":
			tickets = append(tickets, r)
		case "Phase", "Iteration":
			parents = append(parents, r)
		}
	}
	if len(tickets) == 0 {
		return nil
	}

	ticketGroups := map[string][]issueRow{}
	for _, t := range tickets {
		key := strings.TrimSpace(t.Parent)
		if key == "" {
			key = "No Parent"
		}
		ticketGroups[key] = append(ticketGroups[key], t)
	}

	parentMap := map[string]issueRow{}
	for _, p := range parents {
		key := extractParentKey(p)
		parentMap[key] = p
	}
	for key := range ticketGroups {
		if _, ok := parentMap[key]; !ok {
			parentMap[key] = issueRow{Title: key + " parent lane", Parent: key, Phase: "Unknown", Iteration: "Unknown", Difficulty: 0}
		}
	}

	keys := make([]string, 0, len(parentMap))
	for k := range parentMap {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		ai, at, as, an := parentSortTuple(keys[i])
		bi, bt, bs, bn := parentSortTuple(keys[j])
		if ai != bi {
			return ai < bi
		}
		if at != bt {
			return at < bt
		}
		if as != bs {
			return as < bs
		}
		return an < bn
	})

	laneMap := map[string]lane{}
	for _, key := range keys {
		parent := parentMap[key]
		items := ticketGroups[key]
		sort.Slice(items, func(i, j int) bool {
			io := extractTicketOrder(items[i].Title, items[i].Number)
			jo := extractTicketOrder(items[j].Title, items[j].Number)
			if io != jo {
				return io < jo
			}
			return items[i].Number < items[j].Number
		})

		total := parent.Difficulty
		if len(items) > 0 {
			total = 0
			for _, it := range items {
				total += it.Difficulty
			}
		}

		ln := lane{
			ParentKey:       key,
			Phase:           fallback(parent.Phase, "Unknown"),
			Label:           key,
			TotalDifficulty: total,
			IsPhase:         parent.Type == "Phase",
			Tickets:         make([]ticket, 0, len(items)),
		}
		cursor := 0
		for idx, it := range items {
			start := cursor
			end := cursor + it.Difficulty
			cursor = end
			completed := isCompletedStatus(it.Status)
			if completed && start == ln.CompletedEnd {
				ln.CompletedEnd = end
			}
			ln.Tickets = append(ln.Tickets, ticket{
				ID:             it.ID,
				Title:          it.Title,
				URL:            it.URL,
				Estimate:       it.Estimate,
				DependsOn:      it.DependsOn,
				Status:         fallback(it.Status, "Unknown"),
				Completed:      completed,
				Difficulty:     it.Difficulty,
				Start:          start,
				End:            end,
				ExecutionOrder: idx + 1,
				TicketCode:     extractTicketCode(it.Title, it.ID),
			})
		}
		laneMap[key] = ln
	}

	globalCursor := 0
	phaseStarts := map[int]int{}
	phaseEnds := map[int]int{}
	phaseKeys := map[int]string{}

	for _, key := range keys {
		ph, t, _, _ := parentSortTuple(key)
		if ph == 999 {
			continue
		}
		if t == 0 {
			phaseKeys[ph] = key
			continue
		}
		if t != 1 {
			continue
		}

		ln := laneMap[key]
		if _, ok := phaseStarts[ph]; !ok {
			phaseStarts[ph] = globalCursor
		}

		ln.Start = globalCursor
		ln.CompletedEnd += ln.Start
		for i := range ln.Tickets {
			ln.Tickets[i].Start += ln.Start
			ln.Tickets[i].End += ln.Start
		}
		globalCursor += ln.TotalDifficulty
		phaseEnds[ph] = globalCursor
		laneMap[key] = ln
	}

	for _, key := range keys {
		ph, t, _, _ := parentSortTuple(key)
		if ph == 999 || t != 0 {
			continue
		}
		ln := laneMap[key]
		start, hasStart := phaseStarts[ph]
		end, hasEnd := phaseEnds[ph]
		if hasStart && hasEnd && end >= start {
			ln.Start = start
			ln.TotalDifficulty = end - start
			ln.CompletedEnd = start
		} else {
			ln.Start = globalCursor
			ln.CompletedEnd = ln.Start
			globalCursor += ln.TotalDifficulty
		}
		laneMap[key] = ln
	}

	lanes := make([]lane, 0, len(keys))
	for _, key := range keys {
		ln := laneMap[key]
		if ln.IsPhase && ln.TotalDifficulty <= 0 {
			continue
		}
		lanes = append(lanes, ln)
	}
	return lanes
}

func applyPercentages(lanes []lane) (maxDifficulty int, totalTickets int) {
	for i := range lanes {
		laneEnd := lanes[i].Start + lanes[i].TotalDifficulty
		if laneEnd > maxDifficulty {
			maxDifficulty = laneEnd
		}
		totalTickets += len(lanes[i].Tickets)
		for ti := range lanes[i].Tickets {
			if lanes[i].Tickets[ti].End > maxDifficulty {
				maxDifficulty = lanes[i].Tickets[ti].End
			}
		}
	}
	if maxDifficulty <= 0 {
		maxDifficulty = 1
	}
	for li := range lanes {
		for ti := range lanes[li].Tickets {
			t := &lanes[li].Tickets[ti]
			t.StartPct = float64(t.Start) * 100.0 / float64(maxDifficulty)
			t.WidthPct = float64(t.Difficulty) * 100.0 / float64(maxDifficulty)
		}
	}
	return maxDifficulty, totalTickets
}

func extractTicketOrder(title string, fallback int) int {
	m := ticketOrderPattern.FindStringSubmatch(title)
	if len(m) < 2 {
		return fallback
	}
	var n int
	_, err := fmt.Sscanf(m[1], "%d", &n)
	if err != nil {
		return fallback
	}
	return n
}

func extractTicketCode(title, fallback string) string {
	m := ticketCodePattern.FindStringSubmatch(title)
	if len(m) < 2 {
		return fallback
	}
	return m[1]
}

func extractParentKey(r issueRow) string {
	title := r.Title
	if r.Type == "Iteration" {
		if m := itPattern.FindStringSubmatch(title); len(m) > 0 {
			return m[0]
		}
	}
	if r.Type == "Phase" {
		if m := phPattern.FindStringSubmatch(title); len(m) > 0 {
			return m[0]
		}
	}
	if strings.TrimSpace(r.Parent) != "" {
		return strings.TrimSpace(r.Parent)
	}
	return r.ID
}

func parentSortTuple(parentKey string) (int, int, int, string) {
	if m := itPattern.FindStringSubmatch(parentKey); len(m) == 3 {
		var ph, it int
		fmt.Sscanf(m[1], "%d", &ph)
		fmt.Sscanf(m[2], "%d", &it)
		return ph, 1, it, parentKey
	}
	if m := phPattern.FindStringSubmatch(parentKey); len(m) == 2 {
		var ph int
		fmt.Sscanf(m[1], "%d", &ph)
		return ph, 0, 0, parentKey
	}
	return 999, 2, 999, parentKey
}

func buildOutputFileName(projectTitle string, now time.Time) string {
	base := sanitizeFileName(strings.TrimSpace(projectTitle))
	if base == "" {
		base = "プロジェクト"
	}
	return fmt.Sprintf("%s進捗_%s.html", base, now.Format(outputTimeLayout))
}

func sanitizeFileName(v string) string {
	s := invalidFileChars.ReplaceAllString(v, "_")
	s = strings.TrimSpace(s)
	s = strings.Trim(s, ".")
	if s == "" {
		return ""
	}
	return s
}

func writeHTML(path string, data htmlData) error {
	funcMap := template.FuncMap{
		"colorClass": func(n int) string {
			return fmt.Sprintf("c%d", (n-1)%8+1)
		},
		"jsStr": func(s string) template.JS {
			b, _ := json.Marshal(s)
			return template.JS(b)
		},
	}

	tmpl, err := template.New("roadmap").Funcs(funcMap).Parse(htmlTemplate)
	if err != nil {
		return fmt.Errorf("HTMLテンプレート解析失敗: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("HTML生成失敗: %w", err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		return fmt.Errorf("HTML書き込み失敗: %w", err)
	}
	return nil
}

const graphQLQueryUser = `query($owner: String!, $number: Int!, $after: String) {
  user(login: $owner) {
    projectV2(number: $number) {
      title
			items(first: 100, after: $after) {
        nodes {
          content {
            ... on Issue {
              number
              title
              url
              repository {
                nameWithOwner
              }
            }
          }
          fieldValues(first: 30) {
            nodes {
              ... on ProjectV2ItemFieldSingleSelectValue {
                name
                field {
                  ... on ProjectV2SingleSelectField {
                    name
                  }
                }
              }
              ... on ProjectV2ItemFieldTextValue {
                text
                field {
                  ... on ProjectV2FieldCommon {
                    name
                  }
                }
              }
              ... on ProjectV2ItemFieldIterationValue {
                title
                field {
                  ... on ProjectV2IterationField {
                    name
                  }
                }
              }
            }
          }
        }
				pageInfo {
					hasNextPage
					endCursor
				}
      }
    }
  }
}`

const graphQLQueryOrg = `query($owner: String!, $number: Int!, $after: String) {
  organization(login: $owner) {
    projectV2(number: $number) {
      title
			items(first: 100, after: $after) {
        nodes {
          content {
            ... on Issue {
              number
              title
              url
              repository {
                nameWithOwner
              }
            }
          }
          fieldValues(first: 30) {
            nodes {
              ... on ProjectV2ItemFieldSingleSelectValue {
                name
                field {
                  ... on ProjectV2SingleSelectField {
                    name
                  }
                }
              }
              ... on ProjectV2ItemFieldTextValue {
                text
                field {
                  ... on ProjectV2FieldCommon {
                    name
                  }
                }
              }
              ... on ProjectV2ItemFieldIterationValue {
                title
                field {
                  ... on ProjectV2IterationField {
                    name
                  }
                }
              }
            }
          }
        }
				pageInfo {
					hasNextPage
					endCursor
				}
      }
    }
  }
}`

const htmlTemplate = `<!doctype html>
<html lang="ja">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>{{.ProjectTitle}}進捗</title>
  <style>
    *, *::before, *::after { box-sizing: border-box; }
    body {
      margin: 0;
      padding: 20px 24px 40px;
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", "Hiragino Sans", "Yu Gothic", sans-serif;
      font-size: 13px;
      color: #1f2937;
      background: #f8fafc;
    }
    h1 { margin: 0 0 4px; font-size: 22px; }
    .meta { color: #374151; margin-bottom: 3px; }
    .chart-wrap {
      margin-top: 16px;
      background: #fff;
      border: 1px solid #e5e7eb;
      border-radius: 12px;
      padding: 16px 16px 12px;
      overflow-x: auto;
    }
    /* ガントチャート */
	.gantt { display: table; width: 100%; border-collapse: collapse; min-width: 700px; position: relative; table-layout: fixed; }
    .gantt-row { display: table-row; }
    .gantt-row:hover .gantt-bar-cell { background: #f0fdf4; }
    .gantt-label {
      display: table-cell;
      width: 210px;
      max-width: 210px;
      padding: 3px 10px 3px 0;
      vertical-align: middle;
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
      font-size: 11px;
      color: #374151;
      border-bottom: 1px solid #f3f4f6;
    }
    .gantt-label.is-phase {
      font-weight: 700;
      color: #0f172a;
      font-size: 13px;
      padding-top: 14px;
      padding-bottom: 2px;
    }
    .gantt-label.is-parent {
      font-weight: 600;
      color: #374151;
      font-size: 12px;
      padding-left: 14px;
      padding-top: 6px;
    }
    .gantt-label.is-ticket { padding-left: 32px; }
    .gantt-row.is-phase-row .gantt-label,
    .gantt-row.is-phase-row .gantt-bar-cell {
      border-top: 2px solid #64748b;
    }
    .gantt-bar-cell {
      display: table-cell;
      vertical-align: middle;
      padding: 3px 0;
      border-bottom: 1px solid #f3f4f6;
      position: relative;
    }
    .bar-bg {
      position: relative;
      height: 22px;
      background: #f1f5f9;
      border-radius: 4px;
    }
		.progress-line-global {
			position: absolute;
			top: 0;
			width: 3px;
			pointer-events: none;
			background: rgba(153, 27, 27, 0.98);
			border-radius: 0;
			box-shadow: 0 0 0 1px rgba(153, 27, 27, 0.18);
			z-index: 4;
		}
    .bar-parent {
      position: absolute;
      top: 3px; bottom: 3px;
      border-radius: 3px;
      background: rgba(100,116,139,0.22);
      pointer-events: none;
    }
    .bar-ticket {
      position: absolute;
      top: 0; bottom: 0;
      border-radius: 4px;
      cursor: pointer;
      transition: filter .15s;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 10px;
      font-weight: 700;
      color: rgba(255,255,255,0.92);
      text-shadow: 0 1px 2px rgba(0,0,0,.3);
      overflow: hidden;
      text-decoration: none;
    }
    .bar-ticket:hover { filter: brightness(1.12); }
    /* 実行順カラーパレット (vegalite版と近似) */
    .c1  { background: #4e79a7; }
    .c2  { background: #f28e2b; }
    .c3  { background: #e15759; }
    .c4  { background: #76b7b2; }
    .c5  { background: #59a14f; }
    .c6  { background: #edc948; }
    .c7  { background: #b07aa1; }
    .c8  { background: #ff9da7; }
    /* X軸ラベル */
    .axis-row { display: table-row; }
    .axis-label-cell { display: table-cell; }
    .axis-bar-cell  { display: table-cell; }
    .x-axis {
      position: relative;
      height: 24px;
      margin-top: 2px;
    }
    .x-tick {
      position: absolute;
      top: 0;
      transform: translateX(-50%);
      font-size: 10px;
      color: #6b7280;
      white-space: nowrap;
    }
    .x-tick::before {
      content: '';
      position: absolute;
      top: -4px;
      left: 50%;
      transform: translateX(-50%);
      width: 1px;
      height: 4px;
      background: #d1d5db;
    }
    /* ツールチップ */
    #tt {
      position: fixed;
      display: none;
      z-index: 9999;
      background: #1e293b;
      color: #f1f5f9;
      border-radius: 8px;
      padding: 10px 14px;
      font-size: 12px;
      line-height: 1.6;
      max-width: 320px;
      pointer-events: none;
      box-shadow: 0 8px 24px rgba(0,0,0,.3);
    }
    #tt .tt-title { font-weight: 700; margin-bottom: 4px; color: #e2e8f0; }
    #tt .tt-row   { display: flex; gap: 8px; }
    #tt .tt-key   { color: #94a3b8; min-width: 60px; }
    #tt .tt-val   { color: #f8fafc; }
  </style>
</head>
<body>
  <h1>{{.ProjectTitle}}進捗</h1>
  <div class="meta">生成日時: {{.GeneratedAt}}</div>
  <div class="meta">ソース: GitHub Project #{{.ProjectNumber}} ({{.Owner}}/{{.Repo}}) / アイテム数: {{.TotalItems}}</div>
  <div class="meta">難易度マッピング: {{.DifficultyLegend}}</div>

  <div class="chart-wrap">
    <div class="gantt" id="gantt"></div>
    <div style="display:table;width:100%;min-width:700px;">
      <div style="display:table-row;">
        <div style="display:table-cell;width:220px;min-width:160px;"></div>
        <div style="display:table-cell;">
          <div class="x-axis" id="xaxis"></div>
        </div>
      </div>
    </div>
  </div>

  <div id="tt"></div>

  <script>
  (function(){
    const MAX = {{.MaxDifficulty}};
    const lanes = [
      {{range .Lanes}}
      {
        phase: {{jsStr .Phase}},
        label: {{jsStr .Label}},
				start: {{.Start}},
        total: {{.TotalDifficulty}},
				isPhase: {{.IsPhase}},
				completed: {{.CompletedEnd}},
        tickets: [
          {{range .Tickets}}
          {
            id:      {{jsStr .ID}},
            code:    {{jsStr .TicketCode}},
            title:   {{jsStr .Title}},
            url:     {{jsStr .URL}},
            est:     {{jsStr .Estimate}},
            dep:     {{jsStr .DependsOn}},
						status:  {{jsStr .Status}},
						done:    {{.Completed}},
            diff:    {{.Difficulty}},
            start:   {{.Start}},
            end:     {{.End}},
            order:   {{.ExecutionOrder}},
          },
          {{end}}
        ],
      },
      {{end}}
    ];

    const gantt = document.getElementById('gantt');
    const tt = document.getElementById('tt');

    function pct(v){ return (v / MAX * 100).toFixed(4) + '%'; }

		var curPhase = null;
    lanes.forEach(function(lane){
			if(lane.phase !== curPhase) curPhase = lane.phase;
      // 親行
      var pr = document.createElement('div');
			pr.className = lane.isPhase ? 'gantt-row is-phase-row' : 'gantt-row';

      var pl = document.createElement('div');
			pl.className = lane.isPhase ? 'gantt-label is-phase' : 'gantt-label is-parent';
      pl.textContent = lane.label;
      pl.title = lane.label;
      pr.appendChild(pl);

      var pbc = document.createElement('div');
      pbc.className = 'gantt-bar-cell';
      var pbg = document.createElement('div');
      pbg.className = 'bar-bg';
      var pbr = document.createElement('div');
      pbr.className = 'bar-parent';
			pbr.style.left = pct(lane.start);
      pbr.style.width = pct(lane.total);
      pbg.appendChild(pbr);
      pbc.appendChild(pbg);
      pr.appendChild(pbc);
      gantt.appendChild(pr);

      // チケット行
      lane.tickets.forEach(function(tk){
        var row = document.createElement('div');
        row.className = 'gantt-row';

        var lbl = document.createElement('div');
        lbl.className = 'gantt-label is-ticket';
        lbl.textContent = '  ' + tk.title;
        lbl.addEventListener('mouseenter', function(e){
          tt.innerHTML =
            '<div class="tt-title">' + esc(tk.title) + '</div>' +
            '<div class="tt-row"><span class="tt-key">チケット</span><span class="tt-val">' + esc(tk.id) + '</span></div>' +
            '<div class="tt-row"><span class="tt-key">状態</span><span class="tt-val">' + esc(tk.status || '-') + '</span></div>' +
            '<div class="tt-row"><span class="tt-key">Estimate</span><span class="tt-val">' + esc(tk.est || '-') + '</span></div>' +
            '<div class="tt-row"><span class="tt-key">難易度</span><span class="tt-val">' + tk.diff + '</span></div>';
          tt.style.display = 'block';
          moveTT(e);
        });
        lbl.addEventListener('mousemove', moveTT);
        lbl.addEventListener('mouseleave', function(){ tt.style.display = 'none'; });
        row.appendChild(lbl);

        var bc = document.createElement('div');
        bc.className = 'gantt-bar-cell';
        var bg = document.createElement('div');
        bg.className = 'bar-bg';

        var bar = document.createElement('a');
        bar.className = 'bar-ticket c' + ((tk.order - 1) % 8 + 1);
        bar.href = tk.url;
        bar.target = '_blank';
        bar.rel = 'noopener';
        bar.style.left  = pct(tk.start);
        bar.style.width = pct(tk.end - tk.start);
        bar.textContent = tk.code;

        // ツールチップ
        bar.addEventListener('mouseenter', function(e){
          tt.innerHTML =
            '<div class="tt-title">' + esc(tk.title) + '</div>' +
            '<div class="tt-row"><span class="tt-key">チケット</span><span class="tt-val">' + esc(tk.id) + '</span></div>' +
						'<div class="tt-row"><span class="tt-key">状態</span><span class="tt-val">' + esc(tk.status || '-') + '</span></div>' +
            '<div class="tt-row"><span class="tt-key">Estimate</span><span class="tt-val">' + esc(tk.est || '-') + '</span></div>' +
            '<div class="tt-row"><span class="tt-key">難易度</span><span class="tt-val">' + tk.diff + '</span></div>' +
            '<div class="tt-row"><span class="tt-key">範囲</span><span class="tt-val">' + tk.start + ' - ' + tk.end + '</span></div>' +
            '<div class="tt-row"><span class="tt-key">依存</span><span class="tt-val">' + esc(tk.dep || '-') + '</span></div>' +
            '<div class="tt-row"><span class="tt-key">実行順</span><span class="tt-val">' + tk.order + '</span></div>';
          tt.style.display = 'block';
          moveTT(e);
        });
        bar.addEventListener('mousemove', moveTT);
        bar.addEventListener('mouseleave', function(){ tt.style.display = 'none'; });

        bg.appendChild(bar);
        bc.appendChild(bg);
        row.appendChild(bc);
        gantt.appendChild(row);
      });
    });

		// 各行ではなく、ガント全体を跨ぐ1本の進捗線を描画
		// lane.completed == lane.start は「完了なし」を意味するため除外する
		var globalCompleted = 0;
		lanes.forEach(function(lane){
			if(!lane.isPhase && lane.completed > lane.start && lane.completed > globalCompleted) globalCompleted = lane.completed;
		});
		renderGlobalProgressLine(globalCompleted);

    // X軸目盛り
    var xaxis = document.getElementById('xaxis');
    var step = MAX <= 20 ? 5 : MAX <= 50 ? 10 : 25;
    for(var v = 0; v <= MAX; v += step){
      var tick = document.createElement('div');
      tick.className = 'x-tick';
      tick.style.left = pct(v);
      tick.textContent = v;
      xaxis.appendChild(tick);
    }

    function moveTT(e){
      var x = e.clientX + 14, y = e.clientY + 14;
      if(x + 330 > window.innerWidth) x = e.clientX - 330;
      if(y + 160 > window.innerHeight) y = e.clientY - 160;
      tt.style.left = x + 'px';
      tt.style.top  = y + 'px';
    }
		function renderGlobalProgressLine(completed){
			var bars = gantt.querySelectorAll('.bar-bg');
			if(!bars.length) return;

			var firstBar = bars[0];
			var lastBar = bars[bars.length - 1];
			var line = document.createElement('div');
			line.className = 'progress-line-global';
			gantt.appendChild(line);

			function positionLine(){
				var pos = completed > 0 ? completed : 0;
				var ganttRect = gantt.getBoundingClientRect();
				var firstRect = firstBar.getBoundingClientRect();
				var lastRect = lastBar.getBoundingClientRect();

				var x = (firstRect.left - ganttRect.left) + (firstRect.width * pos / MAX);
				var top = firstRect.top - ganttRect.top - 1;
				var height = (lastRect.bottom - firstRect.top) + 2;

				line.style.left = (x - 1.5) + 'px';
				line.style.top = top + 'px';
				line.style.height = height + 'px';
			}

			positionLine();
			window.addEventListener('resize', positionLine);
		}
    function esc(s){
      return String(s).replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;');
    }
  })();
  </script>
</body>
</html>
`
