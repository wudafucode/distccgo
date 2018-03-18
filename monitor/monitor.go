package monitor

import (
    "fmt"
    "time"
    "flag"
    "log"
    "net/http"
    //"net/url"
    "io/ioutil"
    "encoding/json"
    "os"

)
var (
    masternode string
    MFlag     flag.FlagSet
)
type Monitor struct{
    workernodes [] string
}
func init() {
    MFlag.StringVar(&masternode, "masternode", "localhost:4001", "masternode")
   
}
func RunMonitor(argvs []string) bool {
    mon:=Monitor{
        workernodes: make([]string,10),
    }
    MFlag.Parse(argvs)
    log.Printf("local monitor running,masternode:%s",masternode)
    go mon.GetNode(masternode)
	return true
}
func Get(url string) ([]byte, error) {
    client := &http.Client{}
    r, err := client.Get(url)
    if err != nil {
        return nil, err
    }
    defer r.Body.Close()
    b, err := ioutil.ReadAll(r.Body)
    if r.StatusCode >= 400 {
        return nil, fmt.Errorf("%s: %s", url, r.Status)
    }
    if err != nil {
        return nil, err
    }
    return b, nil
}
func (m *Monitor)GetWorker(url string)([]string,error){
    ret := make([]string,10)
    jsonBlob, err := Get(url)
    if err != nil {
        return nil,err
    }
    err = json.Unmarshal(jsonBlob, &ret)
    if err != nil {
        return nil,err
    }
    return ret,nil
}

func (m *Monitor)updateInfo(nodes []string) bool{
   
    log.Printf("get worker:%d",len(nodes))
    m.workernodes = nodes
    var ret string
    for _,v:=range(m.workernodes){
    
        ret=ret+v+" "
    }
    err:=os.Setenv("DISTCCGO_HOSTS",ret)
    if err!=nil{
        log.Println("error",err)
    }

    return true
}
func (m *Monitor)GetNode(masternode string) bool {
	tickChan := time.Tick(time.Duration(5)*time.Second)
	doneChan := make(chan error, 1)
    for{
    	 select {
            case <-tickChan:
                //todo add all the server list
                ret,_:=m.GetWorker("http://" + masternode + "/worker/status")
                if(ret == nil){
                    continue
                }
                m.updateInfo(ret)
                
        	case _ = <-doneChan:
            	log.Println("done")
            	return true
        }


    }
}


