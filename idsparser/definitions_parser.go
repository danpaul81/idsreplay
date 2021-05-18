package idsparser

type Rule struct {
	Commented bool
	Action    string
	Protocol  ruleProtocolMatchType
	SID       string

	// this implementation ignores Source/Destination/Direction
	//Source        Network
	//Destination   Network
	//Bidirectional bool

	//if we find unknown option during processing will set true
	RawOptions     string
	HasUnknownOpts bool
	// store last unkown option for debugging
	LastUnknownOpt string
	RuleOptions    []RuleOptionList
}

type RuleOptionList struct {
	OptionName  ruleOptionMatchType
	OptionValue string
}

// describes kinds of actions
type ruleProtocolMatchType int

const (
	tcp ruleProtocolMatchType = iota
	ipv6
	http
	tls
	udp
	smtp
	dns
	ip
	icmp
	ftp
	tcppkt
	smb
	ssh
	ftpdata
	dcerpc
)

//
var ruleProtocolMatchTypeVals = map[ruleProtocolMatchType]string{
	tcp:     "tcp",
	ipv6:    "ipv6",
	http:    "http",
	tls:     "tls",
	udp:     "udp",
	smtp:    "smtp",
	dns:     "dns",
	ip:      "ip",
	icmp:    "icmp",
	ftp:     "ftp",
	tcppkt:  "tcp-pkt",
	smb:     "smb",
	ssh:     "ssh",
	ftpdata: "ftp-data",
	dcerpc:  "dcerpc",
}

//allProtocolMatchTypeNames returns a slice of valid protocol keywords
func allProtocolMatchTypeNames() []string {
	b := make([]string, len(ruleProtocolMatchTypeVals))
	var i int
	for _, n := range ruleProtocolMatchTypeVals {
		b[i] = n
		i++
	}
	return b
}

//getProtocolMatchTypeIndex returns ruleProtocolMatchType for given protocol string
func getProtocolMatchTypeIndex(rawOpt string) ruleProtocolMatchType {
	for a, b := range ruleProtocolMatchTypeVals {
		if b == rawOpt {
			return a
		}
	}
	return -1
}

// describes kinds of options
type ruleOptionMatchType int

const (
	msg ruleOptionMatchType = iota
	content
	flow
	reference
	fastpattern
	sid
	classtype
	metadata
	httpuri
	httpmethod
	rev
	flags
	rawbytes
	httpdata
	httpheader
	depth
	itype
	icode
	filedata
	nocase
	distance
	tlssni
	httpclientbody
	flowbits
	dnsquery
	pcre
	isdataat
	within
	tlscertsubject
	httphost
	httpstatcode
	httpheadernames
	httpcookie
	dsize
	httpuseragent
	httpcontenttype
	offset
	id
	icmpid
	httprequestline
	urilen
	httpstart
	httpaccept
	httpreferer
	httpacceptlang
	byteextract
	threshold
	tlsversion
	uricontent
	bytejump
	bytetest
	httpstatmsg
	tlsfingerprint
	httpcontentlen
	httpresponseline
	httprawuri
	httpconnection
	tlscertissuer
	sslstate
	httpacceptenc
	tlscertserial
	httprawheader
	streamsize
	httpprotocol
	httpserverbody
	ipproto
	base64decode
	base64data
	detectionfilter
	sslversion
	xbits
	sshproto
	dceiface
	dceopnum
	ttl
)

//
var ruleOptionMatchTypeVals = map[ruleOptionMatchType]string{
	msg:              "msg",
	content:          "content",
	classtype:        "classtype",
	flow:             "flow",
	reference:        "reference",
	fastpattern:      "fast_pattern",
	sid:              "sid",
	flags:            "flags",
	rawbytes:         "rawbytes",
	icmpid:           "icmp_id",
	streamsize:       "stream_size",
	rev:              "rev",
	metadata:         "metadata",
	filedata:         "file_data",
	nocase:           "nocase",
	distance:         "distance",
	itype:            "itype",
	icode:            "icode",
	depth:            "depth",
	tlssni:           "tls_sni",
	flowbits:         "flowbits",
	dnsquery:         "dns_query",
	pcre:             "pcre",
	isdataat:         "isdataat",
	offset:           "offset",
	within:           "within",
	dsize:            "dsize",
	xbits:            "xbits",
	sshproto:         "ssh_proto",
	dceiface:         "dce_iface",
	id:               "id",
	ttl:              "ttl",
	base64decode:     "base64_decode",
	base64data:       "base64_data",
	dceopnum:         "dce_opnum",
	detectionfilter:  "detection_filter",
	ipproto:          "ip_proto",
	byteextract:      "byte_extract",
	bytejump:         "byte_jump",
	bytetest:         "byte_test",
	threshold:        "threshold",
	httpuri:          "http_uri",
	httpmethod:       "http_method",
	httpdata:         "http_data",
	httpheader:       "http_header",
	httpclientbody:   "http_client_body",
	httphost:         "http_host",
	httpstatcode:     "http_stat_code",
	httpheadernames:  "http_header_names",
	httpcookie:       "http_cookie",
	httpuseragent:    "http_user_agent",
	httpcontenttype:  "http_content_type",
	httprequestline:  "http_request_line",
	httpstart:        "http_start",
	httpaccept:       "http_accept",
	httpreferer:      "http_referer",
	httpacceptlang:   "http_accept_lang",
	httpstatmsg:      "http_stat_msg",
	httpcontentlen:   "http_content_len",
	httpresponseline: "http_response_line",
	httpconnection:   "http_connection",
	httprawuri:       "http_raw_uri",
	httpacceptenc:    "http_accept_enc",
	httprawheader:    "http_raw_header",
	httpprotocol:     "http_protocol",
	httpserverbody:   "http_server_body",
	urilen:           "urilen",
	uricontent:       "uricontent",
	sslstate:         "ssl_state",
	sslversion:       "ssl_version",
	tlscertsubject:   "tls_cert_subject",
	tlscertissuer:    "tls_cert_issuer",
	tlsversion:       "tls.version",
	tlsfingerprint:   "tls.fingerprint",
	tlscertserial:    "tls_cert_serial",
}

//allOptionMatchTypeNames returns a slice of valid option keywords
func allOptionMatchTypeNames() []string {
	b := make([]string, len(ruleOptionMatchTypeVals))
	var i int
	for _, n := range ruleOptionMatchTypeVals {
		b[i] = n
		i++
	}
	return b
}

//getOptionMatchTypeIndex returns ruleOptionMatchType for given raw option string
func getOptionMatchTypeIndex(rawOpt string) ruleOptionMatchType {
	for a, b := range ruleOptionMatchTypeVals {
		if b == rawOpt {
			return a
		}
	}
	return -1
}
