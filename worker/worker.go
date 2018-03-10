package worker

import (
    "fmt"
    "time"
    "flag"
    "log"
    "golang.org/x/net/context"
    "google.golang.org/grpc"
    "math/rand"
    "../pb"
)

type Worker struct {
	workername        string
	masternode        string
    servernodes       []string
    localip           string
    leader            string
}
type WorkerOption struct {
    masternode    string
    host          string   
}


func (s *Worker) connectionString() string {
	return fmt.Sprintf("http://%s", s.masternode)
}
var WFlag flag.FlagSet
var Woption   WorkerOption
func init() {
      WFlag.StringVar(&Woption.masternode, "masternode", "localhost:4001", "masternode")
      WFlag.StringVar(&Woption.host, "h", "localhost", "hostname")
}

func NewWorker(workername string,masternode string,localip string)*Worker{
    wk :=&Worker{
    	workername:workername,
    	masternode:masternode,
        localip:localip,
    }
    return wk
}
func RunWorker(argvs []string) bool {

    var workFlag flag.FlagSet
    log.Print("worker running")


    workFlag.Parse(argvs)
    
    name:= fmt.Sprintf("%07x", rand.Int())[0:7]
    wk :=NewWorker(name,Woption.masternode,Woption.host)

    go wk.heartbeat()
    wk.Dameon()
    return true 
}
func (wk *Worker) heartbeat(){
    time.Sleep(time.Duration(10)*time.Second) 
    for{

        wk.doheartbeat()
        time.Sleep(time.Duration(10)*time.Second) 

    }
}
func (wk *Worker) doheartbeat(){
    log.Println("once")
    grpcConection, err := grpc.Dial(wk.masternode, grpc.WithInsecure())
    if err != nil {
        fmt.Errorf("fail to dial: %v", err)
        return 
    }
    defer grpcConection.Close()
    client := pb.NewMsgClient(grpcConection)
    stream, err := client.SendHeartbeat(context.Background())
    if err != nil {
        log.Println("send error",err)
        return 
    }
  

    doneChan := make(chan error, 1)

    go func() {
        for{
            in, err := stream.Recv()
            if err != nil {
                log.Println("recv error",err)
                doneChan <- err
                return
            }
            //log.Println("worker recv ")
            if len(in.GetServernode()) !=0 {
                wk.UpdateServerNode(in.GetServernode())
            }
            if in.GetLeader() != wk.masternode{
                wk.masternode = in.GetLeader()
                doneChan <- err
                log.Printf("new leader:%s",in.GetLeader())
                return
            }
        }
    }()
    if err = stream.Send(wk.CollectHeartbeat()); err != nil {
        log.Println("send error",err)
        return 
    }  
    tickChan := time.Tick(time.Duration(5)*time.Second)
    for{
        select {
            case <-tickChan:
                if err = stream.Send(wk.CollectHeartbeat()); err != nil {
                    log.Println("send error",err)
                    return 
                }    

            case _ = <-doneChan:
                log.Println("done")
                return     
        }
    }
   

}
func (wk *Worker) UpdateServerNode(servernode []string){


}
func (wk *Worker) CollectHeartbeat()*pb.Heartbeat{
    msg := pb.Heartbeat{
           Worknode:wk.localip,
    }
    return &msg
}
/*func (wk *Worker) dameon(){
    for{
        //log.Printf("dameon nothing")
        time.Sleep(time.Duration(10)*time.Second) 
    }
    
}*/


