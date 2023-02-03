package etcd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/olekukonko/tablewriter"
	etcdserverpb "go.etcd.io/etcd/api/v3/etcdserverpb"
)

type Endpoint struct {
	Endpoint string                      `json:"Endpoint"`
	Resp     etcdserverpb.StatusResponse `json:"Status"`
}

type epHealth struct {
	Ep     string `json:"endpoint"`
	Health bool   `json:"health"`
	Took   string `json:"took"`
	Error  string `json:"error,omitempty"`
}

func EndpointStatus(etcdFolderPath string) {
	_file, _ := ioutil.ReadFile(etcdFolderPath + "endpoint_status.json")
	var Endpoints []Endpoint
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
	_file, _ := ioutil.ReadFile(etcdFolderPath + "endpoint_health.json")
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
