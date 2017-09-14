package main 
import (
	"fmt"
	"net"
	//"os"
    "log"
	"strings"
	"encoding/json"
)


func main(){
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

        return 
		//log.Println(conn.RemoteAddr().String(),"receive data string:\n",string(buffer[:n]))
	}




}
