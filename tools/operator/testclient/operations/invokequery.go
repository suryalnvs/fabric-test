package operations

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"regexp"
	"sync"
	"time"
	"net/http"
	"net/url"
	"io/ioutil"

	"github.com/hyperledger/fabric-test/tools/operator/logger"
	"github.com/hyperledger/fabric-test/tools/operator/networkclient"
	//"github.com/hyperledger/fabric-test/tools/operator/launcher/k8s"
	"github.com/hyperledger/fabric-test/tools/operator/paths"
	"github.com/hyperledger/fabric-test/tools/operator/testclient/inputStructs"
)

//InvokeQueryUIObject --
type InvokeQueryUIObject struct {
	LogLevel        string                `json:"logLevel,omitempty"`
	ChaincodeID     string                `json:"chaincodeID,omitempty"`
	InvokeCheck     string                `json:"invokeCheck,omitempty"`
	TransMode       string                `json:"transMode,omitempty"`
	TransType       string                `json:"transType,omitempty"`
	InvokeType      string                `json:"invokeType,omitempty"`
	TargetPeers     string                `json:"targetPeers,omitempty"`
	TLS             string                `json:"TLS,omitempty"`
	NProcPerOrg     string                `json:"nProcPerOrg,omitempty"`
	NRequest        string                `json:"nRequest,omitempty"`
	RunDur          string                `json:"runDur,omitempty"`
	ChannelOpt      ChannelOptions        `json:"channelOpt,omitempty"`
	BurstOpt        BurstOptions          `json:"burstOpt,omitempty"`
	MixOpt          MixOptions            `json:"mixOpt,omitempty"`
	ConstOpt        ConstantOptions       `json:"constantOpt,omitempty"`
	EventOpt        EventOptions          `json:"eventOpt,omitempty"`
	DiscoveryOpt    DiscoveryOptions      `json:"discoveryOpt,omitempty"`
	ListOpt         map[string][]string   `json:"listOpt,omitempty"`
	CCType          string                `json:"ccType,omitempty"`
	CCOpt           CCOptions             `json:"ccOpt,omitempty"`
	Parameters      map[string]Parameters `json:"invoke,omitempty"`
	ConnProfilePath string                `json:"ConnProfilePath,omitempty"`
	TimeOutOpt      TimeOutOptions        `json:"timeoutOpt,timeoutOpt"`
}

//BurstOptions --
type BurstOptions struct {
	BurstFreq0 string `json:"burstFreq0,omitempty"`
	BurstDur0  string `json:"burstDur0,omitempty"`
	BurstFreq1 string `json:"burstFreq1,omitempty"`
	BurstDur1  string `json:"burstDur1,omitempty"`
}

//MixOptions --
type MixOptions struct {
	MixFreq string `json:"mixFreq,omitempty"`
}

//ConstantOptions --
type ConstantOptions struct {
	RecHist   string `json:"recHist,omitempty"`
	ConstFreq string `json:"constFreq,omitempty"`
	DevFreq   string `json:"devFreq,omitempty"`
}

//EventOptions --
type EventOptions struct {
	Type     string `json:"type,omitempty"`
	Listener string `json:"listener,omitempty"`
	TimeOut  string `json:"timeout,omitempty"`
}

//CCOptions --
type CCOptions struct {
	KeyIdx      []int  `json:"keyIdx,omitempty"`
	KeyPayLoad  []int  `json:"keyPayLoad,omitempty"`
	KeyStart    string `json:"keyStart,omitempty"`
	PayLoadMin  string `json:"payLoadMin,omitempty"`
	PayLoadMax  string `json:"payLoadMax,omitempty"`
	PayLoadType string `json:"payLoadType,omitempty"`
}

//DiscoveryOptions --
type DiscoveryOptions struct {
	Localhost string `json:"localHost,omitempty"`
	InitFreq  int    `json:"initFreq,omitempty"`
}

//Parameters --
type Parameters struct {
	Fcn  string   `json:"fcn,omitempty"`
	Args []string `json:"args,omitempty"`
}

type blockchainCount struct {
	peerTransactionCount int
	peerBlockchainHeight int 
}

//InvokeQuery -- To perform invoke/query with the objects created
func (i InvokeQueryUIObject) InvokeQuery(config inputStructs.Config, tls, action string) error {

	var invokeQueryObjects []InvokeQueryUIObject
	var err error
	configObjects := config.Invoke
	if action == "Query" {
		configObjects = config.Query
	}
	for key := range configObjects {
		invkQueryObjects := i.generateInvokeQueryObjects(configObjects[key], config.Organizations, tls, action)
		invokeQueryObjects = append(invokeQueryObjects, invkQueryObjects...)
	}
	err = i.invokeQueryTransactions(invokeQueryObjects)
	if err != nil {
		return err
	}
	return err
}

