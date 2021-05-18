package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/danpaul81/idsreplay/idsparser"
)

var CountRuleMatch int
var CountRuleError int
var CountRuleOptError int
var CountReplay uint

func analyzeRule(IDSruleLine string, rule *idsparser.Rule) bool {
	err := idsparser.ParseRule(IDSruleLine, rule)

	rule.HasUnknownOpts = false
	rule.LastUnknownOpt = ""

	//Todo: when error is "comment line" we shouldnt display anything
	switch err {
	case nil:
		CountRuleMatch++
		idsparser.ParseRuleOptions(rule)
		if rule.HasUnknownOpts {
			CountRuleOptError++
		}
		return (true)
	default:
		CountRuleError++
		log.Printf("Rule Parse Error: %v (TODO Run with --ruleparsedebug to show )\n", err)
		return (false)
	}
}

func main() {

	//process command line args
	ipPtr := flag.String("dest", "127.0.0.1", "IP / hostname address of IDS replay target")
	portPtr := flag.Uint64("dport", 80, "IP port of IDS replay target")
	waitsecPtr := flag.Int("waitsec", 5, "seconds to wait between replay attempts. Note: Not each attempt might be successful")
	replayCountPtr := flag.Uint("count", 0, "# of IDS replay attemps (will count successful TCP connections doing a replay request). 0 for infinite")
	rulePtr := flag.String("rulefile", "emerging-all.rules", "IDS signatures source. Suricata 4 format.")
	flag.Parse()

	if *portPtr > 65535 {
		fmt.Print("ip port out of range:", *portPtr)
		return
	}

	/*
		targetIP := net.ParseIP(*ipPtr)
		if targetIP == nil {
			fmt.Print("invalid target ip", *ipPtr)
			return
		}
	*/
	// process single rule for testing cases or open rule file with multiple rules
	single_rule := false

	var IDSRule idsparser.Rule
	var pIDSRule *idsparser.Rule = &IDSRule

	var HTTPRequestList []idsparser.HTTPrequest

	if single_rule {
		rule := `alert http $HOME_NET any -> $EXTERNAL_NET any (msg:"Testrule"; flow:established,to_server; content:"GET"; http_method; content:"?sbFileName=../"; http_uri; cfast_pattern; reference:url,vmware.com; reference:cve,2020-8209; classtype:demorule; sid:4711; rev:1; metadata:lotofinfo; )`

		if analyzeRule(rule, pIDSRule) && !IDSRule.Commented {
			httpRequest, err := idsparser.ReplayRule(pIDSRule)
			if err == nil {
				HTTPRequestList = append(HTTPRequestList, httpRequest)
			}
		}
	} else {

		file, err := os.Open(*rulePtr)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			rule := scanner.Text()
			if len(rule) != 0 {
				IDSRule.RuleOptions = nil
				IDSRule.SID = ""

				if analyzeRule(rule, pIDSRule) && !IDSRule.Commented {
					httpRequest, err := idsparser.ReplayRule(pIDSRule)
					if err == nil {
						HTTPRequestList = append(HTTPRequestList, httpRequest)
					}
				}
			}
		}
	}

	log.Printf("Rules Processed: %v Matches: %v NoMatch: %v \n", CountRuleMatch+CountRuleError, CountRuleMatch, CountRuleError)
	log.Printf("Rules with at least one unknown option: %v . (TODO Run with --ruleoptdebug to show)\n", CountRuleOptError)
	log.Printf("Found %v HTTP Rules. Added to replay repository: %v", idsparser.CountHTTPrules, idsparser.CountHTTPreplay)
	log.Printf("Will now start IDS signature replay to %v:%v waiting %v sec between attempts. Limit: %v \n", *ipPtr, *portPtr, *waitsecPtr, *replayCountPtr)

	if idsparser.CountHTTPreplay == 0 {
		log.Printf("Oops. No rules to replay. Exiting now")
		return
	}
	var cont bool = true
	if *replayCountPtr > 0 {
		CountReplay = 1
	}

	for cont {
		rand.Seed(time.Now().UnixNano())
		x := rand.Intn(len(HTTPRequestList))
		log.Printf("RAND: %v", x)
		req := "http://" + *ipPtr + ":" + fmt.Sprintf("%v", *portPtr) + "/" + HTTPRequestList[x].HTTPuri

		log.Printf("# %v \t replay SID %v \t Method %v \t URI %v", CountReplay, HTTPRequestList[x].SID, HTTPRequestList[x].HTTPmethod, req)
		_, err := http.Get(req)
		if err != nil {
			log.Printf("HTTP Request %v", err)
		} else {
			if *replayCountPtr > 0 {
				if CountReplay == *replayCountPtr {
					cont = false
				}
				CountReplay++
			}
		}
		time.Sleep(time.Duration(*waitsecPtr) * time.Second)
	}
}
