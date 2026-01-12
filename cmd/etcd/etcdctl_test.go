package etcd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"testing"
)

func TestEndpointStatus(t *testing.T) {
	testData := createHealthyETCDStatus()
	endpoints := []string{"https://192.168.50.11:2379", "https://192.168.50.10:2379", "https://192.168.50.12:2379"}
	memberIDsDec := []string{"2509054861951574500", "7258754974466672000", "7656230591208016000"}
	memberIDsHex, err := decimalToHex(memberIDsDec)
	if err != nil {
		t.Fatal(err)
	}

	// Mock a temporary file with test data
	tmpfile, err := os.Create("/tmp/endpoint_status.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	if _, err := tmpfile.Write([]byte(testData)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Redirect stdout to a buffer
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	EndpointStatus("/tmp/")
	w.Close()
	os.Stdout = old
	var output bytes.Buffer
	io.Copy(&output, r)
	r.Close()

	for _, endpoint := range endpoints {
		if !bytes.Contains(output.Bytes(), []byte(endpoint)) {
			t.Errorf("endpoint %q is missing from the output", endpoint)
		}
	}

	for _, memberIDHex := range memberIDsHex {
		if !bytes.Contains(output.Bytes(), []byte(memberIDHex)) {
			t.Errorf("member ID %q is missing from the output", memberIDHex)
		}
	}
}

func TestEndpointHealth(t *testing.T) {
	testData := createHealthyETCDHealth()
	endpoints := []string{"https://192.168.50.11:2379", "https://192.168.50.10:2379", "https://192.168.50.12:2379"}

	// Mock a temporary file with test data
	tmpfile, err := os.Create("/tmp/endpoint_health.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	if _, err := tmpfile.Write([]byte(testData)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Redirect stdout to a buffer
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	EndpointHealth("/tmp/")
	w.Close()
	os.Stdout = old
	var output bytes.Buffer
	io.Copy(&output, r)
	r.Close()

	for _, endpoint := range endpoints {
		if !bytes.Contains(output.Bytes(), []byte(endpoint)) {
			t.Errorf("endpoint %q is missing from the output", endpoint)
		}
	}
}

func decimalToHex(memberIDsDec []string) ([]string, error) {
	var memberIDsHex []string

	// Convert each decimal member ID to hexadecimal
	for _, dec := range memberIDsDec {
		decInt, err := strconv.ParseUint(dec, 10, 64)
		if err != nil {
			return nil, err
		}
		hexStr := fmt.Sprintf("%x", decInt)
		memberIDsHex = append(memberIDsHex, hexStr)
	}

	return memberIDsHex, nil
}

func createHealthyETCDStatus() string {
	testData := `[
        {
            "Endpoint": "https://192.168.50.11:2379",
            "Status": {
                "header": {
                    "cluster_id": 9124304915214735000,
                    "member_id": 2509054861951574500,
                    "revision": 139602,
                    "raft_term": 8
                },
                "version": "3.5.10",
                "dbSize": 83017728,
                "leader": 7258754974466672000,
                "raftIndex": 162279,
                "raftTerm": 8,
                "raftAppliedIndex": 162279,
                "dbSizeInUse": 58556416
            }
        },
        {
            "Endpoint": "https://192.168.50.10:2379",
            "Status": {
              "header": {
                "cluster_id": 9124304915214735000,
                "member_id": 7258754974466672000,
                "revision": 139602,
                "raft_term": 8
              },
              "version": "3.5.10",
              "dbSize": 82898944,
              "leader": 7258754974466672000,
              "raftIndex": 162279,
              "raftTerm": 8,
              "raftAppliedIndex": 162279,
              "dbSizeInUse": 58609664
            }
        },
        {
            "Endpoint": "https://192.168.50.12:2379",
            "Status": {
              "header": {
                "cluster_id": 9124304915214735000,
                "member_id": 7656230591208016000,
                "revision": 139602,
                "raft_term": 8
              },
              "version": "3.5.10",
              "dbSize": 82976768,
              "leader": 7258754974466672000,
              "raftIndex": 162279,
              "raftTerm": 8,
              "raftAppliedIndex": 162279,
              "dbSizeInUse": 58568704
            }
        }
    ]`
	return testData
}

func createHealthyETCDHealth() string {
	testData := `[
        {
            "endpoint": "https://192.168.50.10:2379",
            "health": true,
            "took": "12.628848ms"
        },
        {
            "endpoint": "https://192.168.50.11:2379",
            "health": true,
            "took": "11.552505ms"
        },
        {
            "endpoint": "https://192.168.50.12:2379",
            "health": true,
            "took": "12.666484ms"
        }
    ]`
	return testData
}