//generateInvokeQueryObjects -- To generate objects for invoke/query
func (i InvokeQueryUIObject) generateInvokeQueryObjects(invkQueryObject inputStructs.InvokeQuery, organizations []inputStructs.Organization, tls, action string) []InvokeQueryUIObject {

	var invokeQueryObjects []InvokeQueryUIObject
	orgNames := strings.Split(invkQueryObject.Organizations, ",")
	invkQueryObjects := i.createInvokeQueryObjectForOrg(orgNames, action, tls, organizations, invkQueryObject)
	invokeQueryObjects = append(invokeQueryObjects, invkQueryObjects...)
	return invokeQueryObjects
}

//createInvokeQueryObjectForOrg -- To craete invoke/query objects for an organization
func (i InvokeQueryUIObject) createInvokeQueryObjectForOrg(orgNames []string, action, tls string, organizations []inputStructs.Organization, invkQueryObject inputStructs.InvokeQuery) []InvokeQueryUIObject {

	var invokeQueryObjects []InvokeQueryUIObject
	invokeParams := make(map[string]Parameters)
	invokeCheck := "TRUE"
	if invkQueryObject.QueryCheck > 0 {
		invokeCheck = "FALSE"
	}
	i = InvokeQueryUIObject{
		LogLevel:    "ERROR",
		InvokeCheck: invokeCheck,
		TransType:   "Invoke",
		InvokeType:  "Move",
		TargetPeers: invkQueryObject.TargetPeers,
		TLS:         tls,
		NProcPerOrg: strconv.Itoa(invkQueryObject.NProcPerOrg),
		NRequest:    strconv.Itoa(invkQueryObject.NRequest),
		RunDur:      strconv.Itoa(invkQueryObject.RunDuration),
		CCType:      invkQueryObject.CCOptions.CCType,
		ChaincodeID: invkQueryObject.ChaincodeName,
		EventOpt: EventOptions{
			Type:     invkQueryObject.EventOptions.Type,
			Listener: invkQueryObject.EventOptions.Listener,
			TimeOut:  strconv.Itoa(invkQueryObject.EventOptions.TimeOut),
		},
		CCOpt: CCOptions{
			KeyIdx:      invkQueryObject.CCOptions.KeyIdx,
			KeyPayLoad:  invkQueryObject.CCOptions.KeyPayload,
			KeyStart:    strconv.Itoa(invkQueryObject.CCOptions.KeyStart),
			PayLoadMin:  strconv.Itoa(invkQueryObject.CCOptions.PayLoadMin),
			PayLoadMax:  strconv.Itoa(invkQueryObject.CCOptions.PayLoadMax),
			PayLoadType: invkQueryObject.CCOptions.PayLoadType,
		},
		TimeOutOpt: TimeOutOptions{
			Request:   invkQueryObject.TimeOutOpt.Request,
			PreConfig: invkQueryObject.TimeOutOpt.PreConfig,
		},
		ChannelOpt: ChannelOptions{
			Name:    invkQueryObject.ChannelName,
			OrgName: orgNames,
		},
		ConnProfilePath: paths.GetConnProfilePath(orgNames, organizations),
	}
	if strings.EqualFold("DISCOVERY", invkQueryObject.TargetPeers) {
		localHost := strings.ToUpper(strconv.FormatBool(invkQueryObject.DiscoveryOptions.Localhost))
		i.DiscoveryOpt = DiscoveryOptions{
			Localhost: localHost,
			InitFreq:  invkQueryObject.DiscoveryOptions.InitFreq,
		}
	}
	if strings.EqualFold("LIST", invkQueryObject.TargetPeers) {
		i.ListOpt = invkQueryObject.ListOptions
	}
	if action == "Query" {
		i.InvokeType = action
		i.CCOpt = CCOptions{KeyIdx: invkQueryObject.CCOptions.KeyIdx, KeyStart: strconv.Itoa(invkQueryObject.CCOptions.KeyStart)}
	}
	invokeParams["move"] = Parameters{
		Fcn:  invkQueryObject.Fcn,
		Args: strings.Split(invkQueryObject.Args, ","),
	}
	invokeParams["query"] = Parameters{
		Fcn:  invkQueryObject.Fcn,
		Args: strings.Split(invkQueryObject.Args, ","),
	}
	i.Parameters = invokeParams
	for key := range invkQueryObject.TxnOptions {
		mode := invkQueryObject.TxnOptions[key].Mode
		options := invkQueryObject.TxnOptions[key].Options
		i.TransMode = mode
		switch mode {
		case "constant":
			i.ConstOpt = ConstantOptions{RecHist: "HIST", ConstFreq: strconv.Itoa(options.ConstFreq), DevFreq: strconv.Itoa(options.DevFreq)}
		case "burst":
			i.BurstOpt = BurstOptions{BurstFreq0: strconv.Itoa(options.BurstFreq0), BurstDur0: strconv.Itoa(options.BurstDur0), BurstFreq1: strconv.Itoa(options.BurstFreq1), BurstDur1: strconv.Itoa(options.BurstDur1)}
		case "mix":
			i.MixOpt = MixOptions{MixFreq: strconv.Itoa(options.MixFreq)}
		}
		invokeQueryObjects = append(invokeQueryObjects, i)
	}
	return invokeQueryObjects
}

