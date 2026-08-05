package database

import (
	"database/sql"
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
	where, args := activityWhere(query)
	sortColumn := map[string]string{"score": "score", "year": "year"}[query.Sort]
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

func (d *Database) GetRejectionBreakdown() (int, map[string]int, error) {
	rows, err := d.db.Query(`SELECT CASE
		WHEN instr(lower(COALESCE(message, '')), 'certification') > 0 THEN 'Content Rating'
		WHEN instr(lower(COALESCE(message, '')), 'rating') > 0 THEN 'Low Rating'
		WHEN instr(lower(COALESCE(message, '')), 'country') > 0 THEN 'Wrong Country'
		WHEN instr(lower(COALESCE(message, '')), 'language') > 0 THEN 'Wrong Language'
		WHEN instr(lower(COALESCE(message, '')), 'genre') > 0 THEN 'Blacklisted Genre'
		WHEN instr(lower(COALESCE(message, '')), 'keyword') > 0 THEN 'Blacklisted Keyword'
		WHEN instr(lower(COALESCE(message, '')), 'runtime') > 0 THEN 'Runtime Out of Range'
		WHEN instr(lower(COALESCE(message, '')), 'year') > 0 THEN 'Year Out of Range'
		WHEN instr(lower(COALESCE(message, '')), 'votes') > 0 THEN 'Insufficient Votes'
		WHEN instr(lower(COALESCE(message, '')), 'network') > 0 THEN 'Blacklisted Network'
		ELSE 'Other' END AS reason, COUNT(*)
		FROM activity_logs WHERE status = ? GROUP BY reason`, enums.ActivityStatusRejected)
	if err != nil {
		return 0, nil, err
	}
	defer func() { _ = rows.Close() }()
	total, breakdown := 0, map[string]int{}
	for rows.Next() {
		var reason string
		var count int
		if err := rows.Scan(&reason, &count); err != nil {
			return 0, nil, err
		}
		breakdown[reason], total = count, total+count
	}
	return total, breakdown, rows.Err()
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
