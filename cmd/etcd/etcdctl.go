package etcd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/olekukonko/tablewriter"
	etcdserverpb "go.etcd.io/etcd/api/v3/etcdserverpb"
)

type epStatus struct {
	Endpoint string                      `json:"Endpoint"`
	Resp     etcdserverpb.StatusResponse `json:"Status"`
}

var alarmType_name = map[int32]string{
	0: "NONE",
	1: "NOSPACE",
	2: "CORRUPT",
}

type alarm struct {
	MemberID uint64 `json:"memberID,omitempty"`
	Alarm    int32  `json:"alarm,omitempty"`
}

type alarmList struct {
	Alarms []alarm `json:"alarms,omitempty"`
}

type epHealth struct {
	Ep     string `json:"endpoint"`
	Health bool   `json:"health"`
	Took   string `json:"took"`
	Error  string `json:"error,omitempty"`
}

type member struct {
	ID   uint64 `json:"id"` // from https://github.com/etcd-io/etcd/blob/main/client/pkg/types/id.go#L25C9-L25C15
	Name string `json:"name,omitempty"`
	//Status is not a field but just "started" unles Name is zero-length as per https://github.com/etcd-io/etcd/blob/4601818f511478980725a215e814e56fb8ee31ef/etcdctl/ctlv3/command/printer.go#L188-L191
	ClientURLs []string `json:"clientURLs,omitempty"`
	PeerURLs   []string `json:"peerURLs"`
	IsLearner  bool     `json:"isLearner,omitempty"`
}

type memberList struct {
	Members []member `json:"members"`
}

func EndpointStatus(etcdFolderPath string) {
	_file, _ := os.ReadFile(etcdFolderPath + "endpoint_status.json")
	var Endpoints []epStatus
	if err := json.Unmarshal([]byte(_file), &Endpoints); err != nil {
		fmt.Fprintln(os.Stderr, "Error when trying to unmarshal file \""+etcdFolderPath+"endpoint_status.json\":", err.Error())
		os.Exit(1)
	}
	var rows [][]string
	var hdr = []string{"endpoint", "ID", "version", "db size/in use", "not used", "is leader", "is learner", "raft term",
		"raft index", "raft applied index", "errors"}
	for _, status := range Endpoints {
		rows = append(rows, []string{
			status.Endpoint,
			fmt.Sprintf("%x", status.Resp.Header.MemberId),
			status.Resp.Version,
			humanize.Bytes(uint64(status.Resp.DbSize)) + "/" + humanize.Bytes(uint64(status.Resp.DbSizeInUse)),
			fmt.Sprint(100-(status.Resp.DbSizeInUse*100/status.Resp.DbSize)) + "%",
			fmt.Sprint(status.Resp.Leader == status.Resp.Header.MemberId),
			fmt.Sprint(status.Resp.IsLearner),
			fmt.Sprint(status.Resp.RaftTerm),
			fmt.Sprint(status.Resp.RaftIndex),
			fmt.Sprint(status.Resp.RaftAppliedIndex),
			fmt.Sprint(strings.Join(status.Resp.Errors, ", ")),
		})
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(hdr)
	table.AppendBulk(rows)
	table.Render()
}

func EndpointHealth(etcdFolderPath string) {
	_file, _ := os.ReadFile(etcdFolderPath + "endpoint_health.json")
	var healthList []epHealth
	if err := json.Unmarshal([]byte(_file), &healthList); err != nil {
		fmt.Fprintln(os.Stderr, "Error when trying to unmarshal file \""+etcdFolderPath+"endpoint_status.json\":", err.Error())
		os.Exit(1)
	}
	var rows [][]string
	var hdr = []string{"endpoint", "health", "took", "error"}
	for _, h := range healthList {
		rows = append(rows, []string{
			h.Ep,
			fmt.Sprintf("%v", h.Health),
			h.Took,
			h.Error,
		})
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(hdr)
	table.AppendBulk(rows)
	table.Render()
}

func AlarmList(etcdFolderPath string) {
	_file, _ := os.ReadFile(etcdFolderPath + "alarm_list.json")
	var reportedAlarmList alarmList
	if err := json.Unmarshal([]byte(_file), &reportedAlarmList); err != nil {
		fmt.Fprintln(os.Stderr, "Error when trying to unmarshal file \""+etcdFolderPath+"alarm_list.json\":", err.Error())
		os.Exit(1)
	}
	for _, reportedAlarm := range reportedAlarmList.Alarms {
		memberAlarm := alarmType_name[reportedAlarm.Alarm]
		if memberAlarm == "" {
			fmt.Fprintf(os.Stderr, "Error when trying to unmarshal file \""+etcdFolderPath+"alarm_list.json\": Member %d shows invalid alert %d\n", reportedAlarm.MemberID, reportedAlarm.Alarm)
			os.Exit(1)
		}
		fmt.Printf("memberID:%d alarm:%s\n", reportedAlarm.MemberID, memberAlarm)
	}

}

func MemberList(etcdFolderPath string) {
	_file, _ := os.ReadFile(etcdFolderPath + "member_list.json")
	var memberList memberList
	if err := json.Unmarshal([]byte(_file), &memberList); err != nil {
		fmt.Fprintln(os.Stderr, "Error when trying to unmarshal file \""+etcdFolderPath+"member_list.json\":", err.Error())
		os.Exit(1)
	}
	var rows [][]string
	var hdr = []string{"ID", "status", "name", "peer addrs", "client addrs", "is learner"}
	for _, m := range memberList.Members {
		status := "started"
		if len(m.Name) == 0 {
			status = "unstarted"
		}
		isLearner := "false"
		if m.IsLearner {
			isLearner = "true"
		}
		rows = append(rows, []string{
			fmt.Sprintf("%x", m.ID),
			status,
			m.Name,
			strings.Join(m.PeerURLs, ","),
			strings.Join(m.ClientURLs, ","),
			isLearner,
		})
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(hdr)
	table.AppendBulk(rows)
	table.Render()
}
