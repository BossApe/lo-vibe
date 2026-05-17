package roadmap

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

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
		diff := storyPointDifficulty(fm["Story Point"], est)
		rows = append(rows, issueRow{
			ID:         fmt.Sprintf("#%d", content.Number),
			Number:     content.Number,
			Title:      content.Title,
			URL:        content.URL,
			Type:       fallback(fm["Type"], "Unknown"),
			Phase:      fallback(fm["Phase"], "Unknown"),
			Sprint:     fallback(fm["Sprint"], "Unknown"),
			Parent:     fallback(fm["Parent"], "No Parent"),
			DependsOn:  fm["Depends on"],
			Status:     fm["Status"],
			Estimate:   fallback(est, "N/A"),
			Difficulty: diff,
		})
	}
	return rows
}

// storyPointDifficulty は Story Point フィールドが正の整数として設定されている場合
// その値を難易度として返す。未設定・非数値の場合は Estimate のマッピングにフォールバックする。
func storyPointDifficulty(spText, est string) int {
	if v := strings.TrimSpace(spText); v != "" {
		if sp, err := strconv.Atoi(v); err == nil && sp > 0 {
			return sp
		}
	}
	diff, ok := estimateScore[est]
	if !ok {
		return 1
	}
	return diff
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
		case "Phase", "Sprint":
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
			parentMap[key] = issueRow{Title: key + " parent lane", Parent: key, Phase: "Unknown", Sprint: "Unknown", Difficulty: 0}
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

	// phaseKeys は将来の拡張用に保持（現在は未使用）
	_ = phaseKeys

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

func extractTicketOrder(title string, fallbackVal int) int {
	m := ticketOrderPattern.FindStringSubmatch(title)
	if len(m) < 2 {
		return fallbackVal
	}
	var n int
	_, err := fmt.Sscanf(m[1], "%d", &n)
	if err != nil {
		return fallbackVal
	}
	return n
}

func extractTicketCode(title, fallbackVal string) string {
	m := ticketCodePattern.FindStringSubmatch(title)
	if len(m) < 2 {
		return fallbackVal
	}
	return m[1]
}

func extractParentKey(r issueRow) string {
	title := r.Title
	if r.Type == "Sprint" {
		if m := spPattern.FindStringSubmatch(title); len(m) > 0 {
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
	if m := spPattern.FindStringSubmatch(parentKey); len(m) == 3 {
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
