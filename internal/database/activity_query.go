package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/mahcks/blockbusterr/pkg/enums"
)

type ActivityQuery struct {
	Status    string
	MediaType string
	Job       string
	Language  string
	Search    string
	Start     *time.Time
	End       *time.Time
	RunID     *int64
	Dedupe    bool
	Sort      string
	Order     string
	Limit     int
	Offset    int
	// Reason narrows rejected items to a specific rejectionReason() category
	// (e.g. "Minimum Year"), matching what the rejection breakdown widget
	// shows. It isn't a stored column, so it can't join the normal SQL WHERE
	// clause - see queryActivityByReason.
	Reason string
}

type ActivityLogGroup struct {
	Log     ActivityLog
	Count   int
	History []ActivityLog
	key     string
}

type ActivityPage struct {
	Groups []ActivityLogGroup
	Total  int
	Offset int
}

const activityColumns = "id, timestamp, run_id, job_id, job_type, media_type, title, language, year, tmdb_id, tvdb_id, imdb_id, poster_url, score, rank, status, message, filter_details"

const activityGroupKey = `media_type || '|' || CASE
	WHEN media_type = 'movie' AND COALESCE(tmdb_id, 0) > 0 THEN 'tmdb:' || tmdb_id
	WHEN media_type = 'show' AND COALESCE(tvdb_id, 0) > 0 THEN 'tvdb:' || tvdb_id
	WHEN media_type = 'show' AND COALESCE(tmdb_id, 0) > 0 THEN 'tmdb:' || tmdb_id
	ELSE lower(title) || ':' || COALESCE(year, 0)
	END || '|' || status || '|' || job_type || CASE WHEN COALESCE(run_id, 0) > 0 THEN '|run:' || run_id ELSE '' END`

func (d *Database) QueryActivity(query ActivityQuery) (ActivityPage, error) {
	if query.Limit <= 0 {
		query.Limit = 50
	}
	if query.Reason != "" {
		// Reason only ever applies to rejected items; force it regardless of
		// what status was also requested, rather than silently returning an
		// empty page for a contradictory combination.
		query.Status = string(enums.ActivityStatusRejected)
		return d.queryActivityByReason(query)
	}
	where, args := activityWhere(query)
	sortColumn := map[string]string{"score": "score", "year": "year", "title": "title COLLATE NOCASE"}[query.Sort]
	if sortColumn == "" {
		sortColumn = "timestamp"
	}
	order := "DESC"
	if strings.EqualFold(query.Order, "asc") {
		order = "ASC"
	}

	if !query.Dedupe {
		var total int
		if err := d.db.QueryRow("SELECT COUNT(*) FROM activity_logs"+where, args...).Scan(&total); err != nil {
			return ActivityPage{}, err
		}
		offset := boundedOffset(query.Offset, query.Limit, total)
		rows, err := d.db.Query("SELECT "+activityColumns+" FROM activity_logs"+where+" ORDER BY "+sortColumn+" "+order+", id "+order+" LIMIT ? OFFSET ?", append(args, query.Limit, offset)...)
		if err != nil {
			return ActivityPage{}, err
		}
		logs, err := scanActivityLogs(rows)
		if err != nil {
			return ActivityPage{}, err
		}
		groups := make([]ActivityLogGroup, 0, len(logs))
		for _, log := range logs {
			groups = append(groups, ActivityLogGroup{Log: log, Count: 1, History: []ActivityLog{log}})
		}
		return ActivityPage{Groups: groups, Total: total, Offset: offset}, nil
	}

	cte := "WITH filtered AS (SELECT " + activityColumns + ", " + activityGroupKey + " AS group_key FROM activity_logs" + where + ")"
	var total int
	if err := d.db.QueryRow(cte+" SELECT COUNT(*) FROM (SELECT 1 FROM filtered GROUP BY group_key)", args...).Scan(&total); err != nil {
		return ActivityPage{}, err
	}
	offset := boundedOffset(query.Offset, query.Limit, total)
	pageSQL := cte + `, ranked AS (
		SELECT *, ROW_NUMBER() OVER (PARTITION BY group_key ORDER BY timestamp DESC, id DESC) AS row_number,
			COUNT(*) OVER (PARTITION BY group_key) AS group_count
		FROM filtered
	) SELECT ` + activityColumns + `, group_key, group_count FROM ranked WHERE row_number = 1
		ORDER BY ` + sortColumn + ` ` + order + `, id ` + order + ` LIMIT ? OFFSET ?`
	rows, err := d.db.Query(pageSQL, append(args, query.Limit, offset)...)
	if err != nil {
		return ActivityPage{}, err
	}
	groups, err := scanActivityGroups(rows)
	if err != nil || len(groups) == 0 {
		return ActivityPage{Groups: groups, Total: total, Offset: offset}, err
	}

	keys := make([]string, len(groups))
	byKey := make(map[string]int, len(groups))
	for index := range groups {
		keys[index], byKey[groups[index].key] = groups[index].key, index
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(keys)), ",")
	historyArgs := append(append([]any{}, args...), stringsToAny(keys)...)
	historyRows, err := d.db.Query(cte+" SELECT "+activityColumns+", group_key FROM filtered WHERE group_key IN ("+placeholders+") ORDER BY timestamp DESC, id DESC", historyArgs...)
	if err != nil {
		return ActivityPage{}, err
	}
	defer func() { _ = historyRows.Close() }()
	for historyRows.Next() {
		log, key, err := scanActivityLogWithKey(historyRows)
		if err != nil {
			return ActivityPage{}, err
		}
		if index, ok := byKey[key]; ok {
			groups[index].History = append(groups[index].History, log)
		}
	}
	if err := historyRows.Err(); err != nil {
		return ActivityPage{}, err
	}
	return ActivityPage{Groups: groups, Total: total, Offset: offset}, nil
}

