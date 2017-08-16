package main 
import (
	"fmt"
	"os"
	"os/exec"
    "log"
	"strings"
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
func dcc_scan_args(argvs []string,outputfile string,input_file string)dcc_exitcode{

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
	        	   if outputfile != ""{
	        	   	return EXIT_DISTCC_FAILED
	        	   }
                   outputfile = argvs[i]
	        }else if strings.HasPrefix(argvs[i],"-o"){
	        	   if outputfile != ""{
	        	   	return EXIT_DISTCC_FAILED
	        	   }
                   outputfile = strings.TrimPrefix(argvs[i],"-o")
                   
	        }		
		}else {
			 if dcc_is_source(argvs[i]){
			 	   input_file = argvs[i]

			 	}else if strings.HasSuffix(argvs[i],".o") {

			 		 if outputfile != ""{
	        	   	 return EXIT_DISTCC_FAILED
	        	    }
                   outputfile = argvs[i]
			 	}
		}	
	}
	if (!seen_opt_c && !seen_opt_s){
		return EXIT_DISTCC_FAILED
	}
	if input_file == ""{
	   return EXIT_DISTCC_FAILED
	}

	return 0
}
func dcc_compile_local(argvs []string,filename string)bool{
	 cmd := exec.Command("cc",os.Args[1:]...)
     output,err:=cmd.CombinedOutput()
     if err!= nil{
     	//fmt.Println(err)
     	log.Fatal(err)
     	log.Fatal(output,err)
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
func dcc_cpp_maybe(argvs[] string,input_fname string,pcpp_fname *string)bool{
	var cpp_argv []string
	if dcc_is_preprocessed(input_fname){
		*pcpp_fname = input_fname
		return true
	}
    cpp_argv = dcc_strip_dasho(argvs)
    dcc_set_action_opt(cpp_argv)
    fmt.Println("dcc_cpp_maybe",cpp_argv)
    cmd := exec.Command("cc",cpp_argv[0:]...)
    _,err:=cmd.CombinedOutput()
     if err!= nil{
     	fmt.Println("dcc_cpp_maybe",err)
     	return false
     }
     return true
}
func dcc_build_somewhere(argvs []string) int{
      
      var outputfile string
      var input_file string
      var cpp_fanme  string
      argvs = dcc_expand_preprocessor_options(argvs)
    
      ret := dcc_scan_args(argvs,outputfile,input_file)
      if ret == EXIT_DISTCC_FAILED{
      	 fmt.Println("local")
      	 dcc_compile_local(argvs,outputfile)
      	 return 0
      }
      dcc_cpp_maybe(argvs,input_file,&cpp_fanme)

      fmt.Println("success")




	return 0
}
func main(){

     //test:= dcc_expand_preprocessor_options(os.Args);
    
	 dcc_build_somewhere(os.Args)
     
    
     
     return 
     cmd := exec.Command("ls","-l aa")
     out,err:=cmd.CombinedOutput()
     if err!= nil{
     	fmt.Println(err)
     }

     fmt.Println(string(out))

}