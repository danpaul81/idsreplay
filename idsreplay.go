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

	//process command line args
	ipPtr := flag.String("dest", "127.0.0.1", "IP / hostname address of IDS replay target")
	portPtr := flag.Uint64("dport", 80, "IP port of IDS replay target")
	waitsecPtr := flag.Int("waitsec", 5, "seconds to wait between replay attempts. Note: Not each attempt might be successful")
	replayCountPtr := flag.Uint("count", 0, "# of IDS replay attemps (will count successful TCP connections doing a replay request). 0 for infinite")
	rulePtr := flag.String("rulefile", "/idsreplay/emerging-all.rules", "IDS signatures source. Suricata 4 format.")
	debugPtr := flag.Bool("debug", false, "run in debug mode")
	sidlistPtr := flag.String("sidlist", "", "comma separated list of rule SID to replay")

	flag.Parse()

	if *portPtr > 65535 {
		fmt.Print("ip port out of range:", *portPtr)
		return
	}

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

	//do we have a list of SID to replay? if not, we will run in random mode
	if len(*sidlistPtr) > 0 {
		log.Printf("Requested to replay following SID list %v \n", *sidlistPtr)
		var replayList []int
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
		log.Printf("Will now start replaying %v item IDS signature list to %v:%v waiting %v sec between attempts. Limit: %v \n", len(replayList), *ipPtr, *portPtr, *waitsecPtr, *replayCountPtr)
		// TODO: merge this code with the following random replay code, maybe create a single function
		var cont bool = true
		var totalAttempts uint = 0

		if *replayCountPtr > 0 {
			totalAttempts = *replayCountPtr * uint(len(replayList))
			log.Printf("will run SID list %v times (resulting in %v total attempts) Note: only successful TCP connects counted. If a single replay fails it wont be counted and we'll continue with next one. \n", *replayCountPtr, totalAttempts)
			CountReplay = 1
		}

		for cont {
			for _, x := range replayList {
				req := "http://" + *ipPtr + ":" + fmt.Sprintf("%v", *portPtr) + "/" + HTTPRequestList[x].HTTPuri

				log.Printf("# %v \t replay SID %v \t Method %v \t URI %v", CountReplay, HTTPRequestList[x].SID, HTTPRequestList[x].HTTPmethod, req)
				_, err := http.Get(req)
				if err != nil {
					log.Printf("HTTP Request %v", err)
				} else {
					if totalAttempts > 0 {
						//log.Printf("%v of %v total", CountReplay, totalAttempts)
						if CountReplay == totalAttempts {
							cont = false
						}
						CountReplay++
					}
				}
				time.Sleep(time.Duration(*waitsecPtr) * time.Second)
			}
		}

		// replay random rules
	} else {
		log.Printf("Will now start random IDS signature replay to %v:%v waiting %v sec between attempts. Limit: %v \n", *ipPtr, *portPtr, *waitsecPtr, *replayCountPtr)
		var cont bool = true
		if *replayCountPtr > 0 {
			CountReplay = 1
		}

		for cont {
			rand.Seed(time.Now().UnixNano())
			x := rand.Intn(len(HTTPRequestList))
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
}
