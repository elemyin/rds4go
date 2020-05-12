package main

import (
	"bytes"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/rds"
	"github.com/go-ini/ini"
	"github.com/jordan-wright/email"
	"html/template"
	"net/http"
	"net/smtp"
	"strings"
	"time"
)

var (
	regionId string
	accessKeyId string
	accessKeySecret string
	DBInstanceId string
	DBName string

	emailFrom string
	emailTo string
	emailAddr string
	emailUsername string
	emailPassword string
	emailHost string
)

type TemplateBody struct {
	DBName  string
	TotalRecordCount  int
	TRData  []rds.SQLSlowLog
}

func init() {
	cfg, err := ini.Load("config.ini")
	if err != nil {
		fmt.Print(err.Error())
		return
	}
	//rds配置
	rds, err := cfg.GetSection("rds")
	if err != nil {
		fmt.Print(err.Error())
		return
	}
	regionId = rds.Key("regionId").String()
	accessKeyId = rds.Key("accessKeyId").String()
	accessKeySecret = rds.Key("accessKeySecret").String()
	DBInstanceId = rds.Key("DBInstanceId").String()
	DBName = rds.Key("DBName").String()
	//email配置
	email, err := cfg.GetSection("email")
	if err != nil {
		fmt.Print(err.Error())
		return
	}
	emailFrom = email.Key("emailFrom").String()
	emailTo = email.Key("emailTo").String()
	emailAddr = email.Key("emailAddr").String()
	emailUsername = email.Key("emailUsername").String()
	emailPassword = email.Key("emailPassword").String()
	emailHost = email.Key("emailHost").String()
}

func main() {
	day, _ := time.ParseDuration("-24h")
	now := time.Now().Add(day).Format("2006-01-02")
	// 创建client实例
	client, err := rds.NewClientWithAccessKey(
		regionId,             // 您的地域ID
		accessKeyId,         // 您的AccessKey ID
		accessKeySecret)        // 您的AccessKey Secret
	if err != nil {
		fmt.Print(err.Error())
		return
	}
	// 创建慢日志统计请求并设置参数
	request := rds.CreateDescribeSlowLogsRequest()
	request.Scheme = "https"
	request.StartTime = now + "Z"
	request.EndTime = request.StartTime
	request.DBName = DBName
	//页码，范围：大于0且不超过Integer的最大值
	request.PageNumber = requests.NewInteger(1)
	//每天条数，范围：30~100
	request.PageSize = requests.NewInteger(100)

	//每个实例的慢查询统计，分别发送邮件
	DBInstanceIds := strings.Split(DBInstanceId, ",")
	for _, v := range DBInstanceIds {
		request.DBInstanceId = v
		response, err := client.DescribeSlowLogs(request)
		if err != nil {
			fmt.Print(err.Error())
			return
		}
		if response.GetHttpStatus() != http.StatusOK {
			fmt.Print(response.GetHttpStatus())
			return
		}
		if response.TotalRecordCount == 0 {
			fmt.Printf("实例：%s，数据库：%s，没有慢查询记录/n", v, DBName)
			continue
		}
		logs := response.Items.SQLSlowLog
		for i := 0; i < len(logs); i++ {
			logs[i].AvgExecutionTime = logs[i].MySQLTotalExecutionTimes / logs[i].MySQLTotalExecutionCounts
		}
		fmt.Printf("response is %#v\n", response)
		body := TemplateBody{DBName: DBName, TotalRecordCount:response.TotalRecordCount, TRData: logs}
		err = sendMail(now, v, body)
		if err != nil {
			fmt.Print(err.Error())
		}
	}
}

//发送邮件
func sendMail(dateStr, dbId string, data TemplateBody) error {
	e := email.NewEmail()
	e.From = emailFrom
	e.To = strings.Split(emailTo,",")
	//e.Bcc = []string{"test_bcc@example.com"}
	//e.Cc = []string{"test_cc@example.com"}
	e.Subject = dateStr + " RDS实例：" + dbId + " 慢查询统计"
	t := template.Must(template.ParseFiles("email.html"))
	body := new(bytes.Buffer)
	//作为变量传递给html模板
	t.Execute(body, data)
	e.HTML = body.Bytes()
	return e.Send(emailAddr, smtp.PlainAuth("", emailUsername, emailPassword, emailHost))
}