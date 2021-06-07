package sspanel

import "encoding/json"

// NodeInfoResponse is the response of node
type NodeInfoResponse struct {
	Group           int     `json:"node_group"`
	Class           int     `json:"node_class"`
	SpeedLimit      float64  `json:"node_speedlimit"`
	TrafficRate     float64 `json:"traffic_rate"`
	MuOnly          int     `json:"mu_only"`
	Sort            int     `json:"sort"`
	RawServerString string  `json:"server"`
	Type            string  `json:"type"`
}

// UserResponse is the response of user
type UserResponse struct {
	ID            int    `json:"id"`
	Email         string `json:"email"`
	Passwd        string `json:"passwd"`
	Port          int    `json:"port"`
	Method        string `json:"method"`
	SpeedLimit    float64 `json:"node_speedlimit"`
	DeviceLimit   int    `json:"node_connector"`
	Protocol      string `json:"protocol"`
	ProtocolParam string `json:"protocol_param"`
	Obfs          string `json:"obfs"`
	ObfsParam     string `json:"obfs_param"`
	ForbiddenIP   string `json:"forbidden_ip"`
	ForbiddenPort string `json:"forbidden_port"`
	UUID          string `json:"uuid"`
	MultiUser     int    `json:"is_multi_user"`
}

// Response is the common response
type Response struct {
	Ret  uint            `json:"ret"`
	Data json.RawMessage `json:"data"`
}

// PostData is the data structure of post data
type PostData struct {
	Data interface{} `json:"data"`
}

// SystemLoad is the data structure of systemload
type SystemLoad struct {
	Uptime string `json:"uptime"`
	Load   string `json:"load"`
}

// OnlineUser is the data structure of online user
type OnlineUser struct {
	UID int    `json:"user_id"`
	IP  string `json:"ip"`
}

// UserTraffic is the data structure of traffic
type UserTraffic struct {
	UID      int   `json:"user_id"`
	Upload   int64 `json:"u"`
	Download int64 `json:"d"`
}

type RuleItem struct {
	ID      int    `json:"id"`
	Content string `json:"regex"`
}

type IllegalItem struct {
	ID  int `json:"list_id"`
	UID int `json:"user_id"`
}
