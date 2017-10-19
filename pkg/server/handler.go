package server

import (
	"encoding/base64"
	"html/template"
	"github.com/golang/glog"
	"bytes"
	"fmt"
	"time"
	"path/filepath"
	"net/http"
	"strings"
)


var (
	htmlHeadTemplate string = `
	<html><head><title>{{.PageTitle}}</title></head><boday><center>
	<h1>{{.PageHead}}</h1>
	<hr width="50%">
	`

	htmlFootTemplate string = `
	<hr width="50%">hostName:  {{.HostName}}
	<br/>
	hostIP: {{.HostIP}}
	<br/>
	ClientIP: {{.ClientIP}}
	<br/>
	OriginalClient: {{.OriginalClient}}
	</center></body></html>
	`

	tableImgTemplate string = `
	<table>
	  <tr><td><img style="width:250px;height:260px" src="data:image/jpg;base64,{{.Image}}"></td></tr>
	  <tr><td align="center">{{.ImageName}}</td></tr>
	 </table>`


	imageTemplate string = `<!DOCTYPE html>
<html lang="en"><head></head>
<body><center><img src="data:image/jpg;base64,{{.Image}}"></center></body>`

	smallImageTemplate string = `<!DOCTYPE html>
<html lang="en"><head></head>
<body><center><table>
	  <tr><td><img style="width:200px;height:180px" src="data:image/jpg;base64,{{.Image}}"></td></tr>
	  <tr><td>{{.ImageName}}</td></tr></table>
      </center></body>`
)

func GetSimpleHtml() string {
	head := "<html><head><title> Welcome to InceptionServer</title></head>"
	body := "<body><center><p style='font-size:30px'> It works. </p></center></body>"
	return head + body
}

func getHead(title string, head string) (string, error) {
	tmp, err := template.New("head").Parse(htmlHeadTemplate)
	if err != nil {
		glog.Errorf("Failed to parse image template %v:%v", imageTemplate, err)
		return "", fmt.Errorf("parse failed")
	}

	var result bytes.Buffer
	data := map[string]interface{}{"PageTitle": title, "PageHead": head}
	if err := tmp.Execute(&result, data); err != nil {
		glog.Errorf("Faile to execute template: %v", err)
		return "", fmt.Errorf("execute failed.")
	}

	return result.String(), nil
}

func getImgTable(fpath string, img []byte) string {
	str := base64.StdEncoding.EncodeToString(img)
	tmp, err := template.New("image").Parse(tableImgTemplate)
	if err != nil {
		glog.Errorf("Failed to parse image template %v:%v", imageTemplate, err)
		return ""
	}

	var table bytes.Buffer
	fname := filepath.Base(fpath)
	data := map[string]interface{}{"Image": str, "ImageName": fname}
	if err := tmp.Execute(&table, data); err != nil {
		glog.Errorf("Faile to execute template: %v", err)
		return ""
	}

	return table.String()
}

func getClientIP(r *http.Request) string {
	return r.RemoteAddr
}

func getOriginalClientInfo(r *http.Request) string {
	orig := r.Header.Get("X-Forwarded-For")
	glog.V(3).Infof("request from %v, %v", r.RemoteAddr, orig)

	if len(orig) > 0 {
		ips := strings.Split(orig, ", ")
		return ips[0]
	}

	return ""
}

func GetImgHtml(fname string, img []byte, predict, foot string, begin time.Time) string {
	head, err := getHead("ShowImage", "Image details")
	if err != nil {
		glog.Errorf("Failed to get head: %v", err)
		return ""
	}

	table := getImgTable(fname, img)
	if table == "" {
		glog.Errorf("Failed to get image.")
		return ""
	}

	du := fmt.Sprintf("%5.2f ms", time.Since(begin).Seconds() * 1000)
	prediction := fmt.Sprintf("<table>%v</table><br>RespTime: %v", predict, du)

	result := fmt.Sprintf("%v %v %v %v", head, table, prediction, foot)

	return result
}