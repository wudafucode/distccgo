package common
import (
    "net"
	"strings"
	"encoding/json"
	"os"
	"log"
	"os/exec"
	"time"
	"io/ioutil"
)
type dcc_exitcode int
const (
	EXIT_DISTCC_FAILED  dcc_exitcode = 100+iota
)
type ServerArg struct{
     Server_side_argv string   `json:"server_side_argv"`
     Cpp_fname        string   `json:"cpp_fname"`
     File_length       int     `json:"file_length"`
}
type Response struct{
	 Ret      bool    `json:"ret"`
	 result   string  `json:"result"`
}
type OutputArg struct{
     Cpp_fname        string   `json:"cpp_fname"`
     File_length       int     `json:"file_length"`
}
func Dcc_x_many_files(filename string,conn net.Conn)bool{

	 file,err:=os.Open(filename)
	 if err != nil{
	 	log.Fatal(err)
	 	return false
	 }

     data,err := ioutil.ReadAll(file)
     if err != nil{
        return false
     }
	 _,err=conn.Write(data)
	 if err!= nil{
		 log.Fatal(err)
		 return false
	 }  
	 return true
}
func Dcc_wait_response(conn net.Conn)(string,bool){
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
func Dcc_compile_local(argvs []string,filename string)bool{
	 cmd := exec.Command("cc",argvs[0:]...)
     output,err:=cmd.CombinedOutput()
     if err!= nil{
     	
     	log.Printf("output:%s,args:%s\n",output,argvs)
     	log.Fatal(err)
     	return false
     }
     return true
}
func Dcc_get_filelength(filename string)(int,error){
	 fileinfo,err:= os.Stat(filename)
	 if err != nil{
	 	log.Fatal(err)
	 	return 0,err
	 }
     return int(fileinfo.Size()),nil
}
func Dcc_is_source(filename string)bool {
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
func Dcc_response(conn net.Conn)bool{
	tres:= Response{};
    tres.Ret = true;
    byt,_:=json.Marshal(tres)
    conn.Write(byt)
    return true
}
func Dcc_r_file(filename string,conn net.Conn,filelength int)bool{
	 buffer := make([]byte,2048)

     f,err := os.OpenFile(filename,os.O_CREATE|os.O_RDWR,0777)
     if err != nil{
     	//log.Printf("open file:%s",filename)
     	log.Println(err)
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


func Dcc_scan_args(argvs []string,poutputfile *string,pinput_file *string)dcc_exitcode{

    seen_opt_s:=false
    seen_opt_c:=false
	//var outputfile string
	//var  input_file string
	for i:=0;i<len(argvs);i++{
		if strings.HasPrefix(argvs[i],"-"){
			if argvs[i] == "-E" {
	 	          return EXIT_DISTCC_FAILED
	        }else if argvs[i]  == "-MD" || argvs[i]  == "-MMD"{

	        }else if argvs[i]  == "-MG" || argvs[i]  == "-MP"{

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
			 if Dcc_is_source(argvs[i]){
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

func Dcc_set_input(argvs []string,fname string)int{
	for i:=0;i<len(argvs);i++{
        if Dcc_is_source(argvs[i]){
        	argvs[i] = fname
        	return 0

        }
	}

	return 0
}

