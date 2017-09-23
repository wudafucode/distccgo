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
func loadavg(){
   for{
  	 cmd := exec.Command("uptime")
     output,err:=cmd.CombinedOutput()
     if err!= nil{
     	log.Printf("output:%s,args:%s\n",output)
     	log.Fatal(err)   
     }
     ldavg1,ldavg5,ldavg10:=GetLoad(string(output))
     fmt.Printf("err:%s",string(output));
     fmt.Printf("1:%6.3f,2:%6.3f,3:%6.3f\r\n",ldavg1,ldavg5,ldavg10);
    

  	 time.Sleep(time.Duration(10)*time.Second)
  }


}

func main(){


    go loadavg()

	netlisten,err:= net.Listen("tcp","localhost:8000")
	if err != nil{
		//fmt.Printf(os.Stderr,"err:%s",err.Error());
		//os.exit(1)
		fmt.Printf("err:%s",err.Error())
		return
	}
	defer netlisten.Close()
	for{
		conn,err:= netlisten.Accept()
		if err != nil{
			continue
		}
		go handleConnection(conn)

	}
}

func handleConnection(conn net.Conn){
	defer conn.Close()
	buffer := make([]byte,2048)
    var outputfile string
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
        
        dcc_r_file(server_arg.Cpp_fname,conn,server_arg.File_length)
        dcc_response(conn)
        
        argv:= strings.Split(server_arg.Server_side_argv," ")  
        ret := dcc_scan_args(argv,&outputfile,&input_file)
        if ret == EXIT_DISTCC_FAILED{

      		log.Printf("error:%s",server_arg.Server_side_argv)
      		return 
        }
        dcc_set_input(argv,server_arg.Cpp_fname)
        dcc_compile_local(argv,outputfile)
        
        
        tmpout := OutputArg{}
        tmpout.File_length,_ = dcc_get_filelength(outputfile)
        tmpout.Cpp_fname  = outputfile
        byt,_:=json.Marshal(tmpout)
        log.Printf("output:1:%s,2:%d,3:%d",tmpout.Cpp_fname,tmpout.File_length,len(byt))
	    _,err=conn.Write(byt)
	    if err != nil{
	    	log.Fatal(err)
	    	return 
	    }
	    dcc_wait_response(conn)
        dcc_x_many_files(outputfile,conn)
        dcc_wait_response(conn)
	    


        return 
		//log.Println(conn.RemoteAddr().String(),"receive data string:\n",string(buffer[:n]))
	}




}