func (i InvokeQueryUIObject) invokeConfig(channelName string, args []string, wg *sync.WaitGroup) error {
	//defer wg.Done()
	_, err := networkclient.ExecuteCommand("node", args, true)
	if err != nil {
		logger.ERROR(fmt.Sprintf("Failed to perform invoke/query on %s channel: %v", channelName, err))
		os.Exit(1)
	}
	return nil
}

//invokeQueryTransactions -- To invoke/query transactions
func (i InvokeQueryUIObject) invokeQueryTransactions(invokeQueryObjects []InvokeQueryUIObject) error {

	var err error
	var jsonObject []byte
	var wg sync.WaitGroup
	pteMainPath := paths.PTEPath()
	for key := range invokeQueryObjects {
		jsonObject, err = json.Marshal(invokeQueryObjects[key])
		if err != nil {
			return err
		}
		startTime := fmt.Sprintf("%s", time.Now())
		args := []string{pteMainPath, strconv.Itoa(key), string(jsonObject), startTime}
		wg.Add(1)
		go func(err error) {
			defer wg.Done()
			err = i.invokeConfig(invokeQueryObjects[key].ChannelOpt.Name, args, &wg)
			if err != nil {
				logger.ERROR("Failed to complete invokes/queries")
				err = fmt.Errorf("Something went wrong in completing invokes/queries")
			}
			blockchain, err := i.fetchMetrics(invokeQueryObjects[key])
			if err != nil {
				logger.ERROR("failed fetching metrics")
				err = fmt.Errorf("Something went wrong in fetching metrics")
			}
			var blockchainHeight int
			var transactionCount int
			channelBlock := blockchain[invokeQueryObjects[key].ChannelOpt.Name]
			for key, value := range channelBlock {
				if blockchainHeight == 0 && transactionCount == 0 {
					blockchainHeight = value.peerBlockchainHeight
					transactionCount = value.peerTransactionCount
				} else {
					if value.peerBlockchainHeight != blockchainHeight || value.peerTransactionCount != transactionCount {
						logger.ERROR("Peers are not in sync")
						err = fmt.Errorf("Something went wrong with peer ", key, " as blockchain height or transactions does not match up")
					}
				}
		    }
		}(err)
	}
	wg.Wait()
	return nil
}

func (i InvokeQueryUIObject) fetchMetrics(invokeQueryObject InvokeQueryUIObject) (map[string]map[string]blockchainCount, error)  {

	var connProfilePath, metrics string
	var channelBlockchainCount = make(map[string]map[string]blockchainCount)
	connectionProfilePath := invokeQueryObject.ConnProfilePath
	orgName := invokeQueryObject.ChannelOpt.OrgName
	channelName := invokeQueryObject.ChannelOpt.Name
	channelBlockchainCount[channelName] = make(map[string]blockchainCount)
	var err error
	for i := range orgName {
		if strings.Contains(connectionProfilePath, ".yaml") || strings.Contains(connectionProfilePath, ".json") {
			connProfilePath = connectionProfilePath
		} else {
			connProfilePath = fmt.Sprintf("%s/connection_profile_%s.yaml", connectionProfilePath, orgName[i])
		}
		connProfConfig, err := ConnProfileInformationForOrg(connProfilePath, orgName[i])
		if err != nil {
			return nil, err
		}
		for _, peerName := range connProfConfig.Channels[channelName].Peers {
			metricsURL, err := url.Parse(connProfConfig.Peers[peerName].MetricsURL)
			if err != nil {
				logger.ERROR("Failed to get peer url from connection profile")
				return nil, err
			}
			resp, err := http.Get(fmt.Sprintf("%s/metrics", metricsURL))
			if err != nil {
				logger.ERROR("Error while hitting the endpoint")
				return nil, err
			}
			defer resp.Body.Close()
			bodyBytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
			metrics = string(bodyBytes)
			blockHeight := strings.Split(metrics, fmt.Sprintf(`ledger_blockchain_height{channel="%s"}`, channelName))
			height := strings.Split(blockHeight[1], "\n")[0]
			num, _ := strconv.Atoi(strings.TrimSpace(height))	
			regex := regexp.MustCompile(fmt.Sprintf(`ledger_transaction_count{chaincode="%s:[0-9A-Za-z]+",channel="%s",transaction_type="ENDORSER_TRANSACTION",validation_code="VALID"}`, invokeQueryObject.ChaincodeID, channelName))
			transactionCount := strings.Split(metrics, fmt.Sprintf(`%s`, regex.FindString(metrics)))
			trxnCount := strings.Split(transactionCount[1], "\n")[0]
			count, _ := strconv.Atoi(strings.TrimSpace(trxnCount))
			channelBlockchainCount[channelName][peerName] = blockchainCount{
				peerBlockchainHeight: num,
				peerTransactionCount: count,
			}
		}
	}
	return channelBlockchainCount, err
}