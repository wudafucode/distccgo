package main 
import (
	"fmt"
	"os"
	"os/exec"
    "log"
	"strings"
	"net"
	"time"
	"encoding/json"
)
type dcc_exitcode int
const (
	EXIT_DISTCC_FAILED  dcc_exitcode = 100+iota
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
func dcc_is_source(filename string)bool {
	splitext := strings.Split(filename,".")
	if len(splitext) == 1{
		return false
	}
	ext := splitext[1]
	switch ext[0]{
       case 'i':
       	return ext == "i" || ext == "ii"
       case 'c':
       	return ext == "c" || ext == "cc" || ext == "cpp" || ext == "cxx" || ext == "cp" || ext=="c++"
       case 'C':
       	return ext == "C"
       case 'm':
       	return ext == "m" || ext =="mm" || ext == "mi" || ext == "mii"
       case 'M':
         return ext == "M"
       default:
          return false	

	}

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
func dcc_scan_args(argvs []string,poutputfile *string,pinput_file *string)dcc_exitcode{

    seen_opt_s:=false
    seen_opt_c:=false
	//var outputfile string
	//var  input_file string
	for i:=0;i<len(argvs);i++{
		if strings.HasPrefix(argvs[i],"-"){
			if argvs[i] == "-E" {
	 	          return EXIT_DISTCC_FAILED
	        }else if argvs[i]  == "-MD" || argvs[i]  == "-MMD"{

	        }else if argvs[i]  == "-MF" || argvs[i]  == "-MT" || argvs[i]  =="-MQ" {
	        	  i++
	        }else if strings.HasPrefix(argvs[i],"-MF") || strings.HasPrefix(argvs[i],"-MT") ||strings.HasPrefix(argvs[i],"-MQ") {
	        	
	        }else if strings.HasPrefix(argvs[i],"-M") {
	              return EXIT_DISTCC_FAILED
	        }else if argvs[i] == "-march=native" || argvs[i] == "-matune=native" {
	        	   return EXIT_DISTCC_FAILED
	        }else if strings.HasPrefix(argvs[i],"-Wa"){
	        	     if strings.Contains(argvs[i],"-a") || strings.Contains(argvs[i],"--MD"){
	        	     	return EXIT_DISTCC_FAILED
	        	     }
	        }else if strings.HasPrefix(argvs[i],"-specs="){
	        	     return EXIT_DISTCC_FAILED
	        }else if argvs[i] == "-S"{
	        	    seen_opt_s = true 
	        }else if argvs[i] == "-fprofile-arcs" || argvs[i] == "-ftest-coverage" || argvs[i] == "--coverage"{
	        	   return EXIT_DISTCC_FAILED
	        }else if strings.HasPrefix(argvs[i],"-frepo"){
	        	   return EXIT_DISTCC_FAILED
	        }else if strings.HasPrefix(argvs[i],"-x"){
	        	   return EXIT_DISTCC_FAILED
	        }else if strings.HasPrefix(argvs[i],"-dr"){
	        	   return EXIT_DISTCC_FAILED
	        }else if argvs[i] == "-c"{
	        	   seen_opt_c = true
	        }else if argvs[i] == "-o"{
	        	   i++;
	        	   if *poutputfile != ""{
	        	   	return EXIT_DISTCC_FAILED
	        	   }
                   *poutputfile = argvs[i]
	        }else if strings.HasPrefix(argvs[i],"-o"){
	        	   if *poutputfile != ""{
	        	   	return EXIT_DISTCC_FAILED
	        	   }
                   *poutputfile = strings.TrimPrefix(argvs[i],"-o")
                   
	        }		
		}else {
			 if dcc_is_source(argvs[i]){
			 	   *pinput_file = argvs[i]

			 	}else if strings.HasSuffix(argvs[i],".o") {

			 		 if *poutputfile != ""{
	        	   	 return EXIT_DISTCC_FAILED
	        	    }
                   *poutputfile = argvs[i]
			 	}
		}	
	}
	if (!seen_opt_c && !seen_opt_s){
		return EXIT_DISTCC_FAILED
	}
	if *pinput_file == ""{
	   return EXIT_DISTCC_FAILED
	}

	return 0
}
func dcc_compile_local(argvs []string,filename string)bool{
	 cmd := exec.Command("cc",argvs[0:]...)
     output,err:=cmd.CombinedOutput()
     if err!= nil{
     	
     	log.Printf("output:%s,args:%s\n",output,argvs)
     	log.Fatal(err)
     	return false
     }
     return true
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
func dcc_cpp_maybe(argvs[] string,input_fname string,pcpp_fname *string)bool{
	var cpp_argv []string
	fmt.Printf("input_fame::%s \n",input_fname)
	if dcc_is_preprocessed(input_fname){
		*pcpp_fname = input_fname
		return true
	}
    cpp_argv = dcc_strip_dasho(argvs)
    dcc_set_action_opt(cpp_argv)
    *pcpp_fname = dcc_preproc_extern(input_fname)
    if len(*pcpp_fname) == 0{
    	log.Fatal(input_fname)
    }
  
    log.Printf("local preprocess:: %s num:%d,cpp_fname::%s\n",cpp_argv,len(cpp_argv),*pcpp_fname)
    //test_argv:={""}
    cmd := exec.Command("cc",cpp_argv[0:]...)
    //cmd := exec.Command("ls","-al")

    data,err:=cmd.CombinedOutput()
     if err!= nil{
     	//log.Printf(err)
     	log.Printf("data:%s\n",data)
     	log.Fatal(err)
     
     	return false
     }
     //log.Printf("data:%s\n",data)
     //return false
     dcc_write_file(*pcpp_fname,data)
     return true
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
   
	 server := "127.0.0.1:8000"
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
func dcc_wait_response(conn net.Conn)(string,bool){
	conn.SetReadDeadline(time.Now().Add(time.Second*3))

    var res Response
    buffer := make([]byte,2048)
    n,err:=conn.Read(buffer)
	if err!= nil{
		log.Fatal(err)
		return "",false
	}
    readbuffer:=buffer[:n]
	if err:=json.Unmarshal(readbuffer,&res); err!= nil{
		log.Printf("fail:%d,read:%d",n,len(buffer))
		log.Fatal(err)
 		return "",false
    }
    if res.Ret == false{
    	return "",false
    }
    conn.SetReadDeadline(time.Time{})
    return res.result,true
}
func dcc_send_argv(server_side_argv []string,outputfile string){
	tmparg:=ServerArg{}
	if len(server_side_argv) == 0{
		return 
	}
	var i int
	for i=0;i<len(server_side_argv)-1;i++{
		tmparg.Server_side_argv += server_side_argv[i] + " "
	}
    tmparg.Server_side_argv += server_side_argv[i]
    conn,err:=dcc_remote_connect()
    if err != nil{
    	return 
    }
    //defer conn.Close()
    byt,err:=json.Marshal(tmparg)
    if err != nil{
    	log.Fatal(err)
    	return 
    }
    _,err=conn.Write(byt)
    if err != nil{
    	log.Fatal(err)
    	return 
    }
    _,ret:=dcc_wait_response(conn)
    if ret == false{
    	return 
    }
   conn.Close()

}
func dcc_build_somewhere(argvs []string) int{
      
      var outputfile string
      var input_file string
      var cpp_fanme  string
      fmt.Println("hello")
      argvs = dcc_expand_preprocessor_options(argvs)
    
      ret := dcc_scan_args(argvs,&outputfile,&input_file)
      if ret == EXIT_DISTCC_FAILED{
      	 fmt.Println("local")
      	 dcc_compile_local(argvs,outputfile)
      	 return 0
      }
      dcc_cpp_maybe(argvs,input_file,&cpp_fanme)
      server_side_argv:= dcc_strip_local_args(argvs)
      dcc_compile_local(server_side_argv,outputfile)
      log.Printf("server_side_argv:%s,\n",server_side_argv)
    




	  return 0
}
type Response struct{
	 Ret      bool    `json:"ret"`
	 result   string  `json:"result"`
}
type ServerArg struct{
     Server_side_argv string   `json:"server_side_argv"`
     Cpp_fname        string   `json:"cpp_fname"`
     File_length       int     `json:"file_length"`
}
func main(){
     


    
     //dcc_build_somewhere(os.Args)
     //return 
	 teststring :=[]string{"gcc","hello"}
	 dcc_send_argv(teststring,"1.cpp")
     
    
   
     return 
   
}