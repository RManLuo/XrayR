package panel

import (
	"encoding/json"
	io "io/ioutil"
	"log"
	"sync"

	"github.com/XrayR-project/XrayR/api"
	"github.com/XrayR-project/XrayR/api/pmpanel"
	"github.com/XrayR-project/XrayR/api/sspanel"
	"github.com/XrayR-project/XrayR/api/v2board"
	"github.com/XrayR-project/XrayR/app/mydispatcher"
	_ "github.com/XrayR-project/XrayR/main/distro/all"
	"github.com/XrayR-project/XrayR/service"
	"github.com/XrayR-project/XrayR/service/controller"
	"github.com/r3labs/diff/v2"
	"github.com/xtls/xray-core/app/proxyman"
	"github.com/xtls/xray-core/app/stats"
	"github.com/xtls/xray-core/common/serial"
	"github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/infra/conf"
)

// Panel Structure
type Panel struct {
	access      sync.Mutex
	panelConfig *Config
	Server      *core.Instance
	Service     []service.Service
	Running     bool
}

func New(panelConfig *Config) *Panel {
	p := &Panel{panelConfig: panelConfig}
	return p
}

func (p *Panel) loadCore(panelConfig *Config) *core.Instance {
	// Log Config
	coreLogConfig := &conf.LogConfig{}
	defaultLogConfig := getDefaultLogConfig()
	if panelConfig.LogConfig != nil {
		if _, err := diff.Merge(defaultLogConfig, panelConfig.LogConfig, defaultLogConfig); err != nil {
			log.Panicf("Read Log config failed: %s", err)
		}
	}
	coreLogConfig.LogLevel = defaultLogConfig.Level
	coreLogConfig.AccessLog = defaultLogConfig.AccessPath
	coreLogConfig.ErrorLog = defaultLogConfig.ErrorPath

	// DNS config
	coreDnsConfig := &conf.DNSConfig{}
	if panelConfig.DnsConfigPath != "" {
		if data, err := io.ReadFile(panelConfig.DnsConfigPath); err != nil {
			log.Panicf("Failed to read dns.json at: %s", panelConfig.DnsConfigPath)
		} else {
			if err = json.Unmarshal(data, coreDnsConfig); err != nil {
				log.Panicf("Failed to unmarshal dns.json")
			}
		}
	}
	dConfig, err := coreDnsConfig.Build()
	if err != nil {
		log.Panicf("Failed to understand dns.json, Please check: https://xtls.github.io/config/base/dns/ for help: %s", err)
	}
	// Policy config
	levelPolicyConfig := parseConnectionConfig(panelConfig.ConnetionConfig)
	corePolicyConfig := &conf.PolicyConfig{}
	corePolicyConfig.Levels = map[uint32]*conf.Policy{0: levelPolicyConfig}
	pConfig, _ := corePolicyConfig.Build()
	config := &core.Config{
		App: []*serial.TypedMessage{
			serial.ToTypedMessage(coreLogConfig.Build()),
			serial.ToTypedMessage(&mydispatcher.Config{}),
			serial.ToTypedMessage(&stats.Config{}),
			serial.ToTypedMessage(&proxyman.InboundConfig{}),
			serial.ToTypedMessage(&proxyman.OutboundConfig{}),
			serial.ToTypedMessage(pConfig),
			serial.ToTypedMessage(dConfig),
		},
	}
	server, err := core.New(config)
	if err != nil {
		log.Panicf("failed to create instance: %s", err)
	}
	log.Printf("Xray Core Version: %s", core.Version())

	return server
}

// Start Start the panel
func (p *Panel) Start() {
	p.access.Lock()
	defer p.access.Unlock()
	log.Print("Start the panel..")
	// Load Core
	server := p.loadCore(p.panelConfig)
	if err := server.Start(); err != nil {
		log.Panicf("Failed to start instance: %s", err)
	}
	p.Server = server
	// Load Nodes config
	for _, nodeConfig := range p.panelConfig.NodesConfig {
		var apiClient api.API
		switch nodeConfig.PanelType {
		case "SSpanel":
			apiClient = sspanel.New(nodeConfig.ApiConfig)
		case "V2board":
			apiClient = v2board.New(nodeConfig.ApiConfig)
		case "PMpanel":
			apiClient = pmpanel.New(nodeConfig.ApiConfig)
		default:
			log.Panicf("Unsupport panel type: %s", nodeConfig.PanelType)
		}
		var controllerService service.Service
		// Regist controller service
		defaultControllerConfig := getDefaultControllerConfig()
		if nodeConfig.ControllerConfig != nil {
			if _, err := diff.Merge(defaultControllerConfig, nodeConfig.ControllerConfig, defaultControllerConfig); err != nil {
				log.Panicf("Read Controller Config Failed")
			}
		}
		controllerService = controller.New(server, apiClient, defaultControllerConfig)
		p.Service = append(p.Service, controllerService)

	}

	// Start all the service
	for _, s := range p.Service {
		err := s.Start()
		if err != nil {
			log.Panicf("Panel Start fialed: %s", err)
		}
	}
	p.Running = true
	return
}

// Close Close the panel
func (p *Panel) Close() {
	p.access.Lock()
	defer p.access.Unlock()
	for _, s := range p.Service {
		err := s.Close()
		if err != nil {
			log.Panicf("Panel Close fialed: %s", err)
		}
	}
	p.Server.Close()
	p.Running = false
	return
}

func parseConnectionConfig(c *ConnetionConfig) (policy *conf.Policy) {
	defaultConnetionConfig := getDefaultConnetionConfig()
	if c != nil {
		if _, err := diff.Merge(defaultConnetionConfig, c, defaultConnetionConfig); err != nil {
			log.Panicf("Read ConnetionConfig failed: %s", err)
		}
	}
	policy = &conf.Policy{
		StatsUserUplink:   true,
		StatsUserDownlink: true,
		Handshake:         &defaultConnetionConfig.Handshake,
		ConnectionIdle:    &defaultConnetionConfig.ConnIdle,
		UplinkOnly:        &defaultConnetionConfig.UplinkOnly,
		DownlinkOnly:      &defaultConnetionConfig.DownlinkOnly,
		BufferSize:        &defaultConnetionConfig.BufferSize,
	}

	return
}
