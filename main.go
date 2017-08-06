package main 
import (
	"fmt"
	"os"
	"os/exec"
	//"strings"
)
func main(){
    
     fmt.Println(os.Args)
     cmd := exec.Command("cc",os.Args[1:]...)
     out,err:=cmd.CombinedOutput()
     if err!= nil{
     	fmt.Println(err)
     }
     fmt.Println(string(out))

}