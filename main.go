package main 
import (
	"fmt"
	"os"
	"os/exec"

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
func dcc_scan_args(argvs []string)dcc_exitcode{

    seen_opt_s:=false
    seen_opt_c:=false
	var outputfile string
	var  input_file string
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
func dcc_build_somewhere() int{





	return 0
}
func main(){

     test:= dcc_expand_preprocessor_options(os.Args);
     fmt.Println(test)
     //eturn ;
     fmt.Println(os.Args)
    
     
     return 
     cmd := exec.Command("cc",os.Args[1:]...)
     out,err:=cmd.CombinedOutput()
     if err!= nil{
     	fmt.Println(err)
     }
     fmt.Println(string(out))

}