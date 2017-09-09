package main 
import (
	"fmt"
	"net"
	"os"
    "log"
	//"strings"
	"encoding/json"
)


func maint(){
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
func CreateFile(conn net.Conn,filename string,filelength int)bool{
     
	 buffer := make([]byte,2048)

     f,err := os.OpenFile(filename,os.O_CREATE|os.O_RDWR,0777)
     if err != nil{
        return false
     }
     defer f.Close()
	for{
		n,err:=conn.Read(buffer)
		if err!= nil{
			log.Println(conn.RemoteAddr().String(),"connection err",err)
			return false
		}
		 _,err =f.Write(buffer[0:n])
        if err!= nil{
           return false
        }
        filelength= filelength - n
        if filelength<=0{
        	break
        }
	}

	return true
}
func handleConnection(conn net.Conn){
	defer conn.Close()
	buffer := make([]byte,2048)

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
        tres:= Response{};
        tres.Ret = true;
        byt,_:=json.Marshal(tres)
        conn.Write(byt)
        
        CreateFile(conn,server_arg.Cpp_fname,server_arg.File_length)

		//log.Println(conn.RemoteAddr().String(),"receive data string:\n",string(buffer[:n]))
	}




}