// maxReasonScan bounds how many rejected rows queryActivityByReason will pull
// into memory to classify. Rejection reasons aren't a stored column (see
// rejectionReason), so filtering by one can't be pushed into SQL; this caps
// the cost of doing it in Go instead for a self-hosted, single-tenant app.
const maxReasonScan = 20000

// queryActivityByReason mirrors QueryActivity's filtering/sorting/pagination/
// grouping semantics, but for a Reason that only exists once rows are
// classified in Go. It fetches the candidate rows ordered by recency (so
// group representative/history selection matches QueryActivity's "latest
// wins" rule exactly), classifies and filters them, then re-applies the
// requested sort to the resulting groups before paginating.
func (d *Database) queryActivityByReason(query ActivityQuery) (ActivityPage, error) {
	where, args := activityWhere(query)
	order := "DESC"
	if strings.EqualFold(query.Order, "asc") {
		order = "ASC"
	}

	rows, err := d.db.Query("SELECT "+activityColumns+" FROM activity_logs"+where+" ORDER BY timestamp DESC, id DESC LIMIT ?", append(append([]any{}, args...), maxReasonScan)...)
	if err != nil {
		return ActivityPage{}, err
	}
	logs, err := scanActivityLogs(rows)
	if err != nil {
		return ActivityPage{}, err
	}

	matched := make([]ActivityLog, 0, len(logs))
	for _, log := range logs {
		if rejectionReason(log.FilterDetails, log.Message) == query.Reason {
			matched = append(matched, log)
		}
	}

	if !query.Dedupe {
		total := len(matched)
		offset := boundedOffset(query.Offset, query.Limit, total)
		end := min(offset+query.Limit, total)
		page := matched[offset:end]
		groups := make([]ActivityLogGroup, 0, len(page))
		for _, log := range page {
			groups = append(groups, ActivityLogGroup{Log: log, Count: 1, History: []ActivityLog{log}})
		}
		return ActivityPage{Groups: groups, Total: total, Offset: offset}, nil
	}

	// matched is already timestamp DESC, so the first row seen per key is the
	// group representative and history accumulates in the right order - both
	// match what the SQL grouping path (ROW_NUMBER partitioned by group_key
	// ordered by timestamp DESC) does for QueryActivity.
	byKey := map[string]*ActivityLogGroup{}
	groups := make([]*ActivityLogGroup, 0, len(matched))
	for _, log := range matched {
		key := activityGroupKeyFor(log)
		group, ok := byKey[key]
		if !ok {
			group = &ActivityLogGroup{Log: log, key: key}
			byKey[key] = group
			groups = append(groups, group)
		}
		group.Count++
		group.History = append(group.History, log)
	}

	sortGroups(groups, query.Sort, order)

	total := len(groups)
	offset := boundedOffset(query.Offset, query.Limit, total)
	end := min(offset+query.Limit, total)
	page := make([]ActivityLogGroup, end-offset)
	for i, group := range groups[offset:end] {
		page[i] = *group
	}
	return ActivityPage{Groups: page, Total: total, Offset: offset}, nil
}

func activityGroupKeyFor(log ActivityLog) string {
	var idPart string
	switch {
	case log.MediaType == "movie" && log.TMDBID > 0:
		idPart = fmt.Sprintf("tmdb:%d", log.TMDBID)
	case log.MediaType == "show" && log.TVDBID > 0:
		idPart = fmt.Sprintf("tvdb:%d", log.TVDBID)
	case log.MediaType == "show" && log.TMDBID > 0:
		idPart = fmt.Sprintf("tmdb:%d", log.TMDBID)
	default:
		idPart = fmt.Sprintf("%s:%d", strings.ToLower(log.Title), log.Year)
	}
	key := log.MediaType + "|" + idPart + "|" + log.Status + "|" + log.JobType
	if log.RunID > 0 {
		key += fmt.Sprintf("|run:%d", log.RunID)
	}
	return key
}

