package main 
import (
        "testing"
        "strings"
        "fmt"
        "os"
)

func Test_dcc_is_source(t *testing.T){
	var result string
	s:=[15]struct{
		input   string 
		expect  string 
	}{{ "hello.c", "source",},
       {"hello.cc", "source",},
       {"hello.cxx", "source",},
       {"hello.cpp", "source",},
       {"hello.m", "source",},
       {"hello.M", "source",},
       {"hello.mm", "source",},
       {"hello.mi", "source",},
       {"hello.mii", "source",},
   	   {"hello.2.4.4.i","not-source",},
       {".foo", "not-source",},
       {"gcc", "not-source",},
       {"hello.ii","source",},
       {"boot.s", "not-source",},
       {"boot.S", "not-source",},}
    for i:=0;i<len(s);i++{
    	ret:=dcc_is_source(s[i].input)
    	if ret == true{
    		result= "source"
    	}else{
    		result= "not-source"
    	}
    	if result != s[i].expect{
    		t.Error("not passed",s[i]);
    	}
    }
    t.Log("dcc_is_source passed")
	//_= dcc_is_source("1.cpp")
}
func Test_dcc_is_preprocessed(t *testing.T){
	var result string
	s:=[15]struct{
		input   string 
		expect  string 
	}{{ "hello.c", "not-preprocessed",},
       {"hello.cc", "not-preprocessed",},
       {"hello.cxx","not-preprocessed",},
       {"hello.cpp", "not-preprocessed",},
       {"hello.m", "not-preprocessed",},
       {"hello.M", "not-preprocessed",},
       {"hello.mm", "not-preprocessed",},
       {"hello.mi", "preprocessed",},
       {"hello.mii", "preprocessed",},
   	   {"hello.2.4.4.i","not-preprocessed",},
       {".foo", "not-preprocessed",},
       {"gcc", "not-preprocessed",},
       {"hello.ii","preprocessed",},
       {"boot.s", "preprocessed",},
       {"boot.S", "not-preprocessed",},}
    for i:=0;i<len(s);i++{
    	ret:=dcc_is_preprocessed(s[i].input)
    	if ret == true{
    		result= "preprocessed"
    	}else{
    		result= "not-preprocessed"
    	}
    	if result != s[i].expect{
    		t.Error("not passed",s[i]);
    	}
    }
    t.Log("dcc_is_preprocessed passed")
	//_= dcc_is_source("1.cpp")


}
func Test_dcc_strip_args(t *testing.T){
  
     s:=[14]struct{
      input   string 
      expect  string 
    }{
       {"gcc -c hello.c", "gcc -c hello.c"},
       {"cc -Dhello hello.c -c", "cc hello.c -c"},
       {"gcc -g -O2 -W -Wall -Wshadow -Wpointer-arith -Wcast-align -c -o h_strip.o h_strip.c",
        "gcc -g -O2 -W -Wall -Wshadow -Wpointer-arith -Wcast-align -c -o h_strip.o h_strip.c"},
        //# invalid but should work
       {"cc -c hello.c -D", "cc -c hello.c"},
       {"cc -c hello.c -D -D", "cc -c hello.c"},
       {"cc -c hello.c -I ../include", "cc -c hello.c"}, 
       {"cc -c -I ../include hello.c", "cc -c hello.c"},
       {"cc -c -I. -I.. -I../include -I/home/mbp/garnome/include -c -o foo.o foo.c",
        "cc -c -c -o foo.o foo.c"},
       {"cc -c -DDEBUG -DFOO=23 -D BAR -c -o foo.o foo.c",
         "cc -c -c -o foo.o foo.c"},
        //# New options stripped in 0.11
       {"cc -o nsinstall.o -c -DOSTYPE=\"Linux2.4\" -DOSARCH=\"Linux\" -DOJI -D_BSD_SOURCE -I../dist/include -I../dist/include -I/home/mbp/work/mozilla/mozilla-1.1/dist/include/nspr -I/usr/X11R6/include -fPIC -I/usr/X11R6/include -Wall -W -Wno-unused -Wpointer-arith -Wcast-align -pedantic -Wno-long-long -pthread -pipe -DDEBUG -D_DEBUG -DDEBUG_mbp -DTRACING -g -I/usr/X11R6/include -include ../config-defs.h -DMOZILLA_CLIENT -Wp,-MD,.deps/nsinstall.pp nsinstall.c",
         "cc -o nsinstall.o -c -fPIC -Wall -W -Wno-unused -Wpointer-arith -Wcast-align -pedantic -Wno-long-long -pthread -pipe -g nsinstall.c"},
    }
    var cmdline string
    var j int
    for i:=0;i<len(s);i++{

        cmdline = ""
        result:=dcc_strip_local_args(strings.Split(s[i].input," "))

        for j=0;j<len(result)-1;j++{
         
            cmdline = cmdline + result[j] + " "
        }
        cmdline = cmdline + result[j]
        if cmdline != s[i].expect{
         t.Error("not passed,expect:",s[i].input);
         t.Error("not passed,expect:",s[i].expect);
         t.Error("not passed,result:",cmdline);
        }
    }
 
    t.Log("dcc_strip_args passed")


}
type Compilation_Case struct{

}

func (this *Compilation_Case)sourceFilename()string{
   return "testtmp.c"  
}
func (this *Compilation_Case)headerFilename()string{
   return "testhdr.h"  
}
func (this *Compilation_Case)headerSource()string{
   return ""  
}
func (this *Compilation_Case)compileCmd()string{
  return "distcc"+" -o testtmp.o " + " -c "+this.sourceFilename()

}

type CompileHello_Case struct{
      Compilation_Case
}
func (this *CompileHello_Case)headerSource()string{
      return "#define HELLO_WORLD \"hello world\" "
}
func (this *CompileHello_Case)source()string{
        return "#include<stdio.h>\n"+ this.headerFilename() + "\n"+"int main(void) {\nputs(HELLO_WORLD);\n return 0;}"
       //return "12%s3" + this.headerSource()
}
type Compile interface{
    
     headerSource()string
     headerFilename()string
     sourceFilename()string
     source()string

     compileCmd()string

}
//type build struct{

//}
//func (this *build)source()string{
func build(this Compile)bool{
     //fmt.Printf(this.compileCmd())
     fmt.Printf(this.source())
     f,err := os.Create(this.sourceFilename())
     if err != nil{
        panic(err)
     }
     _,err =f.Write([]byte(this.source()))
    if err!= nil{
      panic(err)
    }

    f.Close()
    f,err = os.Create(this.headerFilename())
    if err != nil{
       panic(err)
    }
     _,err =f.Write([]byte(this.headerSource()))
    if err!= nil{
      panic(err)
    }
    f.Close()
  
     return true
}
func Test_CompileHello_Case(t *testing.T){
      var test CompileHello_Case
      ret:=build(&test)
      if ret == false{
        t.Error("not passed,expect:");
      }

}
