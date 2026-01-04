package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/markmilligan/goth-hw/internal/salesforce"
	"github.com/markmilligan/goth-hw/templates"
)

const defaultPageSize = 25

type ObjectOptions struct {
	NewLeadsOnly      bool
	IncludeClosedLost bool
	IncludeClosedWon  bool
	SortField         string
	SortDir           string
}

var activitySubjects = []string{
	"trial-signup",
	"book-a-demo",
	"contact-sales",
	"newsletter-signup",
}

type Handlers struct {
	sfClient *salesforce.Client
}

func New(sfClient *salesforce.Client) *Handlers {
	return &Handlers{sfClient: sfClient}
}

func (h *Handlers) Home(w http.ResponseWriter, r *http.Request) {
	templates.Home().Render(r.Context(), w)
}

func (h *Handlers) ObjectPage(w http.ResponseWriter, r *http.Request) {
	objTypeStr := r.PathValue("type")
	objType, valid := salesforce.ValidObjectType(objTypeStr)
	if !valid {
		http.Error(w, "Invalid object type", http.StatusBadRequest)
		return
	}

	page := parseIntOrDefault(r.URL.Query().Get("page"), 0)
	filters := h.extractFilters(r, objType)
	opts := h.extractObjectOptions(r, objType)

	data, err := h.fetchObjectData(objType, filters, page, opts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Check if this is an HTMX request for just the table
	if r.Header.Get("HX-Request") == "true" && r.Header.Get("HX-Target") == "object-table-container" {
		templates.ObjectTable(data).Render(r.Context(), w)
		return
	}

	templates.ObjectPage(data).Render(r.Context(), w)
}

func (h *Handlers) ObjectRows(w http.ResponseWriter, r *http.Request) {
	objTypeStr := r.PathValue("type")
	objType, valid := salesforce.ValidObjectType(objTypeStr)
	if !valid {
		http.Error(w, "Invalid object type", http.StatusBadRequest)
		return
	}

	filters := h.extractFilters(r, objType)
	opts := h.extractObjectOptions(r, objType)

	data, err := h.fetchObjectData(objType, filters, 0, opts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	templates.ObjectRows(data).Render(r.Context(), w)
}

func (h *Handlers) extractFilters(r *http.Request, objType salesforce.ObjectType) map[string]string {
	filters := make(map[string]string)
	fields := salesforce.ObjectFields[objType]

	for _, field := range fields {
		if val := r.URL.Query().Get(field.Name); val != "" {
			filters[field.Name] = val
		}
	}

	return filters
}

func (h *Handlers) extractObjectOptions(r *http.Request, objType salesforce.ObjectType) ObjectOptions {
	isHTMX := r.Header.Get("HX-Request") == "true"

	opts := ObjectOptions{
		SortField: r.URL.Query().Get("sortField"),
		SortDir:   r.URL.Query().Get("sortDir"),
	}

	// Set defaults based on object type
	if opts.SortField == "" {
		switch objType {
		case salesforce.ObjectOpportunity:
			opts.SortField = "CloseDate"
			opts.SortDir = "ASC"
		case salesforce.ObjectAccount:
			opts.SortField = "Name"
			opts.SortDir = "ASC"
		case salesforce.ObjectContact:
			opts.SortField = "FirstName"
			opts.SortDir = "ASC"
		default:
			opts.SortField = "CreatedDate"
			opts.SortDir = "DESC"
		}
	}
	if opts.SortDir == "" {
		opts.SortDir = "DESC"
	}

	// Lead-specific options
	if objType == salesforce.ObjectLead {
		if isHTMX {
			opts.NewLeadsOnly = r.URL.Query().Get("newLeadsOnly") == "on"
		} else {
			opts.NewLeadsOnly = true
			if r.URL.Query().Has("newLeadsOnly") {
				opts.NewLeadsOnly = r.URL.Query().Get("newLeadsOnly") == "on"
			}
		}
	}

	// Opportunity-specific options (exclude closed by default)
	if objType == salesforce.ObjectOpportunity {
		if isHTMX {
			opts.IncludeClosedLost = r.URL.Query().Get("includeClosedLost") == "on"
			opts.IncludeClosedWon = r.URL.Query().Get("includeClosedWon") == "on"
		} else {
			// Default to false (exclude) unless explicitly set
			opts.IncludeClosedLost = r.URL.Query().Get("includeClosedLost") == "on"
			opts.IncludeClosedWon = r.URL.Query().Get("includeClosedWon") == "on"
		}
	}

	return opts
}

func (h *Handlers) fetchObjectData(objType salesforce.ObjectType, filters map[string]string, page int, opts ObjectOptions) (templates.ObjectPageData, error) {
	soql := salesforce.BuildSOQL(objType, filters, defaultPageSize, page*defaultPageSize, opts.SortField, opts.SortDir, opts.NewLeadsOnly, opts.IncludeClosedLost, opts.IncludeClosedWon)

	result, err := h.sfClient.Query(soql)
	if err != nil {
		return templates.ObjectPageData{}, err
	}

	return templates.ObjectPageData{
		ObjectType:        objType,
		Fields:            salesforce.ObjectFields[objType],
		Records:           result.Records,
		Filters:           filters,
		Page:              page,
		PageSize:          defaultPageSize,
		TotalSize:         result.TotalSize,
		InstanceURL:       h.sfClient.InstanceURL(),
		NewLeadsOnly:      opts.NewLeadsOnly,
		SortField:         opts.SortField,
		SortDir:           opts.SortDir,
		IncludeClosedLost: opts.IncludeClosedLost,
		IncludeClosedWon:  opts.IncludeClosedWon,
	}, nil
}

func (h *Handlers) ActivityPage(w http.ResponseWriter, r *http.Request) {
	page := parseIntOrDefault(r.URL.Query().Get("page"), 0)
	
	// Default to true (exclude nuon.co), only false if explicitly unchecked
	excludeNuonCo := true
	if r.URL.Query().Has("excludeNuonCo") {
		excludeNuonCo = r.URL.Query().Get("excludeNuonCo") == "on"
	}

	newLeadsOnly := r.URL.Query().Get("newLeadsOnly") == "on"
	allStatusTypes := r.URL.Query().Get("allStatusTypes") == "on"

	// Default CTAs to true on initial load, respect HTMX requests
	ctasOnly := true
	if r.Header.Get("HX-Request") == "true" {
		ctasOnly = r.URL.Query().Get("ctasOnly") == "on"
	} else if r.URL.Query().Has("ctasOnly") {
		ctasOnly = r.URL.Query().Get("ctasOnly") == "on"
	}

	filters := templates.ActivityFilters{
		FirstName: r.URL.Query().Get("filterFirstName"),
		LastName:  r.URL.Query().Get("filterLastName"),
		Email:     r.URL.Query().Get("filterEmail"),
		Company:   r.URL.Query().Get("filterCompany"),
		Subject:   r.URL.Query().Get("filterSubject"),
	}

	data, err := h.fetchActivityData(page, excludeNuonCo, newLeadsOnly, allStatusTypes, ctasOnly, filters)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		templates.ActivityTable(data).Render(r.Context(), w)
		return
	}

	templates.ActivityPage(data).Render(r.Context(), w)
}

func (h *Handlers) ActivityRows(w http.ResponseWriter, r *http.Request) {
	// Hidden inputs send "on" or "" as values
	excludeNuonCo := r.URL.Query().Get("excludeNuonCo") == "on"
	newLeadsOnly := r.URL.Query().Get("newLeadsOnly") == "on"
	allStatusTypes := r.URL.Query().Get("allStatusTypes") == "on"
	ctasOnly := r.URL.Query().Get("ctasOnly") == "on"

	filters := templates.ActivityFilters{
		FirstName: r.URL.Query().Get("filterFirstName"),
		LastName:  r.URL.Query().Get("filterLastName"),
		Email:     r.URL.Query().Get("filterEmail"),
		Company:   r.URL.Query().Get("filterCompany"),
		Subject:   r.URL.Query().Get("filterSubject"),
	}

	data, err := h.fetchActivityData(0, excludeNuonCo, newLeadsOnly, allStatusTypes, ctasOnly, filters)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	templates.ActivityRows(data).Render(r.Context(), w)
}

func (h *Handlers) fetchActivityData(page int, excludeNuonCo, newLeadsOnly, allStatusTypes, ctasOnly bool, filters templates.ActivityFilters) (templates.ActivityPageData, error) {
	var whereParts []string
	whereParts = append(whereParts, "Who.Type = 'Lead'")

	if ctasOnly {
		quoted := make([]string, len(activitySubjects))
		for i, s := range activitySubjects {
			quoted[i] = fmt.Sprintf("'%s'", s)
		}
		whereParts = append(whereParts, fmt.Sprintf("Subject IN (%s)", strings.Join(quoted, ", ")))
	}

	// Add Subject text filter to SOQL (it's a Task field, not polymorphic)
	if filters.Subject != "" {
		whereParts = append(whereParts, fmt.Sprintf("Subject LIKE '%%%s%%'", escapeSOQL(filters.Subject)))
	}

	whereClause := strings.Join(whereParts, " AND ")

	// Fetch all matching records for client-side filtering and pagination
	fetchLimit := 2000
	soql := fmt.Sprintf(`SELECT Id, CreatedDate, Subject, 
		TYPEOF Who WHEN Lead THEN FirstName, LastName, Email, Company, Status END
		FROM Task 
		WHERE %s
		ORDER BY CreatedDate DESC 
		LIMIT %d`,
		whereClause,
		fetchLimit,
	)

	result, err := h.sfClient.Query(soql)
	if err != nil {
		return templates.ActivityPageData{}, err
	}



	// First, apply all client-side filters and collect all matching records
	allFiltered := make([]templates.ActivityRecord, 0)
	for _, rec := range result.Records {
		who, _ := rec["Who"].(map[string]any)
		email := getStrNested(who, "Email")
		status := getStrNested(who, "Status")
		firstName := getStrNested(who, "FirstName")
		lastName := getStrNested(who, "LastName")
		company := getStrNested(who, "Company")
		subject := getStr(rec, "Subject")

		// Client-side filter for nuon.co
		if excludeNuonCo && strings.Contains(strings.ToLower(email), "nuon.co") {
			continue
		}

		// Client-side status filtering
		if newLeadsOnly && status != "New" {
			continue
		}
		if !newLeadsOnly && !allStatusTypes && status == "Unqualified" {
			continue
		}

		// Client-side text filters
		if filters.FirstName != "" && !strings.Contains(strings.ToLower(firstName), strings.ToLower(filters.FirstName)) {
			continue
		}
		if filters.LastName != "" && !strings.Contains(strings.ToLower(lastName), strings.ToLower(filters.LastName)) {
			continue
		}
		if filters.Email != "" && !strings.Contains(strings.ToLower(email), strings.ToLower(filters.Email)) {
			continue
		}
		if filters.Company != "" && !strings.Contains(strings.ToLower(company), strings.ToLower(filters.Company)) {
			continue
		}

		allFiltered = append(allFiltered, templates.ActivityRecord{
			ID:          getStr(rec, "Id"),
			CreatedDate: formatCreatedDate(getStr(rec, "CreatedDate")),
			Subject:     subject,
			FirstName:   firstName,
			LastName:    lastName,
			Email:       email,
			Company:     company,
		})
	}

	// Now paginate the filtered results
	totalFiltered := len(allFiltered)
	start := page * defaultPageSize
	end := start + defaultPageSize
	if start > totalFiltered {
		start = totalFiltered
	}
	if end > totalFiltered {
		end = totalFiltered
	}
	records := allFiltered[start:end]

	return templates.ActivityPageData{
		Records:        records,
		Page:           page,
		PageSize:       defaultPageSize,
		TotalSize:      totalFiltered,
		InstanceURL:    h.sfClient.InstanceURL(),
		ExcludeNuonCo:  excludeNuonCo,
		NewLeadsOnly:   newLeadsOnly,
		AllStatusTypes: allStatusTypes,
		CTAsOnly:       ctasOnly,
		Filters:        filters,
	}, nil
}

func escapeSOQL(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "'", "\\'")
	return s
}

func formatCreatedDate(isoDate string) string {
	// Salesforce returns format like "2024-01-15T14:30:00.000+0000"
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05.000+0000",
		"2006-01-02T15:04:05.000Z",
		"2006-01-02T15:04:05Z",
	}
	
	for _, format := range formats {
		if t, err := time.Parse(format, isoDate); err == nil {
			return t.Format("Jan 2, 2006 3:04 PM")
		}
	}
	return isoDate
}

func getStr(m map[string]any, key string) string {
	if v, ok := m[key]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}

func getStrNested(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	return getStr(m, key)
}

func parseIntOrDefault(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return val
}
