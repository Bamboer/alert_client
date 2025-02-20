package notification

import (
        "time"
        "strings"
        "bytes"
        "encoding/json"
        "fmt"
        "grafana/pkg/client"
        "grafana/pkg/configer"
        "net/http"
)
/*
func init() {
        SNS["dingding"] = DSend
}*/

var (
        reminders []string
        gr        interface{}
)

type dingding struct {
        dApi   string
        client *http.Client
}

func Newdingding(url string) *dingding {
        return &dingding{
                dApi:   url,
                client: &http.Client{},
        }
}

func DSend(state string, msg client.SimpleInfo, b []byte) error {
        conf := configer.ConfigParse()
        dclient := Newdingding(conf.Dingding)
        if err := dclient.Send(state, msg, b); err != nil {
                info.Println(err)
                return err
        }
        return nil
}
func Text(msg string) error {
        conf := configer.ConfigParse()
        dclient := Newdingding(conf.Dingding)
        if err := dclient.SendText(msg); err != nil {
                info.Println(err)
                return err
        }
        return nil
}

func (d *dingding) SendText(msg string) error {
        data := make(map[string]interface{})
        data["msgtype"] = "text"
        data["at"] = map[string]interface{}{"atMobiles": reminders, "isAtAll": true}
        data["text"] = map[string]string{"content": msg}

        mdata, err := json.Marshal(data)
        if err != nil {
                info.Println("Marshal error: ", err)
                return err
        }
        reader := bytes.NewReader(mdata)
        req, err := http.NewRequest("POST", d.dApi, reader)
        req.Header.Set("Content-Type", "application/json; charset=utf-8")
        resp, err := d.client.Do(req)
        defer resp.Body.Close()
        if err != nil {
                info.Println("err: ", err)
                return err
        }
        err = json.NewDecoder(resp.Body).Decode(&gr)
        if err != nil {
                info.Println("err: ", err)
                return err
        }
        info.Println(gr)
        return nil
}

func (d *dingding) Send(state string, msg client.SimpleInfo, b []byte) error {
        // state: alert status
        // msg: send message body
        // b: png format image

        data := make(map[string]interface{})
        data["msgtype"] = "markdown"
        data["at"] = map[string]interface{}{"atMobiles": reminders, "isAtAll": true}
        data["markdown"] = d.RenderMsg(state, msg)

        mdata, err := json.Marshal(data)
        if err != nil {
                info.Println("Marshal error: ", err)
                return err
        }
        reader := bytes.NewReader(mdata)
        req, err := http.NewRequest("POST", d.dApi, reader)
        if err != nil{
           info.Println(err)
        }
        req.Header.Set("Content-Type", "application/json; charset=utf-8")
        resp, err := d.client.Do(req)
        defer resp.Body.Close()
        if err != nil {
                info.Println("err: ", err)
                return err
        }
        err = json.NewDecoder(resp.Body).Decode(&gr)
        if err != nil {
                info.Println("err: ", err)
                return err
        }
        info.Println("dingding send result: ",gr)
        return nil
}

func (d *dingding) RenderMsg(state string, msg client.SimpleInfo) map[string]string {
        var content map[string]string
        if state == "alerting" {
                content = map[string]string{"title": "Alarm", "text": fmt.Sprintf("### Alarm: %s \n\n> 1.Metric: %s\n\n> 2.Value: %v\n\n> 3.Dashboard: %s \n\n> 4.AlertingNum: %d\n\n> ![screenshot](%s)\n> ###### %s UTC发布[详情](%s) \n", msg.Name, msg.AlertMetrics, msg.AlertValues, msg.DbSlug, *msg.AlertNum, msg.RenderURL,strings.Split(time.Now().UTC().String(),".")[0], msg.RenderURL)}
        } else if state == "ok" {
                content = map[string]string{"title": "Recovery", "text": fmt.Sprintf("### Alarm: %s Recovery !\n\n> 1.Metric: %s\n\n> 2.Value: %v\n\n> 3.Dashboard: %s\n\n> 4.AlertingNum: %d\n\n> ![screenshot](%s)\n>###### %s UTC发布[详情](%s)\n", msg.Name, msg.AlertMetrics, msg.AlertValues, msg.DbSlug, *msg.AlertNum, msg.ImgURL, strings.Split(time.Now().UTC().String(),".")[0], msg.ImgURL)}
        }
        info.Println("dinigding render message: ",content)
        return content
}
