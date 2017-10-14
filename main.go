package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/coreos/go-systemd/daemon"
)

func getExternalIPMethod1() string {
	out, err := exec.Command("/usr/bin/dig", "+short", "myip.opendns.com", "@resolver1.opendns.com").Output()
	if err != nil {
		log.Printf("Error get my external IP Address - Method #1 dig: %s", err)
		return ""
	}
	return strings.TrimSpace(string(out))
}
func getExternalIPMethod2() string {
	resp, err := http.Get("http://ipinfo.io/ip")
	if err != nil {
		log.Printf("Error get my external IP Address - Method #2 ipinfo: %s", err)
		return ""
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error get my external IP Address - Method #2 ipinfo: %s", err)
		return ""
	}
	return strings.TrimSpace(string(body))

}
func getMyExternalIP() string {

	var ip string
	ip = getExternalIPMethod1()
	if ip != "" {
		return ip
	}
	ip = getExternalIPMethod2()
	return ip

}

func setIPinAWS(recordSetName string, ip string, zoneID string) bool {
	svc := route53.New(session.New())
	input := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String("UPSERT"),
					ResourceRecordSet: &route53.ResourceRecordSet{
						ResourceRecords: []*route53.ResourceRecord{
							{
								Value: aws.String(ip),
							},
						},
						Name: aws.String(recordSetName),
						TTL:  aws.Int64(300),
						Type: aws.String("A"),
					},
				},
			},
			Comment: aws.String("Update of AWSmyIP"),
		},
		HostedZoneId: aws.String(zoneID),
	}
	result, err := svc.ChangeResourceRecordSets(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case route53.ErrCodeNoSuchHostedZone:
				log.Println(route53.ErrCodeNoSuchHostedZone, aerr.Error())
			case route53.ErrCodeNoSuchHealthCheck:
				log.Println(route53.ErrCodeNoSuchHealthCheck, aerr.Error())
			case route53.ErrCodeInvalidChangeBatch:
				log.Println(route53.ErrCodeInvalidChangeBatch, aerr.Error())
			case route53.ErrCodeInvalidInput:
				log.Println(route53.ErrCodeInvalidInput, aerr.Error())
			case route53.ErrCodePriorRequestNotComplete:
				log.Println(route53.ErrCodePriorRequestNotComplete, aerr.Error())
			default:
				log.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			log.Println(err.Error())
		}
		return false
	}
	return *result.ChangeInfo.Status == route53.ChangeStatusPending
}

var zoneID *string
var recordSet *string
var noDaemon *bool
var lastIP string

const intervalSleep = 15 * time.Minute

func init() {

	zoneID = flag.String("zoneID", "", "Zone ID")
	recordSet = flag.String("recordSet", "", "Record Set")
	noDaemon = flag.Bool("noDaemon", false, "No daemon flag")
}

func main() {
	log.Println("AWS Auto-update IP Remote Address")
	flag.Parse()
	if *zoneID == "" {
		log.Fatal("Need specify: zoneID")
		return
	}
	if *recordSet == "" {
		log.Fatal("Need specify: recordSet")
		return
	}
	if !(*noDaemon) {
		daemon.SdNotify(false, "READY=1")
		go func() {
			for {
				daemon.SdNotify(false, "WATCHDOG=1")
				time.Sleep(10 * time.Second)
			}
		}()
	}

	for {
		ip := getMyExternalIP()
		if ip == "" {
			log.Printf("IP empty, nothing to update")
			time.Sleep(intervalSleep)
			continue
		}
		log.Printf("IP: %s", ip)
		if ip == lastIP {
			log.Print("Same IP, nothing to do")
			time.Sleep(intervalSleep)
			continue
		}
		if !setIPinAWS(*recordSet, ip, *zoneID) {
			log.Print("Error, setting IP")
			time.Sleep(intervalSleep)
			continue
		}
		lastIP = ip
		time.Sleep(intervalSleep)
	}
}
