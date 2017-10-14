package main

import (
	"flag"
	"log"
	"os/exec"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
)

func getMyExternalIP() string {
	out, err := exec.Command("/usr/bin/dig", "+short", "myip.opendns.com", "@resolver1.opendns.com").Output()
	if err != nil {
		log.Printf("Error get my external IP Address: %s", err)
		return ""
	}
	return strings.TrimSpace(string(out))
}

func setIPinAWS(recordSetName string, ip string, zoneID string) {
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
		return
	}
	log.Println(result)

}

var zoneID *string
var recordSet *string

func init() {

	zoneID = flag.String("zoneID", "", "Zone ID")
	recordSet = flag.String("recordSet", "", "Record Set")
}

func main() {
	log.Println("AWS Auto-update IP Remote Address")
	flag.Parse()
	ip := getMyExternalIP()
	log.Printf("IP: %s", ip)
	setIPinAWS(*recordSet, ip, *zoneID)
}
