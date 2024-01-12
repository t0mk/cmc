package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/tidwall/pretty"
)

type CcmApiResponse struct {
	Status interface{} `json:"status"`
	Data   interface{} `json:"data"`
}

func getCMCkeyFromEnv() string {
	cmcKey := os.Getenv("CMC_KEY")
	if cmcKey == "" {
		log.Fatal("CMC_KEY not set in environment")
	}
	return cmcKey
}

type Context struct {
	Debug bool
}

var (
	// CMC_KEY is the API key for CoinMarketCap
	cmcKey string
)

const (
	sandboxUrlBase = "https://sandbox-api.coinmarketcap.com"
	prodUrlBase    = "https://pro-api.coinmarketcap.com"
)

func getRaw(path string, args map[string]string) ([]byte, error) {
	client := &http.Client{}

	q := url.Values{}
	for k, v := range args {
		q.Add(k, v)
	}

	urlBase := prodUrlBase
	path = strings.TrimPrefix(path, "/")
	endpoint := fmt.Sprintf("%s/%s", urlBase, path)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accepts", "application/json")
	req.Header.Add("X-CMC_PRO_API_KEY", cmcKey)
	req.URL.RawQuery = q.Encode()

	log.Println("get to ", req.URL.String())

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		if areBytesJson(respBody) {
			return nil, fmt.Errorf("bad status code: %s\n%s", resp.Status, prettyColoredJsonBytes(respBody))
		}
		return nil, fmt.Errorf("bad status code: %s\n%s", resp.Status, string(respBody))
	}
	return respBody, err
}

func areBytesJson(b []byte) bool {
	var js map[string]interface{}
	return json.Unmarshal(b, &js) == nil
}

type CmcExpression struct {
	Path string
	Args string
	Cmd  CmcMode
}

func SplitOnDot(s string) (string, string) {
	sli := strings.Split(s, ".")
	if len(sli) == 1 {
		return sli[0], ""
	}
	return sli[0], sli[1]
}

func NewCmcExpression(cmd string) CmcExpression {
	last := cmd[len(cmd)-1:]
	if last == "?" {
		path, args := SplitOnDot(cmd[:len(cmd)-1])
		return CmcExpression{Path: path, Args: args, Cmd: CmcModeHelp}
	}
	if last == "+" {
		path, args := SplitOnDot(cmd[:len(cmd)-1])
		return CmcExpression{Path: path, Args: args, Cmd: CmcModeTree}
	}
	if last == "!" {
		path, args := SplitOnDot(cmd[:len(cmd)-1])
		return CmcExpression{Path: path, Args: args, Cmd: CmcModeExpand}
	}
	path, args := SplitOnDot(cmd)
	return CmcExpression{Path: path, Args: args, Cmd: CmcModeGet}
}

type CmcCmd struct {
	Path string `arg:"" help:"Path to get"`
}

type CmcMode int64

const (
	CmcModeGet CmcMode = iota
	CmcModeExpand
	CmcModeTree
	CmcModeHelp
)

func CmcRun(c string) error {
	x := NewCmcExpression(c)

	switch x.Cmd {
	case CmcModeGet:
		return GetRun(&x)
	case CmcModeExpand:
		return ExpandRun(&x)
	case CmcModeTree:
		return TreeRun(&x)
	case CmcModeHelp:
		return HelpRun(&x)
	}
	return nil
}

func HelpRun(c *CmcExpression) error {
	p, err := api.GetLeafNode(c.Path)
	if err != nil {
		return err
	}
	fmt.Println(p)
	fmt.Println(p.Args)
	return nil
}

func TreeRun(c *CmcExpression) error {
	p, err := api.GetNode(c.Path)
	if err != nil {
		return err
	}
	e, err := api.ExpandExpression(c.Path)
	if err != nil {
		return err
	}
	fmt.Println(e)
	fmt.Println(p)
	return nil
}

func ExpandRun(c *CmcExpression) error {
	p, err := api.ExpandExpression(c.Path)
	if err != nil {
		return err
	}
	fmt.Println(p)
	return nil
}

func GetRun(c *CmcExpression) error {
	inargs := map[string]string{}
	if len(c.Args) > 1 {
		targs := strings.Split(c.Args, ",")
		for _, t := range targs {
			kv := strings.Split(t, "=")
			inargs[kv[0]] = kv[1]
		}
	}

	p, err := api.GetLeafNode(c.Path)
	if err != nil {
		return err
	}

	rargs := map[string]string{}
	for k, v := range inargs {
		foundArg, err := p.Args.FindByPrefix(k)
		if err != nil {
			return err
		}
		rargs[foundArg] = v
	}

	resp, err := getRaw(p.Url, rargs)
	if err != nil {
		return err
	}

	var ccmResp CcmApiResponse
	err = json.Unmarshal(resp, &ccmResp)
	if err != nil {
		return err
	}

	dataBytes, err := json.Marshal(ccmResp.Data)
	if err != nil {
		return err
	}

	fmt.Println(prettyColoredJsonBytes(dataBytes))

	return nil
}

type RawCmd struct {
	Endpoint string `arg:"" help:"Endpoint to query, e.g v1/cryptocurrency/quotes/latest"`
}

func prettyColoredJsonBytes(resp []byte) string {
	return string(pretty.Color(pretty.Pretty(resp), nil))
}

func (r *RawCmd) Run() error {
	resp, err := getRaw(r.Endpoint, nil)
	if err != nil {
		return err
	}
	fmt.Println(prettyColoredJsonBytes(resp))
	return nil
}

func main() {

	if len(os.Args) == 1 {
		fmt.Println("usage: cmc <path>[.args][?|+|!]")
		fmt.Println("")
		fmt.Println("you must export your Coinmarketcap API key to envvar CMC_KEY")
		fmt.Println("")
		fmt.Println("path")
		fmt.Println(" - is a dot-separated path to a resource, e.g. v1/cryptocurrency/quotes/latest")
		fmt.Println(" - can be shorted to the first unique prefix, e.g. v1/cryptocurrency/quotes/latest can be v1/c/q/l")
		fmt.Println(" - can be expanded to the full path, e.g. v1/c/q/l! will print v1/cryptocurrency/quotes/latest")
		fmt.Println("")
		fmt.Println("args")
		fmt.Println(" - are comma-separated key=value pairs, e.g. v2/tools/price-conversion.amount=1,convert=eth,symbol=rpl")
		fmt.Println(" - can be shorted to the first unique prefix just like path, e.g. v2/t/p.s=rpl,convert=eth,a=1")
		fmt.Println("")
		fmt.Println("commands at the end of the path/args")
		fmt.Println(" - ? prints documentation for arguments for endpoint on the path")
		fmt.Println(" - + prints the tree of endpoints under the path")
		fmt.Println(" - ! expands the full path from the short path")
		fmt.Println("")
		fmt.Println("examples")
		fmt.Println(" $ /cmc \"v2/t/p?\"")
		os.Exit(0)
	}

	cmcKey = getCMCkeyFromEnv()
	err := CmcRun(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
}
