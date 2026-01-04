package salesforce

import (
	"fmt"
	"strings"
)

type ObjectType string

const (
	ObjectAccount     ObjectType = "Account"
	ObjectContact     ObjectType = "Contact"
	ObjectOpportunity ObjectType = "Opportunity"
	ObjectLead        ObjectType = "Lead"
)

type ObjectField struct {
	Name     string
	Label    string
	Multiline bool
}

var ObjectFields = map[ObjectType][]ObjectField{
	ObjectAccount: {
		{Name: "Name", Label: "Name"},
		{Name: "Industry", Label: "Industry"},
		{Name: "Type", Label: "Type"},
		{Name: "Website", Label: "Website"},
		{Name: "Description", Label: "Description", Multiline: true},
	},
	ObjectContact: {
		{Name: "FirstName", Label: "First Name"},
		{Name: "LastName", Label: "Last Name"},
		{Name: "Email", Label: "Email"},
		{Name: "Title", Label: "Title"},
		{Name: "Description", Label: "Description", Multiline: true},
		{Name: "Account.Name", Label: "Account"},
	},
	ObjectOpportunity: {
		{Name: "Name", Label: "Name"},
		{Name: "StageName", Label: "Stage"},
		{Name: "Amount", Label: "Amount"},
		{Name: "CloseDate", Label: "Close Date"},
		{Name: "Type", Label: "Type"},
		{Name: "Probability", Label: "Probability"},
		{Name: "NextStep", Label: "Next Step"},
		{Name: "LeadSource", Label: "Lead Source"},
		{Name: "Description", Label: "Description", Multiline: true},
	},
	ObjectLead: {
		{Name: "FirstName", Label: "First Name"},
		{Name: "LastName", Label: "Last Name"},
		{Name: "Email", Label: "Email"},
		{Name: "Company", Label: "Company"},
		{Name: "Status", Label: "Status"},
	},
}

func BuildSOQL(objType ObjectType, filters map[string]string, limit, offset int, sortField, sortDir string, newLeadsOnly, includeClosedLost, includeClosedWon bool) string {
	fields := ObjectFields[objType]
	fieldNames := make([]string, 0, len(fields)+1)
	fieldNames = append(fieldNames, "Id")
	for _, f := range fields {
		fieldNames = append(fieldNames, f.Name)
	}

	query := fmt.Sprintf("SELECT %s FROM %s", strings.Join(fieldNames, ", "), objType)

	var whereClauses []string
	for field, value := range filters {
		if value != "" {
			whereClauses = append(whereClauses, fmt.Sprintf("%s LIKE '%%%s%%'", field, escapeSOQL(value)))
		}
	}

	// For Leads, add status filter
	if objType == ObjectLead && newLeadsOnly {
		whereClauses = append(whereClauses, "Status = 'New'")
	}

	// For Opportunities, exclude closed stages unless explicitly included
	if objType == ObjectOpportunity {
		var excludedStages []string
		if !includeClosedLost {
			excludedStages = append(excludedStages, "'Closed Lost'")
		}
		if !includeClosedWon {
			excludedStages = append(excludedStages, "'Closed Won'")
		}
		if len(excludedStages) > 0 {
			whereClauses = append(whereClauses, fmt.Sprintf("StageName NOT IN (%s)", strings.Join(excludedStages, ", ")))
		}
	}

	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Default sort
	if sortField == "" {
		sortField = "CreatedDate"
	}
	if sortDir == "" {
		sortDir = "DESC"
	}
	query += fmt.Sprintf(" ORDER BY %s %s", sortField, sortDir)
	query += fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)

	return query
}

func escapeSOQL(s string) string {
	s = strings.ReplaceAll(s, "'", "\\'")
	s = strings.ReplaceAll(s, "\\", "\\\\")
	return s
}

func ValidObjectType(s string) (ObjectType, bool) {
	switch ObjectType(s) {
	case ObjectAccount, ObjectContact, ObjectOpportunity, ObjectLead:
		return ObjectType(s), true
	default:
		return "", false
	}
}
