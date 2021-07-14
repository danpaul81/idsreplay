package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/danpaul81/idsreplay/idsparser"
)

const Version string = "0.2.1"

var CountRuleMatch int
var CountRuleError int
var CountRuleOptError int
var CountReplay uint

func analyzeRule(IDSruleLine string, rule *idsparser.Rule, debug bool) bool {
	// first parse rule with regex match
	err := idsparser.ParseRule(IDSruleLine, rule, debug)

	rule.HasUnknownOpts = false
	rule.LastUnknownOpt = ""

	switch err.(type) {
	case nil:
		CountRuleMatch++
		// second analyse options of rule
		idsparser.ParseRuleOptions(rule)
		if rule.HasUnknownOpts {
			CountRuleOptError++
			if debug {
				log.Printf("DEBUG: SID %v last unknown ignored option: %v \n", rule.SID, rule.LastUnknownOpt)
			}
		}
		return (true)
	case *idsparser.ErrorCommentLine:
		return (false)
	default:
		CountRuleError++

		if debug {
			log.Printf("DEBUG: Rule Parse Error: %v Rule: %v \n", err, IDSruleLine)
		} else {
			log.Printf("Rule Parse Error: %v (run with --debug=true to show rule )\n", err)
		}
		return (false)
	}
}

func main() {

	log.Printf("idsparser version %v \n", Version)

	//process command line args
	ipPtr := flag.String("dest", "127.0.0.1", "IP / hostname address of IDS replay target")
	portPtr := flag.Uint64("dport", 80, "IP port of IDS replay target")
	waitsecPtr := flag.Int("waitsec", 5, "seconds to wait between replay attempts. Note: Not each attempt might be successful")
	replayCountPtr := flag.Uint("count", 0, "# of IDS replay attemps (will count successful TCP connections doing a replay request). 0 for infinite")
	rulePtr := flag.String("rulefile", "/idsreplay/emerging-all.rules", "IDS signatures source. Suricata 4 format.")
	debugPtr := flag.Bool("debug", false, "run in debug mode")
	sidlistPtr := flag.String("sidlist", "", "comma separated list of rule SID to replay. if none of these SID is suitable random sid will be replayed")

	flag.Parse()

	if *portPtr > 65535 {
		fmt.Print("ip port out of range:", *portPtr)
		return
	}
	log.Printf("start processing rules. this may take a while \n")
	// process single rule for testing cases or open rule file with multiple rules
	single_rule := false

	var IDSRule idsparser.Rule
	var pIDSRule *idsparser.Rule = &IDSRule

	var HTTPRequestList []idsparser.HTTPrequest

	if single_rule {
		rule := `alert http $HOME_NET any -> $EXTERNAL_NET any (msg:"Testrule"; flow:established,to_server; content:"GET"; http_method; content:"?sbFileName=../"; http_uri; cfast_pattern; reference:url,vmware.com; reference:cve,2020-8209; classtype:demorule; sid:4711; rev:1; metadata:lotofinfo; )`

		if analyzeRule(rule, pIDSRule, *debugPtr) && !IDSRule.Commented {
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

				if analyzeRule(rule, pIDSRule, *debugPtr) && !IDSRule.Commented {
					httpRequest, err := idsparser.ReplayRule(pIDSRule)
					if err == nil {
						HTTPRequestList = append(HTTPRequestList, httpRequest)
					}
				}
			}
		}
	}

	log.Printf("Rules Processed: %v Matches: %v NoMatch: %v \n", CountRuleMatch+CountRuleError, CountRuleMatch, CountRuleError)
	log.Printf("Rules with at least one unknown option: %v . Unknown options will be ignored but rule may be added to repository (run with --debug=true to show)\n", CountRuleOptError)
	log.Printf("Found %v HTTP Rules. Added to replay repository: %v as these have valid & known options \n", idsparser.CountHTTPrules, idsparser.CountHTTPreplay)

	if idsparser.CountHTTPreplay == 0 {
		log.Printf("Oops. No rules to replay in repository. Exiting now")
		return
	}

	var replayList []int
	//build replay list if --sidlist option is set
	if len(*sidlistPtr) > 0 {
		log.Printf("Requested to replay following SID list %v \n", *sidlistPtr)

		sidList := strings.Split(*sidlistPtr, ",")
		for _, v := range sidList {
			v = strings.TrimSpace(v)
			var validSID bool = false
			// is requested SID available in HTTPRequestList [a.k.a. replay repository]
			for i := range HTTPRequestList {
				if HTTPRequestList[i].SID == v {
					validSID = true
					replayList = append(replayList, i)
				}
			}
			if !validSID {
				log.Printf("requested SID %v not found in replay repository. Will be ignored in replay \n", v)
			}
		}
	}

	var totalAttempts uint = 0
	var cont bool = true

	if *replayCountPtr > 0 {
		if len(replayList) > 0 {
			totalAttempts = *replayCountPtr * uint(len(replayList))
			log.Printf("replay SID list %v times (%v total attempts) to %v:%v waiting %v sec between attempts \n Note: only successful TCP connects counted. If a single replay fails it wont be counted and we'll continue with next one. \n", *replayCountPtr, totalAttempts, *ipPtr, *portPtr, *waitsecPtr)
			CountReplay = 1
		} else {
			totalAttempts = *replayCountPtr
			log.Printf("random replay (limit %v) to %v:%v waiting %v sec between attempts \n Note: only successful TCP connects counted. If a single replay fails it wont be counted and we'll continue with next one. \n", totalAttempts, *ipPtr, *portPtr, *waitsecPtr)
			CountReplay = 1
		}
	} else {
		if len(replayList) > 0 {
			totalAttempts = 0
			log.Printf("replay SID list (no limit) to %v:%v waiting %v sec between attempts\n", *ipPtr, *portPtr, *waitsecPtr)
			CountReplay = 0
		} else {
			totalAttempts = 0
			log.Printf("random replay (no limit) to %v:%v waiting %v sec between attempts\n", *ipPtr, *portPtr, *waitsecPtr)
			CountReplay = 0
		}

	}

	var nextItemIndex int = 0
	var nextItem int = 0
	client := &http.Client{}

	for cont {

		// do we need to follow a SID list? set nextItem accordingly
		if len(replayList) > 0 {
			nextItem = replayList[nextItemIndex]
			nextItemIndex++
			if nextItemIndex == len(replayList) {
				nextItemIndex = 0
			}
			// choose random nextItem
		} else {
			rand.Seed(time.Now().UnixNano())
			nextItem = rand.Intn(len(HTTPRequestList))
		}

		url := "http://" + *ipPtr + ":" + fmt.Sprintf("%v", *portPtr) + "/" + HTTPRequestList[nextItem].HTTPuri
		req, err := http.NewRequest(HTTPRequestList[nextItem].HTTPmethod, url, nil)
		if err != nil {
			log.Printf("URL Composition %v", err)
		}

		req.Header.Add("X-idsreplay-sid", HTTPRequestList[nextItem].SID)

		log.Printf("# %v \t replay SID %v \t Method %v \t URI %v", CountReplay, HTTPRequestList[nextItem].SID, HTTPRequestList[nextItem].HTTPmethod, url)
		_, err = client.Do(req)

		if err != nil {
			log.Printf("HTTP Request %v", err)
		} else {
			if totalAttempts != 0 {
				if CountReplay == totalAttempts {
					cont = false
				} else {
					CountReplay++
				}
			}
		}
		time.Sleep(time.Duration(*waitsecPtr) * time.Second)
	}
}
