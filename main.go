package main 
import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
    "log"
	"strings"
	"net"
	"time"
	"encoding/json"
	"io/ioutil"
    //"math/rand"
    "flag"
    "./server"
    "./common"
	"./worker"
	"./monitor"
	"net/http"
)

func copy_extra_args(presultargs* []string,args string)int {
	 splitargs := strings.Split(args,",")
	 for i:=1;i<len(splitargs);i++{
	 	 *presultargs=append(*presultargs,splitargs[i])
         if strings.HasPrefix(splitargs[i],"-MD") || strings.HasPrefix(splitargs[i],"-MMD"){
         	*presultargs=append(*presultargs,"-MF")
         	i++
         	if(i < len(splitargs)){
         		*presultargs=append(*presultargs,splitargs[i])
         	}
         		
         	
         }
 	 }
	return 0
}
func dcc_expand_preprocessor_options(argvs []string )[]string {
	resultargs:= []string{}

	for i:=1;i<len(argvs);i++{
		if strings.HasPrefix(argvs[i],"-Wp,"){
	        copy_extra_args(&resultargs,argvs[i])
		}else{
			resultargs=append(resultargs,argvs[i])
		}

	}
	fmt.Println(resultargs)
    return resultargs;	
}

func dcc_is_preprocessed(filename string)bool {
	splitext := strings.Split(filename,".")
	if len(splitext) == 1{
		return false
	}
	ext := splitext[1]
	switch ext[0]{
       case 's':
       	return ext == "s" 
       case 'i':
       	return ext == "i" || ext == "ii"
       case 'm':
       	return ext == "mi" || ext == "mii"
       default:
       	return false
    }


}


func dcc_strip_dasho(argvs []string)[]string{
	var result []string
	for i:=0;i<len(argvs);{
		if argvs[i] == "-o"{
			i=i+2
		}else if(strings.HasPrefix(argvs[i],"-o")){
			i++
		}else{
             result = append(result,argvs[i])
			 i++
		}
	}
	return result

}
func dcc_set_action_opt(argvs []string){
	 for i:=0;i<len(argvs);i++{
	 	  if argvs[i] == "-c" || argvs[i] == "-S"{
	 	  	 argvs[i]= "-E"
	 	  }

	 }

}
func dcc_preproc_extern(args string)string{
	
	splitext := strings.Split(args,".")
	if len(splitext) == 1{
		return ""
	}
	ext := splitext[1]
	if ext == "i" || ext == "c"{
		return splitext[0]+".i"
	}
	if ext == "c" || ext == "cc" || ext == "cpp" || ext == "cxx" || ext == "cp" || ext == "c++" || ext == "C" || ext == "i"{
	   	return splitext[0]+".ii"
	 }
    if ext == "mi" || ext == "m"{
    	return splitext[0]+".mi"
    }
    if ext == "mii" || ext == "mm" || ext == "M"{
    	return splitext[0]+".mii"
    }
    if ext == "s" || ext == "S"{
    	return splitext[0]+".s"
    }
    return ""
}
func dcc_cpp_maybe(argvs[] string,input_fname string,pcpp_fname *string)([]byte,bool){
	var cpp_argv []string
	log.Printf("input_fame::%s \n",input_fname)
	if dcc_is_preprocessed(input_fname){
		*pcpp_fname = input_fname
		file,err:=os.Open(input_fname)
	    if err != nil{
	 	   log.Fatal(err)
	 	    return nil,false
	    }
        data,err := ioutil.ReadAll(file)
        if err != nil{
         return nil,false
        }
		return data,true
	}
    cpp_argv = dcc_strip_dasho(argvs)
    dcc_set_action_opt(cpp_argv)
    *pcpp_fname = dcc_preproc_extern(input_fname)
    if len(*pcpp_fname) == 0{
    	log.Fatal(input_fname)
    	return nil,false
    }
  
    log.Printf("local preprocess:: %s num:%d,cpp_fname::%s\n",cpp_argv,len(cpp_argv),*pcpp_fname)
    
    cmd := exec.Command("cc",cpp_argv[0:]...)
  

    data,err:=cmd.CombinedOutput()
     if err!= nil{
     	//log.Printf(err)
     	log.Printf("data:%s\n",data)
     	log.Fatal(err)
     
     	return nil,false
     }
     //dcc_write_file(*pcpp_fname,data)
     return data,true
}
func dcc_write_file(fname string,data []byte){
	f,err := os.Create(fname)
	if err != nil{
		panic(err)
	}
	defer f.Close()
	_,err =f.Write(data)
	if err!= nil{
		panic(err)
	}
	f.Sync()


} 
func dcc_strip_local_args(argvs []string)[]string{
	var result []string
	for i:=0;i<len(argvs);i++{
		 if argvs[i] == "-D" || argvs[i] == "-I" || argvs[i] == "-U"||
		    argvs[i] == "-L" || argvs[i] == "l" || argvs[i] == "-MF"||
		    argvs[i] == "-MT" || argvs[i] == "-MQ" || argvs[i] == "-include"||
		    argvs[i] == "imacros" || argvs[i] == "-iprefix" || argvs[i] =="-iwithpreifx"||
		    argvs[i] == "idirafter"{
                 i++
                 continue
		    }
		 if strings.HasPrefix(argvs[i],"-Wp,") || strings.HasPrefix(argvs[i],"-Wl,") || strings.HasPrefix(argvs[i],"-D") ||
		    strings.HasPrefix(argvs[i],"-U") || strings.HasPrefix(argvs[i],"-I") || strings.HasPrefix(argvs[i],"-l") ||
		    strings.HasPrefix(argvs[i],"-MF") || strings.HasPrefix(argvs[i],"-MT") || strings.HasPrefix(argvs[i],"-MQ"){
            
            continue
		 }
         if argvs[i] == "-undef" || argvs[i] == "nostdinc" || argvs[i] == "nostdinc++" ||
            argvs[i] == "-MD" || argvs[i] == "-MMD" || argvs[i] == "-MG" || argvs[i] == "-MP"{
            	continue        
            }
         result = append(result,argvs[i])
	}
	return result


}

