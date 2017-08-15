package main 
import (
        "testing"
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