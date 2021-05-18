package idsparser

//when processing a rule is multiple occurence allowed
// content:xxx always
// http_uri could also be composed by multiple options, but first version doesnt support this
type replayRuleSettings struct {
	ruleOption      ruleOptionMatchType
	multipleAllowed bool
	sticky          bool
}

// Rules to ignore when building replay -> non-traffic related
var ReplayGlobalIgnoreRules = [...]ruleOptionMatchType{
	msg,
	sid,
	rev,
	flow,
	fastpattern,
	classtype,
	reference,
	metadata,
}

type HTTPrequest struct {
	SID        string
	HTTPmethod string
	HTTPuri    string
}

// store allowed rules for http replay and if they are allowed to appear multiple times
var ReplayHTPPrules = [...]replayRuleSettings{
	{content, true, false},
	{httpmethod, false, false},
	//first version doesnt support uri re-compose, so only single httpuri option
	{httpuri, false, false},

	//{httpheader, true, false},
}
