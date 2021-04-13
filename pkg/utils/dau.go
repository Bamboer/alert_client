package main

import(
  "fmt"
  "time"
  "context"
  "net/http"
  "net/url"
  "io/ioutil"
  "encoding/json"
  "path/filepath"
  "html/template"
  "grafana/pkg/configer"
  "github.com/aws/aws-sdk-go-v2/aws"
  "github.com/aws/aws-sdk-go-v2/config"
  "github.com/aws/aws-sdk-go-v2/service/cloudwatch"
  "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)


var(
  GraphiteURL = "http://10.50.24.197:7001/render"
  User = "8bf3584d"
  Pwd  = "6f30e011"
  _,wk := time.Now().ISOWeek()
  DReport =  DailyReport{
         Timer: time.Now(),
         WK:    wk,
    WeekDay: DayData{},
}
)

type Tres struct{
 Datapoints [][]float64
 Target  string
}

type DailyReport struct{
  Timer  time.Time
  WK     int
  WeekDay map[int]DayData
}

type DayData struct{
  Access  int
  Health  int
}

func DAU(b bytes.Buffer)error{
  conf := configer.ConfigParse()
  tpPath := conf.DauTpPath
  elb := conf.AWSELBName
  region := conf.AWSRegion
  absPath,err := filepath.Abs(tpPath)
  if err != nil{
     fmt.Println(err)
  }
  tp,err := template.ParseFiles(absPath)
  if err != nil{
     fmt.Println(err)
  }
  access,err := Access()
  if err != nil{
    fmt.Println(err)
  }
  health,err := Health()
  if err != nil{
    fmt.Println(err)
  }
  t := time.Now()
  td := t.Weekday()
  tnow := int(time.Date(t.Year(),t.Month(),t.Day(),0,0,0,0,time.UTC).Unix())
  for i := 1;i <= int(td);i++{
    t1 := tnow - i*86400
    
  }
  
}

func Access(region,elb string)(map[int]int,error){
  data := map[int]int{}
  t := time.Now()
  st := time.Date(t.Year(),t.Month(),t.Day()-6,0,0,0,0,time.Local)
  et := time.Date(t.Year(),t.Month(),t.Day(),0,0,0,0,time.Local)
  cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
  if err != nil {
          fmt.Println("unable to load SDK config, %v", err)
          return data,err
  }
  client := cloudwatch.NewFromConfig(cfg)
  input := &cloudwatch.GetMetricStatisticsInput{
      StartTime : aws.Time(st),
      EndTime : aws.Time(et),
      MetricName: aws.String("RequestCount"),
      Namespace: aws.String("AWS/ELB"),
      Period: aws.Int32(86400),
      Dimensions: []types.Dimension{{Name: aws.String("LoadBalancerName"),Value: aws.String(elb)}},
      Statistics: []types.Statistic{types.StatisticSum},
  }
  output,err := client.GetMetricStatistics(context.Background(),input)
  if err != nil {
    return data,err
  }
  for _,v := range output.Datapoints{
    data[int((*v.Timestamp).Unix())] = int(*v.Sum)
  }
  //fmt.Println(time.Unix(int64(1.617408e+09),0))
  return data,nil
}

func Health()(map[int]int,error){
  gr := []Tres{}
  data := map[int]int{}
  c,_ := NewRender(GraphiteURL,User,Pwd)
  req,err := http.NewRequest("GET",c.uri.String(),nil)
  if err != nil{
    return data,err
  }
  req.SetBasicAuth(c.user,c.password)

  q := req.URL.Query()
  q.Add("target",`alias(summarize(averageSeries(ec2-cn-north-1-svoice-idg-rel.timers.application.dummy-client.*.*vdt.health.mean, *), "1d", "avg", false), "Overall ")`)
  q.Add("from","-144hours")
  q.Add("format","json")
  req.URL.RawQuery = q.Encode()
//  fmt.Println("Render url: ", req.URL.String())
  resp, err := c.client.Do(req)
  if err != nil {
      return data,err
  }
  defer resp.Body.Close()

  b,err := ioutil.ReadAll(resp.Body)
  if err != nil{
      return data,err
  }
  if err := json.Unmarshal(b,&gr);err != nil{
      return data,err
  }
  for _, v := range gr[0].Datapoints{
      data[int(v[1])] = int(v[0])
  }
  return data,nil
//  t := time.Unix(int64(1.617408e+09),0)
}


type Render struct{
  uri    *url.URL
  user   string
  password string
  client *http.Client
}

func NewRender(uri, user,password string)(*Render,error){
   url,err := url.Parse(uri)
   if err != nil{
     fmt.Println("Error: ",err)
     return nil,err
   }
   return &Render{
       uri : url,
       user: user,
       password: password,
       client: &http.Client{},
   },nil
}

