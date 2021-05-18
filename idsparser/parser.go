package idsparser

import (
	"errors"
	"regexp"
	"strings"
)

func inSlice(str string, strings []string) bool {
	for _, k := range strings {
		if str == k {
			return true
		}
	}
	return false
}

func ParseRule(ruleLine string, rule *Rule) error {

	ruleLine = strings.TrimSpace(ruleLine)
	regExRule := `^(?P<comment>^#*[[:space:]#]*)(?P<header>[^()]+\((?P<options>.*)\)$)`

	var compRegExRule = regexp.MustCompile(regExRule)

	if compRegExRule.MatchString(ruleLine) {
		match := compRegExRule.FindStringSubmatch(ruleLine)
		//fmt.Println(compRegEx.SubexpNames()[1])

		// is rule commented?
		if strings.Contains(match[1], "#") {
			rule.Commented = true
		} else {
			rule.Commented = false
		}

		// split header to get action and protocol
		header := strings.Fields(match[2])

		if inSlice(header[0], []string{"alert", "drop"}) {
			rule.Action = header[0]
		} else {
			return (errors.New("unknown header action"))
		}

		if inSlice(header[1], allProtocolMatchTypeNames()) {
			rule.Protocol = getProtocolMatchTypeIndex(header[1])
			rule.RawOptions = match[3]
		} else {
			return (errors.New("unknown header protocol"))
		}

		return (nil)
	} else {
		if string(ruleLine[0]) != "#" {
			return (errors.New("single rule: no rule regex match"))
		} else {
			return (errors.New("commented line"))
		}
	}
}

func ParseRuleOptions(rule *Rule) {
	rawOptions := strings.Split(rule.RawOptions, ";")
	for _, v := range rawOptions {
		v := strings.TrimSpace(v)
		if len(v) > 0 {
			opt := strings.SplitN(v, ":", 2)
			//found a known rule option?
			if inSlice(opt[0], allOptionMatchTypeNames()) {

				optindex := getOptionMatchTypeIndex(opt[0])

				switch len(opt) {

				//opt rule keyword like 'nocase' (without setting)
				case 1:
					rule.RuleOptions = append(rule.RuleOptions, RuleOptionList{optindex, ""})

				//opt rule with settings like 'sid:123456'
				case 2:
					if opt[0] == "sid" {
						rule.SID = opt[1]
					}
					rule.RuleOptions = append(rule.RuleOptions, RuleOptionList{optindex, strings.Trim(opt[1], "\"")})
				}

			} else {
				//unkown rule option. Mark Rule, store (last) found option and proceed.
				// ToDo: Splitting rules by ; causes issues with pcre option
				rule.HasUnknownOpts = true
				rule.LastUnknownOpt = opt[0]
			}

		}
	}

}