// sortGroups re-orders group representatives in place. Groups arrive in
// timestamp-DESC order (see queryActivityByReason); for the "timestamp" sort
// that's already correct for DESC and just needs reversing for ASC.
func sortGroups(groups []*ActivityLogGroup, field, order string) {
	asc := order == "ASC"
	less := func(i, j int) bool { return false }
	switch field {
	case "score":
		less = func(i, j int) bool { return groups[i].Log.Score < groups[j].Log.Score }
	case "year":
		less = func(i, j int) bool { return groups[i].Log.Year < groups[j].Log.Year }
	case "title":
		less = func(i, j int) bool {
			return strings.ToLower(groups[i].Log.Title) < strings.ToLower(groups[j].Log.Title)
		}
	default:
		if asc {
			for l, r := 0, len(groups)-1; l < r; l, r = l+1, r-1 {
				groups[l], groups[r] = groups[r], groups[l]
			}
		}
		return
	}
	sort.SliceStable(groups, func(i, j int) bool {
		if asc {
			return less(i, j)
		}
		return less(j, i)
	})
}

// rejectionFilterCheck mirrors jobs.FilterCheck's JSON shape without importing
// the jobs package, which itself imports database (would be a cycle).
type rejectionFilterCheck struct {
	Name   string `json:"name"`
	Passed bool   `json:"passed"`
}

// GetRejectionBreakdown tallies why rejected titles failed. It prefers the
// exact filter recorded in filter_details at evaluation time over guessing
// from the free-text message, which is fragile and can misclassify a title
// whose own text happens to contain a keyword like "rating" or "genre".
// Only entries logged before filter_details existed fall back to the message.
func (d *Database) GetRejectionBreakdown() (int, map[string]int, error) {
	rows, err := d.db.Query(`SELECT COALESCE(filter_details, ''), COALESCE(message, '') FROM activity_logs WHERE status = ?`, enums.ActivityStatusRejected)
	if err != nil {
		return 0, nil, err
	}
	defer func() { _ = rows.Close() }()

	total, breakdown := 0, map[string]int{}
	for rows.Next() {
		var filterDetails, message string
		if err := rows.Scan(&filterDetails, &message); err != nil {
			return 0, nil, err
		}
		breakdown[rejectionReason(filterDetails, message)]++
		total++
	}
	return total, breakdown, rows.Err()
}

func rejectionReason(filterDetails, message string) string {
	if filterDetails != "" {
		var checks []rejectionFilterCheck
		if err := json.Unmarshal([]byte(filterDetails), &checks); err == nil {
			for _, check := range checks {
				if !check.Passed && check.Name != "" {
					return humanizeReasonName(check.Name)
				}
			}
		}
	}
	return rejectionReasonFromMessage(message)
}

// humanizeReasonName presents raw filter check identifiers (e.g.
// "blacklisted_genres", used by some job executors) the same way as the
// human-written names most filters already use (e.g. "Blocked genre").
func humanizeReasonName(name string) string {
	if !strings.Contains(name, "_") {
		return name
	}
	words := strings.Split(name, "_")
	for i, word := range words {
		if word == "" {
			continue
		}
		words[i] = strings.ToUpper(word[:1]) + word[1:]
	}
	return strings.Join(words, " ")
}

func rejectionReasonFromMessage(message string) string {
	lower := strings.ToLower(message)
	switch {
	case strings.Contains(lower, "certification"):
		return "Content certification"
	case strings.Contains(lower, "rating"):
		return "Minimum Rating"
	case strings.Contains(lower, "country"):
		return "Allowed Countries"
	case strings.Contains(lower, "language"):
		return "Allowed Languages"
	case strings.Contains(lower, "genre"):
		return "Blacklisted Genres"
	case strings.Contains(lower, "keyword"):
		return "Blacklisted Keywords"
	case strings.Contains(lower, "runtime"):
		return "Runtime"
	case strings.Contains(lower, "year"):
		return "Release Year"
	case strings.Contains(lower, "votes"):
		return "Minimum Votes"
	case strings.Contains(lower, "network"):
		return "Blacklisted Networks"
	default:
		return "Other"
	}
}

