package main
import (
  "time"
  "strconv"
  "image/png"
  "grafana/pkg/client"
  "grafana/pkg/notification"
  "grafana/pkg/configer"
)

var (
  alertDict = map[int] alertinfo{}
  dbInfo    = map[string] map[string]string{}
  timeSeries     = map[int][]time.Time {}
  alertingNum    = 0
)

type alertinfo struct {
  Name           string
  AlertMetrics    []string
  AlertValues     []interface{}
  PanelId        int
  OrgId          int
  DbUid          string
  DbSlug         string
  Frequency      int
  TempVar        map[string] string
}


func run() error{
  ticker := time.NewTicker(30 * time.Second)
  for _ = range ticker.C{

  }
}

func Alerting()error{
//Recovery alert item
   for alerId,alertV := range(alertDict){
       alert_info,err := client.GetAlert(strconv.Itoa(alertId))
       if err != nil {
          info.Println(err)
          return err
       }
       if *alert_info.State == "ok"{
           notification.Emit(alertId,alertV)
       }
   }

// Get alert list
   data,err := client.GetAlerts()
   if err != nil{
     info.Println(err)
     return err
   }
   for _,alert := range(data){
      if alert.State == "alerting"{
         if _,ok := alertDict[alert.Id];ok{
            continue
         }
         m := alertinfo{}
         m.Name = alert.Name
         m.PanelId = alert.PanelId
         m.DbUid  =  alert.DashboardUid
         m.DbSlug =  alert.DashboardSlug
//Scan the alerting data
         for _,v = range(alert.EvalData.EvalMatches){
            m.AlertMetrics = append(m.AlertMetrics,v.Metric)
            m.AlertValues = append(m.AlertValues,v.Value)
         }
//Get alert item detail info
         alert_info,err := client.GetAlert(strconv.Itoa(alert.Id))
         if err != nil{
             info.Println("Get alert error: ",err)
         }
         m.OrgId = *alert_info.OrgId
         m.Frequency = *alert_info.Frequency

         if k,ok := dbInfo[m.DbUid];ok{
            m.TmpVar = k
         }else{
// Get Dashboard templating variables for render image url
            db,err := client.GetDashboard(m.DbUid)
            if err != nil{
               info.Println("Get Dashboard error: ",err)
            }

            s := map[string]string{}
            Temvars = *db.Dashboard.Templating["list"]
            for _,var := range(Temvars){
               if var.Current.Selected{
                  s[var.Name] = var.Current.Text
               }
            }
            dbInfo[m.DbUid] = s
            m.TempVar = s
         }
         alertDict[alert.Id] = m
         alertNum = len(alertDict)
         RenderImage(m)
         notification.Emit(m)
      }
   }
}


func RenderImage(m alertinfo){
//Generator time series for render image
   t1 := int(time.Now().Unix())*1000
   t2 := t1-3600000
   grafana_conf := configer.ConfigParse()

   uri,err := url.Parse(grafana_conf.Grafana_uri)
   if err != nil{
     info.Println("URL parse error: ",err)
   }
   token := "Bearer " + grafana_conf.Grafana_token
   uri.Path = "/render/d-solo/" + m.DbUid + "/" + m.DbSlug
   req, err := http.NewRequest("GET", uri.String(),nil)
   if err != nil{
     info.Println("request generator error: ",err)
   }
// request header add 
   req.Header.Add("Authorization", token)
   req.Header.Add("Content-Type", "application/json")
   req.Header.Add("Accept", "application/json")
// query data add
   q := req.URL.Query()
   q.Add("orgid", strconv.Itoa(m.OrgId))
   q.Add("from", strconv.Itoa(t2))
   q.Add("to", strconv.Itoa(t1))
   q.Add("panelId", strconv.Itoa(m.PanelId))
   q.Add("width", "1000")
   q.Add("height", "500")
   for k,v := range(m.TempVar){
       q.Add(k, v)
   }
   req.URL.RawQuery = q.Encode()
   info.Println(req.URL.String())
//request
   resp, err := c.client.Do(req)
   if err != nil {
           fmt.Println(err)
   }
   defer resp.Body.Close()
   reader := bufio.NewReader(resp.Body)
   file,err := os.Create("test.png")
   if err != nil {
      info.Println("Create error: ",err)
   }
   n,err := reader.WriteTo(file)
   if err != nil{
     info.Println("Reader write error: ",err)
   }
   info.Println(n)
}
