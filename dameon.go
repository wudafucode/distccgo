package main 
import (
	"fmt"
	"net"
	"time"
	"os/exec"
    "log"
	"strings"
	"encoding/json"
	"regexp"
	"strconv"
    //"math/rand"
	"flag"
    //"os"
)
func GetLoad(loadstr string)(float64,float64,float64){
     reg := regexp.MustCompile(`\d+(\.\d+)`)

    
     loadbuf:=reg.FindAllString(loadstr,-1)
     if len(loadbuf) != 3{
     	return 0,0,0
     }
     ldavg1,_  := strconv.ParseFloat(loadbuf[0],32)
     ldavg5,_  := strconv.ParseFloat(loadbuf[1],32)
     ldavg10,_ := strconv.ParseFloat(loadbuf[2],32)
     return ldavg1,ldavg5,ldavg10
}
func senddata(balanceserver  string,ldavg1 float64,ldavg5 float64,ldavg10 float64){

    tcpaddr,err := net.ResolveTCPAddr("tcp4",balanceserver)
    if err!= nil{
        log.Printf("error:%s",err.Error())
        return 
    }
    conn,err:= net.DialTCP("tcp",nil,tcpaddr)
    if err!= nil{
        log.Printf("errpr:%s",err.Error())
        return 
    }
    tmparg:=CpuArg{}
    tmparg.Ldavg1 = ldavg1
    tmparg.Ldavg5 = ldavg5
    tmparg.Ldavg10 = ldavg10

    byt,_:=json.Marshal(tmparg)

     _,err=conn.Write(byt)
     if err!= nil{
         log.Fatal(err)
         return 
     }  

}
func loadavg(balanceserver  string){
    
   for{
  	 cmd := exec.Command("uptime")
     output,err:=cmd.CombinedOutput()
     if err!= nil{
     	log.Printf("output:%s,args:%s\n",output)
     	log.Fatal(err)   
     }
     ldavg1,ldavg5,ldavg10:=GetLoad(string(output))
     log.Printf("err:%s",string(output));
     log.Printf("1:%6.3f,2:%6.3f,3:%6.3f\r\n",ldavg1,ldavg5,ldavg10);
     senddata(balanceserver,ldavg1,ldavg5,ldavg10)
  	 time.Sleep(time.Duration(10)*time.Second)
  }


}
func loadbalance(){


}

func main(){
   
	log.SetFlags(log.Ldate|log.Ltime |log.LUTC |log.Lshortfile)
    serverip:= flag.String("s","","serverip")
    controler:= flag.Bool("c",false,"controler")
    localip:=flag.String("l","","localip")

    flag.Parse()
   
    if *serverip == "" || *localip == ""{
        log.Printf("serverip:%s,localip:%s",*serverip,*localip)
        return
    }
    if *controler == true{
        go loadbalance()
    }

 
	addr:= net.ParseIP(*serverip)
    if addr == nil{
        log.Printf("invalid ip address")
    	return 
    }
    balanceserver:= *serverip +":8001"
	go loadavg(balanceserver)
    
    
    localinfo:= *localip + ":8000"
	netlisten,err:= net.Listen("tcp",localinfo)
	if err != nil{
	
		fmt.Printf("err:%s",err.Error())
		return
	}
	defer netlisten.Close()
	for{
		conn,err:= netlisten.Accept()
		if err != nil{
			continue
		}
        log.Printf(conn.RemoteAddr().String())
		go handleConnection(conn)
		//go handlefun(conn)

	}
}
func dcc_split(fname string)string{
     buf:= strings.Split(fname,"/")
     return buf[len(buf)-1]
}
func dcc_prep(argvs []string){
    for i:=0;i<len(argvs);i++{
         if dcc_is_source(argvs[i]){
            argvs[i] = dcc_split(argvs[i])
        }else if(strings.HasSuffix(argvs[i],".o")){
            argvs[i] = dcc_split(argvs[i])
        }
    }

}


func handlefun(conn net.Conn){
	 defer conn.Close()
	 buffer := make([]byte,2048)
	 for{
        n,err:=conn.Read(buffer)
        if err!= nil{
        	continue
        }
        //readbuffer:=buffer[:n]
        conn.Write(buffer[:n])
        fmt.Println(conn.RemoteAddr().String())
        log.Printf("recv 4:%d",n)
	 	time.Sleep(time.Duration(1)*time.Second)
	 }
	 

}
func handleConnection(conn net.Conn){
	defer conn.Close()
	buffer := make([]byte,2048)
    var remote_outputfile string
    var input_file string
	server_arg :=ServerArg{}
	for{
		n,err:=conn.Read(buffer)
		if err!= nil{
			log.Println(conn.RemoteAddr().String(),"connection err",err)
			return 
		}
		readbuffer:=buffer[:n]
		if err:=json.Unmarshal(readbuffer,&server_arg); err!= nil{
			log.Printf("fail:%d,read:%d",n,len(buffer))
			log.Fatal(err)
     		return 
        }
        
        log.Printf("1:%s,2:%s,3:%d,4:%d",server_arg.Server_side_argv,server_arg.Cpp_fname,server_arg.File_length,n)

        
        dcc_response(conn)
        server_arg.Cpp_fname = dcc_split(server_arg.Cpp_fname)
        dcc_r_file(server_arg.Cpp_fname,conn,server_arg.File_length)
        dcc_response(conn)
        
        argv:= strings.Split(server_arg.Server_side_argv," ") 
       
        ret := dcc_scan_args(argv,&remote_outputfile,&input_file)
        if ret == EXIT_DISTCC_FAILED{

      		log.Printf("error:%s",server_arg.Server_side_argv)
      		return 
        }
        dcc_prep(argv) 
        dcc_set_input(argv,server_arg.Cpp_fname)
        dcc_compile_local(argv,remote_outputfile)
        
        
        tmpout := OutputArg{}
        local_outputfile := dcc_split(remote_outputfile)
        tmpout.File_length,_ = dcc_get_filelength(local_outputfile)
        tmpout.Cpp_fname  = remote_outputfile
        byt,_:=json.Marshal(tmpout)
        log.Printf("output:1:%s,2:%d,3:%d",tmpout.Cpp_fname,tmpout.File_length,len(byt))
	    _,err=conn.Write(byt)
	    if err != nil{
	    	log.Fatal(err)
	    	return 
	    }
	    dcc_wait_response(conn)
        dcc_x_many_files(local_outputfile,conn)
        dcc_wait_response(conn)
	    


        return 
		//log.Println(conn.RemoteAddr().String(),"receive data string:\n",string(buffer[:n]))
	}




}