func activityWhere(query ActivityQuery) (string, []any) {
	clauses := []string{}
	args := []any{}
	add := func(clause string, values ...any) {
		clauses, args = append(clauses, clause), append(args, values...)
	}
	if query.Status != "" {
		add("status = ?", query.Status)
	}
	if query.MediaType != "" {
		add("media_type = ?", query.MediaType)
	}
	if query.Job != "" {
		add("(job_id = ? OR job_type = ?)", query.Job, query.Job)
	}
	if query.Language != "" {
		add("language = ?", query.Language)
	}
	if query.Search != "" {
		add("instr(lower(title), lower(?)) > 0", query.Search)
	}
	if query.Start != nil {
		add("timestamp >= ?", *query.Start)
	}
	if query.End != nil {
		add("timestamp < ?", *query.End)
	}
	if query.RunID != nil {
		add("run_id = ?", *query.RunID)
	}
	if len(clauses) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(clauses, " AND "), args
}

func boundedOffset(offset, limit, total int) int {
	if offset < 0 || total == 0 {
		return 0
	}
	if offset >= total {
		return ((total - 1) / limit) * limit
	}
	return offset
}

func stringsToAny(values []string) []any {
	result := make([]any, len(values))
	for index := range values {
		result[index] = values[index]
	}
	return result
}

func scanActivityLogs(rows *sql.Rows) ([]ActivityLog, error) {
	defer func() { _ = rows.Close() }()
	logs := []ActivityLog{}
	for rows.Next() {
		log, err := scanActivityLog(rows)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	return logs, rows.Err()
}

func scanActivityGroups(rows *sql.Rows) ([]ActivityLogGroup, error) {
	defer func() { _ = rows.Close() }()
	groups := []ActivityLogGroup{}
	for rows.Next() {
		log, key, count, err := scanActivityGroup(rows)
		if err != nil {
			return nil, err
		}
		groups = append(groups, ActivityLogGroup{Log: log, Count: count, key: key, History: []ActivityLog{}})
	}
	return groups, rows.Err()
}

type rowScanner interface{ Scan(dest ...any) error }

func scanActivityLog(row rowScanner) (ActivityLog, error) {
	log, _, err := scanActivity(row, false)
	return log, err
}

func scanActivityLogWithKey(row rowScanner) (ActivityLog, string, error) {
	return scanActivity(row, true)
}

func scanActivityGroup(row rowScanner) (ActivityLog, string, int, error) {
	log, values, destinations := activityScanTargets()
	var key string
	var count int
	destinations = append(destinations, &key, &count)
	if err := row.Scan(destinations...); err != nil {
		return ActivityLog{}, "", 0, err
	}
	applyActivityScan(log, values)
	return *log, key, count, nil
}

func scanActivity(row rowScanner, withKey bool) (ActivityLog, string, error) {
	log, values, destinations := activityScanTargets()
	var key string
	if withKey {
		destinations = append(destinations, &key)
	}
	if err := row.Scan(destinations...); err != nil {
		return ActivityLog{}, "", err
	}
	applyActivityScan(log, values)
	return *log, key, nil
}

type activityNullableValues struct {
	runID, tmdbID, tvdbID, rank                                sql.NullInt64
	jobID, language, imdbID, posterURL, message, filterDetails sql.NullString
	score                                                      sql.NullFloat64
}

func activityScanTargets() (*ActivityLog, *activityNullableValues, []any) {
	log := &ActivityLog{}
	values := &activityNullableValues{}
	destinations := []any{&log.ID, &log.Timestamp, &values.runID, &values.jobID, &log.JobType, &log.MediaType, &log.Title, &values.language, &log.Year, &values.tmdbID, &values.tvdbID, &values.imdbID, &values.posterURL, &values.score, &values.rank, &log.Status, &values.message, &values.filterDetails}
	return log, values, destinations
}

func applyActivityScan(log *ActivityLog, values *activityNullableValues) {
	if values.runID.Valid {
		log.RunID = values.runID.Int64
	}
	if values.jobID.Valid {
		log.JobID = values.jobID.String
	}
	if values.language.Valid {
		log.Language = values.language.String
	}
	if values.tmdbID.Valid {
		log.TMDBID = int(values.tmdbID.Int64)
	}
	if values.tvdbID.Valid {
		log.TVDBID = int(values.tvdbID.Int64)
	}
	if values.imdbID.Valid {
		log.IMDBID = values.imdbID.String
	}
	if values.posterURL.Valid {
		log.PosterURL = values.posterURL.String
	}
	if values.score.Valid {
		log.Score = values.score.Float64
	}
	if values.rank.Valid {
		log.Rank = int(values.rank.Int64)
	}
	if values.message.Valid {
		log.Message = values.message.String
	}
	if values.filterDetails.Valid {
		log.FilterDetails = values.filterDetails.String
	}
}
