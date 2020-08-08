package client

import (
	"ddns/common"
	"errors"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"
	"strings"
)

func Aliyun(ipAddr string) (err error) {
	ayc := AliyunConf{}
	// 获取配置
	err = common.LoadAndUnmarshal(ConfPath+"/aliyun.json", &ayc)
	if err != nil {
		return
	}

	if ayc.AccessKeyId == "" || ayc.AccessKeySecret == "" || ayc.Domain == "" || ayc.SubDomain == "" {
		err = errors.New("请打开配置文件 " + ConfPath + "/aliyun.json 核对你的 accesskey_id, accesskey_secret, domain, sub_domain 并重新启动")
		return
	}

	recordId, _, recordIP, err := ayc.GetParseRecordId()
	if err != nil {
		return
	}
	recordType := ""
	if strings.Contains(ipAddr, ":") {
		recordType = "AAAA"
	} else {
		recordType = "A"
	}
	if recordId != ayc.RecordId || recordType != ayc.RecordType {
		ayc.RecordId = recordId
		ayc.RecordType = recordType
		err = common.MarshalAndSave(ayc, ConfPath+"/aliyun.json")
		if err != nil {
			return
		}
	}
	if recordIP == ipAddr {
		err = errors.New("阿里云记录的 IP 和当前获取的 IP 一致")
		return
	}

	err = ayc.UpdateParseRecord(ipAddr)
	if err != nil {
		return
	}
	return
}

func (ayc AliyunConf) GetParseRecordId() (recordId, recordType, recordIP string, err error) {
	client, err := alidns.NewClientWithAccessKey("cn-hangzhou", ayc.AccessKeyId, ayc.AccessKeySecret)
	if err != nil {
		return
	}

	request := alidns.CreateDescribeDomainRecordsRequest()
	request.Scheme = "https"

	request.DomainName = ayc.Domain

	response, err := client.DescribeDomainRecords(request)
	if err != nil {
		return
	}

	for i := range response.DomainRecords.Record {
		if response.DomainRecords.Record[i].RR == ayc.SubDomain {
			recordId = response.DomainRecords.Record[i].RecordId
			recordType = response.DomainRecords.Record[i].Type
			recordIP = response.DomainRecords.Record[i].Value
			break
		}
	}
	if recordId == "" || recordType == "" || recordIP == "" {
		err = errors.New("解析记录不存在")
	}
	return
}

func (ayc AliyunConf) UpdateParseRecord(ipAddr string) (err error) {
	client, err := alidns.NewClientWithAccessKey("cn-hangzhou", ayc.AccessKeyId, ayc.AccessKeySecret)
	if err != nil {
		return
	}

	request := alidns.CreateUpdateDomainRecordRequest()
	request.Scheme = "https"

	request.RecordId = ayc.RecordId
	request.RR = ayc.SubDomain
	request.Type = ayc.RecordType
	request.Value = ipAddr

	_, err = client.UpdateDomainRecord(request)
	if err != nil {
		return
	}
	return
}
