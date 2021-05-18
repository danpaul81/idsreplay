package idsparser

import (
	"errors"
	"fmt"
)

var CountHTTPrules int
var CountHTTPreplay int

func isOnReplayGlobalIgnore(checkOpt ruleOptionMatchType) bool {
	for _, v := range ReplayGlobalIgnoreRules {
		if v == checkOpt {
			return (true)
		}
	}
	return (false)
}

func isOnReplayHTTPrules(checkOpt ruleOptionMatchType) bool {
	for _, v := range ReplayHTPPrules {
		if v.ruleOption == checkOpt {
			return (true)
		}
	}
	return (false)
}

// currently designed only for http requests. when doing other requests this needs redesign
func ReplayRule(rule *Rule) (HTTPrequest, error) {

	var HTTPreq HTTPrequest
	var pHTTPreq *HTTPrequest = &HTTPreq

	HTTPreq.SID = rule.SID

	switch rule.Protocol {
	case http:

		//method default is GET
		HTTPreq.HTTPmethod = "GET"
		CountHTTPrules++

		if PrepareReplayHTTP(rule, pHTTPreq) {
			CountHTTPreplay++
			return HTTPreq, nil
		} else {
			return HTTPreq, errors.New("no valid http request built")
		}
	default:
		return HTTPreq, errors.New("not http")
	}

}

func PrepareReplayHTTP(rule *Rule, httpreq *HTTPrequest) bool {

	var allValidOptions bool
	var contentopen bool
	var contentbuf string

	allValidOptions = true
	contentopen = false

	for _, v := range rule.RuleOptions {
		if !isOnReplayGlobalIgnore(v.OptionName) {
			if isOnReplayHTTPrules(v.OptionName) {

				// do we have a "content" option?
				if ruleOptionMatchTypeVals[v.OptionName] == "content" {
					contentopen = true
					contentbuf = contentbuf + v.OptionValue
				} else {
					//if not content its a keyword for open content buffer
					if contentopen {
						switch ruleOptionMatchTypeVals[v.OptionName] {
						case "http_uri":
							httpreq.HTTPuri = httpreq.HTTPuri + contentbuf
							contentopen = false
							contentbuf = ""
						case "http_method":
							httpreq.HTTPmethod = contentbuf
							contentopen = false
							contentbuf = ""
						}
					} else {
						fmt.Printf("SID: %v \t Found option %v without open content. %v\n", rule.SID, ruleOptionMatchTypeVals[v.OptionName], rule.RuleOptions)
					}
				}
			} else {
				//option is neither on global ignore nor on http known option list
				//we'll continue building request, but it will not be executed
				allValidOptions = false
			}

		}

	}

	// for the case there is a content field without a follwing, valid keyword we will ignore this rule
	if contentopen && allValidOptions {
		allValidOptions = false
	}

	//if allValidOptions {
	//	fmt.Printf("READY SID: %v \t HTTPMethod: %v \t \t HTTPURI %v \n", rule.SID, httpreq.HTTPmethod, httpreq.HTTPuri)
	//}
	return (allValidOptions)
}