func dcc_remote_connect()(net.Conn,error){
   
	 server := dcc_pick_host_from_list_and_lock_it()
	 tcpaddr,err := net.ResolveTCPAddr("tcp4",server)
	 if err!= nil{
	 	fmt.Printf("error:%s",err.Error())
	 	return nil,err
	 }
	 conn,err:= net.DialTCP("tcp",nil,tcpaddr)
	 if err!= nil{
	 	fmt.Printf("errpr:%s",err.Error())
	 	return nil,err
	 }
     return conn,nil
}





func dcc_send_argv(server_side_argv []string,outputfile string,data []byte)net.Conn{
	tmparg:=common.ServerArg{}
	if len(server_side_argv) == 0{
		return nil
	}
	var i int
	for i=0;i<len(server_side_argv)-1;i++{
		tmparg.Server_side_argv += server_side_argv[i] + " "
	}
    tmparg.Server_side_argv += server_side_argv[i]
    conn,err:=dcc_remote_connect()
    if err != nil{
    	return nil
    }
    //defer conn.Close()
 
    tmparg.Cpp_fname = outputfile

    //tmparg.File_length,_ = dcc_get_filelength(tmparg.Cpp_fname)
    tmparg.File_length=len(data)

    byt,_:=json.Marshal(tmparg)
   
    _,err=conn.Write(byt)
    if err != nil{
    	log.Fatal(err)
    	return nil
    }
    _,ret:=common.Dcc_wait_response(conn)
    if ret == false{
    	return nil
    }
    _,err=conn.Write(data)
	 if err!= nil{
		 log.Fatal(err)
		 return nil
	}  
    //dcc_x_many_files(tmparg.Cpp_fname,conn)
    _,ret=common.Dcc_wait_response(conn)
    if ret == false{
    	return nil
    }
    return conn

}
func dcc_recv_output(conn net.Conn){
	buffer := make([]byte,2048)
	n,err:=conn.Read(buffer)
		if err!= nil{
			log.Printf(conn.RemoteAddr().String(),"connection err",err)
			return 
	}
	readbuffer:=buffer[:n]
	tmpout := common.OutputArg{}
	if err:=json.Unmarshal(readbuffer,&tmpout); err!= nil{
		log.Printf("fail:%d,read:%d",n,len(buffer))
		log.Fatal(err)
 		return 
    }
    log.Printf("1:%s,2:%s",tmpout.Cpp_fname,tmpout.File_length)
    common.Dcc_response(conn)
    common.Dcc_r_file(tmpout.Cpp_fname,conn,tmpout.File_length)
    common.Dcc_response(conn)
}
func dcc_build_somewhere(argvs []string) int{
      
      var outputfile string
      var input_file string
      var cpp_fanme  string
      
      argvs = dcc_expand_preprocessor_options(argvs)
    
      ret := common.Dcc_scan_args(argvs,&outputfile,&input_file)
      if ret == common.EXIT_DISTCC_FAILED{
      	 log.Printf("local")
      	 common.Dcc_compile_local(argvs,outputfile)
      	 return 0
      }
      data,flag:=dcc_cpp_maybe(argvs,input_file,&cpp_fanme)
      if flag == false{
      	return 1
      }

      server_side_argv:= dcc_strip_local_args(argvs)
      conn:=dcc_send_argv(server_side_argv,cpp_fanme,data)

      //dcc_compile_local(server_side_argv,outputfile)
      log.Printf("server_side_argv:%s,\n",server_side_argv)
      if conn == nil{
      	return 1
      }
      dcc_recv_output(conn)
    
	  return 0
}
func dcc_pick_host_from_list_and_lock_it()string {
    resp, err := http.Get("http://localhost:8001/worker")
    if err!=nil{
  	  log.Printf("there is no useful distccgo_hosts")
  	  return ""
    }
    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)
    if err!=nil{
    	log.Printf("read err")
    	return ""
    }
    ret :=string(body)
    fmt.Println(ret)
    return ret + ":8000"
}



