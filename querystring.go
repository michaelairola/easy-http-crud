package easyHttpCrud

import (
	"errors"
	// "fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// RULES OF PARSE QUERY
/*
	reserved words:
		- page: establishes offset of query
		- query: will be written at face value

*/
func parseQueryStringFactory(queryType string) func(interface{}, url.Values) (string, []interface{}, string, int, int, []error) {
	return func(record interface{}, urlQueries url.Values) (string, []interface{}, string, int, int, []error) {
		intType := reflect.TypeOf(0)
		charType := reflect.TypeOf("")
		boolType := reflect.TypeOf(false)
		timeType := reflect.TypeOf(time.Now())

		queries := make([]string, 0)
		args := make([]interface{}, 0)
		var order string
		order = "id desc"
		page := 0
		pageLimit := 50
		warnings := make([]error, 0)

		for key, matches := range urlQueries {
			if key == "order" && len(matches) != 0 {
				order = matches[0]
			}
			if key == "page" && len(matches) != 0 {
				page, _ = strconv.Atoi(matches[0])
				continue
			}
			if key == "page_length" && len(matches) != 0 {
				var err error
				pageLimit, err = strconv.Atoi(matches[0])
				if err != nil {
					pageLimit = 50
				}
				continue
			}
			matchStr, t, hasCol := recordHasColumn(key, record)
			if !hasCol && key != "query" {
				warnings = append(warnings, errors.New("query error!"))
				continue
			}
			var query string
			var keyArgs []interface{}
			var warns []error
			if key == "" {
				continue
			} else if key == "query" {
				query = matches[0]
			} else if t == intType {
				query, keyArgs, warns = parseIntQuery(matchStr, matches)
			} else if t == charType {
				query, keyArgs = parseCharQuery(matchStr, matches)
			} else if t == boolType {
				query, keyArgs, warns = parseBoolQuery(matchStr, matches)
			} else if t == timeType {
				query, keyArgs = parseTimeQuery(matchStr, matches)
			}
			if query != "" {
				queries = append(queries, query)
			}
			if len(keyArgs) != 0 {
				args = append(args, keyArgs...)
			}
			if len(warns) != 0 {
				warnings = append(warnings, warns...)
			}
		}
		var Query string
		if len(queries) == 0 {
			Query = ""
		} else if len(queries) == 1 {
			Query = queries[0]
		} else {
			Query = "(" + strings.Join(queries, ") AND (") + ")"
		}
		return Query, args, order, page, pageLimit, warnings
	}
}

var ParseQueryStr = parseQueryStringFactory("")

func parseIntQuery(matchStr string, matches []string) (string, []interface{}, []error) {
	var queries []string
	var args []interface{}
	var notToggle bool
	var warnings []error
	for _, match := range matches {
		var queryStr string
		if match == "" {
			continue
		}
		if match == "not" {
			notToggle = true
			continue
		}
		num, err := strconv.Atoi(match)
		if err != nil {
			warn := errors.New("parsing integer for column " + matchStr + ". Unable to parse " + match + ".")
			warnings = append(warnings, warn)
			continue
		}
		if notToggle {
			queryStr = "!= ?"
			notToggle = false
		} else {
			queryStr = "= ?"
		}
		query := matchStr + " " + queryStr
		queries = append(queries, query)
		args = append(args, num)
	}
	return strings.Join(queries, " OR "), args, warnings
}
func parseCharQuery(matchStr string, matches []string) (string, []interface{}) {
	var exactToggle, notToggle bool
	var queries []string
	var args []interface{}
	for _, match := range matches {
		if match == "exact" {
			exactToggle = true
			continue
		} else if match == "not" {
			notToggle = true
			continue
		}
		newQuery := "LIKE ?"
		if notToggle {
			newQuery = "NOT " + newQuery
			notToggle = false
		}
		newQuery = matchStr + " " + newQuery
		queries = append(queries, newQuery)
		if !exactToggle {
			match = "%" + match + "%"
		} else {
			exactToggle = false
		}
		args = append(args, match)
	}
	return strings.Join(queries, " OR "), args
}
func parseBoolQuery(matchStr string, matches []string) (string, []interface{}, []error) {
	var queries []string
	var args []interface{}
	var warnings []error
	for _, match := range matches {
		v, err := strconv.ParseBool(match)
		if err != nil {
			warn := errors.New("parsing boolean for column " + matchStr + ". Unable to parse " + match + ".")
			warnings = append(warnings, warn)
			continue
		}
		queries = append(queries, matchStr+" = ?")
		args = append(args, v)
	}
	return strings.Join(queries, " OR "), args, warnings
}
func parseTimeQuery(matchStr string, matches []string) (string, []interface{}) {
	var queries []string
	var args []interface{}
	for _, match := range matches {
		if match == "lt" {
			queries = append(queries, matchStr+" < ?")
		} else if match == "gt" {
			queries = append(queries, matchStr+" > ?")
		} else if match == "NULL" {
			queries = append(queries, matchStr+" IS NULL")
		} else if match != "" {
			args = append(args, match)
		}
	}
	return strings.Join(queries, " AND "), args
}
