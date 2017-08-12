package main 
import (
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"strings"
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
func dcc_expand_preprocessor_options(argvs []string )int {
	resultargs:= []string{"hle"}

	for i:=1;i<len(argvs);i++{
		if strings.HasPrefix(argvs[i],"-Wp,"){
	        copy_extra_args(&resultargs,argvs[i])
		}else{
			resultargs=append(resultargs,argvs[i])
		}

	}
	fmt.Println(resultargs)
    return 0;	
}
func dcc_build_somewhere() int{





	return 0
}
func main(){

     //dcc_expand_preprocessor_options(os.Args);
     //eturn ;
     fmt.Println(os.Args)
     t := reflect.TypeOf(os.Args)
     fmt.Println("Type:",t.Name())
     return 
     cmd := exec.Command("cc",os.Args[1:]...)
     out,err:=cmd.CombinedOutput()
     if err!= nil{
     	fmt.Println(err)
     }
     fmt.Println(string(out))

}