func dcc_pick_host_from_list_and_lock_itss()string {
	var distccgo_hosts string
	distccgo_hosts = os.Getenv("DISTCCGO_HOSTS")
	ip_hosts := strings.Split(distccgo_hosts," ")
	if len(ip_hosts) == 0{
		log.Printf("there is no useful distccgo_hosts")
	}
	//index:=rand.Intn(len(ip_hosts))
	index := time.Now().Nanosecond()%len(ip_hosts)
	return  ip_hosts[index]+":8000"
}
func test(){
	var testFlag = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
    //var testFlag flag.FlagSet
    var masternode string
    var host string
    testFlag.StringVar(&masternode, "masternode", "localhost:4001", "masternode")
    testFlag.StringVar(&host, "h", "127.0.0.2", "hostname")
    flag.Parse()

    args := flag.Args()
    testFlag.Parse(args[1:])   
    
    log.Printf("worker running masternode:%s,host:%s",masternode,host)

}
func usage(){
	fmt.Fprintf(os.Stderr,"Usage: %s [arguments] <data-path> \n", os.Args[0])
}
func interruptListener()<-chan struct{}{
	c:=make(chan struct{})
	go func(){
		interruptChannel :=make(chan os.Signal,1)
	    signal.Notify(interruptChannel,os.Interrupt)
	    <-interruptChannel
	    close(c)
	}()
	return c
}
func main(){
     
    log.SetFlags(log.Ldate|log.Ltime |log.LUTC|log.Lshortfile)
    //test()
    flag.Parse()

    args := flag.Args()
	if len(args) < 1 {
		usage()
		return 
	}
     
    if args[0] == "worker"{
     		done:=interruptListener()
     		worker.RunWorker(args[1:])
     		<-done
     		return 
    }else if args[0] == "server"{
    		done:=interruptListener()
     		server.RunServer(args[1:])
     		<-done
     		return 
    }else if args[0] == "monitor"{
    	done:=interruptListener()
 		monitor.RunMonitor(args[1:])
 		<-done
 		return

    }
     //dcc_build_somewhere(os.Args)
     //dcc_pick_host_from_list_and_lock_itss()
     dcc_build_somewhere(args[0:])
     return 
	
}